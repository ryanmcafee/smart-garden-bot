-- Seed data for Smart Garden Bot database
-- This migration adds initial reference data and sample content

-- =============================================
-- SUBSCRIPTION PLANS
-- =============================================

INSERT INTO billing.subscription_plans (
    id,
    name,
    description,
    price_monthly,
    price_yearly,
    currency,
    max_gardens,
    max_zones_per_garden,
    max_devices_per_garden,
    max_api_calls_per_month,
    features
) VALUES 
(
    uuid_generate_v4(),
    'Starter',
    'Perfect for small home gardens and beginners',
    9.99,
    99.99,
    'USD',
    1,
    5,
    10,
    10000,
    '{
        "weather_integration": true,
        "basic_automation": true,
        "mobile_app": true,
        "email_notifications": true,
        "basic_analytics": true,
        "api_access": false,
        "advanced_ml": false,
        "custom_integrations": false
    }'::jsonb
),
(
    uuid_generate_v4(),
    'Professional',
    'Ideal for serious gardeners and small farms',
    29.99,
    299.99,
    'USD',
    5,
    20,
    50,
    100000,
    '{
        "weather_integration": true,
        "basic_automation": true,
        "advanced_automation": true,
        "mobile_app": true,
        "email_notifications": true,
        "sms_notifications": true,
        "basic_analytics": true,
        "advanced_analytics": true,
        "plant_disease_detection": true,
        "api_access": true,
        "advanced_ml": false,
        "custom_integrations": false,
        "priority_support": true
    }'::jsonb
),
(
    uuid_generate_v4(),
    'Enterprise',
    'For commercial operations and advanced users',
    99.99,
    999.99,
    'USD',
    NULL, -- Unlimited
    NULL, -- Unlimited
    NULL, -- Unlimited  
    1000000,
    '{
        "weather_integration": true,
        "basic_automation": true,
        "advanced_automation": true,
        "mobile_app": true,
        "email_notifications": true,
        "sms_notifications": true,
        "webhook_notifications": true,
        "basic_analytics": true,
        "advanced_analytics": true,
        "plant_disease_detection": true,
        "yield_prediction": true,
        "api_access": true,
        "advanced_ml": true,
        "custom_integrations": true,
        "priority_support": true,
        "dedicated_support": true,
        "custom_training": true
    }'::jsonb
);

-- =============================================
-- SAMPLE USERS (for development/testing)
-- =============================================

-- Insert sample users (only in development)
INSERT INTO auth.users (
    id,
    email,
    name,
    auth0_user_id,
    metadata
) VALUES 
(
    '550e8400-e29b-41d4-a716-446655440001',
    'demo@smartgardenbot.com',
    'Demo User',
    'auth0|demo123',
    '{
        "profile_completed": true,
        "preferred_units": "metric",
        "timezone": "America/New_York"
    }'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440002',
    'professional@smartgardenbot.com',
    'Professional Gardener',
    'auth0|pro456',
    '{
        "profile_completed": true,
        "preferred_units": "metric",
        "timezone": "America/Los_Angeles",
        "experience_level": "professional",
        "garden_type": "commercial"
    }'::jsonb
);

-- =============================================
-- SAMPLE GARDENS AND ZONES
-- =============================================

-- Demo user's backyard garden
INSERT INTO gardens.gardens (
    id,
    user_id,
    name,
    description,
    location,
    latitude,
    longitude,
    timezone,
    area_sqm,
    metadata
) VALUES (
    '660e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440001',
    'Backyard Garden',
    'My main vegetable and herb garden in the backyard',
    'San Francisco, CA, USA',
    37.7749,
    -122.4194,
    'America/Los_Angeles',
    50.0,
    '{
        "soil_type": "clay_loam",
        "sun_exposure": "full_sun",
        "established_date": "2023-03-15",
        "irrigation_type": "drip"
    }'::jsonb
);

-- Professional's commercial garden
INSERT INTO gardens.gardens (
    id,
    user_id,
    name,
    description,
    location,
    latitude,
    longitude,
    timezone,
    area_sqm,
    metadata
) VALUES (
    '660e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440002',
    'Commercial Greenhouse',
    'Temperature-controlled greenhouse for year-round production',
    'Portland, OR, USA',
    45.5152,
    -122.6784,
    'America/Los_Angeles',
    200.0,
    '{
        "structure_type": "greenhouse",
        "climate_controlled": true,
        "automation_level": "high",
        "crop_rotation": true
    }'::jsonb
);

