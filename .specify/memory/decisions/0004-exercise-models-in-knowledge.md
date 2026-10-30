# 0004: exercise models live in the knowledge database

Status: Ratified by Davide on 2026-10-30.

## Context

Cycle 005 needs the mannequin in the app. The model is one
file today, but the catalog will grow: more models, and users
picking a favorite. A model the user picks is data, not app
code: it must be addable without an app release, and it must
travel with the same versioning as the animations it poses.

## Decision

Models live in the knowledge database, in a new table
`exercise_models`, next to `exercise_animations`. One row per
model: ID, display name, model bytes, licence, source, md5.
The backend lists models (metadata only) and serves model
bytes by ID. The app downloads the selected model once,
caches it, and passes it to the player page. The GLB files,
their licence texts, and a manifest live in the repository under
`knowledge/features/models/` and ride the existing
`features/` COPY into the knowledge image; bootstrap stores
them, mirroring the animations step.

## Alternatives considered

- Bundle the model in app assets. Simpler for one immutable
  model: offline, zero runtime moving parts. Rejected: every
  new model, and every model swap, would need an app release,
  and the model would live in two places (app repository and
  gallery) instead of one versioned source.
- Host the models on a CDN or static host. Rejected: a new
  hosting dependency, a second versioning story, and network
  required for a first view, for data the knowledge image
  already versions.

## Consequences

- The app stays thin: page and Three.js bundled, all player
  data from the API. The picker UI lands with the second
  model; the pipeline is designed for many from day one.
- Every model must share the skeleton the animation
  documents drive. Bootstrap rejects a model missing a
  required joint. A model with a different rig needs
  re-rigging before it enters the table.
- Every model carries its licence in its row, and its
  licence text in the repository. No verified licence, no row.
- Cost: model bytes in Postgres (about 617 KB per model),
  a binary endpoint, and app-side download, cache, and
  Blob plumbing. Justified by the multi-model future, not by
  the single model alone.
