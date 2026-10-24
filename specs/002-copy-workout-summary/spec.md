# Feature Specification: Copy Workout Summary

**Feature Branch**: `feat/copy-workout-summary`

**Created**: 2026-10-23

**Status**: Implemented (shipped 2026-10-23)

**Input**: User description: "Add a copy action to the overflow
menu of a training. The action generates a written summary of the
workout (exercises performed, type of training, rests) and copies
the summary to the clipboard, so the user pastes the summary into
a health application instead of sending a screenshot of the
finished workout every time."

## Clarifications

### Session 2026-10-23

The owner directed the cycle to proceed on the proposed defaults
of this draft. The session records the outcome per question.

- Q: Which fields does the summary carry? → A: The proposed set:
  training name, date, methodology, total duration, and for each
  exercise the name, the sets, the repetitions or duration, the
  load, and the rests. The training description and the user
  notes stay out.
- Q: For a training with recorded session actuals, the summary
  reports the actual values; otherwise the prescribed values.
  Does the text state which of the two the numbers are? → A: No
  marker. The wording stays neutral and never claims a
  performance, so the same format serves a planned training and
  a completed one.
- Q: Is the action available on every training, or on completed
  trainings only? → A: On every training. The neutral wording of
  the previous answer removes the distinction.
- Q: The summary language follows the application language of
  the user. → A: Confirmed. No per-action language choice.

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Copy the summary of a finished workout (Priority: P1)

A user finishes a workout and opens the training detail. In the
overflow menu the user taps the copy action. The application puts
a plain text summary on the clipboard and confirms the copy. The
user switches to the health application and pastes the summary
into the workout note.

**Why this priority**: The user performs this transcription after
every workout today, through a screenshot and manual reading. The
action removes a repeated chore at the exact moment the chore
arises. Nothing else in this cycle matters if the pasted text is
incomplete or wrong.

**Independent Test**: Complete a training, copy its summary, and
paste the clipboard content into any text field. The pasted text
names every exercise of the training with its sets, repetitions or
duration, load, and rests, under the training name and type.

**Acceptance Scenarios**:

1. **Given** a completed training, **When** the user taps the copy
   action in the overflow menu, **Then** the clipboard holds the
   summary text and the interface confirms the copy.
2. **Given** a training with a warm-up, a work section, and a
   cooldown, **When** the user copies the summary, **Then** the
   text keeps the section order of the training and groups the
   exercises under their sections.
3. **Given** a training on a device without network access,
   **When** the user taps the copy action, **Then** the summary
   still reaches the clipboard: the action works on the data the
   application already holds.

### Edge Cases

- A training with a circuit or AMRAP block repeats the same
  exercises for rounds or for time. The summary states the block
  structure in words instead of repeating the exercise list once
  per round.
- A bodyweight exercise carries no load. The summary omits the
  load for that exercise instead of printing a zero value.
- The user taps the copy action twice. The clipboard holds the
  same summary; no duplicate or appended text appears.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: The overflow menu of the training detail MUST offer
  a copy summary action, next to the existing actions.
- **FR-002**: The action MUST generate a plain text summary of the
  training and MUST place the text on the system clipboard.
- **FR-003**: The summary MUST state the training name, the date,
  the methodology, and the total duration.
- **FR-004**: The summary MUST list every exercise of the training
  in training order, with the sets, the repetitions or the
  duration, the load where the exercise carries one, and the
  rests.
- **FR-005**: The interface MUST confirm the copy to the user
  immediately after the action.
- **FR-006**: The action MUST work without network access, from
  the training data the application already holds.
- **FR-007**: The summary language MUST follow the application
  language of the user.
- **FR-008**: When recorded session actuals exist for the
  training, the summary MUST report the actual values, under the
  answers to clarifications Q2 and Q3.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: For a completed training, the pasted summary
  contains 100% of the exercises of the training, each with sets
  and repetitions or duration, and no exercise outside the
  training.
- **SC-002**: The copy action takes one tap in the menu and shows
  its confirmation, replacing a screenshot and a manual
  transcription that today follow every workout.
- **SC-003**: With the network disabled, the action still
  produces the complete summary in 100% of attempts.

## Assumptions

- The target of the paste is a free text note field of a health
  application. The summary is plain text, with no rich formatting
  and no attachment.
- The summary derives only from the training record. No new data
  collection enters this cycle.
- Cycle `specs/001-session-actuals` decides what an actual value
  is. This cycle consumes that record when the record exists and
  stays correct without it.
