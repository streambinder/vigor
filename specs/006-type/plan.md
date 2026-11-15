# Cycle 006 plan: Type — exercise modality in workout generation

## Shape

Knowledge first, then backend, then app, then verification.
One code commit after the docs commit, same two-commit shape
as cycle 005. The docs commit carries the spec (Draft), this
plan, the tasks, and ADR 0005 (Proposed).

## Knowledge

- New files `knowledge/features/types.json` and
  `knowledge/features/flags.json`: one entry per Type and
  per Flag, with a canonical ID and multilingual aliases,
  as equipment does in `equipment.json`. `mixed` is a
  request-only entry in `types.json`.
- New fields `type` and `flags` on the exercise record in
  `knowledge/features/exercises.json`: one Type of five
  per exercise, zero or more Flags. Classification by rule
  where the signal is unambiguous. Loadable equipment
  means `free_weights`. Machine or cable equipment means
  `machines_cables`. Bands as the primary source mean
  `bands`. Apparatus work means `bodyweight`. No equipment,
  and no plyometric marker, means `bodyweight`. Plyometric
  exercises are named in a curated list, reviewed in the
  implementation PR. The existing `mode` field is checked
  first.
- Bootstrap seeds the two vocabularies. It stores the
  fields next to `equipment` and `difficulty`. It rejects
  a record with a missing or unknown Type or Flag. The
  exercise model gains both fields. Existing production
  rows are updated through the bootstrap, not by hand.
- The embedded exercise text gains a Type line and a
  Flags line, through `GenExercise`. Embeddings regenerate
  at image build time through the existing two-pass
  bootstrap; no new pipeline.
- The classification lands as data, reviewed by count.
  The PR description carries counts per Type and per Flag,
  the size of the `special` bucket, and the full
  plyometric list.

## Backend

- New `Type` and `Flags` fields on `PostTrainingRequest`
  (JSON fields `type` and `flags`), threaded through
  `GenerateTraining` as parameters at the same rank as
  `methodology`. Unknown value: 400. Absent Type:
  `mixed`. Absent Flags: no flag filter.
- Free-text derivation deduces Type and Flags from the
  prompt in `DeriveFreeTextParams`, as it does for
  methodology, goals, and equipment. The derivation
  prompt carries the formal list of allowed values with
  aliases. An explicit choice stays authoritative.
- Retrieval filter in `retrieveBySimilarity`, next to the
  `MobilityOnly` filter. The Type value admits only
  exercises of that Type. A requested Flag admits only
  exercises carrying it. The filter applies to the work
  pool only. `mixed` applies no Type filter. Warmup and
  cooldown retrieval is unchanged. Random order stays
  inside the filtered pool.
- Availability endpoint: given an equipment list, it
  returns the available Types with counts, using the
  same predicate as the equipment filter. A Type with
  no servable exercise is not available. Modifiers do
  not enter the computation.
- Validation: a Type that no exercise can serve with the
  user equipment is rejected with 400, and the error
  lists the available Types. An empty filtered pool is
  rejected the same way.
- The selection prompt shows Type and Flags for each
  candidate exercise, and states the Type rule as
  reinforcement of the filter.
- Go tests: retrieval admits only the requested Type and
  Flags. `mixed` and absent Type behave as today.
  Unknown Type or Flag is rejected. Derivation keeps an
  explicit choice. Bootstrap rejects a bad Type or Flag.
  The availability endpoint returns only servable Types.

## App

- New Type selector in the generation modal, next to the
  equipment mode. Options: `free_weights`,
  `machines_cables`, `bodyweight`, `bands`, `special`,
  `mixed`. Only available Types are shown, from the
  availability endpoint. Mixed is the default. The request
  carries the chosen value in the `type` field.
- No Flags selector in this cycle. Flags come from free
  text, through derivation.
- No change to the equipment mode behavior. `gym` mode
  still admits bodyweight exercises when Type is `mixed`.
  Type `free_weights` is what excludes them.
- App tests: the modal sends the chosen Type, shows only
  available Types, and the default request carries
  `mixed`.

## Verification

- Backend suite green, knowledge tests green, app
  analyzer and tests green.
- Integration harness green: one trajectory per Type
  value, plus the default (no Type) trajectory, plus
  one trajectory with the `plyometric` Flag.
- Classification review: counts per Type and per Flag
  in the PR, and the full plyometric list. Spot-check
  the boundary cases. Apparatus work (rings, TRX,
  pull-up bar) is typed `bodyweight`. Cable and machine
  work is typed `machines_cables`. Bands as primary
  source are typed `bands`. A vest-loaded pull-up is
  typed `bodyweight` with the vest modifier, not
  `free_weights`. If the `special` bucket passes 5 to
  10 percent, consider splitting its largest entry.
- Davide verifies on device: one generated workout per
  Type value, and confirms the work exercises match the
  requested modality (SC-011).

## Tree

```text
knowledge/
├── features/
│   ├── exercises.json    # new `type` and `flags` per exercise
│   ├── types.json        # formal Type list: id + aliases
│   └── flags.json        # formal Flag list: id + aliases
└── main.go               # bootstrap seeds vocab, stores fields,
                          # rejects bad values, updates live rows
backend/
├── handler/
│   ├── dto/
│   │   └── training.go   # PostTrainingRequest gains Type, Flags
│   └── exercise_types.go # availability endpoint (new)
├── llm/
│   ├── dag.go            # DeriveFreeTextParams deduces Type, Flags
│   ├── prompt/
│   │   └── gen_training_exercises.qtpl  # candidates carry Type, Flags
│   └── rag/
│       ├── gen_exercise.qtpl           # Type, Flags lines in text
│       └── retrieval.go  # Type, Flag filter in retrieveBySimilarity
├── model/
│   └── exercise.go       # Exercise gains Type, Flags
└── service/
    └── training.go       # GenerateTraining gains the parameters
app/
└── lib/
    └── widgets/
        └── training_generation_modal.dart  # Type selector
```
