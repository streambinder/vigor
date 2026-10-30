# Cycle 005 tasks

- [ ] T000 Cycle 004 convergence bookkeeping: observe SC-004
      (36 rows in the production knowledge database) and check
      cycle 004 T008 in its tasks file.
- [x] T001 Backend endpoints serving `exercise_animations` by
      exercise ID and the animated exercises with their
      stills, with tests (document byte-identical, 404 for
      missing, stills payload matches the animated rows with
      well-formed SVGs, auth enforced).
- [x] T002 Knowledge `exercise_models` table and bootstrap:
      mannequin GLB, licence text, and manifest committed
      under `knowledge/features/models/`; joint validation
      fails bootstrap on a missing required joint.
- [x] T003 Knowledge `exercise_stills` table, stills generator
      (peak frame, 3/4 projection, capsule figure plus prop,
      no background), and bootstrap; malformed SVG fails
      bootstrap; all 36 stills reviewed on a preview page.
- [x] T004 Backend model endpoints: list (metadata only) and
      binary by ID (ETag = md5), with tests (bytes
      md5-identical, auth enforced).
- [x] T005 App player page and dynamic model loading:
      single-exercise Three.js page, bundled Three.js build,
      model fetched once per ID and cached, passed to the page
      as bytes, choice persisted through the preferences
      service, per the cycle 004 player contract.
- [x] T006 `ExerciseModal` and row thumbnail integration:
      WebView viewport in place of the GIF when an animation
      exists, GIF or placeholder fallback otherwise; still
      thumbnail, never the GIF, for animated exercises, with
      a small play affordance overlaid by the app.
- [x] T007 Licence per model: licence text archived in the
      repository next to each GLB and recorded in its table row.
      No model without a verified licence enters the table.
- [ ] T008 Device verification by Davide (animated exercise,
      non-animated exercise, the two GIF-less exercises, model
      motion identical to the gallery, still thumbnails on
      animated rows) and CI green job by job, including
      the app jobs.

## Convergence notes

- Implementation ships in four commits: knowledge, backend,
  app, docs. Every file appears in exactly one commit.
- Plan discrepancy: plan.md said no Dockerfile change, but
  the stills generator needed a build step, so the
  knowledge commit extended the animations-builder stage.
- T000 stays open: the production knowledge database is
  not readable from the implementation machine, so the
  36-row observation and the cycle 004 T008 check remain
  for Davide.
- T008 stays open: device verification belongs to Davide.
- Verification done before filing: knowledge tests pass
  at 94.3% coverage, backend handler tests pass, app
  analyze is clean with 57 app tests passing, the
  integration harness passes all 11 trajectories, and
  all 36 stills were generated from the canonical pool
  and reviewed on a preview page outside the repository.
