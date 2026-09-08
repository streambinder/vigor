# Muscle calibration migration runbook

One-time migration for the movement-family → muscle calibration rework.
The `backfill` tool in this directory recomputes proficiencies; this file
covers the manual steps around it for local Docker Postgres and production.

## What changed

- `proficiencies.movement_family` → `proficiencies.muscle` (app DB)
- knowledge `exercises.progressions` (jsonb) → `difficulty` (int) + `is_mobility` (bool)
- knowledge `methodologies.work` JSON gains `min_difficulty` / `max_difficulty` / `mobility_only`
- `knowledge/features/movement_families.json` deleted

## Order of operations (do not reorder)

The old backend reads `proficiencies.movement_family` and
`exercises.progressions`; the new backend reads the new columns.
Running the old backend against the renamed schema (or vice versa)
breaks generation.

1. Stop the backend (local: `docker compose stop backend`, prod: scale to 0).
2. App DB — run the manual rename (no AutoMigrate for this):

   ```sql
   ALTER TABLE proficiencies RENAME COLUMN movement_family TO muscle;
   ```

3. Knowledge DB — re-seed from the transformed JSON. From `knowledge/`:

   ```sh
   DATABASE_URL=<redacted> go run .
   ```

   Its built-in AutoMigrate adds `difficulty` / `is_mobility`; the
   bootstrap upserts all exercises and methodologies. Afterwards, drop
   the stale column manually:

   ```sql
   ALTER TABLE exercises DROP COLUMN progressions;
   ```

   Do not drop it before the re-seed succeeds.

4. App DB — run the backfill (connects without AutoMigrate, idempotent):

   ```sh
   # from backend/
   DATABASE_URL=<redacted> KNOWLEDGE_URL=<redacted> go run ./tools/backfill
   ```

   For every (training, user) pair with feedback on a completed training it
   deletes existing proficiency rows and re-records one row per muscle from
   the training's work activities (primary muscle, difficulty + modifier
   impact). Re-running it is safe.

5. Verify:

   ```sql
   SELECT muscle, COUNT(*) FROM proficiencies GROUP BY muscle;
   -- expect only: chest, back, shoulders, arms, core, glutes, legs
   SELECT COUNT(*) FROM exercises WHERE difficulty IS NULL; -- expect 0
   ```

6. Deploy the new backend and start it. Its AutoMigrate is a no-op for
   these tables once steps 2–3 are done.

## Rollback

- App DB: `ALTER TABLE proficiencies RENAME COLUMN muscle TO movement_family;`
  then re-run the backfill from the previous release (it recomputes
  family rows from the same feedback).
- Knowledge DB: re-run `go run .` from the previous release's `knowledge/`
  directory, then `ALTER TABLE exercises DROP COLUMN difficulty;` and
  `ALTER TABLE exercises DROP COLUMN is_mobility;`.
- Redeploy the previous backend release.
