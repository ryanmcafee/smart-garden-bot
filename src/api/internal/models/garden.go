package models

import (
	"time"

	"github.com/google/uuid"
)

// Garden represents a garden in the system
type Garden struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name" validate:"required,min=2,max=255"`
	Description *string   `json:"description,omitempty" db:"description"`
	Location    string    `json:"location" db:"location" validate:"required,max=255"`
	Latitude    *float64  `json:"latitude,omitempty" db:"latitude" validate:"omitempty,min=-90,max=90"`
	Longitude   *float64  `json:"longitude,omitempty" db:"longitude" validate:"omitempty,min=-180,max=180"`
	Timezone    string    `json:"timezone" db:"timezone" validate:"required"`
	AreaSqm     *float64  `json:"area_sqm,omitempty" db:"area_sqm" validate:"omitempty,gt=0"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Metadata    Metadata  `json:"metadata,omitempty" db:"metadata"`
	
	// Related data (loaded separately)
	Zones   []Zone   `json:"zones,omitempty"`
	Devices []Device `json:"devices,omitempty"`
}

// CreateGardenRequest represents the request to create a new garden
type CreateGardenRequest struct {
	Name        string    `json:"name" validate:"required,min=2,max=255"`
	Description *string   `json:"description,omitempty"`
	Location    string    `json:"location" validate:"required,max=255"`
	Latitude    *float64  `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude   *float64  `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	Timezone    string    `json:"timezone" validate:"required"`
	AreaSqm     *float64  `json:"area_sqm,omitempty" validate:"omitempty,gt=0"`
	Metadata    Metadata  `json:"metadata,omitempty"`
}

// UpdateGardenRequest represents the request to update a garden
type UpdateGardenRequest struct {
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description *string   `json:"description,omitempty"`
	Location    *string   `json:"location,omitempty" validate:"omitempty,max=255"`
	Latitude    *float64  `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude   *float64  `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	Timezone    *string   `json:"timezone,omitempty"`
	AreaSqm     *float64  `json:"area_sqm,omitempty" validate:"omitempty,gt=0"`
	Metadata    Metadata  `json:"metadata,omitempty"`
}

// Zone represents a zone within a garden
type Zone struct {
	ID                       uuid.UUID `json:"id" db:"id"`
	GardenID                 uuid.UUID `json:"garden_id" db:"garden_id"`
	Name                     string    `json:"name" db:"name" validate:"required,min=2,max=255"`
	Description              *string   `json:"description,omitempty" db:"description"`
	PlantType                *string   `json:"plant_type,omitempty" db:"plant_type"`
	AreaSqm                  *float64  `json:"area_sqm,omitempty" db:"area_sqm" validate:"omitempty,gt=0"`
	SoilType                 *string   `json:"soil_type,omitempty" db:"soil_type"`
	SunExposure              *string   `json:"sun_exposure,omitempty" db:"sun_exposure"`
	MoistureThreshold        *int      `json:"moisture_threshold,omitempty" db:"moisture_threshold" validate:"omitempty,min=0,max=100"`
	DefaultWateringDuration  int       `json:"default_watering_duration" db:"default_watering_duration" validate:"required,gt=0"`
	PositionX                *float64  `json:"position_x,omitempty" db:"position_x"`
	PositionY                *float64  `json:"position_y,omitempty" db:"position_y"`
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" db:"updated_at"`
	Metadata                 Metadata  `json:"metadata,omitempty" db:"metadata"`
	
	// Related data
	WateringSchedules []WateringSchedule `json:"watering_schedules,omitempty"`
	LatestReading     *SensorReading     `json:"latest_reading,omitempty"`
}

// CreateZoneRequest represents the request to create a new zone
type CreateZoneRequest struct {
	Name                    string   `json:"name" validate:"required,min=2,max=255"`
	Description             *string  `json:"description,omitempty"`
	PlantType               *string  `json:"plant_type,omitempty"`
	AreaSqm                 *float64 `json:"area_sqm,omitempty" validate:"omitempty,gt=0"`
	SoilType                *string  `json:"soil_type,omitempty"`
	SunExposure             *string  `json:"sun_exposure,omitempty"`
	MoistureThreshold       *int     `json:"moisture_threshold,omitempty" validate:"omitempty,min=0,max=100"`
	DefaultWateringDuration int      `json:"default_watering_duration" validate:"required,gt=0"`
	PositionX               *float64 `json:"position_x,omitempty"`
	PositionY               *float64 `json:"position_y,omitempty"`
	Metadata                Metadata `json:"metadata,omitempty"`
}

// UpdateZoneRequest represents the request to update a zone
type UpdateZoneRequest struct {
	Name                    *string  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description             *string  `json:"description,omitempty"`
	PlantType               *string  `json:"plant_type,omitempty"`
	AreaSqm                 *float64 `json:"area_sqm,omitempty" validate:"omitempty,gt=0"`
	SoilType                *string  `json:"soil_type,omitempty"`
	SunExposure             *string  `json:"sun_exposure,omitempty"`
	MoistureThreshold       *int     `json:"moisture_threshold,omitempty" validate:"omitempty,min=0,max=100"`
	DefaultWateringDuration *int     `json:"default_watering_duration,omitempty" validate:"omitempty,gt=0"`
	PositionX               *float64 `json:"position_x,omitempty"`
	PositionY               *float64 `json:"position_y,omitempty"`
	Metadata                Metadata `json:"metadata,omitempty"`
}