-- Zones for backyard garden
INSERT INTO gardens.zones (
    id,
    garden_id,
    name,
    description,
    plant_type,
    area_sqm,
    soil_type,
    moisture_threshold,
    default_watering_duration,
    position_x,
    position_y,
    metadata
) VALUES 
(
    '770e8400-e29b-41d4-a716-446655440001',
    '660e8400-e29b-41d4-a716-446655440001',
    'Tomatoes',
    'Heritage tomato varieties - Cherokee Purple, Brandywine',
    'Vegetables',
    15.0,
    'clay_loam',
    40,
    20,
    0.0,
    0.0,
    '{
        "varieties": ["Cherokee Purple", "Brandywine", "San Marzano"],
        "planting_date": "2024-04-15",
        "expected_harvest": "2024-07-01",
        "support_type": "cages"
    }'::jsonb
),
(
    '770e8400-e29b-41d4-a716-446655440002',
    '660e8400-e29b-41d4-a716-446655440001',
    'Herbs',
    'Culinary herbs - basil, oregano, thyme, rosemary',
    'Herbs',
    8.0,
    'clay_loam',
    35,
    10,
    15.0,
    0.0,
    '{
        "varieties": ["Genovese Basil", "Greek Oregano", "English Thyme", "Rosemary"],
        "harvest_method": "continuous",
        "companion_planting": true
    }'::jsonb
),
(
    '770e8400-e29b-41d4-a716-446655440003',
    '660e8400-e29b-41d4-a716-446655440001',
    'Lettuce Bed',
    'Mixed greens and lettuce varieties',
    'Leafy Greens',
    12.0,
    'clay_loam',
    45,
    15,
    0.0,
    8.0,
    '{
        "varieties": ["Buttercrunch", "Red Sail", "Arugula", "Spinach"],
        "succession_planting": true,
        "days_to_maturity": 45
    }'::jsonb
);

-- =============================================
-- SAMPLE DEVICES
-- =============================================

-- Weather station for backyard garden
INSERT INTO gardens.devices (
    id,
    garden_id,
    name,
    device_type,
    manufacturer,
    model,
    firmware_version,
    mac_address,
    ip_address,
    endpoint_url,
    status,
    last_seen_at,
    battery_level,
    signal_strength,
    configuration,
    metadata
) VALUES (
    '880e8400-e29b-41d4-a716-446655440001',
    '660e8400-e29b-41d4-a716-446655440001',
    'Main Weather Station',
    'weather_station',
    'Ecowitt',
    'WS2902C',
    '1.6.8',
    '00:11:22:33:44:55',
    '192.168.1.100',
    'http://192.168.1.100/api/v1',
    'active',
    CURRENT_TIMESTAMP - INTERVAL '5 minutes',
    85,
    92,
    '{
        "sampling_interval": 60,
        "data_retention_days": 365,
        "sensors": {
            "temperature": true,
            "humidity": true,
            "pressure": true,
            "wind": true,
            "rain": true,
            "uv": true,
            "solar": true
        }
    }'::jsonb,
    '{
        "installation_date": "2023-03-20",
        "mounting_height": "2.5m",
        "calibration_date": "2024-01-15"
    }'::jsonb
);

-- Soil moisture sensors
INSERT INTO gardens.devices (
    id,
    garden_id,
    name,
    device_type,
    manufacturer,
    model,
    status,
    last_seen_at,
    battery_level,
    signal_strength,
    configuration,
    metadata
) VALUES 
(
    '880e8400-e29b-41d4-a716-446655440002',
    '660e8400-e29b-41d4-a716-446655440001',
    'Tomato Soil Sensor',
    'sensor',
    'Ecowitt',
    'WH51',
    'active',
    CURRENT_TIMESTAMP - INTERVAL '10 minutes',
    78,
    88,
    '{
        "measurement_depth": "15cm",
        "sampling_interval": 300,
        "low_battery_threshold": 20
    }'::jsonb,
    '{
        "zone_id": "770e8400-e29b-41d4-a716-446655440001",
        "installation_date": "2023-04-01"
    }'::jsonb
),
(
    '880e8400-e29b-41d4-a716-446655440003',
    '660e8400-e29b-41d4-a716-446655440001',
    'Herb Garden Soil Sensor',
    'sensor',
    'Ecowitt',
    'WH51',
    'active',
    CURRENT_TIMESTAMP - INTERVAL '15 minutes',
    82,
    91,
    '{
        "measurement_depth": "10cm",
        "sampling_interval": 300,
        "low_battery_threshold": 20
    }'::jsonb,
    '{
        "zone_id": "770e8400-e29b-41d4-a716-446655440002",
        "installation_date": "2023-04-01"
    }'::jsonb
);

