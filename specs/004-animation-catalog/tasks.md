# Tasks: Animation Catalog Generation

**Input**: Design documents from `specs/004-animation-catalog/`

**Prerequisites**: plan.md, spec.md (investigation recorded
2026-10-27; the generation-at-build form is the owner
direction)

## Phase 1: User Story 1 - One image carries catalog and animations (P1)

- [x] T001 Vendor the recipe set, the mannequin skeleton, and
      the exercise mapping under `knowledge/recipes/`, with
      the generator entry point in
      `knowledge/recipes/generate.py` (orchestration,
      validation, minified emission) and the provenance note
      in `knowledge/recipes/README.md`
- [x] T002 Verify the vendored generator against the
      evaluated pool: canonical hashes equal for all 36
      animations, before and after lint normalization of the
      vendored code
- [x] T003 Add the generation stage to
      `knowledge/Dockerfile` (pinned Python, numpy, scipy)
      and copy the generated documents into the data stage
- [x] T004 Add the `exercise_animations` table and the
      `bootstrapAnimations` step to `knowledge/main.go`:
      inject when the directory is present, skip when it is
      absent, fail on a malformed document
- [x] T005 [P] Ingestion tests in `knowledge/main_test.go`:
      directory present with a non-document file, directory
      missing, malformed document, store error

## Phase 2: Polish and verification

- [x] T006 Knowledge test suite and coverage gate locally;
      full Go suite through the generation trajectory
      harness runner before publishing
- [x] T007 Lint and format: ruff with the super-linter
      configuration on the recipes, gofumpt and golangci-lint
      on the knowledge module
- [x] T009 Restructure the generator: engine.py plus one module
      per exercise plus family generics, discovery from
      exercises.json, mapping.json removed, fingers.py removed
      (22-bone documents); verified: body bones and root
      identical for all 36, finger tracks removed, displayName
      from the catalog, source per exercise identifier
- [x] T011 Move the shared recipe machinery (engine,
      emit_frames, the five generics) under `recipes/` and
      rename the pool builder to `build.py`; verified:
      generated documents byte-identical to the pre-move set,
      CI green job by job including the knowledge image build
- [x] T010 Reorganize the generator under
      `knowledge/features/animations/`, with the 36 exercise
      modules under `recipes/`; verified: generated documents
      byte-identical to the pre-move set, CI green job by job
      including the knowledge image build
- [ ] T008 CI green on the pull request, verified job by job
      including the knowledge image build; production row
      count observed after the rollout (SC-004)
