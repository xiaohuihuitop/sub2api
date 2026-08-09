-- Content moderation is scoped by Platform, which is the account-pool authority.
-- This migration is intentionally separate from 196: 196 is already used by the
-- model-pricing override migration in this branch.

-- Seed the platform mapping before converting historical moderation rows. The
-- later catalog migration adds model rules and account ownership, while this
-- early step guarantees every legacy group has a stable platform id for the
-- strict audit-log conversion below.
INSERT INTO platforms (
    code,
    name,
    account_platform,
    status,
    endpoint_capabilities,
    legacy_group_id
)
SELECT
    'legacy-group-' || g.id::text,
    g.name,
    g.platform,
    CASE WHEN g.status = 'active' THEN 'active' ELSE 'disabled' END,
    CASE
        WHEN g.platform = 'openai' THEN '["chat_completions", "responses"]'::jsonb
        ELSE '["chat_completions"]'::jsonb
    END,
    g.id
FROM groups g
WHERE g.deleted_at IS NULL
  AND g.platform IN ('openai', 'anthropic', 'gemini', 'antigravity', 'grok')
  AND NOT EXISTS (
      SELECT 1 FROM platforms p WHERE p.legacy_group_id = g.id
  )
ON CONFLICT (code) DO NOTHING;

ALTER TABLE content_moderation_logs
    ADD COLUMN IF NOT EXISTS platform_id BIGINT REFERENCES platforms(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS platform_name VARCHAR(255) NOT NULL DEFAULT '';

DO $$
DECLARE
    missing_group_id BIGINT;
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'content_moderation_logs' AND column_name = 'group_id'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'platforms' AND column_name = 'legacy_group_id'
    ) THEN
        IF EXISTS (
            SELECT 1
            FROM content_moderation_logs l
            WHERE l.platform_id IS NULL
              AND l.group_id IS NOT NULL
              AND NOT EXISTS (
                  SELECT 1 FROM platforms p
                  WHERE p.legacy_group_id = l.group_id
              )
            LIMIT 1
        ) THEN
            SELECT l.group_id INTO missing_group_id
            FROM content_moderation_logs l
            WHERE l.platform_id IS NULL
              AND l.group_id IS NOT NULL
              AND NOT EXISTS (
                  SELECT 1 FROM platforms p
                  WHERE p.legacy_group_id = l.group_id
              )
            LIMIT 1;
            RAISE EXCEPTION 'content moderation log group % cannot be mapped to a platform', missing_group_id;
        END IF;

        UPDATE content_moderation_logs l
        SET platform_id = p.id,
            platform_name = p.name
        FROM platforms p
        WHERE l.platform_id IS NULL
          AND p.legacy_group_id = l.group_id;
    END IF;

    IF EXISTS (
        SELECT 1 FROM settings
        WHERE key = 'content_moderation_config'
          AND value::jsonb ? 'group_ids'
    ) THEN
        IF EXISTS (
            SELECT 1
            FROM settings s
            CROSS JOIN LATERAL jsonb_array_elements_text(COALESCE(s.value::jsonb->'group_ids', '[]'::jsonb)) ids(value)
            WHERE s.key = 'content_moderation_config'
              AND COALESCE((s.value::jsonb->>'all_groups')::boolean, TRUE) = FALSE
              AND NOT EXISTS (
                  SELECT 1 FROM platforms p
                  WHERE p.legacy_group_id = ids.value::bigint
              )
            LIMIT 1
        ) THEN
            RAISE EXCEPTION 'content moderation config contains a group without a platform mapping';
        END IF;

        UPDATE settings s
        SET value = (
            (s.value::jsonb - 'all_groups' - 'group_ids')
            || jsonb_build_object(
                'all_platforms', COALESCE((s.value::jsonb->>'all_groups')::boolean, TRUE),
                'platform_ids', COALESCE((
                    SELECT jsonb_agg(p.id ORDER BY p.id)
                    FROM jsonb_array_elements_text(COALESCE(s.value::jsonb->'group_ids', '[]'::jsonb)) ids(value)
                    JOIN platforms p ON p.legacy_group_id = ids.value::bigint
                ), '[]'::jsonb)
            )
        )
        WHERE s.key = 'content_moderation_config';
    END IF;
END $$;

DROP INDEX IF EXISTS idx_content_moderation_logs_group_created_at;
CREATE INDEX IF NOT EXISTS idx_content_moderation_logs_platform_created_at
    ON content_moderation_logs(platform_id, created_at DESC);

ALTER TABLE content_moderation_logs
    DROP CONSTRAINT IF EXISTS content_moderation_logs_group_id_fkey,
    DROP COLUMN IF EXISTS group_id,
    DROP COLUMN IF EXISTS group_name;
