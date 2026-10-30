# 0003: app renders exercise animations in a WebView with Three.js

Status: Ratified by Davide on 2026-10-30.

## Context

Cycle 005 puts the parametric exercise animations in the app:
tap an exercise image, see the mannequin animate the recipe.
The animation documents are custom keyframe JSON (schema
`vigor-exercise-animation/1`, 22 bones, root track), already
rendered by the gallery player in Three.js.

## Decision

The app renders the animations in a WebView (`webview_flutter`)
with a minimal single-exercise Three.js page, reusing the
gallery player's contract, not its page.

## Alternatives considered

- Thermion (Filament in Dart). Its bone-frame API takes raw
  per-bone data, the closest native shape to our documents.
  Rejected for now: it needs a spike to prove the bone-index
  mapping, plus a Dart-side player and native-asset binaries.
- flutter_scene. Joints are writable scene-graph nodes, so
  manual bone driving is plausible, and the app already meets
  its Flutter 3.47 requirement. Rejected for now: pre-1.0 with
  breaking minor releases, Flutter GPU flags per platform, and
  the mapping from our relative quaternions to its joints is
  unproven.
- flutter_3d_controller, model_viewer_plus. Disqualified: no
  bone setter, autoplay of embedded clips only. Neither can
  consume the custom keyframe documents.

## Consequences

- The app player is the only path already proven end-to-end:
  the gallery renders these exact documents.
- Cost: one WebView engine instance per exercise view, and the
  model ships in app assets (licence gate, cycle 005 T004).
- A native player stays possible later. The documents and the
  backend endpoint do not change if the renderer changes.
