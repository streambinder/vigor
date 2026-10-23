# Tasks: Session Actuals at Feedback Time

**Input**: Design documents from `specs/001-session-actuals/`

**Prerequisites**: plan.md, spec.md (clarifications resolved
2026-10-23)

## Phase 1: Foundational (record and write path)

- [x] T001 [P] Add the ActivityActual type, the additive feedback
      column, and the typed accessor in backend/model/training.go
- [x] T002 Add the activityActuals field to the complete request
      in backend/handler/dto/training.go
- [x] T003 Validate actuals against the training activities and
      upsert them in backend/service/training.go; thread the new
      argument through CompleteTraining and UpdateTrainingFeedback
- [x] T004 Map the invalid-actuals error to a 400 response in
      backend/handler/training.go
- [x] T005 [P] Unit tests for validation and upsert in
      backend/service/training_actuals_test.go
- [x] T006 [P] Unit tests for the model accessor in
      backend/model/training_actuals_test.go

## Phase 2: User Story 1 - Record actual repetitions and load (P1)

- [x] T007 [US1] Fold actuals into the performed history in
      backend/llm/dag.go (signals and history prompt read one
      folded copy; skipped activities drop out)
- [x] T008 [P] [US1] Unit tests for the folding in
      backend/llm/dag_actuals_test.go
- [x] T009 [US1] Regenerate the Dart models with the repository codegen
      and add the actuals field to the app request DTO and
      service call
- [x] T010 [US1] Per-activity actuals editor in the feedback
      modal (prefilled with the prescription, skip toggle) in
      app/lib/utils/feedback_modal.dart
- [x] T011 [US1] Show recorded actuals and deviations in the
      training detail in
      app/lib/screens/training_details_screen.dart
- [x] T012 [P] [US1] Widget tests for the editor prefill and the
      deviation rendering in app/test/

## Phase 3: User Story 2 - Rename the exercise performed (P2)

- [x] T013 [US2] Name field in the actuals editor, saved as the
      display label of the activity in this training
- [x] T014 [US2] The training detail and the folded history show
      the recorded name; the library name stays untouched

## Phase 4: Polish and verification

- [x] T015 Full backend suite and coverage gate locally; flutter
      analyze and test locally
- [ ] T016 Pull request with the trajectory-harness gap declared;
      CI green before the merge decision returns to the owner