-- =============================================
-- SAMPLE WATERING SCHEDULES
-- =============================================

INSERT INTO gardens.watering_schedules (
    id,
    zone_id,
    name,
    cron_expression,
    duration_minutes,
    skip_if_rain_probability,
    skip_if_soil_moisture_above,
    is_active,
    next_run_at,
    metadata
) VALUES 
(
    '990e8400-e29b-41d4-a716-446655440001',
    '770e8400-e29b-41d4-a716-446655440001',
    'Tomato Morning Watering',
    '0 6 * * *', -- Daily at 6 AM
    20,
    70,
    60,
    true,
    date_trunc('day', CURRENT_TIMESTAMP) + INTERVAL '1 day' + INTERVAL '6 hours',
    '{
        "created_by": "user",
        "watering_method": "drip",
        "flow_rate": "2L/min"
    }'::jsonb
),
(
    '990e8400-e29b-41d4-a716-446655440002',
    '770e8400-e29b-41d4-a716-446655440002',
    'Herb Light Watering',
    '0 7,19 * * *', -- Twice daily at 7 AM and 7 PM
    10,
    60,
    70,
    true,
    date_trunc('day', CURRENT_TIMESTAMP) + INTERVAL '1 day' + INTERVAL '7 hours',
    '{
        "created_by": "user",
        "watering_method": "sprinkler",
        "gentle_mode": true
    }'::jsonb
),
(
    '990e8400-e29b-41d4-a716-446655440003',
    '770e8400-e29b-41d4-a716-446655440003',
    'Lettuce Cool Season Watering',
    '0 8,16 * * *', -- Twice daily at 8 AM and 4 PM
    15,
    50,
    55,
    true,
    date_trunc('day', CURRENT_TIMESTAMP) + INTERVAL '1 day' + INTERVAL '8 hours',
    '{
        "created_by": "user",
        "watering_method": "mist",
        "seasonal_adjustment": true
    }'::jsonb
);

-- =============================================
-- SAMPLE SENSOR READINGS (last 7 days)
-- =============================================

-- Generate sample sensor readings for the past week
INSERT INTO sensors.readings (
    device_id,
    zone_id,
    timestamp,
    temperature,
    humidity,
    soil_moisture,
    soil_temperature,
    atmospheric_pressure,
    light_intensity,
    signal_strength,
    data_quality
)
SELECT 
    '880e8400-e29b-41d4-a716-446655440002'::uuid, -- Tomato soil sensor
    '770e8400-e29b-41d4-a716-446655440001'::uuid, -- Tomato zone
    generate_series(
        CURRENT_TIMESTAMP - INTERVAL '7 days',
        CURRENT_TIMESTAMP,
        INTERVAL '5 minutes'
    ) as timestamp,
    -- Temperature: realistic daily cycle (15-25°C)
    20 + 5 * sin(2 * pi() * extract(hour from generate_series) / 24) + random() * 2 - 1,
    -- Humidity: inverse correlation with temperature (40-80%)
    65 - 3 * sin(2 * pi() * extract(hour from generate_series) / 24) + random() * 10 - 5,
    -- Soil moisture: decreases over time, increases with watering events
    greatest(30, 60 - extract(epoch from (CURRENT_TIMESTAMP - generate_series)) / 3600 * 0.5 + 
        case when extract(hour from generate_series) = 6 then 20 else 0 end + 
        random() * 5 - 2.5),
    -- Soil temperature: slightly delayed temperature response
    18 + 4 * sin(2 * pi() * (extract(hour from generate_series) - 2) / 24) + random() * 1.5 - 0.75,
    -- Atmospheric pressure: relatively stable
    1013.25 + random() * 10 - 5,
    -- Light intensity: follows sun cycle
    greatest(0, 50000 * sin(pi() * greatest(0, extract(hour from generate_series) - 6) / 12)) + random() * 5000,
    -- Signal strength: mostly good with occasional variations
    85 + random() * 10,
    'good'
WHERE generate_series <= CURRENT_TIMESTAMP;

