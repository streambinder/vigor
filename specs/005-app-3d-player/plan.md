# Cycle 005 plan: app 3D player on exercise image tap

## Shape

Knowledge first, then backend, then app, then verification.
One code commit after the docs commit, same two-commit shape
as cycle 004. The docs commit carries the spec (Draft), this
plan, the tasks, and ADRs 0003 and 0004 (Proposed).

## Knowledge

- New directory `knowledge/features/models/`: one GLB per
  model, its licence text next to it, and a manifest with ID,
  display name, licence, and source. The mannequin enters
  here, byte-identical to the gallery model.
- New table `exercise_models` (knowledge-local, like
  `exercise_animations`), new bootstrap step mirroring the
  animations step, new entry in the `AutoMigrate` list. A
  model missing a required joint fails bootstrap.
- New stills generator: from each animation document, take
  the peak frame, forward-kinematics through the rig
  (`rig.py` rest pose and quaternion math), project 3/4
  orthographic, emit one SVG per animated exercise into
  `knowledge/features/stills/`: capsule figure plus the
  document's prop as simple geometry, no background. New
  table `exercise_stills` (exercise ID plus SVG text), new
  bootstrap step, new `AutoMigrate` entry. A malformed SVG
  fails bootstrap.
- No Dockerfile change: the data-builder stage already
  copies `knowledge/features` into the image.

## Backend

- New handler in `backend/handler`: read one row from
  `exercise_animations` through the existing knowledge
  connection, by exercise ID. Return the stored document as-is.
  Unknown ID, or exercise without animation: 404.
- One more handler: return the animated exercises with their
  stills (ID plus SVG), for the row thumbnails. Fetched once
  by the app and cached.
- Two more handlers: list models (metadata only, no bytes)
  and serve model bytes by ID, with content type
  `model/gltf-binary`, content length, and the md5 as ETag.
- New routes under the existing authenticated API.
- Go tests: document returned byte-identical to the stored
  row; 404 for a missing ID; stills payload matches the
  animated rows with well-formed SVGs; model bytes
  md5-identical to the stored row; unauthenticated request
  rejected.

## App

- Add `webview_flutter` (first-party) as a dependency.
- New assets: a single-exercise Three.js player page and the
  Three.js build. The model is not an asset: it comes from
  the API. The page takes the animation document and the
  model bytes and renders them per the cycle 004 player
  contract: root track at the pelvis, pelvis x and z pinned,
  22 bone tracks, prop from the document, hand mesh stripped
  at load as in the gallery.
- On modal open for an animated exercise, fetch the document
  and the selected model (cached per model ID, downloaded
  once). Pass both to the page. Persist the model choice
  through the existing preferences service.
- `ExerciseModal`: replace the GIF area with the WebView
  viewport when an animation exists for the exercise. Any
  fetch failure, and any exercise without animation, falls
  back to the current GIF or placeholder, unchanged.
- Row thumbnails: an animated exercise shows its still
  through the existing SVG rendering (`SvgPicture.string`,
  no new dependency), never the GIF, with a small play
  affordance overlaid by the app. Other rows are unchanged.
  The still SVGs stay pure silhouettes: no background, no
  badge inside the file.
- In-page controls only: play and pause, phase scrub, speed,
  camera presets. No Flutter-to-JS bridge beyond passing the
  document and the model in.

## Verification

- Backend suite green, app analyzer and tests green.
- Gallery stays the motion reference: the 36 documents are
  unchanged, so gallery motion and app motion must match.
- Model bytes served by the backend are md5-identical to
  the gallery model.
- All 36 stills render on a preview page for visual review
  before merge: recognizable silhouettes, no background,
  each matching its animation's peak pose.
- Davide verifies on device: one animated exercise, one
  non-animated exercise, and the two exercises without a GIF.
- Licence: each model's licence text is archived in the repository
  next to its GLB, and recorded in its table row.

## Tree

```text
knowledge/
├── features/
│   ├── models/
│   │   ├── quaternius-mannequin.glb  # byte-identical to gallery
│   │   ├── LICENSE-quaternius.txt    # CC0 1.0 Universal text
│   │   └── manifest.json             # id, display name, licence, source
│   └── stills/
│       └── <exercise_id>.svg         # generated at build, one per animated exercise
└── main.go                   # exercise_models + exercise_stills tables + bootstrap
backend/
└── handler/
    ├── exercise_animation.go # animation endpoint + animated exercises with stills
    └── exercise_model.go     # model list + binary endpoints
app/
├── assets/
│   └── player/
│       ├── player.html       # single-exercise Three.js page
│       └── three.min.js      # bundled Three.js build
└── lib/
    ├── services/
    │   └── preferences_service.dart  # persisted model choice
    └── utils/
        └── exercise_modal.dart       # WebView viewport + GIF fallback
```
