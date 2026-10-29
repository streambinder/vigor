# Implementation Plan: Animation Catalog Generation

**Branch**: `feat/animation-catalog` | **Date**: 2026-10-27 |
**Spec**: [spec.md](spec.md)

## Summary

The knowledge image build gains one stage and the bootstrap
gains one step. A Python stage runs the committed parametric
recipes against the committed mannequin skeleton and writes one
minified animation document per mapped catalog exercise,
validating each document before the stage ends. The data
stage copies the documents next to the feature files, and the
bootstrap injects them into a new `exercise_animations` table
while it bakes the database dump. The repository commits only
sources: recipes, mapping, model. The documents exist only
inside the built image, versioned with the catalog by
construction.

## Technical Context

**Language/Version**: Python 3.12 (build stage only), Go
(knowledge bootstrap)

**Primary Dependencies**: numpy 1.26.4 and scipy 1.17.1,
pinned, in the generation stage; the final image gains no
new runtime dependency

**Storage**: one new table, `exercise_animations`, keyed by
exercise identifier, payload the minified document as text

**Testing**: generator validation as a build gate; knowledge
bootstrap tests on the fake wire server for the ingestion
paths; the repository CI jobs as the regression net

**Target Platform**: production knowledge image

**Project Type**: data pipeline inside an existing deployable
unit

**Constraints**: deterministic output under the pinned
toolchain; the bootstrap keeps working without the generated
documents (local runs, tests); no consumer reads the new
table in this cycle

## Constitution Check

- One Product, Bounded Components: the generation lives
  inside the knowledge unit, which already owns the catalog.
  No new component or service appears.
- CI Is the Source of Truth: the generator validation runs
  in the image build, and the ingestion paths run in the
  knowledge test job. A red build publishes nothing.
- Minimal Owned Surface: the owned surface is the recipe
  set, which is the product logic. Emission, validation, and
  storage reuse the existing build and bootstrap machinery.
- Data Compatibility: the change is additive. One new table,
  no change to existing records or consumers.
- Tests and the Coverage Gate: the new bootstrap step
  carries tests for its success, skip, malformed, and store
  error paths.

## Project Structure

### Documentation (this feature)

```text
specs/004-animation-catalog/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```text
knowledge/
├── Dockerfile                # generation stage, documents copied to the data stage
├── main.go                   # exercise_animations table, bootstrapAnimations step
├── main_test.go              # ingestion path tests
└── features/
    └── animations/
        ├── README.md             # provenance, generation, format
        ├── generate.py           # orchestration, validation, minified emission
        ├── mannequin.json        # the model skeleton, build input
        ├── build.py              # pool builder, discovery from exercises.json
        ├── rig.py                # model loading, rest pose, quaternion math, IK, swing
        └── recipes/
            ├── engine.py             # shared construction machinery
            ├── emit_frames.py        # pose frames to animation documents
            ├── standing.py           # squat-family generic builder
            ├── hinge.py              # hinge-family generic builder
            ├── lying.py              # lying-press generic builder
            ├── hanging.py            # hanging generic builder
            ├── upright.py            # upright-arm generic builder
            └── <exercise_id>.py      # one builder module per exercise
```
