# Feature Specification: Session Actuals at Feedback Time

**Feature Branch**: `feat/session-actuals`

**Created**: 2026-10-23

**Status**: Implemented (shipped 2026-10-23)

**Input**: User description: "At feedback time, let the user edit
the repetitions performed and the loads used for each exercise, and
rename the exercise performed. Example: the training prescribed 15
repetitions of bench press with 40 kg; the user performed the set
with 30 kg."

## Clarifications

### Session 2026-10-23

- Q: At which granularity does the user record the actual values?
  One value per activity, or one value per single set? → A: One
  value per activity, valid for every set of the block. A per-set
  record is a later cycle if a real need appears.
- Q: What does renaming mean? A display label on this training, or
  a re-link to a different library exercise? → A: A display label
  on the record of this training only. Statistics and progression
  keep the original exercise identity.
- Q: Does the recording cover duration-based activities with an
  actual duration, symmetric to repetitions? → A: Yes.
- Q: Can the user mark an activity as not performed at all? → A:
  Yes, a per-activity not-performed flag. A partially performed
  block (fewer sets than prescribed) stays out of scope for this
  cycle: the block repeats record does not exist, and the
  qualitative feedback already carries that signal.
- Q: Do the recorded actual values feed the next generations as
  the performed history? → A: Yes. Progression starts from the
  load the user actually lifted.

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Record the load and repetitions actually performed (Priority: P1)

A user completes a training and opens the feedback step. The
feedback form lists each exercise with the prescribed repetitions
and load, prefilled as the actual values. The user changes only
the values that differ from the prescription and saves. The
training now records what the user actually did, next to what the
training prescribed.

**Why this priority**: Today the system keeps only the
prescription and a qualitative signal per exercise. Every
progression decision after a session where the user deviated rests
on numbers nobody recorded. This story closes that gap; the rename
and the remaining stories build on it.

**Independent Test**: Complete a training, open the feedback step,
change one load and one repetition count, save, and reopen the
training. The detail shows the actual values for the changed
exercises and the prescribed values everywhere else.

**Acceptance Scenarios**:

1. **Given** a completed training with a prescribed bench press of
   15 repetitions at 40 kg, **When** the user records 15
   repetitions at 30 kg in the feedback step and saves, **Then**
   the training stores the actual values 15 and 30 kg for that
   exercise and still shows the prescription 15 at 40 kg.
2. **Given** a training where the user performed every exercise as
   prescribed, **When** the user opens the feedback step and saves
   without changes, **Then** the flow takes no more steps than the
   feedback flow takes today and the recorded actuals equal the
   prescription.
3. **Given** a training with recorded actuals, **When** the user
   reopens the feedback step, **Then** the form shows the recorded
   actuals, and a second save updates them without duplicating the
   record.

### User Story 2 - Rename the exercise performed (Priority: P2)

During the session the user performed a variant of the prescribed
exercise. At feedback time the user edits the exercise name for
this training, so the record names the movement the user actually
performed.

**Why this priority**: A wrong name in the record corrupts the
history the same way a wrong load does. The request arrived
together with the actual values, and the feedback step is the one
moment where the user still remembers the substitution.

**Independent Test**: At feedback time, rename one exercise and
save. The training detail shows the new name for that exercise;
the library and every other training keep the original name.

**Acceptance Scenarios**:

1. **Given** a completed training, **When** the user renames an
   exercise in the feedback step and saves, **Then** the training
   record shows the new name and the exercise library entry stays
   unchanged.
2. **Given** a renamed exercise, **When** the user clears the name
   field, **Then** the form rejects the empty name and keeps the
   previous value.

### Edge Cases

- The user deviates on a duration-based activity instead of a
  repetition-based one. The behavior follows the answer to
  clarification Q3.
- The user performed fewer sets than the block prescribes. This
  cycle records no per-set data; the qualitative per-exercise
  feedback carries the signal, as today.
- The feedback arrives twice for the same training, from a retry
  or a second device. The second save replaces the first record;
  the training never carries two actual records for one activity.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: The feedback step MUST present every exercise of the
  completed training with its prescribed repetitions or duration
  and its prescribed load.
- **FR-002**: The feedback form MUST prefill the actual values
  with the prescribed values, so the user edits only deviations.
- **FR-003**: The user MUST be able to change the actual
  repetitions, the actual duration where applicable, and the
  actual load of an exercise before saving the feedback.
- **FR-004**: Saving the feedback MUST store the actual values
  next to the prescription. The prescription MUST stay unchanged:
  a generated training is a historical record.
- **FR-005**: The training detail MUST show the recorded actual
  values for a completed training, and MUST make a deviation from
  the prescription visible as a deviation.
- **FR-006**: The user MUST be able to rename the performed
  exercise for this training in the same feedback step. The rename
  MUST NOT change the exercise library or any other training.
- **FR-007**: A repeated feedback save for the same training MUST
  update the existing actual record, never duplicate it.
- **FR-008**: The recorded actual values MUST be available to the
  generation pipeline as the performed history of the user, under
  the answer to clarification Q5.
- **FR-009**: Existing feedback behavior (the qualitative signal
  per exercise and the session rating) MUST keep working unchanged
  for trainings with and without recorded actuals.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: A user who deviated on one load completes the
  feedback step with at most one edited field and one save, and
  the training detail afterwards shows the actual load.
- **SC-002**: In a test session with deviations on load,
  repetitions, and one exercise name, 100% of the deviations are
  present in the stored record after one save, and 0 prescribed
  values changed.
- **SC-003**: The feedback flow for a session without deviations
  takes the same number of user actions as the flow takes today.
- **SC-004**: After the change ships, the next generated training
  for a user with recorded actuals cites the performed load, not
  the prescribed load, in its progression rationale.

## Assumptions

- The feedback step stays the single entry point for post-session
  corrections. No live, mid-session recording enters this cycle.
- Actual values are per activity in the record: one value set per
  activity, valid for every set of its block.
- The qualitative per-exercise feedback that exists today stays;
  actual values complement the signal, they do not replace it.
- A rename touches the record of one training. Statistics and
  progression keep the original exercise identity, unless
  clarification Q2 decides otherwise.
