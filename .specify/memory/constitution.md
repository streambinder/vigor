# Vigor Constitution

**Status**: Draft. The repository owner ratifies this document.

The canonical standard for this repository is the unciae golden
standard (`streambinder/unciae`, file `AGENTS.md`). This constitution
restates the rules of that standard that bind this repository, in the
form that the Spec Kit workflow checks. When a rule here and the
canonical standard disagree, the canonical standard wins and this
document gets amended.

## Core Principles

### I. One Product, Bounded Components

The repository builds one product from bounded components: the Go
backend, the Flutter application, the knowledge base, and the
cockpit. Each component keeps its own layout and toolchain. The root
stays minimal: a new root entry needs a justification. Configuration
files live in their conventional locations, never ad hoc.
(unciae §1, §2)

### II. CI Is the Source of Truth

A change is done when the continuous integration pipeline is green.
Local checks mirror the pipeline; local checks never replace the
pipeline. Do not push on red. (unciae §1, §17, §18)

### III. One Linting System

super-linter is the linter. Its configuration files live only under
`.github/linters/`. No linter configuration sits at the repository
root, not even as a symlink. Format code with the canonical formatter
of the language before every push: gofumpt for Go, the Flutter SDK
formatter for Dart, prettier for JSON, YAML, and Markdown.
(unciae §3)

### IV. Tests and the Coverage Gate

Unit tests live next to the code that the tests cover. Integration
tests live separately and carry their build tag. CI runs the tests
on every push. A failing test blocks the merge. CI measures unit
test coverage on every push. The minimum coverage is 90%. Coverage
below 90% blocks the merge. Non-test code never changes to satisfy
a test; the test changes instead. (unciae §15)

### V. Minimal Owned Surface

The repository owns its product logic and nothing else. For every
other functionality, first simplify the problem away, then delegate
to a specialized library with an established reputation. Pin exact
dependency versions where possible. (unciae §16, §16.1)

### VI. Commits and Pull Requests

Commits follow Conventional Commits: one logical change per commit,
signed, with no artificial-intelligence attribution trailers. Commit
messages and pull request descriptions never name a verification
subject; describe the class of the subject and the outcome instead.
The owner merges, except where the owner has delegated a merge in
writing for a stated class of change. (unciae §7, §19, §24)

### VII. Containers and Releases

Build for linux/arm64 only. Pin every action and base image.
Publish never runs on a pull request. A merge to the default branch
deploys the backend; the deploy is part of the change, not an
afterthought. (unciae §5, §10, §11)

### VIII. Writing and Consistency

Technical writing follows ASD-STE100: short sentences, one term for
one concept, active voice. A change follows the conventions of the
file that the change touches. Do not reformat parts that the change
does not touch. (unciae §22, §23)

## Project Constraints

These constraints are decisions of the repository owner for this
project. Each constraint stands until the owner replaces it with a
recorded decision.

### Production Care

The backend serves real users in production. A merge to the default
branch rolls out to production. Before a change is declared ready,
the author verifies the change against the live service shape:
routes, data, and configuration as production runs them. After a
deploy, verify the running service and the real content. Never
write on production data to test.

### Health Data

Health records belong to one user each and stay server-side. No
feature reads another user's health data. Analytics access to the
production database never includes the health tables.

### Generation Integrity

The training generation pipeline and the knowledge base are the
product core. A change to either one is evaluated against the full
generation trajectory before the change is declared ready, not
against a single node or a single output. Knowledge entries carry
a verifiable source.

### Data Compatibility

Production data is a contract. A schema change is additive, or the
change carries a recorded decision and a migration path before
implementation. A generated training that a user completed is a
historical record: later changes never rewrite its prescription.

### Specification Cycle

Every feature change and every domain change follows the
specification cycle: specify, plan, tasks, implement, verify,
converge. The artifacts live in `specs/NNN-<short-name>/` at the
repository root. A completed feature directory is an immutable
historical record. A bugfix and a small change skip the cycle; the
pull request body states the rationale. A significant decision goes
to the decision register under `.specify/memory/decisions/`, in
MADR format, one file per decision. A decision is a draft until
the owner ratifies it. An accepted decision is reused; a superseded
decision stays in the register and names its successor.

## Working Agreements

This section restates the operative form of the principles, so the
repository stays self-contained for the points that the golden
standard states in full. A new session applies these agreements
without further sources.

### Branches and commits

- A feature branch carries the name `feat/<short-name>`. The name
  also covers fixes and chores.
- A commit follows Conventional Commits. The subject holds at most
  72 characters, uses the imperative, carries no terminal period,
  and contains none of the characters `()[]"'`/,:;` inside the
  subject text.
- A commit body holds one short paragraph. Each body line holds at
  most 100 characters. The body states why, not what, and ends with
  a period.
- One logical change makes one commit.
- A fix for a red feature branch amends the failing commit. The
  push uses force with an explicit lease that names the remote
  commit. Nobody amends or force-pushes the default branch.

### Text and formatting

- Prettier formats every authored Markdown file. The author runs
  prettier, in the major version the lint image bundles, on each
  authored file before the push.
- Codespell checks every text file in CI. Authored text passes
  codespell before the push.
- Spec Kit manages the files under `.specify/`, except
  `.specify/memory/`. Nobody edits a managed file by hand. The lint
  pipeline excludes the managed subtrees through one targeted rule.

## Governance

The owner ratifies this constitution and every amendment. The agent
drafts amendments and records the ratified text. Every pull request
review checks compliance with this constitution. A rule that proves
wrong is amended through a recorded decision; the rule is never
silently ignored.

**Version**: 1.0.0 | **Ratified**: pending | **Last Amended**: —
