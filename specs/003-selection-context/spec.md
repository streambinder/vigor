# Feature Specification: Selection Context Metadata

**Feature Branch**: `feat/selection-context`

**Created**: 2026-10-24

**Status**: Draft

**Input**: Owner direction: "the exercise selection node chooses
from the exercise name alone, while the knowledge base carries
instructions and more; give the node more to choose with." The
owner selected the compact metadata variant over full
instruction text after the investigation below.

## Clarifications

### Investigation 2026-10-24 (hypothesis challenge)

The premise was checked against the code and the knowledge
base before any design.

- What the node sees today, per candidate: the exercise ID, the
  primary muscle as the grouping key, the mode annotation
  (timer-only or reps-only), a recent flag, and a weighted flag.
  The warmup pool adds the equipment list. Nothing else.
- The instructions already work upstream: the retrieval
  embeddings are computed over the name, the equipment, the
  muscles, and the full instruction text, so the candidate pool
  itself (16 to 40 work exercises by session duration, 6
  warmup, 2 per muscle of cooldown) is instruction aware. Only
  the final choice is metadata blind.
- What the knowledge base holds and the node never sees: the
  difficulty (full coverage, today only a hard retrieval
  filter), the full equipment list for work exercises, and the
  form cues (60 percent coverage, about 77 characters each).
- Production evidence: no recorded selection failure is
  attributable to missing instructions. The two known
  selection-adjacent failures were a deterministic filter drop
  and a load-node mode assignment, both fixed at their own
  layer. The expected gain is discrimination among near-equal
  candidates whose IDs understate the movement, not a repair.
- Full instruction text in the prompt was costed and set
  aside: about 90 tokens per candidate, 4,000 to 6,000 extra
  input tokens per call before retries, against a comparison
  task. The compact set costs about 15 tokens per candidate.

### Decisions

- The enrichment is compact metadata: difficulty, full
  equipment, and form cues where the record carries them.
- The metadata guides the choice; it constrains nothing.
  Difficulty bounds and equipment availability stay enforced
  deterministically before the pool is built.
- All three pools (work, warmup, cooldown) render the same
  metadata shape.

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Choose between near-equal candidates (Priority: P1)

Two candidates in a pool train the same muscle and read almost
alike by ID, but one is markedly harder, needs equipment the
other does not, or moves through a pattern its cues describe
and the ID hides. The selection node sees that difference and
picks with it in view, and its short rationale can reflect it.

**Why this priority**: This is the only gap the investigation
confirmed: discrimination at the final choice among candidates
the retrieval already judged relevant.

**Independent Test**: Render the selection prompt for a pool
that contains such a pair; the rendered candidate lines carry
the difficulty, the equipment, and the cues of each record.

**Acceptance Scenarios**:

1. **Given** a work candidate with a difficulty, an equipment
   list, and cues, **When** the selection prompt is rendered,
   **Then** the candidate line carries all three.
2. **Given** a work candidate without cues or equipment,
   **When** the prompt is rendered, **Then** the line carries
   the difficulty and no empty segments.
3. **Given** a candidate with no metadata at all, **When**
   the prompt is rendered, **Then** the line is the bare ID,
   exactly as before this cycle.

### Edge Cases

- A record whose difficulty is zero renders no difficulty
  segment; zero is the unrated default, not a rating.
- Cues render only when the record carries them; a candidate
  without cues is not marked as lacking them.
- The explicit-program path renders the same pool format; a
  pinned exercise keeps its mandatory status, and the metadata
  does not license a substitution.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: The selection prompt MUST carry, for every
  candidate in the work, warmup, and cooldown pools, the
  record difficulty when it is above zero.
- **FR-002**: The selection prompt MUST carry the full
  equipment list of every candidate that has one, in every
  pool.
- **FR-003**: The selection prompt MUST carry the form cues of
  every candidate that has them, joined as one segment.
- **FR-004**: The system prompt MUST state that the
  parenthetical metadata now includes difficulty, equipment,
  and cues, and that difficulty bounds and equipment
  availability are already enforced before the pool.
- **FR-005**: The selection output schema, the deterministic
  filters, and every other node MUST stay unchanged.
- **FR-006**: A candidate without metadata MUST render as the
  bare ID, with no empty parenthetical.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Prompt rendering tests cover FR-001 to FR-006,
  including the bare candidate and the warmup and cooldown
  shapes.
- **SC-002**: The added context averages no more than about
  25 tokens per candidate on knowledge-base-shaped records,
  verified against the shipped knowledge base statistics
  (difficulty about 2 tokens, equipment about 4, cues about
  20 where present at 60 percent coverage).
- **SC-003**: The full backend test suite passes, and the
  prompt package keeps the coverage gate.
- **SC-004**: The first production generations after the
  rollout show no new trajectory error class attributable to
  the selection node, observed through the daily trajectory
  analysis.

## Assumptions

- The difficulty values in the knowledge base are ratings, and
  zero marks an unrated record.
- Cues coverage stays partial; the prompt treats a missing
  cues segment as no information, never as a defect of the
  candidate.
- Selection quality is judged on the rationale and exclusion
  output and on production observation; the generation
  trajectory harness remains unavailable in the build
  environment, as recorded for cycle 001.
