-- 135_update_scheduled_test_gpt52_model.sql
-- gpt-5.2 is no longer used for scheduled account tests.
UPDATE scheduled_test_plans
SET model_id = 'gpt-5.5',
    updated_at = NOW()
WHERE model_id = 'gpt-5.2';
