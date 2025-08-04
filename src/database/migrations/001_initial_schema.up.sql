-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";
CREATE EXTENSION IF NOT EXISTS "timescaledb";

-- Create schemas for logical separation
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS gardens;
CREATE SCHEMA IF NOT EXISTS sensors;
CREATE SCHEMA IF NOT EXISTS billing;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Set default search path
SET search_path TO public, auth, gardens, sensors, billing, analytics;

-- =============================================
-- AUTH SCHEMA - User authentication and authorization
-- =============================================

-- Users table
CREATE TABLE auth.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    auth0_user_id VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Indexes
    CONSTRAINT users_email_check CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- API Keys for machine-to-machine authentication
CREATE TABLE auth.api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(10) NOT NULL, -- First 8 chars for identification
    scopes TEXT[] DEFAULT ARRAY[]::TEXT[], -- Array of allowed scopes
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure user can't have duplicate key names
    CONSTRAINT api_keys_user_name_unique UNIQUE (user_id, name)
);

-- User sessions for tracking active sessions
CREATE TABLE auth.user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token_id VARCHAR(255) NOT NULL UNIQUE,
    refresh_token_hash VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =============================================
-- GARDENS SCHEMA - Garden and zone management
-- =============================================

-- Gardens table
CREATE TABLE gardens.gardens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    area_sqm DECIMAL(10, 2), -- Area in square meters
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Metadata for flexible extensions
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Constraints
    CONSTRAINT gardens_lat_check CHECK (latitude >= -90 AND latitude <= 90),
    CONSTRAINT gardens_lng_check CHECK (longitude >= -180 AND longitude <= 180),
    CONSTRAINT gardens_area_positive CHECK (area_sqm > 0)
);

-- Zones within gardens
CREATE TABLE gardens.zones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    garden_id UUID NOT NULL REFERENCES gardens.gardens(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    plant_type VARCHAR(255),
    area_sqm DECIMAL(10, 2),
    soil_type VARCHAR(100),
    
    -- Watering configuration
    moisture_threshold INTEGER CHECK (moisture_threshold >= 0 AND moisture_threshold <= 100),
    default_watering_duration INTEGER DEFAULT 30, -- minutes
    
    -- Position within garden
    position_x DECIMAL(10, 2),
    position_y DECIMAL(10, 2),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Constraints
    CONSTRAINT zones_user_garden_name_unique UNIQUE (garden_id, name),
    CONSTRAINT zones_area_positive CHECK (area_sqm > 0),
    CONSTRAINT zones_duration_positive CHECK (default_watering_duration > 0)
);

-- IoT Devices (weather stations, controllers, etc.)
CREATE TABLE gardens.devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    garden_id UUID NOT NULL REFERENCES gardens.gardens(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    device_type VARCHAR(100) NOT NULL, -- 'weather_station', 'watering_controller', 'sensor'
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    firmware_version VARCHAR(50),
    
    -- Connection details
    mac_address MACADDR,
    ip_address INET,
    endpoint_url VARCHAR(500),
    
    -- Device status
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'inactive', 'maintenance', 'error'
    last_seen_at TIMESTAMP WITH TIME ZONE,
    battery_level INTEGER CHECK (battery_level >= 0 AND battery_level <= 100),
    signal_strength INTEGER CHECK (signal_strength >= 0 AND signal_strength <= 100),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Configuration and metadata
    configuration JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    
    -- Constraints
    CONSTRAINT devices_garden_name_unique UNIQUE (garden_id, name)
);

-- Watering schedules
CREATE TABLE gardens.watering_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    zone_id UUID NOT NULL REFERENCES gardens.zones(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    
    -- Schedule configuration (cron format)
    cron_expression VARCHAR(100) NOT NULL,
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes > 0),
    
    -- Conditions
    skip_if_rain_probability INTEGER DEFAULT 70 CHECK (skip_if_rain_probability >= 0 AND skip_if_rain_probability <= 100),
    skip_if_soil_moisture_above INTEGER CHECK (skip_if_soil_moisture_above >= 0 AND skip_if_soil_moisture_above <= 100),
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Next run calculation
    next_run_at TIMESTAMP WITH TIME ZONE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb
);

-- =============================================
-- SENSORS SCHEMA - Time-series sensor data
-- =============================================

