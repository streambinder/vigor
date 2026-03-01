-- per-user training feedback migration
-- run AFTER deploying the new backend (AutoMigrate creates training_feedbacks table)
-- backfills old feedback data for training owners, then drops legacy columns

BEGIN;

-- backfill training_feedbacks from old inline data for training owners
INSERT INTO training_feedbacks (id, training_id, user_id, quality, quality_reason, message, activity_feedback, created_at, updated_at)
SELECT
    gen_random_uuid(),
    t.id,
    t.user_id,
    (t.feedback->>'quality')::boolean,
    COALESCE(t.feedback->>'qualityReason', ''),
    COALESCE(t.feedback->>'message', ''),
    COALESCE((
        SELECT jsonb_object_agg(a.exercise_id, a.feedback)
        FROM activities a
        JOIN blocks b ON b.id = a.block_id
        JOIN routines r ON r.id = b.routine_id
        WHERE r.training_id = t.id AND r.type = 'work' AND a.feedback != ''
    ), '{}'::jsonb),
    COALESCE(t.completed_at, now()),
    COALESCE(t.completed_at, now())
FROM trainings t
WHERE t.completed_at IS NOT NULL
  AND t.feedback IS NOT NULL
  AND t.feedback != 'null'::jsonb
  AND t.feedback != '{}'::jsonb
ON CONFLICT (training_id, user_id) DO NOTHING;

-- drop legacy columns
ALTER TABLE trainings DROP COLUMN IF EXISTS feedback;
ALTER TABLE activities DROP COLUMN IF EXISTS feedback;

COMMIT;
