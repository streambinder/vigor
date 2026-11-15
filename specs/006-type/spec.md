# Cycle 006: Type — exercise modality in workout generation

Status: Ratified by Davide on 2026-11-15.

## Context

Workout generation takes three inputs that shape exercise
choice: methodology, equipment, and goal. Methodology governs
structure only: blocks, rest, reps versus duration, density.
Equipment governs availability only: which tools the user can
use. Goal governs load and reps only: ranges and progression.

None of these governs the modality of the exercises. A gym
user training for hypertrophy can receive push-ups and planks
mixed with bench press and rows. The mix is not a rare edge
case. It is the default outcome of the current pipeline.

The mechanism is in the retrieval filter. An exercise passes
when it needs no equipment, or when all its equipment is in
the user list. A gym user with full equipment passes every
exercise, including the 360 with no equipment at all. Without
keywords, the retrieval orders the pool with `ORDER BY
RANDOM()`. Bodyweight versus weights is then a coin toss.
The selection node sees only a `weighted` annotation, as a
tie-breaker between near-equal options. Nothing forbids the
mix.

A gym user reported the same symptom in August 2026:
bodyweight exercises judged useless, appearing in a gym
workout.

Modifiers are a separate layer from equipment. A weighted
vest, a dip belt, or ankle weights add load to an exercise
without changing its modality. The vest already exists as a
modifier in `modifiers.json`. It applies to bodyweight
exercises only. It never turns a bodyweight exercise into a
weights exercise.

## Hypothesis challenge

Hypothesis: workout generation needs a Type axis, separate
from methodology, that regulates exercise modality.

Evidence for the hypothesis:

- The ACSM FITT-VP principle names Type as a prescription
  variable of its own. Type is the modality of exercise.
  Frequency, intensity, time, volume, and progression
  regulate how much and how hard. Methodology regulates how
  the session is organized. Type regulates what the user
  trains with.
- The NSCA puts modality first in program design. Exercise
  order and organization methods follow as separate steps.
  Its task analysis lists training methods and modalities
  as two distinct lists.
- Calisthenics, traditional weights, and plyometrics are
  distinct modalities with distinct adaptations. Low-load
  and high-load training build equal hypertrophy to failure
  (Schoenfeld et al. 2017, 21 studies). Heavy loads stay
  superior for maximal strength. Plyometrics build power
  and rate of force development, not maximal strength.
  The modality changes the expected outcome, not only the
  flavor of the session.
- The code has no signal to forbid the mix, as shown above.
  Equipment cannot express it: availability is not
  preference. A user with everything available can still
  want weights only. Methodology cannot express it: a
  superset session can be all weights or all bodyweight.
- An exhaustive literature review (15 November 2026) found
  no single official taxonomy of modalities. The synthesis
  of ACSM, NSCA, Fleck and Kraemer, and Suchomel gives
  five exercise modalities, plus transversal flags. The
  closest empirical classification is Wiedenmann et al.
  2025: free weights, machines, bodyweight, elastic bands,
  mixed. It also shows elastic bands as a missing modality
  in the first draft.

Evidence against the hypothesis, and risks:

- A new input adds a choice to the generation modal. A
  wrong default would change behavior for existing users.
  The default must preserve current behavior.
- The catalog needs one Type and Flags per exercise: 1,489
  records. Most follow by rule from equipment. Plyometrics
  need a curated list, since no equipment signal
  identifies them.
- Plyometrics cut across the other values. Complex training
  pairs a heavy lift with a jump in one session. An
  exclusive Type cannot express that pair directly.

Alternatives considered:

- Extend the methodology with modality constraints, as
  `mobility_only` does today. Rejected: the axes are
  orthogonal. Strength plus calisthenics and strength plus
  weights are two different workouts with one structure.
  Mixing the axes in one field confuses structure choice
  with exercise choice.
- Rely on equipment alone, with a weights-only equipment
  preset. Rejected: availability is not preference, and
  bodyweight exercises need no equipment, so no equipment
  list ever filters them out.
- Rely on the selection prompt only, telling the model to
  prefer weighted exercises for hypertrophy. Rejected: the
  prompt already sees the `weighted` annotation and does
  not use it as a constraint. Prompt-only is weak.
