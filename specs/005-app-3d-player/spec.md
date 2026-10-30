# Cycle 005: app 3D player on exercise image tap

Status: Ratified by Davide on 2026-10-30.

## Context

Cycle 004 merged and deployed. The knowledge image carries the
`exercise_animations` table: one minified animation document per
animated catalog exercise, 36 documents, keyed by exercise ID.

In the app today, an exercise row shows a 72px thumbnail of the
catalog GIF (`exercise.reference`) through `CachedExerciseImage`.
Tapping the row opens `ExerciseModal`, which shows the same GIF
large, plus cues and instructions tabs.

The model itself lives nowhere in the pipeline today: not in the
knowledge image, not in the repository. It exists only as a local
research copy and inside the gallery artifact.

## Hypothesis challenge

Hypothesis: tapping an exercise image should show the parametric
3D animation instead of the GIF.

Evidence for the hypothesis:

- The GIFs are generic exercisedb clips. Some show a different
  variant than the catalog record (leg press side point of view,
  assisted knee raise). The recipes are authored to the record.
- Two animated exercises have no reference GIF at all (overhead
  press, long plank hold). For them the 3D view is strictly better
  than the current placeholder.
- The gallery player already renders these exact documents, and
  Davide approved the motion visually.

Evidence against the hypothesis, and risks:

- The mannequin is abstract. A human performer reads more
  naturally than a rigged model.
- A WebView costs a browser engine instance per view.
- The Quaternius licence is unresolved (CC0 versus QAL v1.0).
  Shipping the model in the app is distribution, so the licence
  is a pre-ship gate.

Alternatives considered:

- Keep the GIFs. Zero work, but keeps the variant mismatch and
  the two empty placeholders.
- A native player (Thermion bone-frame API, flutter_scene
  writable joints). Both are viable, but both need a spike to
  prove the bone mapping. Rejected by the decision to pick the
  best supported way: WebView is the only path already proven
  end-to-end.
- Link out to the gallery. Breaks the in-app flow and needs the
  hosted artifact plus network.

Outcome: build the WebView player inside the modal, with GIF
fallback. The licence gate blocks shipping, not building.

The wiring of the produced artifacts into the knowledge image
is not part of this cycle. Cycle 004 already shipped and
deployed it. This cycle consumes it.

## Requirements

- **FR-001**: The backend serves the stored animation document
  by catalog exercise ID, under the existing app authentication.
  An exercise without an animation answers 404.
- **FR-002**: `ExerciseModal` shows the 3D viewport in place of
  the GIF when an animation exists for the exercise. Otherwise
  the modal is unchanged.
- **FR-003**: The viewport renders the Quaternius mannequin with
  the document's 22 bone tracks, its root track (pelvis height
  and rotation, pelvis x and z pinned), and its prop. Controls
  are in-page: play and pause, phase scrub, speed, camera
  presets with 3/4 as default.
- **FR-004**: The player code (Three.js, page) ships in app
  assets. The player data does not: the animation JSON and the
  model bytes both come from the backend API, served from the
  knowledge database. Only the page itself is bundled.
- **FR-005**: Exercises without an animation keep the exact
  current behavior: GIF, or placeholder when no reference
  exists.
- **FR-006**: The knowledge database carries models in a new
  table `exercise_models`: one row per model with its ID,
  display name, model bytes, licence, source, and md5. The
  backend lists models (metadata only) and serves model bytes
  by ID, cacheable, with the md5 as ETag. The app downloads
  the selected model once, caches it, passes it to the page,
  and persists the choice through the existing preferences
  service. Cycle 005 ships one model through this pipeline.
  The picker UI lands with the second model: the pipeline is
  designed for many from day one, the UI waits until a choice
  exists.
- **FR-007**: Every model in `exercise_models` shares the
  skeleton the animation documents drive (the vrmc-core-22
  compatible joint set). Bootstrap rejects a model missing a
  required joint, mirroring the malformed-JSON rule for
  animations. A model with a different rig needs re-rigging
  before it enters the table.
