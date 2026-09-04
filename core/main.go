// Command claude-hunter watches a Claude Code projects directory and
// streams a rolling-window token-usage snapshot to stdout as NDJSON.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pradeep/claude-hunter/core/snapshot"
	"github.com/pradeep/claude-hunter/core/usage"
	"github.com/pradeep/claude-hunter/core/watcher"
	"github.com/pradeep/claude-hunter/core/window"
)

const (
	burnRateSampleDuration    = 10 * time.Minute
	defaultEmitInterval       = 250 * time.Millisecond
	defaultCeilingEffectiveMi = 88.0
	agentRecencyDuration      = 5 * time.Minute
)

type cliOptions struct {
	projectsDir            string
	windowDuration         time.Duration
	emitInterval           time.Duration
	ceilingEffectiveTokens int64
}

func main() {
	options := parseFlags()
	if err := run(options); err != nil {
		log.Fatalf("claude-hunter: %v", err)
	}
}

func parseFlags() cliOptions {
	defaultProjectsDir := defaultClaudeProjectsDir()

	projectsDir := flag.String("projects-dir", defaultProjectsDir, "Claude Code projects root")
	windowHours := flag.Float64("window-hours", 5.0, "rolling window duration in hours")
	emitIntervalMillis := flag.Int64("emit-interval-ms", int64(defaultEmitInterval/time.Millisecond), "snapshot emission interval in milliseconds")
	ceilingEffectiveMillions := flag.Float64("ceiling-millions", defaultCeilingEffectiveMi, "estimated effective-token ceiling in millions")
	flag.Parse()

	return cliOptions{
		projectsDir:            *projectsDir,
		windowDuration:         time.Duration(*windowHours * float64(time.Hour)),
		emitInterval:           time.Duration(*emitIntervalMillis) * time.Millisecond,
		ceilingEffectiveTokens: int64(*ceilingEffectiveMillions * 1_000_000),
	}
}

func defaultClaudeProjectsDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".claude", "projects")
}

func run(options cliOptions) error {
	rootContext, cancelRoot := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancelRoot()

	observer, err := watcher.NewProjectsObserver(options.projectsDir)
	if err != nil {
		return fmt.Errorf("observer: %w", err)
	}
	go func() {
		if runErr := observer.Run(rootContext); runErr != nil {
			log.Printf("observer stopped: %v", runErr)
		}
	}()

	rollingWindow := window.NewRollingWindow(options.windowDuration)

	emitTicker := time.NewTicker(options.emitInterval)
	defer emitTicker.Stop()
	pruneTicker := time.NewTicker(30 * time.Second)
	defer pruneTicker.Stop()

	snapshotEncoder := json.NewEncoder(os.Stdout)

	for {
		select {
		case <-rootContext.Done():
			return nil

		case lineEvent := <-observer.Events():
			record, parseErr := usage.ParseLine(lineEvent.Line)
			if parseErr != nil || record == nil {
				continue
			}
			rollingWindow.Add(*record)

		case observerErr := <-observer.Errors():
			log.Printf("observer error: %v", observerErr)

		case <-pruneTicker.C:
			rollingWindow.PruneBefore(time.Now().Add(-options.windowDuration))

		case tickAt := <-emitTicker.C:
			rollingWindow.PruneBefore(tickAt.Add(-options.windowDuration))
			totals := rollingWindow.Totals()
			burnRate := rollingWindow.BurnRatePerMinute(tickAt, burnRateSampleDuration)
			payload := snapshot.Snapshot{
				Kind:      "snapshot",
				Timestamp: tickAt.UTC(),
				Model:     rollingWindow.LatestModel(),
				Window: snapshot.WindowSummary{
					InputTokens:              totals.InputTokens,
					OutputTokens:             totals.OutputTokens,
					CacheCreateTokens:        totals.CacheCreateTokens,
					CacheReadTokens:          totals.CacheReadTokens,
					EffectiveTokens:          totals.EffectiveTokens,
					CostUSD:                  totals.CostUSD,
					BurnRatePerMinute:        burnRate,
					WindowStart:              tickAt.Add(-options.windowDuration).UTC(),
					WindowEnd:                tickAt.UTC(),
					PercentOfCeilingEstimate: percentOfCeiling(totals.EffectiveTokens, options.ceilingEffectiveTokens),
					SecondsToLimit:           secondsToLimit(totals.EffectiveTokens, options.ceilingEffectiveTokens, burnRate),
					PerModel:                 toSnapshotBreakdown(rollingWindow.PerModel()),
				},
				Agents:   toSnapshotAgents(rollingWindow.PerSession(tickAt.Add(-agentRecencyDuration))),
				Projects: toSnapshotProjects(rollingWindow.PerProject()),
			}
			if encodeErr := snapshotEncoder.Encode(payload); encodeErr != nil {
				return fmt.Errorf("encode snapshot: %w", encodeErr)
			}
		}
	}
}

func percentOfCeiling(effectiveTokens, ceilingTokens int64) float64 {
	if ceilingTokens <= 0 {
		return 0
	}
	return float64(effectiveTokens) / float64(ceilingTokens) * 100.0
}

// secondsToLimit projects when the effective-token ceiling will be reached at
// the current burn rate. Returns -1 when the burn rate is zero or the ceiling
// is already exceeded (the answer is undefined in either case).
func secondsToLimit(effectiveTokens, ceilingTokens int64, burnRatePerMinute float64) int64 {
	if ceilingTokens <= 0 || burnRatePerMinute <= 0 {
		return -1
	}
	remaining := float64(ceilingTokens - effectiveTokens)
	if remaining <= 0 {
		return 0
	}
	minutes := remaining / burnRatePerMinute
	return int64(minutes * 60)
}

func toSnapshotProjects(projects []window.ProjectActivity) []snapshot.ProjectActivity {
	converted := make([]snapshot.ProjectActivity, len(projects))
	for index, entry := range projects {
		converted[index] = snapshot.ProjectActivity{
			Project:      entry.Project,
			CWD:          entry.CWD,
			Sessions:     entry.Sessions,
			TotalTokens:  entry.TotalTokens,
			CostUSD:      entry.CostUSD,
			LastActiveAt: entry.LastActiveAt.UTC(),
		}
	}
	return converted
}

func toSnapshotAgents(sessions []window.SessionActivity) []snapshot.SessionActivity {
	converted := make([]snapshot.SessionActivity, len(sessions))
	for index, entry := range sessions {
		converted[index] = snapshot.SessionActivity{
			SessionID:         entry.SessionID,
			Model:             entry.Model,
			ContextTokens:     entry.ContextTokens,
			ContextWindowSize: entry.ContextWindowSize,
			TotalTokens:       entry.TotalTokens,
			SidechainTurns:    entry.SidechainTurns,
			LastActiveAt:      entry.LastActiveAt.UTC(),
		}
	}
	return converted
}

func toSnapshotBreakdown(perModel []window.ModelBreakdown) []snapshot.ModelBreakdown {
	converted := make([]snapshot.ModelBreakdown, len(perModel))
	for index, entry := range perModel {
		converted[index] = snapshot.ModelBreakdown{
			Model:             entry.Model,
			InputTokens:       entry.InputTokens,
			OutputTokens:      entry.OutputTokens,
			CacheCreateTokens: entry.CacheCreateTokens,
			CacheReadTokens:   entry.CacheReadTokens,
			EffectiveTokens:   entry.EffectiveTokens,
			CostUSD:           entry.CostUSD,
		}
	}
	return converted
}
