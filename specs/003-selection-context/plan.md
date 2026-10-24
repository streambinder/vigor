# Implementation Plan: Selection Context Metadata

**Branch**: `feat/selection-context` | **Date**: 2026-10-24 |
**Spec**: [spec.md](spec.md)

## Summary

The exercise selection prompt renders each pool candidate as a
bare ID with small annotations. This cycle adds the compact
metadata the knowledge record already carries and the node
never sees: the difficulty, the full equipment list, and the
form cues where present. One parenthetical per candidate keeps
the shape: the existing annotations or the primary muscle
first, then the difficulty marker, the equipment, and the cues
segment, separated by a middle dot. The system prompt gains one
line that names the new metadata and one that fixes its role:
guidance for the choice, never a constraint. Nothing else in
the pipeline changes.

## Technical Context

**Language/Version**: Go (backend only)

**Primary Dependencies**: the quicktemplate templates under
llm/prompt, regenerated with the pinned qtc version; the
knowledge records as already loaded into model.Exercise

**Storage**: none; the fields already reach the node inside
the pool records

**Testing**: prompt rendering tests in the prompt package; the
full backend suite and the package coverage gate as the
regression net; the generation trajectory harness stays
declared unavailable in this environment (see the spec)

**Target Platform**: production backend

**Project Type**: mobile + API (backend side only)

**Constraints**: the output schema of the node, the
deterministic contraindication and proficiency filters, and
the pool retrieval stay byte-identical in behaviour; the added
prompt text averages about 15 tokens per candidate

## Constitution Check

- Generation Integrity: the change enriches the input of one
  node; it moves no decision into or out of the deterministic
  layer, and the trajectory of a generation keeps its shape.
- Data Compatibility: no schema change on any record.
- Tests and the Coverage Gate: the rendering is a pure
  function of the record and carries the tests.
- Minimal Owned Surface: no new dependency; the template
  helper joins the existing prompt utilities.

## Project Structure

### Documentation (this feature)

```text
specs/003-selection-context/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── llm/prompt/gen_training_exercises.qtpl      # pool rendering, system line
├── llm/prompt/gen_training_exercises.qtpl.go   # regenerated
├── llm/prompt/utils.go                         # the parenthetical helper
└── llm/prompt/gen_training_exercises_test.go   # rendering tests
```

**Structure Decision**: one helper builds the parenthetical
content from a lead part (the annotations for work candidates,
the primary muscle for warmup and cooldown) plus the metadata
parts, and returns an empty string when there is nothing to
show, so the templates keep a single conditional print.