- **FR-008**: For an exercise with an animation, the row
  thumbnail shows the exercise's still, never the GIF, even
  when a reference GIF exists. Tapping the still opens the
  modal with the 3D viewport. For an exercise without an
  animation, the thumbnail is unchanged: GIF, or no-image
  placeholder when no reference exists. In short: an
  exercise with an animation shows no GIF anywhere in the
  app. The still signals available 3D on tap, distinct from
  the no-image state; the app overlays a small play
  affordance on still thumbnails. The still SVGs stay pure
  silhouettes: no background, no badge inside the file.
- **FR-009**: Stills live in the knowledge database in a new
  table `exercise_stills`: one row per animated exercise,
  exercise ID plus the SVG text. The stills are generated,
  not hand-drawn: a generator runs at image build time from
  the animation documents, one SVG per animated exercise,
  into `knowledge/features/stills/`. Each still draws the
  animation's peak frame (the keyframe with maximum total
  joint rotation from rest) as a capsule figure in 3/4
  orthographic projection, the same framing as the gallery
  default camera, plus the document's prop as simple
  geometry. Bootstrap stores the stills, mirroring the
  animations step; a malformed SVG fails bootstrap. The
  backend serves the animated exercises with their stills in
  one endpoint (ID plus SVG), fetched once by the app and
  cached. The modal's document fetch stays the source of
  truth at open time. A flag baked into stored activity
  detail was considered and rejected: stored rows would need
  a backfill, while the stills endpoint works for every
  training, old and new. The attached silhouette set is a
  style reference only: the stills are generated originals,
  not copies of that artwork.

## Acceptance scenarios

- **SC-001**: Tap an animated exercise image. The modal shows
  the mannequin animating the recipe. Play, phase, speed, and
  camera controls work.
- **SC-002**: Tap a non-animated exercise image. The modal is
  unchanged (GIF or placeholder).
- **SC-003**: The animation endpoint answers 404 for a
  non-animated ID. The app shows the GIF and does not error.
- **SC-004**: The 36 production documents render with the same
  motion as the gallery, verified visually by Davide on device.
- **SC-005**: The model list endpoint returns the mannequin's
  metadata, the binary endpoint returns bytes identical to
  the gallery model (md5 match), and the app renders the same
  motion with the knowledge-served model as with the gallery
  model.
- **SC-006**: The row thumbnail for an animated exercise
  shows its still, not the GIF, even when a reference GIF
  exists. Tapping it opens the modal with the 3D viewport.
  A non-animated exercise keeps its GIF thumbnail and GIF
  modal.
- **SC-007**: All 36 stills render as recognizable exercise
  silhouettes with no background, reviewed on a preview page
  before merge, and each still matches its animation's peak
  pose.

## Assumptions and open points

- **A-001**: The Quaternius licence is resolved: the mannequin
  is CC0 1.0 Universal, verified from the in-zip licence text
  of its pack. Each future model carries its own licence in
  its table row, and its licence text is archived in the repository
  next to its GLB. A model without a verified licence never
  enters the table.
- **A-002**: The app player is a fresh minimal single-exercise
  page implementing the known contract (schema
  `vigor-exercise-animation/1`, rig `vrmc-core-22`). It does not
  copy the gallery page: the gallery carries grid, multi-cell,
  and global-control UI the app does not need. The gallery
  stays the visual reference.
- **A-003**: `webview_flutter` (first-party) is the WebView
  implementation.
- **A-004**: The backend reads `exercise_animations`,
  `exercise_models`, and `exercise_stills` through the
  existing knowledge connection (`KNOWLEDGE_URL`). No backend
  migration: the knowledge schema is image-owned, and the
  models and stills tables join the existing `AutoMigrate`
  list next to `exerciseAnimation`. The models ride the
  existing `features/` COPY in the knowledge Dockerfile: no
  Dockerfile change. Bootstrap gains a models step and a
  stills step mirroring the animations step, reading
  `features/models/` and the generated stills.

## Out of scope

- A native player.
- New animations. The 36 stay 36.
- New models beyond the one mannequin. The pipeline is ready,
  the catalog is not.
- Stills for exercises without an animation. A still is
  generated with its animation, when one lands.
- The picker UI. It lands with the second model.
- Changes to the modal's cues and instructions tabs.
- Changes to thumbnail behavior. Tap still opens the modal.