-- Similar readings for herb sensor
INSERT INTO sensors.readings (
    device_id,
    zone_id,
    timestamp,
    temperature,
    humidity,
    soil_moisture,
    soil_temperature,
    atmospheric_pressure,
    signal_strength,
    data_quality
)
SELECT 
    '880e8400-e29b-41d4-a716-446655440003'::uuid, -- Herb soil sensor
    '770e8400-e29b-41d4-a716-446655440002'::uuid, -- Herb zone
    generate_series(
        CURRENT_TIMESTAMP - INTERVAL '7 days',
        CURRENT_TIMESTAMP,
        INTERVAL '5 minutes'
    ) as timestamp,
    -- Similar temperature pattern
    20 + 5 * sin(2 * pi() * extract(hour from generate_series) / 24) + random() * 2 - 1,
    -- Similar humidity pattern
    65 - 3 * sin(2 * pi() * extract(hour from generate_series) / 24) + random() * 10 - 5,
    -- Different soil moisture pattern (herbs need different watering)
    greatest(35, 65 - extract(epoch from (CURRENT_TIMESTAMP - generate_series)) / 3600 * 0.3 + 
        case when extract(hour from generate_series) in (7, 19) then 15 else 0 end + 
        random() * 4 - 2),
    -- Soil temperature
    18 + 4 * sin(2 * pi() * (extract(hour from generate_series) - 2) / 24) + random() * 1.5 - 0.75,
    -- Atmospheric pressure
    1013.25 + random() * 10 - 5,
    -- Signal strength
    80 + random() * 15,
    'good'
WHERE generate_series <= CURRENT_TIMESTAMP;

-- =============================================
-- SAMPLE WATERING EVENTS
-- =============================================

-- Generate some historical watering events
INSERT INTO sensors.watering_events (
    id,
    zone_id,
    schedule_id,
    started_at,
    ended_at,
    planned_duration_minutes,
    actual_duration_minutes,
    water_volume_liters,
    trigger_type,
    trigger_reason,
    soil_moisture_before,
    soil_moisture_after,
    temperature,
    status,
    metadata
) VALUES 
(
    uuid_generate_v4(),
    '770e8400-e29b-41d4-a716-446655440001',
    '990e8400-e29b-41d4-a716-446655440001',
    CURRENT_TIMESTAMP - INTERVAL '1 day' + INTERVAL '6 hours',
    CURRENT_TIMESTAMP - INTERVAL '1 day' + INTERVAL '6 hours 20 minutes',
    20,
    20,
    40.0,
    'scheduled',
    'Daily morning watering schedule',
    42.5,
    78.2,
    18.5,
    'completed',
    '{
        "weather_conditions": {
            "temperature": 18.5,
            "humidity": 72,
            "wind_speed": 5.2,
            "cloud_cover": 30
        }
    }'::jsonb
),
(
    uuid_generate_v4(),
    '770e8400-e29b-41d4-a716-446655440002',
    '990e8400-e29b-41d4-a716-446655440002',
    CURRENT_TIMESTAMP - INTERVAL '12 hours',
    CURRENT_TIMESTAMP - INTERVAL '12 hours' + INTERVAL '10 minutes',
    10,
    10,
    15.0,
    'scheduled',
    'Evening herb watering',
    48.1,
    72.5,
    22.1,
    'completed',
    '{
        "weather_conditions": {
            "temperature": 22.1,
            "humidity": 65,
            "wind_speed": 3.8,
            "cloud_cover": 15
        }
    }'::jsonb
);

-- =============================================
-- SAMPLE ANALYTICS DATA
-- =============================================

-- Sample insights
INSERT INTO analytics.insights (
    id,
    garden_id,
    zone_id,
    insight_type,
    title,
    description,
    recommendations,
    confidence_score,
    priority,
    status,
    expires_at
) VALUES 
(
    uuid_generate_v4(),
    '660e8400-e29b-41d4-a716-446655440001',
    '770e8400-e29b-41d4-a716-446655440001',
    'watering_optimization',
    'Optimize Tomato Watering Schedule',
    'Based on soil moisture patterns and weather forecasts, your tomato zone could benefit from adjusted watering times and duration.',
    '[
        {
            "action": "Reduce morning watering duration from 20 to 15 minutes",
            "reason": "Soil moisture retention is better than expected",
            "impact": "Save 25% water while maintaining optimal moisture levels"
        },
        {
            "action": "Add evening watering during hot days (>30°C)",
            "reason": "High evaporation rates detected on hot afternoons",
            "impact": "Prevent plant stress and improve yield"
        }
    ]'::jsonb,
    0.8745,
    'medium',
    'active',
    CURRENT_TIMESTAMP + INTERVAL '30 days'
),
(
    uuid_generate_v4(),
    '660e8400-e29b-41d4-a716-446655440001',
    '770e8400-e29b-41d4-a716-446655440002',
    'plant_health',
    'Herb Garden Moisture Levels Optimal',
    'Your herb garden is maintaining excellent moisture levels. The current watering schedule is working perfectly.',
    '[
        {
            "action": "Continue current watering schedule",
            "reason": "Optimal moisture levels maintained consistently",
            "impact": "Healthy herb growth and efficient water usage"
        }
    ]'::jsonb,
    0.9234,
    'low',
    'active',
    CURRENT_TIMESTAMP + INTERVAL '14 days'
);