- Keep `traditional` as one value for free weights,
  machines, and cables. Rejected: the three have different
  strength effects (Wiedenmann 2025: machines 1.34,
  free weights 1.15) and different availability. One value
  cannot filter by apparatus.
- Keep `plyometric` as an exclusive Type value. Rejected:
  plyometrics are defined by the stretch-shortening cycle,
  not by an apparatus. They combine with weights in complex
  training, and they can be bodyweight, weighted, or
  aquatic. A flag expresses them without a false choice.
- Treat the weighted vest as equipment. Rejected: it is a
  modifier, and it already exists as one in
  `modifiers.json`. Modifiers add load within a modality.
  They never change the modality.
- Do nothing. Rejected: the complaint is real, it repeats,
  and the literature backs the missing dimension.

Outcome: build Type as a separate axis, at the same rank as
methodology. Request values: `free_weights`,
`machines_cables`, `bodyweight`, `bands`, `special`,
`mixed`. Each exercise carries one of the first five.
`mixed` is the default and preserves current behavior.
Each exercise also carries zero or more Flags:
`plyometric`, `ballistic`, `isometric`,
`eccentric_emphasis`, `olympic_weightlifting`, `bfr`,
`variable_resistance`, `unstable_surface`, `unilateral`.
Complex training is expressed as `mixed` in this cycle,
with the plyometric flag on the explosive exercises.

## Requirements

- **FR-001**: Every catalog exercise carries exactly one
  Type: `free_weights`, `machines_cables`, `bodyweight`,
  `bands`, or `special`. It also carries zero or more
  Flags: `plyometric`, `ballistic`, `isometric`,
  `eccentric_emphasis`, `olympic_weightlifting`, `bfr`,
  `variable_resistance`, `unstable_surface`, `unilateral`.
  The values live in the exercise record, next to
  `equipment` and `difficulty`. Bootstrap rejects an
  exercise with a missing or unknown Type or Flag.
- **FR-002**: The supported Types and Flags live in formal
  lists: `knowledge/features/types.json` and
  `knowledge/features/flags.json`. Each entry has a
  canonical ID and multilingual aliases, as equipment
  does in `equipment.json`. The catalog and the request
  use canonical IDs only. UI labels come from the
  Italian aliases.
- **FR-003**: The generation request accepts `Type`, as the
  JSON field `type`, and `Flags`, as the JSON field
  `flags`. Type values: the five exercise values plus
  `mixed`. `mixed` is the default: a request without
  Type behaves exactly as today. An unknown Type or
  Flag is rejected with 400.
- **FR-004**: Type availability derives from equipment.
  A Type is available when at least one exercise of that
  Type is servable with the user equipment. The check
  uses the same predicate as the equipment filter. A
  backend endpoint returns the available Types with
  counts for an equipment list. The app shows only
  available Types. `mixed` means the available Types,
  not all five. Modifiers do not enter availability.
- **FR-005**: Retrieval filters the work pool by Type and
  by requested Flags. A Type admits only exercises of
  that Type. A requested Flag admits only exercises
  carrying it. `mixed` applies no Type filter. Warmup
  and cooldown pools are unchanged. Random order stays
  inside the filtered pool.
- **FR-006**: Free-text derivation deduces Type and Flags
  from the prompt, as it already does for methodology,
  goals, and equipment. The derivation prompt carries
  the formal list of allowed values with aliases. An
  explicit choice stays authoritative: derivation fills
  only what the request left unset.
- **FR-007**: The prompts to the model carry Type and
  Flags. The embedded exercise text carries a Type line
  and a Flags line. The selection prompt shows Type and
  Flags for each candidate exercise, and states the rule
  as reinforcement. The derivation prompt carries the
  formal allowed values. Enforcement stays the retrieval
  filter: the prompt reinforces it, it does not replace
  it.
- **FR-008**: The app generation modal offers a Type
  selector next to the equipment mode. Options:
  `free_weights`, `machines_cables`, `bodyweight`,
  `bands`, `special`, `mixed`. Only available Types are
  shown. Mixed is the default. There is no Flags
  selector in this cycle: Flags come from free text.
- **FR-009**: Type is independent of methodology and goal.
  Any methodology combines with any Type. The weekly
  methodology mix is unchanged. Goal definitions are
  unchanged.
