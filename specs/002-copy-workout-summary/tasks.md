# Tasks: Copy Workout Summary

**Input**: Design documents from `specs/002-copy-workout-summary/`

**Prerequisites**: plan.md, spec.md (clarifications resolved
2026-10-23)

## Phase 1: User Story 1 - Copy the summary of a finished workout (P1)

- [x] T001 Implement the pure summary builder in
      app/lib/utils/workout_summary.dart (header, sections in
      training order, block structure in words, bodyweight load
      omitted, actuals override, skipped activities omitted)
- [x] T002 [P] Unit tests for the builder in
      app/test/workout_summary_test.dart
- [x] T003 Add the copy summary action to the training detail
      overflow menu in
      app/lib/screens/training_details_screen.dart, with the
      clipboard write and the confirmation notification
- [x] T004 Add the action and summary labels to app/l10n
      (English and Italian; other locales fall back)

## Phase 2: Polish and verification

- [x] T005 Full application analyze and test locally
- [x] T006 Pull request; CI green before the merge decision
      returns to the owner
