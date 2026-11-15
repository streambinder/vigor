# 0005: Type is a separate axis from methodology

Status: Ratified by Davide on 2026-11-15.

## Context

Cycle 006 needs to regulate exercise modality in workout
generation. Methodology governs structure only. Equipment
governs availability only. Goal governs load and reps only.
A gym user training for hypertrophy receives bodyweight
exercises mixed with weights, because no input forbids the
mix. The literature treats modality as a prescription
variable of its own. The ACSM FITT-VP principle names
Type. The NSCA puts modality first in program design.
An exhaustive review found no single official taxonomy.
The synthesis gives five exercise modalities plus
transversal flags.

## Decision

Type is a separate generation input, at the same rank as
methodology. It carries one value per request:
`free_weights`, `machines_cables`, `bodyweight`, `bands`,
`special`, `mixed`. `mixed` is the default and preserves
current behavior. It means the available Types, not all
five.

Every catalog exercise carries one Type of the first five,
and zero or more Flags: `plyometric`, `ballistic`,
`isometric`, `eccentric_emphasis`, `olympic_weightlifting`,
`bfr`, `variable_resistance`, `unstable_surface`,
`unilateral`. Retrieval filters the work pool by Type and
by requested Flags with a deterministic filter. The
selection prompt states the rule as reinforcement only.
Free-text derivation deduces Type and Flags as it does
for methodology, goals, and equipment. The app offers a
Type selector next to the equipment mode, showing only
the Types the user equipment can serve.

Type availability derives from equipment: a Type is
available when at least one exercise of that Type is
servable with the user equipment. Modifiers never enter
availability. The supported Types and Flags live in
formal lists with canonical IDs and multilingual aliases,
as equipment does. The prompts to the model carry Type
and Flags for each exercise.

Complex training, a heavy lift paired with a jump, is
expressed as `mixed` in this cycle, with the plyometric
flag on the explosive exercises. A dedicated plyometric
Type value is not part of this cycle.

## Alternatives considered

- Extend the methodology with modality constraints, as
  `mobility_only` does today. Rejected: the axes are
  orthogonal. Strength plus calisthenics and strength plus
  weights are two different workouts with one structure.
  Mixing the axes in one field confuses structure choice
  with exercise choice, and breaks the coverage counting
  that assumes methodologies differ by structure.
- Rely on equipment alone, with a weights-only equipment
  preset. Rejected: availability is not preference. An
  exercise that needs no equipment is never filtered out
  by any equipment list, so a weights-only preset cannot
  exclude bodyweight work.
- Rely on the selection prompt only. Rejected: the prompt
  already sees the `weighted` annotation and does not use
  it as a constraint. A prompt rule without a retrieval
  filter cannot guarantee exclusion.
- Keep `traditional` as one value. Rejected: free weights,
  machines, and cables have different strength effects
  and different availability. One value cannot filter by
  apparatus.
- Keep `plyometric` as an exclusive Type value. Rejected:
  plyometrics are defined by the stretch-shortening cycle,
  not by an apparatus. They combine with weights in
  complex training. A flag expresses them without a
  false choice.
- Treat the weighted vest as equipment. Rejected: it is a
  modifier, and it already exists as one in
  `modifiers.json`. Modifiers add load within a modality.
  They never change the modality.
- Maintain an equipment-to-Type mapping by hand. Rejected:
  availability derived from the catalog is self
  maintaining. A hand mapping drifts from the catalog
  and mishandles support apparatus such as a bench or
  a rack.
- Do nothing. Rejected: the complaint is real, it repeats
  across users, and the literature backs the missing
  dimension.

## Consequences

- The generation request gains two fields, and the catalog
  gains two fields per exercise, plus two vocabulary
  files. The classification is data work: 1,489 records,
  mostly by rule, with a curated plyometric list.
- The retrieval filter guarantees exclusion, where the
  prompt can only suggest it. The filter is the
  enforcement; the prompt is reinforcement.
- The app shows only the Types the user equipment can
  serve. A user with no equipment sees bodyweight only.
- Modifiers are unchanged. Methodology, goal, and weekly
  mix definitions are unchanged. Any methodology combines
  with any Type.
- Cost: one more choice in the generation modal, and two
  more fields to maintain per exercise. Justified by the
  reported mix, which no existing input can express.