-- Sensor readings (time-series data)
CREATE TABLE sensors.readings (
    id UUID DEFAULT uuid_generate_v4(),
    device_id UUID NOT NULL REFERENCES gardens.devices(id) ON DELETE CASCADE,
    zone_id UUID REFERENCES gardens.zones(id) ON DELETE SET NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Environmental measurements
    temperature DECIMAL(5, 2), -- Celsius
    humidity DECIMAL(5, 2), -- Percentage
    soil_moisture DECIMAL(5, 2), -- Percentage
    soil_temperature DECIMAL(5, 2), -- Celsius
    ph_level DECIMAL(4, 2), -- pH scale
    light_intensity DECIMAL(10, 2), -- Lux
    uv_index DECIMAL(4, 2),
    
    -- Weather measurements
    atmospheric_pressure DECIMAL(8, 2), -- hPa
    wind_speed DECIMAL(6, 2), -- km/h
    wind_direction INTEGER CHECK (wind_direction >= 0 AND wind_direction <= 360),
    rainfall DECIMAL(8, 2), -- mm
    
    -- Electrical measurements
    battery_voltage DECIMAL(5, 3), -- Volts
    solar_voltage DECIMAL(5, 3), -- Volts
    
    -- Data quality indicators
    signal_strength INTEGER CHECK (signal_strength >= 0 AND signal_strength <= 100),
    data_quality VARCHAR(20) DEFAULT 'good', -- 'excellent', 'good', 'fair', 'poor'
    
    -- Additional sensor data (flexible JSON storage)
    additional_data JSONB DEFAULT '{}'::jsonb,
    
    PRIMARY KEY (device_id, timestamp)
);

-- Convert to hypertable for time-series optimization
SELECT create_hypertable('sensors.readings', 'timestamp', 'device_id', number_partitions => 4);

-- Watering events log
CREATE TABLE sensors.watering_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    zone_id UUID NOT NULL REFERENCES gardens.zones(id) ON DELETE CASCADE,
    schedule_id UUID REFERENCES gardens.watering_schedules(id) ON DELETE SET NULL,
    
    -- Event details
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    planned_duration_minutes INTEGER NOT NULL,
    actual_duration_minutes INTEGER,
    
    -- Water usage
    water_volume_liters DECIMAL(10, 2),
    
    -- Trigger information
    trigger_type VARCHAR(50) NOT NULL, -- 'scheduled', 'manual', 'sensor_threshold', 'weather_based'
    trigger_reason TEXT,
    
    -- Conditions at time of watering
    soil_moisture_before DECIMAL(5, 2),
    soil_moisture_after DECIMAL(5, 2),
    temperature DECIMAL(5, 2),
    weather_conditions JSONB,
    
    -- Status
    status VARCHAR(50) DEFAULT 'completed', -- 'started', 'completed', 'cancelled', 'failed'
    error_message TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Weather data cache
CREATE TABLE sensors.weather_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_key VARCHAR(255) NOT NULL, -- Could be coordinates, city name, etc.
    provider VARCHAR(50) NOT NULL, -- 'openweathermap', 'weatherapi', etc.
    data_type VARCHAR(50) NOT NULL, -- 'current', 'forecast', 'historical'
    
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Weather data (stored as JSON for flexibility)
    data JSONB NOT NULL,
    
    -- Index for efficient lookups
    CONSTRAINT weather_data_unique UNIQUE (location_key, provider, data_type, timestamp)
);

-- =============================================
-- BILLING SCHEMA - Subscription and payment data
-- =============================================

-- Subscription plans
CREATE TABLE billing.subscription_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    
    -- Pricing
    price_monthly DECIMAL(10, 2),
    price_yearly DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Limits
    max_gardens INTEGER,
    max_zones_per_garden INTEGER,
    max_devices_per_garden INTEGER,
    max_api_calls_per_month INTEGER,
    
    -- Features
    features JSONB DEFAULT '{}'::jsonb,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User subscriptions
CREATE TABLE billing.subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES billing.subscription_plans(id),
    
    -- Stripe integration
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255),
    
    -- Subscription details
    status VARCHAR(50) NOT NULL, -- 'active', 'cancelled', 'past_due', 'unpaid'
    billing_cycle VARCHAR(20) NOT NULL, -- 'monthly', 'yearly'
    
    -- Dates
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    current_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    
    -- Trial
    trial_start TIMESTAMP WITH TIME ZONE,
    trial_end TIMESTAMP WITH TIME ZONE,
    
    -- Usage tracking
    usage_data JSONB DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Payment history
CREATE TABLE billing.payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id UUID NOT NULL REFERENCES billing.subscriptions(id) ON DELETE CASCADE,
    
    -- Stripe integration
    stripe_payment_intent_id VARCHAR(255) UNIQUE,
    stripe_invoice_id VARCHAR(255),
    
    -- Payment details
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL, -- 'succeeded', 'failed', 'pending', 'cancelled'
    
    -- Dates
    attempted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    succeeded_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    
    -- Failure information
    failure_code VARCHAR(100),
    failure_message TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb
);

-- =============================================
-- ANALYTICS SCHEMA - AI/ML and analytics data
-- =============================================

-- Plant disease detection results
CREATE TABLE analytics.plant_disease_detections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    zone_id UUID NOT NULL REFERENCES gardens.zones(id) ON DELETE CASCADE,
    
    -- Image information
    image_url VARCHAR(500),
    image_hash VARCHAR(255),
    
    -- Detection results
    detected_diseases JSONB NOT NULL DEFAULT '[]'::jsonb,
    confidence_scores JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    -- ML model information
    model_version VARCHAR(50),
    model_name VARCHAR(100),
    
    -- Processing metadata
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processing_time_ms INTEGER,
    
    -- User feedback for model improvement
    user_confirmed BOOLEAN,
    user_feedback TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Vector embeddings for similarity search
