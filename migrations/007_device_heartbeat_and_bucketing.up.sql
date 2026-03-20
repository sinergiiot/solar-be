ALTER TABLE devices
ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS last_telemetry_at TIMESTAMPTZ;

ALTER TABLE telemetry_raw
ADD COLUMN IF NOT EXISTS bucket_start TIMESTAMPTZ;

UPDATE telemetry_raw
SET bucket_start = CASE
    WHEN EXTRACT(HOUR FROM event_time AT TIME ZONE 'UTC') < 12
        THEN date_trunc('day', event_time AT TIME ZONE 'UTC')
    ELSE date_trunc('day', event_time AT TIME ZONE 'UTC') + INTERVAL '12 hour'
END
WHERE bucket_start IS NULL;

ALTER TABLE telemetry_raw
ALTER COLUMN bucket_start SET NOT NULL;

ALTER TABLE telemetry_raw
DROP CONSTRAINT IF EXISTS telemetry_raw_device_id_event_time_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_telemetry_raw_device_bucket
ON telemetry_raw(device_id, bucket_start);

CREATE INDEX IF NOT EXISTS idx_devices_user_last_seen
ON devices(user_id, last_seen_at DESC);
