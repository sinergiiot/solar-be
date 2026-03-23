UPDATE notification_preferences SET plan_tier = 'paid' WHERE plan_tier IN ('pro', 'enterprise');

ALTER TABLE notification_preferences
  DROP CONSTRAINT IF EXISTS notification_preferences_plan_tier_check;

ALTER TABLE notification_preferences
  ADD CONSTRAINT notification_preferences_plan_tier_check
  CHECK (plan_tier IN ('free', 'paid'));
