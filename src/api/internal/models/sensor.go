package models

import (
	"time"

	"github.com/google/uuid"
)

// SensorReading represents sensor data
type SensorReading struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	DeviceID             uuid.UUID  `json:"device_id" db:"device_id"`
	ZoneID               *uuid.UUID `json:"zone_id,omitempty" db:"zone_id"`
	Timestamp            time.Time  `json:"timestamp" db:"timestamp"`
	Temperature          *float64   `json:"temperature,omitempty" db:"temperature"`
	Humidity             *float64   `json:"humidity,omitempty" db:"humidity"`
	SoilMoisture         *float64   `json:"soil_moisture,omitempty" db:"soil_moisture"`
	SoilTemperature      *float64   `json:"soil_temperature,omitempty" db:"soil_temperature"`
	PHLevel              *float64   `json:"ph_level,omitempty" db:"ph_level"`
	LightIntensity       *float64   `json:"light_intensity,omitempty" db:"light_intensity"`
	UVIndex              *float64   `json:"uv_index,omitempty" db:"uv_index"`
	AtmosphericPressure  *float64   `json:"atmospheric_pressure,omitempty" db:"atmospheric_pressure"`
	WindSpeed            *float64   `json:"wind_speed,omitempty" db:"wind_speed"`
	WindDirection        *int       `json:"wind_direction,omitempty" db:"wind_direction" validate:"omitempty,min=0,max=360"`
	Rainfall             *float64   `json:"rainfall,omitempty" db:"rainfall"`
	BatteryVoltage       *float64   `json:"battery_voltage,omitempty" db:"battery_voltage"`
	SolarVoltage         *float64   `json:"solar_voltage,omitempty" db:"solar_voltage"`
	SignalStrength       *int       `json:"signal_strength,omitempty" db:"signal_strength" validate:"omitempty,min=0,max=100"`
	DataQuality          string     `json:"data_quality" db:"data_quality" validate:"oneof=excellent good fair poor"`
	AdditionalData       Metadata   `json:"additional_data,omitempty" db:"additional_data"`
}

// CreateSensorReadingRequest represents the request to create a sensor reading
type CreateSensorReadingRequest struct {
	DeviceID            uuid.UUID  `json:"device_id" validate:"required"`
	ZoneID              *uuid.UUID `json:"zone_id,omitempty"`
	Timestamp           *time.Time `json:"timestamp,omitempty"`
	Temperature         *float64   `json:"temperature,omitempty"`
	Humidity            *float64   `json:"humidity,omitempty"`
	SoilMoisture        *float64   `json:"soil_moisture,omitempty"`
	SoilTemperature     *float64   `json:"soil_temperature,omitempty"`
	PHLevel             *float64   `json:"ph_level,omitempty"`
	LightIntensity      *float64   `json:"light_intensity,omitempty"`
	UVIndex             *float64   `json:"uv_index,omitempty"`
	AtmosphericPressure *float64   `json:"atmospheric_pressure,omitempty"`
	WindSpeed           *float64   `json:"wind_speed,omitempty"`
	WindDirection       *int       `json:"wind_direction,omitempty" validate:"omitempty,min=0,max=360"`
	Rainfall            *float64   `json:"rainfall,omitempty"`
	BatteryVoltage      *float64   `json:"battery_voltage,omitempty"`
	SolarVoltage        *float64   `json:"solar_voltage,omitempty"`
	SignalStrength      *int       `json:"signal_strength,omitempty" validate:"omitempty,min=0,max=100"`
	DataQuality         *string    `json:"data_quality,omitempty" validate:"omitempty,oneof=excellent good fair poor"`
	AdditionalData      Metadata   `json:"additional_data,omitempty"`
}

// SensorReadingsQuery represents query parameters for sensor readings
type SensorReadingsQuery struct {
	DeviceID  *uuid.UUID `form:"device_id"`
	ZoneID    *uuid.UUID `form:"zone_id"`
	GardenID  *uuid.UUID `form:"garden_id"`
	StartTime *time.Time `form:"start_time"`
	EndTime   *time.Time `form:"end_time"`
	Limit     int        `form:"limit" validate:"min=1,max=1000"`
	Offset    int        `form:"offset" validate:"min=0"`
	OrderBy   string     `form:"order_by" validate:"oneof=timestamp temperature humidity soil_moisture"`
	OrderDir  string     `form:"order_dir" validate:"oneof=asc desc"`
}

