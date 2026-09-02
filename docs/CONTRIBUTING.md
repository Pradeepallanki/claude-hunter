# Contributing to Claude Hunter

Thanks for considering a contribution. This project is small on
purpose; keeping it small is a design goal.

Before touching any code, please read
[`coding_standards.md`](coding_standards.md). Every rule in there is
non-negotiable — SRP per file, TDD first, feature-based packages, and
descriptive names. If a proposed change conflicts with a rule, the
change is wrong, not the rule.

---

## Development workflow

1. **Fork and branch.** Branch names describe intent — `fix/burn-rate-zero`,
   `feature/per-model-breakdown`, `docs/troubleshooting`.
2. **Write the failing test first.** For any change to `core/`, this is
   a hard requirement: `git blame` should show the test file appearing
   in the same commit as (or before) the source file.
3. **Implement the minimum code to pass.** No extra features, no
   speculative abstractions.
4. **Refactor with tests as a safety net.**
5. **Confirm the whole suite is green.** `cd core && go test ./...`.
6. **Rebuild the bundled binaries** if you changed anything under
   `core/` — see [`BUILDING.md`](BUILDING.md).
7. **Update documentation.** New flags belong in
   [`CONFIGURATION.md`](CONFIGURATION.md); new snapshot fields in
   [`ARCHITECTURE.md`](ARCHITECTURE.md); newly-observed failure modes
   in [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md).
8. **Open a PR** with a clear title, a short *what and why* summary,
   and a manual test plan.

---

## Commit and PR conventions

- **One logical change per commit.** Bundling unrelated fixes into
  one commit makes reviewers work harder than necessary and makes
  reverts painful.
- **Present-tense imperative subject line, ≤ 72 chars.** Example:
  `window: attribute cost per model in the 5h rolling summary`.
- **Body wraps at 72 chars** and explains *why*. The *what* is in the
  diff.
- **PR titles** follow the same convention as commit subjects.

---

## Tests

- **`core/`** uses standard `testing`. Every package that has runtime
  logic has a `_test.go` next to the source file. Integration tests
  that touch the filesystem use `t.TempDir()`.
- **`vscode/`** currently ships no automated tests — end-to-end
  verification is manual through the diagnostic log. Adding a small
  vitest / mocha suite for `hunter_process.ts` and
  `tooltip_renderer.ts` is welcome.
- **`intellij/`** ditto — a plugin-verifier smoke test is a good
  starting point.

If a test genuinely fails after your change, the default assumption
is that the code is wrong. Modifying the assertion to make the test
pass without a written justification is forbidden — see
[`coding_standards.md § 2`](coding_standards.md#2-test-driven-development).

---

## Code review focus

Reviewers will look for, in this order:

1. **Standards compliance.** SRP violations, forbidden names, missing
   tests. These get flagged first because they are cheap to catch
   and expensive to fix later.
2. **Correctness.** Does the change do what the PR claims?
3. **Simplicity.** Could it be smaller? Any speculative abstractions?
4. **Documentation.** New behaviour reflected in the right doc?

---

## Reporting issues

Open a GitHub issue with the information from
[`TROUBLESHOOTING.md § Filing a bug`](TROUBLESHOOTING.md#filing-a-bug).
"It doesn't work" is not enough for us to help.

For security-sensitive reports, please email the maintainers directly
rather than filing a public issue.

---

## License of contributions

By opening a pull request you agree that your contribution is
licensed under the MIT License, the same license as the rest of the
project ([`LICENSE`](../LICENSE)).