// Device represents an IoT device (weather stations, controllers, etc.)
type Device struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	GardenID         uuid.UUID  `json:"garden_id" db:"garden_id"`
	Name             string     `json:"name" db:"name" validate:"required,min=2,max=255"`
	DeviceType       string     `json:"device_type" db:"device_type" validate:"required,oneof=weather_station watering_controller sensor"`
	Manufacturer     *string    `json:"manufacturer,omitempty" db:"manufacturer"`
	Model            *string    `json:"model,omitempty" db:"model"`
	FirmwareVersion  *string    `json:"firmware_version,omitempty" db:"firmware_version"`
	MacAddress       *string    `json:"mac_address,omitempty" db:"mac_address"`
	IPAddress        *string    `json:"ip_address,omitempty" db:"ip_address"`
	EndpointURL      *string    `json:"endpoint_url,omitempty" db:"endpoint_url"`
	Status           string     `json:"status" db:"status" validate:"oneof=active inactive maintenance error"`
	LastSeenAt       *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`
	BatteryLevel     *int       `json:"battery_level,omitempty" db:"battery_level" validate:"omitempty,min=0,max=100"`
	SignalStrength   *int       `json:"signal_strength,omitempty" db:"signal_strength" validate:"omitempty,min=0,max=100"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	Configuration    Metadata   `json:"configuration,omitempty" db:"configuration"`
	Metadata         Metadata   `json:"metadata,omitempty" db:"metadata"`
}

// CreateDeviceRequest represents the request to create a new device
type CreateDeviceRequest struct {
	Name            string   `json:"name" validate:"required,min=2,max=255"`
	DeviceType      string   `json:"device_type" validate:"required,oneof=weather_station watering_controller sensor"`
	Manufacturer    *string  `json:"manufacturer,omitempty"`
	Model           *string  `json:"model,omitempty"`
	FirmwareVersion *string  `json:"firmware_version,omitempty"`
	MacAddress      *string  `json:"mac_address,omitempty"`
	IPAddress       *string  `json:"ip_address,omitempty"`
	EndpointURL     *string  `json:"endpoint_url,omitempty"`
	Configuration   Metadata `json:"configuration,omitempty"`
	Metadata        Metadata `json:"metadata,omitempty"`
}

// UpdateDeviceRequest represents the request to update a device
type UpdateDeviceRequest struct {
	Name            *string  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Status          *string  `json:"status,omitempty" validate:"omitempty,oneof=active inactive maintenance error"`
	FirmwareVersion *string  `json:"firmware_version,omitempty"`
	IPAddress       *string  `json:"ip_address,omitempty"`
	EndpointURL     *string  `json:"endpoint_url,omitempty"`
	BatteryLevel    *int     `json:"battery_level,omitempty" validate:"omitempty,min=0,max=100"`
	SignalStrength  *int     `json:"signal_strength,omitempty" validate:"omitempty,min=0,max=100"`
	Configuration   Metadata `json:"configuration,omitempty"`
	Metadata        Metadata `json:"metadata,omitempty"`
}

// WateringSchedule represents a watering schedule for a zone
type WateringSchedule struct {
	ID                        uuid.UUID  `json:"id" db:"id"`
	ZoneID                    uuid.UUID  `json:"zone_id" db:"zone_id"`
	Name                      string     `json:"name" db:"name" validate:"required,min=2,max=255"`
	CronExpression            string     `json:"cron_expression" db:"cron_expression" validate:"required"`
	DurationMinutes           int        `json:"duration_minutes" db:"duration_minutes" validate:"required,gt=0"`
	SkipIfRainProbability     int        `json:"skip_if_rain_probability" db:"skip_if_rain_probability" validate:"min=0,max=100"`
	SkipIfSoilMoistureAbove   *int       `json:"skip_if_soil_moisture_above,omitempty" db:"skip_if_soil_moisture_above" validate:"omitempty,min=0,max=100"`
	IsActive                  bool       `json:"is_active" db:"is_active"`
	CreatedAt                 time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at" db:"updated_at"`
	NextRunAt                 *time.Time `json:"next_run_at,omitempty" db:"next_run_at"`
	LastRunAt                 *time.Time `json:"last_run_at,omitempty" db:"last_run_at"`
	Metadata                  Metadata   `json:"metadata,omitempty" db:"metadata"`
}

// CreateWateringScheduleRequest represents the request to create a watering schedule
type CreateWateringScheduleRequest struct {
	Name                      string   `json:"name" validate:"required,min=2,max=255"`
	CronExpression            string   `json:"cron_expression" validate:"required"`
	DurationMinutes           int      `json:"duration_minutes" validate:"required,gt=0"`
	SkipIfRainProbability     int      `json:"skip_if_rain_probability" validate:"min=0,max=100"`
	SkipIfSoilMoistureAbove   *int     `json:"skip_if_soil_moisture_above,omitempty" validate:"omitempty,min=0,max=100"`
	Metadata                  Metadata `json:"metadata,omitempty"`
}

// UpdateWateringScheduleRequest represents the request to update a watering schedule
type UpdateWateringScheduleRequest struct {
	Name                      *string  `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	CronExpression            *string  `json:"cron_expression,omitempty"`
	DurationMinutes           *int     `json:"duration_minutes,omitempty" validate:"omitempty,gt=0"`
	SkipIfRainProbability     *int     `json:"skip_if_rain_probability,omitempty" validate:"omitempty,min=0,max=100"`
	SkipIfSoilMoistureAbove   *int     `json:"skip_if_soil_moisture_above,omitempty" validate:"omitempty,min=0,max=100"`
	IsActive                  *bool    `json:"is_active,omitempty"`
	Metadata                  Metadata `json:"metadata,omitempty"`
}

// ManualWateringRequest represents a request for manual watering
type ManualWateringRequest struct {
	DurationMinutes int     `json:"duration_minutes" validate:"required,gt=0,max=120"`
	Reason          *string `json:"reason,omitempty"`
}