// TimeSeriesDataPoint represents a single data point in a time series
type TimeSeriesDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TimeSeriesResponse represents time-series data response
type TimeSeriesResponse struct {
	Metric    string                `json:"metric"`
	Unit      string                `json:"unit,omitempty"`
	DeviceID  uuid.UUID             `json:"device_id"`
	ZoneID    *uuid.UUID            `json:"zone_id,omitempty"`
	StartTime time.Time             `json:"start_time"`
	EndTime   time.Time             `json:"end_time"`
	DataPoints []TimeSeriesDataPoint `json:"data_points"`
}

// WateringEvent represents a watering event log
type WateringEvent struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	ZoneID                 uuid.UUID  `json:"zone_id" db:"zone_id"`
	ScheduleID             *uuid.UUID `json:"schedule_id,omitempty" db:"schedule_id"`
	StartedAt              time.Time  `json:"started_at" db:"started_at"`
	EndedAt                *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	PlannedDurationMinutes int        `json:"planned_duration_minutes" db:"planned_duration_minutes"`
	ActualDurationMinutes  *int       `json:"actual_duration_minutes,omitempty" db:"actual_duration_minutes"`
	WaterVolumeLiters      *float64   `json:"water_volume_liters,omitempty" db:"water_volume_liters"`
	TriggerType            string     `json:"trigger_type" db:"trigger_type" validate:"oneof=scheduled manual sensor_threshold weather_based"`
	TriggerReason          *string    `json:"trigger_reason,omitempty" db:"trigger_reason"`
	SoilMoistureBefore     *float64   `json:"soil_moisture_before,omitempty" db:"soil_moisture_before"`
	SoilMoistureAfter      *float64   `json:"soil_moisture_after,omitempty" db:"soil_moisture_after"`
	Temperature            *float64   `json:"temperature,omitempty" db:"temperature"`
	WeatherConditions      Metadata   `json:"weather_conditions,omitempty" db:"weather_conditions"`
	Status                 string     `json:"status" db:"status" validate:"oneof=started completed cancelled failed"`
	ErrorMessage           *string    `json:"error_message,omitempty" db:"error_message"`
	Metadata               Metadata   `json:"metadata,omitempty" db:"metadata"`
}

// CreateWateringEventRequest represents the request to create a watering event
type CreateWateringEventRequest struct {
	ZoneID                 uuid.UUID  `json:"zone_id" validate:"required"`
	ScheduleID             *uuid.UUID `json:"schedule_id,omitempty"`
	PlannedDurationMinutes int        `json:"planned_duration_minutes" validate:"required,gt=0"`
	TriggerType            string     `json:"trigger_type" validate:"required,oneof=scheduled manual sensor_threshold weather_based"`
	TriggerReason          *string    `json:"trigger_reason,omitempty"`
	SoilMoistureBefore     *float64   `json:"soil_moisture_before,omitempty"`
	Temperature            *float64   `json:"temperature,omitempty"`
	WeatherConditions      Metadata   `json:"weather_conditions,omitempty"`
	Metadata               Metadata   `json:"metadata,omitempty"`
}

// UpdateWateringEventRequest represents the request to update a watering event
type UpdateWateringEventRequest struct {
	ActualDurationMinutes *int     `json:"actual_duration_minutes,omitempty" validate:"omitempty,gt=0"`
	WaterVolumeLiters     *float64 `json:"water_volume_liters,omitempty"`
	SoilMoistureAfter     *float64 `json:"soil_moisture_after,omitempty"`
	Status                *string  `json:"status,omitempty" validate:"omitempty,oneof=started completed cancelled failed"`
	ErrorMessage          *string  `json:"error_message,omitempty"`
	Metadata              Metadata `json:"metadata,omitempty"`
}

