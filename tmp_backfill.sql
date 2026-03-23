DO $$ 
DECLARE
    uid UUID;
    pid UUID;
    d DATE;
    sim_cast float;
    sim_actual float;
    rand float;
    total_acc float := 0;
BEGIN
    SELECT id INTO uid FROM users WHERE email='wijayasenaakbar@gmail.com';
    SELECT id INTO pid FROM solar_profiles WHERE user_id=uid LIMIT 1;
    
    FOR i IN reverse 90..1 LOOP
        d := CURRENT_DATE - i;
        IF NOT EXISTS (SELECT 1 FROM forecasts WHERE user_id=uid AND solar_profile_id=pid AND date=d) THEN
            rand := random();
            sim_cast := 15.0 + (rand * 10.0);
            sim_actual := sim_cast * (0.8 + (random() * 0.4));
            
            INSERT INTO forecasts (user_id, solar_profile_id, date, predicted_kwh, weather_factor, cloud_cover, efficiency, delta_wf, baseline_type)
            VALUES (uid, pid, d, sim_cast, 0.7 + (rand * 0.2), floor(10 + rand*40), 0.8, 0, 'synthetic');

            INSERT INTO actual_daily (user_id, solar_profile_id, date, actual_kwh, source)
            VALUES (uid, pid, d, sim_actual, 'iot') ON CONFLICT DO NOTHING;

            total_acc := total_acc + sim_actual;
        END IF;
    END LOOP;

    -- Update MWh accumulators
    INSERT INTO mwh_accumulators (user_id, solar_profile_id, cumulative_kwh, milestone_reached)
    VALUES (uid, pid, total_acc, CASE WHEN total_acc >= 1000 THEN true ELSE false END)
    ON CONFLICT (user_id, solar_profile_id) 
    DO UPDATE SET cumulative_kwh = mwh_accumulators.cumulative_kwh + EXCLUDED.cumulative_kwh,
                  milestone_reached = CASE WHEN mwh_accumulators.cumulative_kwh + EXCLUDED.cumulative_kwh >= 1000 THEN true ELSE false END;
END $$;