-- =============================================
-- FUNCTIONS FOR SAMPLE DATA GENERATION
-- =============================================

-- Function to generate realistic weather data
CREATE OR REPLACE FUNCTION generate_sample_weather_data()
RETURNS void AS $$
DECLARE
    location_key text := 'san_francisco_ca';
    base_date timestamp with time zone := CURRENT_TIMESTAMP - INTERVAL '30 days';
    current_date timestamp with time zone;
    temp_base decimal := 18.0;
    humidity_base decimal := 65.0;
    pressure_base decimal := 1013.25;
BEGIN
    -- Generate weather data for the past 30 days
    FOR i IN 0..29 LOOP
        current_date := base_date + (i || ' days')::interval;
        
        INSERT INTO sensors.weather_data (
            location_key,
            provider,
            data_type,
            timestamp,
            expires_at,
            data
        ) VALUES (
            location_key,
            'openweathermap',
            'current',
            current_date,
            current_date + INTERVAL '1 hour',
            json_build_object(
                'temperature', temp_base + (random() * 6 - 3),
                'humidity', humidity_base + (random() * 20 - 10),
                'pressure', pressure_base + (random() * 20 - 10),
                'wind_speed', random() * 15,
                'wind_direction', random() * 360,
                'visibility', 10,
                'uv_index', CASE 
                    WHEN extract(hour from current_date) BETWEEN 6 AND 18 
                    THEN random() * 8 + 2 
                    ELSE 0 
                END,
                'cloud_cover', random() * 100,
                'weather_condition', CASE 
                    WHEN random() < 0.7 THEN 'clear'
                    WHEN random() < 0.9 THEN 'cloudy'
                    ELSE 'rainy'
                END
            )
        );
    END LOOP;
    
    RAISE NOTICE 'Generated sample weather data for % days', 30;
END;
$$ LANGUAGE plpgsql;

-- Execute the weather data generation
SELECT generate_sample_weather_data();

-- Drop the temporary function
DROP FUNCTION generate_sample_weather_data();

-- =============================================
-- COMMENTS AND DOCUMENTATION
-- =============================================

COMMENT ON SCHEMA auth IS 'User authentication, authorization, and session management';
COMMENT ON SCHEMA gardens IS 'Garden, zone, device, and schedule management';
COMMENT ON SCHEMA sensors IS 'Time-series sensor data and watering events';
COMMENT ON SCHEMA billing IS 'Subscription plans, billing, and payment processing';
COMMENT ON SCHEMA analytics IS 'AI/ML insights, embeddings, and analytics data';

COMMENT ON TABLE auth.users IS 'User accounts with Auth0 integration';
COMMENT ON TABLE auth.api_keys IS 'API keys for machine-to-machine authentication';
COMMENT ON TABLE gardens.gardens IS 'User gardens with location and metadata';
COMMENT ON TABLE gardens.zones IS 'Watering zones within gardens';
COMMENT ON TABLE gardens.devices IS 'IoT devices (sensors, controllers, weather stations)';
COMMENT ON TABLE sensors.readings IS 'Time-series sensor data (hypertable)';
COMMENT ON TABLE analytics.embeddings IS 'Vector embeddings for similarity search and ML';

-- =============================================
-- COMPLETION MESSAGE
-- =============================================

DO $$
BEGIN
    RAISE NOTICE '=================================================';
    RAISE NOTICE 'Smart Garden Bot database initialization complete!';
    RAISE NOTICE '=================================================';
    RAISE NOTICE 'Schemas created: auth, gardens, sensors, billing, analytics';
    RAISE NOTICE 'Sample data includes:';
    RAISE NOTICE '  - 2 users with different subscription levels';
    RAISE NOTICE '  - 2 gardens with multiple zones each';
    RAISE NOTICE '  - IoT devices and sensors';
    RAISE NOTICE '  - 7 days of sensor readings';
    RAISE NOTICE '  - Watering schedules and events';
    RAISE NOTICE '  - AI/ML insights and recommendations';
    RAISE NOTICE '  - 30 days of weather data';
    RAISE NOTICE '=================================================';
END $$;