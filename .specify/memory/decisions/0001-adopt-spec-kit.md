# Adopt Specification-Driven Development on Vigor

## Status

Proposed. The repository owner ratifies this decision.

## Context and Problem Statement

One agent executes most of the work on this repository, and the
application serves real users in production. Decisions live in
chat threads and in the agent memory, outside the repository. A new
session cannot resume the work from the repository files alone: the
session must reconstruct the state from conversation history.
Feature requests arrive from daily use of the production
application, and each request deserves a persistent path from the
request to the specification, to the implementation, and to the
verification. The fleet standard (unciae §25) prescribes
Specification-Driven Development and names this adoption trigger:
an existing project adopts the section at its next feature change.

## Considered Options

1. Keep the current method. Decisions stay in chat and agent
   memory, and feature work starts from conversation alone.
2. Adopt GitHub Spec Kit with the framework layout. Decisions go
   to a register under `.specify/memory/decisions/`, in MADR
   format, and feature artifacts live in `specs/NNN-<short-name>/`.
3. Adopt the specification cycle with a custom layout. Decisions
   and specifications go to a `docs/` directory, against the
   layout of the framework.

## Decision Outcome

Chosen option: option 2.

Vigor adopts GitHub Spec Kit v1.2.2, pinned, with the `generic`
integration and the command files under `.specify/commands/`. The
repository commits the Spec Kit artifacts, per the framework
guidelines; the machine-local files stay excluded through the
`.gitignore` that the CLI manages. The constitution derives from
the unciae golden standard, and the canonical standard prevails on
conflict. A completed feature specification stays as an immutable
historical record. The first cycles are `specs/001-session-actuals`
and `specs/002-copy-workout-summary`, opened by the feature
requests that triggered this adoption.

Justification: option 1 leaves the problem unchanged. Option 3
fights the framework at every upgrade for no functional gain.
Option 2 matches the layout that the fleet pilot already proved on
a sibling project, so one operative form serves every adopted
project.

## Consequences

- Positive: a new session resumes from repository files alone.
- Positive: every significant decision gains an authority, a
  status, and a supersession chain.
- Positive: a feature request from daily use becomes a tracked
  cycle with verifiable success criteria before code exists.
- Negative: the project carries method artifacts that must stay
  current, or the register decays into noise.
- Neutral: the `generic` integration gives command files as
  documents. The agent follows the documents; the harness gains no
  native commands. The value is the artifact discipline, not the
  tooling.
