# Cycle 005 tasks

- [ ] T000 Cycle 004 convergence bookkeeping: observe SC-004
      (36 rows in the production knowledge database) and check
      cycle 004 T008 in its tasks file.
- [ ] T001 Backend endpoints serving `exercise_animations` by
      exercise ID and the animated exercises with their
      stills, with tests (document byte-identical, 404 for
      missing, stills payload matches the animated rows with
      well-formed SVGs, auth enforced).
- [ ] T002 Knowledge `exercise_models` table and bootstrap:
      mannequin GLB, licence text, and manifest committed
      under `knowledge/features/models/`; joint validation
      fails bootstrap on a missing required joint.
- [ ] T003 Knowledge `exercise_stills` table, stills generator
      (peak frame, 3/4 projection, capsule figure plus prop,
      no background), and bootstrap; malformed SVG fails
      bootstrap; all 36 stills reviewed on a preview page.
- [ ] T004 Backend model endpoints: list (metadata only) and
      binary by ID (ETag = md5), with tests (bytes
      md5-identical, auth enforced).
- [ ] T005 App player page and dynamic model loading:
      single-exercise Three.js page, bundled Three.js build,
      model fetched once per ID and cached, passed to the page
      as bytes, choice persisted through the preferences
      service, per the cycle 004 player contract.
- [ ] T006 `ExerciseModal` and row thumbnail integration:
      WebView viewport in place of the GIF when an animation
      exists, GIF or placeholder fallback otherwise; still
      thumbnail, never the GIF, for animated exercises, with
      a small play affordance overlaid by the app.
- [ ] T007 Licence per model: licence text archived in the
      repository next to each GLB and recorded in its table row.
      No model without a verified licence enters the table.
- [ ] T008 Device verification by Davide (animated exercise,
      non-animated exercise, the two GIF-less exercises, model
      motion identical to the gallery, still thumbnails on
      animated rows) and CI green job by job, including
      the app jobs.
