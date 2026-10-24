# Tasks: Selection Context Metadata

**Input**: Design documents from `specs/003-selection-context/`

**Prerequisites**: plan.md, spec.md (investigation recorded
2026-10-24; the compact variant is the ratified direction)

## Phase 1: User Story 1 - Choose between near-equal candidates (P1)

- [ ] T001 Add the parenthetical helper to
      backend/llm/prompt/utils.go (lead part, difficulty above
      zero, equipment, cues segment; empty when nothing shows)
- [ ] T002 Render the helper in all three pools and add the
      metadata lines to the system prompt in
      backend/llm/prompt/gen_training_exercises.qtpl;
      regenerate the compiled template
- [ ] T003 [P] Rendering tests in
      backend/llm/prompt/gen_training_exercises_test.go:
      full metadata, partial metadata, bare candidate,
      warmup and cooldown shapes, system prompt lines

## Phase 2: Polish and verification

- [ ] T004 Full backend suite and prompt package coverage
      locally, with the stub tier environment
- [ ] T005 Pull request with the trajectory-harness gap
      declared; CI green before the merge decision returns to
      the owner
