# Implementation Plan: Copy Workout Summary

**Branch**: `feat/copy-workout-summary` | **Date**: 2026-10-23 |
**Spec**: [spec.md](spec.md)

## Summary

The training detail overflow menu gains a copy summary action.
The action builds a plain text summary of the training on the
device, from the record the application already holds, and puts
the text on the system clipboard with an immediate confirmation.
Recorded session actuals (cycle `specs/001-session-actuals`)
replace the prescribed values where they exist; a skipped
activity leaves the summary. No network, no backend change.

## Technical Context

**Language/Version**: Dart/Flutter (application only)

**Primary Dependencies**: the framework clipboard service; the
labels the application already resolves for methodology and
routine phases

**Storage**: none

**Testing**: unit tests on the summary builder as a pure
function; the full application suite as the regression net

**Target Platform**: Android application

**Project Type**: mobile + API (application side only)

**Constraints**: the summary derives only from the training
record and the user feedback record the detail already loads;
the wording never claims a performance, so one format serves
planned and completed trainings

## Constitution Check

- Data Compatibility: no schema change, no write path change.
- Production Care: the change ships in the application build;
  the backend rollout is untouched.
- Tests and the Coverage Gate: the builder is a pure function
  and carries the tests; the widget wiring stays thin.
- Minimal Owned Surface: no new dependency.

## Project Structure

### Documentation (this feature)

```text
specs/002-copy-workout-summary/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```text
app/
├── lib/utils/workout_summary.dart  # the pure summary builder
├── lib/screens/training_details_screen.dart  # menu action
├── l10n/app_en.arb, app_it.arb     # action and summary labels
└── test/workout_summary_test.dart
```

**Structure Decision**: the builder takes the training, the
recorded actuals, and the resolved labels as inputs and returns
the text. Localization and date formatting stay at the call
site, so the builder tests run without a widget tree or a
locale setup.