CREATE TABLE analytics.embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(100) NOT NULL, -- 'plant_image', 'weather_pattern', 'growth_cycle'
    entity_id UUID NOT NULL,
    
    -- Vector data
    embedding vector(512), -- 512-dimensional vector (adjust as needed)
    
    -- Metadata
    model_name VARCHAR(100) NOT NULL,
    model_version VARCHAR(50),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint to prevent duplicate embeddings
    CONSTRAINT embeddings_entity_model_unique UNIQUE (entity_type, entity_id, model_name, model_version)
);

-- Garden insights and recommendations
CREATE TABLE analytics.insights (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    garden_id UUID NOT NULL REFERENCES gardens.gardens(id) ON DELETE CASCADE,
    zone_id UUID REFERENCES gardens.zones(id) ON DELETE CASCADE,
    
    -- Insight details
    insight_type VARCHAR(100) NOT NULL, -- 'watering_optimization', 'plant_health', 'yield_prediction'
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    
    -- Recommendations
    recommendations JSONB DEFAULT '[]'::jsonb,
    
    -- Confidence and priority
    confidence_score DECIMAL(5, 4) CHECK (confidence_score >= 0 AND confidence_score <= 1),
    priority VARCHAR(20) DEFAULT 'medium', -- 'low', 'medium', 'high', 'critical'
    
    -- Status
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'dismissed', 'resolved'
    
    -- User interaction
    viewed_at TIMESTAMP WITH TIME ZONE,
    dismissed_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- =============================================
-- INDEXES for performance optimization
-- =============================================

-- Auth schema indexes
CREATE INDEX idx_users_email ON auth.users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_auth0_id ON auth.users(auth0_user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_api_keys_user_id ON auth.api_keys(user_id);
CREATE INDEX idx_api_keys_hash ON auth.api_keys(key_hash);
CREATE INDEX idx_user_sessions_user_id ON auth.user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires ON auth.user_sessions(expires_at);

-- Gardens schema indexes
CREATE INDEX idx_gardens_user_id ON gardens.gardens(user_id);
CREATE INDEX idx_gardens_location ON gardens.gardens USING gist(ll_to_earth(latitude, longitude));
CREATE INDEX idx_zones_garden_id ON gardens.zones(garden_id);
CREATE INDEX idx_devices_garden_id ON gardens.devices(garden_id);
CREATE INDEX idx_devices_type ON gardens.devices(device_type);
CREATE INDEX idx_devices_status ON gardens.devices(status);
CREATE INDEX idx_watering_schedules_zone_id ON gardens.watering_schedules(zone_id);
CREATE INDEX idx_watering_schedules_next_run ON gardens.watering_schedules(next_run_at) WHERE is_active = true;

-- Sensors schema indexes (time-series optimized)
CREATE INDEX idx_readings_timestamp ON sensors.readings(timestamp DESC);
CREATE INDEX idx_readings_zone_timestamp ON sensors.readings(zone_id, timestamp DESC);
CREATE INDEX idx_watering_events_zone_id ON sensors.watering_events(zone_id);
CREATE INDEX idx_watering_events_started_at ON sensors.watering_events(started_at DESC);
CREATE INDEX idx_weather_data_location_provider ON sensors.weather_data(location_key, provider, data_type);
CREATE INDEX idx_weather_data_expires ON sensors.weather_data(expires_at);

-- Billing schema indexes
CREATE INDEX idx_subscriptions_user_id ON billing.subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON billing.subscriptions(status);
CREATE INDEX idx_subscriptions_stripe_id ON billing.subscriptions(stripe_subscription_id);
CREATE INDEX idx_payments_subscription_id ON billing.payments(subscription_id);
CREATE INDEX idx_payments_status ON billing.payments(status);

-- Analytics schema indexes
CREATE INDEX idx_plant_disease_zone_id ON analytics.plant_disease_detections(zone_id);
CREATE INDEX idx_plant_disease_processed_at ON analytics.plant_disease_detections(processed_at DESC);
CREATE INDEX idx_embeddings_entity ON analytics.embeddings(entity_type, entity_id);
CREATE INDEX idx_embeddings_vector ON analytics.embeddings USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX idx_insights_garden_id ON analytics.insights(garden_id);
CREATE INDEX idx_insights_zone_id ON analytics.insights(zone_id);
CREATE INDEX idx_insights_type_status ON analytics.insights(insight_type, status);

-- =============================================
-- FUNCTIONS and TRIGGERS
-- =============================================

-- Function to update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers to relevant tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON auth.users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_gardens_updated_at BEFORE UPDATE ON gardens.gardens FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_zones_updated_at BEFORE UPDATE ON gardens.zones FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_devices_updated_at BEFORE UPDATE ON gardens.devices FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_watering_schedules_updated_at BEFORE UPDATE ON gardens.watering_schedules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_subscription_plans_updated_at BEFORE UPDATE ON billing.subscription_plans FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON billing.subscriptions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();