// WeatherData represents cached weather data
type WeatherData struct {
	ID           uuid.UUID `json:"id" db:"id"`
	LocationKey  string    `json:"location_key" db:"location_key"`
	Provider     string    `json:"provider" db:"provider"`
	DataType     string    `json:"data_type" db:"data_type" validate:"oneof=current forecast historical"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	Data         Metadata  `json:"data" db:"data"`
}

// CurrentWeatherData represents current weather conditions
type CurrentWeatherData struct {
	Location            string    `json:"location"`
	Temperature         float64   `json:"temperature"`         // Celsius
	FeelsLike           *float64  `json:"feels_like,omitempty"` // Celsius
	Humidity            float64   `json:"humidity"`            // Percentage
	Pressure            float64   `json:"pressure"`            // hPa
	WindSpeed           float64   `json:"wind_speed"`          // km/h
	WindDirection       int       `json:"wind_direction"`      // degrees
	WindGust            *float64  `json:"wind_gust,omitempty"` // km/h
	Visibility          *float64  `json:"visibility,omitempty"` // km
	UVIndex             *float64  `json:"uv_index,omitempty"`
	CloudCover          *int      `json:"cloud_cover,omitempty"` // percentage
	Rainfall            float64   `json:"rainfall"`              // mm
	Snowfall            *float64  `json:"snowfall,omitempty"`    // mm
	WeatherCode         string    `json:"weather_code"`
	WeatherDescription  string    `json:"weather_description"`
	Timestamp           time.Time `json:"timestamp"`
	Sunrise             *time.Time `json:"sunrise,omitempty"`
	Sunset              *time.Time `json:"sunset,omitempty"`
}

// WeatherForecast represents weather forecast data
type WeatherForecast struct {
	Location  string               `json:"location"`
	Timestamp time.Time            `json:"timestamp"`
	Days      []DailyForecast      `json:"days"`
	Hours     []HourlyForecast     `json:"hours,omitempty"`
}

// DailyForecast represents a single day's forecast
type DailyForecast struct {
	Date                time.Time `json:"date"`
	TemperatureMax      float64   `json:"temperature_max"`      // Celsius
	TemperatureMin      float64   `json:"temperature_min"`      // Celsius
	Humidity            float64   `json:"humidity"`             // Percentage
	WindSpeed           float64   `json:"wind_speed"`           // km/h
	WindDirection       int       `json:"wind_direction"`       // degrees
	RainProbability     int       `json:"rain_probability"`     // percentage
	RainfallTotal       float64   `json:"rainfall_total"`       // mm
	UVIndex             *float64  `json:"uv_index,omitempty"`
	WeatherCode         string    `json:"weather_code"`
	WeatherDescription  string    `json:"weather_description"`
	Sunrise             *time.Time `json:"sunrise,omitempty"`
	Sunset              *time.Time `json:"sunset,omitempty"`
}

// HourlyForecast represents hourly forecast data
type HourlyForecast struct {
	Timestamp           time.Time `json:"timestamp"`
	Temperature         float64   `json:"temperature"`          // Celsius
	FeelsLike           *float64  `json:"feels_like,omitempty"` // Celsius
	Humidity            float64   `json:"humidity"`             // Percentage
	WindSpeed           float64   `json:"wind_speed"`           // km/h
	WindDirection       int       `json:"wind_direction"`       // degrees
	RainProbability     int       `json:"rain_probability"`     // percentage
	Rainfall            float64   `json:"rainfall"`             // mm
	WeatherCode         string    `json:"weather_code"`
	WeatherDescription  string    `json:"weather_description"`
}

// WeatherQuery represents query parameters for weather data
type WeatherQuery struct {
	Location string `form:"location" validate:"required"`
	Provider string `form:"provider,omitempty" validate:"omitempty,oneof=openweathermap weatherapi noaa open-meteo"`
	Days     int    `form:"days,omitempty" validate:"omitempty,min=1,max=14"`
	Hours    int    `form:"hours,omitempty" validate:"omitempty,min=1,max=168"`
}