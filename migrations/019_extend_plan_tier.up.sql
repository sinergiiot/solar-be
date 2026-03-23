-- Update CHECK constraint on notification_preferences
ALTER TABLE notification_preferences
  DROP CONSTRAINT IF EXISTS notification_preferences_plan_tier_check;

ALTER TABLE notification_preferences
  ADD CONSTRAINT notification_preferences_plan_tier_check
  CHECK (plan_tier IN ('free', 'pro', 'enterprise'));

-- Migration of existing data
UPDATE notification_preferences SET plan_tier = 'pro' WHERE plan_tier = 'paid';