- **FR-010**: The backend rejects a Type that no exercise
  can serve with the user equipment, with 400. The
  error lists the available Types. An empty filtered
  pool is rejected the same way, instead of generating
  an incoherent workout.

## Acceptance scenarios

- **SC-001**: A gym request with goal hypertrophy and Type
  `free_weights` generates a work pool of `free_weights`
  exercises only. No bodyweight exercise appears in the
  work blocks.
- **SC-002**: A request with Type `bodyweight` generates
  bodyweight exercises only.
- **SC-003**: A request with Type `bands` generates
  elastic-band exercises only.
- **SC-004**: A request with Type `machines_cables`
  generates machine and cable exercises only.
- **SC-005**: A request with the `plyometric` Flag
  generates exercises carrying that Flag only, across
  Types. A weighted vest does not change the Type of
  any exercise.
- **SC-006**: A request with Type `mixed`, and a request
  without Type, both behave exactly as today.
- **SC-007**: Free text asking for weights only ("solo
  pesi", "only weights") derives Type `free_weights`.
  Free text asking for plyometrics only derives the
  `plyometric` Flag. An explicit Type choice is not
  overridden by the text.
- **SC-008**: Warmup and cooldown are unaffected by Type:
  a `free_weights` workout still warms up and cools
  down as today.
- **SC-009**: All 1,489 catalog exercises carry a valid
  Type and valid Flags. Bootstrap rejects a record
  with a missing or unknown Type or Flag.
- **SC-010**: Availability follows equipment. A user with
  no equipment and only a weighted vest modifier is
  offered `bodyweight` only. A user with a full gym
  is offered all five Types.
- **SC-011**: Davide verifies on device: one generated
  workout per Type value, and confirms the work
  exercises match the requested modality.

## Assumptions and open points

- **A-001**: The Type value set is fixed for this cycle.
  `free_weights` covers barbell, dumbbell, kettlebell,
  medicine ball, and other free loads. `machines_cables`
  covers guided machines and pulleys. `bodyweight`
  covers body-mass exercises, including apparatus work
  such as pull-up bar, rings, and TRX. `bands` covers
  elastic resistance as the primary source. `special`
  covers isokinetic, flywheel, aquatic, vibration, and
  manual or partner resistance. `mixed` is request-only.
- **A-002**: Type classification follows rules where the
  signal is unambiguous, and a curated list where it is
  not. Loadable equipment means `free_weights`. Machine
  or cable equipment means `machines_cables`. Bands as
  the primary source mean `bands`. Apparatus work means
  `bodyweight`. No equipment, and no plyometric marker,
  means `bodyweight`. Plyometric exercises are named in
  a curated list, reviewed in the implementation PR.
  The existing `mode` field is checked first: it may
  already carry part of the signal.
- **A-003**: Embeddings regenerate at knowledge image
  build time, through the existing two-pass bootstrap.
  Adding Type and Flags lines to the embedded text needs
  no new pipeline: the build already re-embeds the
  catalog.
- **A-004**: The `weighted` annotation in the selection
  prompt stays. It becomes redundant for filtered pools
  and keeps its tie-breaker role in `mixed` pools. There
  is no `weighted` Flag: loaded bodyweight is covered
  by the modifier system (`modifiers.json`).
- **A-005**: Type availability derives from equipment
  only. Modifiers never enter availability: a weighted
  vest changes load and progression inside `bodyweight`,
  not the available Types.
- **A-006**: The formal Type and Flag lists live in
  `types.json` and `flags.json`, with canonical IDs and
  multilingual aliases. The catalog and the request use
  canonical IDs only. `mixed` is a request-only entry
  in `types.json`.

## Out of scope

- A Flags selector in the app. Flags come from free text
  in this cycle.
- Complex training as a first-class combination. Complex
  training is `mixed` in this cycle, with the
  `plyometric` Flag on the explosive exercises.
- Changes to the nine methodologies, to goal definitions,
  and to the weekly methodology mix.
- Changes to warmup and cooldown selection.
- Changes to the modifier system. Modifiers are
  orthogonal to Type.
- New exercises. The 1,489 stay 1,489, only typed and
  flagged.
- Changes to the modal beyond the Type selector.
