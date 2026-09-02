# Coding Standards — Non-Negotiable

> These standards apply to every change in this repository — new code, refactors,
> and bug fixes alike. They are not suggestions. Any AI assistant or human
> contributor must read this document before writing code and again before
> opening a PR.
>
> If a proposed change conflicts with a rule here, the change is wrong, not
> the rule. Push back and correct course.

---

## 1. Single Responsibility Principle

**Rule:** every `.go` file, every `struct`, every function does exactly one
thing. If you cannot describe the purpose of a file in one short sentence
without using the word "and", split it.

### Why

When a bug shows up in production, we should be able to point at the file
where it lives without a debugger. Files that mix concerns force the debugger
to become the map. Small, focused files are the map.

### How to apply

- One responsibility per file. If `executor.go` starts orchestrating steps,
  parsing bodies, and formatting HTTP responses, split into `executor.go`,
  `body_reader.go`, `response_writer.go`.
- One responsibility per struct. If a struct holds both cache state and the
  HTTP client that fills the cache, that's two structs.
- One responsibility per function. If a function does "fetch and transform
  and write", it's three functions with a fourth that composes them.
- Prefer short files. If a file crosses ~300 lines, ask whether it earned
  those lines or whether it collected responsibilities that don't belong
  together.

### Smell checks

- File name is a noun phrase joined by "and" or "or".
- Function name contains "and" or handles unrelated failure modes.
- Struct has fields that are only touched by disjoint subsets of its methods.
- Test file has helper setup that only some tests use.

Any of these means split.

---

## 2. Test-Driven Development

**Rule:** write the failing test first. Then write the minimum code to make
it pass. Then refactor. In that order. Every time.

### Why

Tests written after code get shaped by the code — they document what the
code happens to do, not what it should do. Tests written first document the
contract, and the code is forced to meet it. When a test fails, the test is
right until proven otherwise.

### How to apply

- Write the test. Run it. It **must** fail — if it passes before you write
  the code, the test is wrong.
- Write the minimum code to make it pass. No extra features.
- Refactor with the test as a safety net.
- When a test fails during later work, **do not modify the test to make it
  pass** unless you have first proven the test itself is wrong. The default
  assumption is: the code is wrong, fix the code.
- Tests describe intent. If the test says the function should return an
  error on empty input, and the code returns nil, the code is the bug.

### What is forbidden

- Adjusting an assertion so a failing test passes without a written
  justification for why the test was wrong.
- Deleting a failing test to unblock a change.
- Skipping / `t.Skip`-ing tests to "come back later".
- Adding a test only after the feature works ("post-hoc coverage"). This
  produces false confidence and misses edge cases the test-first flow would
  have caught.

### When a test genuinely is wrong

State it in the commit message: what the test asserted, why that assertion
was incorrect, what the correct assertion is. Then change it. This should
be rare.

---

## 3. Package and File Layout

**Rule:** packages are organised by feature, not by technical layer. Names
are readable at the call site. Sub-packages exist only when a feature has
genuine sub-features worth isolating.

### Why

When investigating a Goja-specific issue, the developer should navigate
straight to a `goja` folder without stopping at "utils" or "helpers" or
"common" first. Feature-oriented packages make the codebase self-indexing.

### How to apply

- Package name = the feature. `scripting`, `executor`, `shipper`,
  `secret`. Not `utils`, `helpers`, `common`, `types`.
- Sub-package = sub-feature of the parent. If Goja is one implementation of
  the scripting engine, it lives at `internal/scripting/goja/`. A future
  sidecar implementation lives at `internal/scripting/sidecar/`. Both are
  under `scripting` because they are variants of the same feature.
- Package names at the call site should read naturally.
  `scripting.Engine` — good. `engineutil.EngineInterface` — bad.
- Do not create a package for a single file that could live in its parent.
- Do not split for the sake of splitting. Two files in the same package
  that share types belong together.

### File naming

- Lowercase, snake_case where multi-word: `runtime_pool.go`, not
  `RuntimePool.go` or `runtimepool.go`.
- Name describes the contents: `engine.go` for the engine, `pool.go` for
  the pool, `contract.go` for the script contract types.
- Test files pair with source: `engine.go` ↔ `engine_test.go`.
- Do not name files `misc.go`, `util.go`, `helpers.go`, `common.go`.

### Anti-patterns

- `internal/util/`, `internal/common/`, `internal/helpers/` — banned.
- `pkg/types/` — types belong with the code that owns them.
- Deep nesting without justification: `internal/scripting/engine/impl/goja/`
  when `internal/scripting/goja/` says the same thing.

---

## 4. Naming — Variables and Functions

**Rule:** names describe intent. Length is unbounded within Go's own
limits. Single-letter and abbreviated names are forbidden except for the
narrow list below.

### Why

Code is read many more times than it is written. `activePipelineCount` is
free at read time; `n` costs a mental lookup every time. We optimise for
reading.

### How to apply

- Variables: describe what the value represents. `partnerResponseBytes`,
  not `resp` or `data`. `secretsResolvedForPipeline`, not `secrets` when
  the scope has multiple secret-shaped things.
- Functions: verb phrase describing the action. `LoadScriptForPipeline`,
  not `Load` in a context where multiple things get loaded.
- Booleans read as questions. `isScriptCompiled`, `hasReachedTimeout`,
  `shouldRetry` — not `flag`, `ok`, `done` when the meaning is specific.
- Constants: describe the value's role, not its literal value.
  `DefaultExecutionTimeoutMs` — not `FIVE_THOUSAND`.

### Allowed short names

- `i`, `j` — loop indexes over integer ranges.
- `k`, `v` — key/value in a map range where the semantic name adds
  nothing.
- `err` — the conventional Go error variable.
- `ctx` — the conventional Go context variable.
- `r`, `w` — HTTP handler request and response writer at the outermost
  handler signature only. Inside the function body, rename to something
  meaningful if the handler grows past a few lines.
- Receiver names: 1–2 letters is idiomatic Go (`e *Executor`, `p *Pool`).

Nothing else is short by default. If in doubt, spell it out.

### Forbidden

- `a`, `b`, `c`, `x`, `y`, `z` as data variables.
- `tmp`, `temp`, `foo`, `bar`, `thing`, `obj`, `res`, `resp2`.
- `data`, `info`, `stuff` — always too vague; describe what data.
- Hungarian notation (`strName`, `intCount`) — the type is already visible.
- Abbreviations that are not universally understood in the domain.
  `PplnID` is not clearer than `PipelineID`.

---

## Enforcement — pre-flight checklist before any code change

Before writing code, confirm:

1. **SRP** — I know exactly which file each new responsibility lives in,
   and each file has one job.
2. **TDD** — The first change I make is a failing test. I've written it,
   run it, and confirmed it fails for the right reason.
3. **Layout** — Every new package name is a feature. No `util`, no
   `common`, no `helpers`. Sub-packages only where sub-features exist.
4. **Naming** — Every identifier I've written describes its intent. No
   forbidden short names in this file.

If any of the four is not yet true, stop and fix before continuing.

---

## When these standards conflict with existing code

The existing code does not grandfather in violations. If you touch a file
that violates a standard, either:

- Fix the violation in the same change (preferred when the fix is local), or
- Note the violation and file a follow-up (when the fix would balloon scope).

Never propagate a violation forward by matching the surrounding style.