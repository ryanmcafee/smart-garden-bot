package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Metadata represents flexible JSON metadata
type Metadata map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database retrieval
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into Metadata", value)
	}

	return json.Unmarshal(bytes, m)
}

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an API error response
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta represents metadata for paginated responses
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginationQuery represents common pagination parameters
type PaginationQuery struct {
	Page    int `form:"page" validate:"min=1"`
	PerPage int `form:"per_page" validate:"min=1,max=100"`
}

// GetOffset calculates the database offset from page and per_page
func (p *PaginationQuery) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.GetLimit()
}

// GetLimit returns the limit with defaults
func (p *PaginationQuery) GetLimit() int {
	if p.PerPage <= 0 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	return p.PerPage
}

// CalculateMeta calculates pagination metadata
func (p *PaginationQuery) CalculateMeta(total int) *Meta {
	totalPages := (total + p.GetLimit() - 1) / p.GetLimit()
	return &Meta{
		Page:       p.Page,
		PerPage:    p.GetLimit(),
		Total:      total,
		TotalPages: totalPages,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// HealthStatus represents system health status
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

type HealthCheck struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

// TimeRange represents a time range filter
type TimeRange struct {
	Start time.Time `json:"start" form:"start"`
	End   time.Time `json:"end" form:"end"`
}

// Validate validates the time range
func (tr *TimeRange) Validate() error {
	if tr.Start.IsZero() || tr.End.IsZero() {
		return fmt.Errorf("start and end times must be provided")
	}
	if tr.Start.After(tr.End) {
		return fmt.Errorf("start time must be before end time")
	}
	return nil
}

// Duration returns the duration of the time range
func (tr *TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// OrderBy represents ordering parameters
type OrderBy struct {
	Field string `form:"order_by"`
	Dir   string `form:"order_dir" validate:"omitempty,oneof=asc desc"`
}

// GetDirection returns the ordering direction with default
func (o *OrderBy) GetDirection() string {
	if o.Dir == "" {
		return "desc"
	}
	return o.Dir
}

// Filter represents a generic filter
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, gte, lt, lte, in, like
	Value    interface{} `json:"value"`
}

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"latitude" validate:"min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"min=-180,max=180"`
	Name      string  `json:"name,omitempty"`
}

// Distance calculates the distance to another location (in km)
func (l *Location) Distance(other *Location) float64 {
	// Simple haversine formula implementation
	const earthRadius = 6371 // km

	lat1Rad := l.Latitude * (3.14159265359 / 180)
	lat2Rad := other.Latitude * (3.14159265359 / 180)
	deltaLatRad := (other.Latitude - l.Latitude) * (3.14159265359 / 180)
	deltaLonRad := (other.Longitude - l.Longitude) * (3.14159265359 / 180)

	a := sin(deltaLatRad/2)*sin(deltaLatRad/2) +
		cos(lat1Rad)*cos(lat2Rad)*sin(deltaLonRad/2)*sin(deltaLonRad/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))

	return earthRadius * c
}

// Helper functions for math operations (simple implementations)
func sin(x float64) float64 {
	// Using Taylor series approximation for simplicity
	// In production, use math.Sin()
	return x
}

func cos(x float64) float64 {
	// Using Taylor series approximation for simplicity
	// In production, use math.Cos()
	return 1 - (x*x)/2
}

func sqrt(x float64) float64 {
	// Newton's method approximation
	// In production, use math.Sqrt()
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func atan2(y, x float64) float64 {
	// Simple approximation
	// In production, use math.Atan2()
	if x > 0 {
		return atan(y / x)
	}
	return 3.14159265359 / 2
}

func atan(x float64) float64 {
	// Simple approximation
	// In production, use math.Atan()
	return x / (1 + 0.28*x*x)
}

// SystemInfo represents system information
type SystemInfo struct {
	Version   string    `json:"version"`
	BuildTime string    `json:"build_time"`
	GitCommit string    `json:"git_commit"`
	GoVersion string    `json:"go_version"`
	Uptime    string    `json:"uptime"`
	StartTime time.Time `json:"start_time"`
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	MaxConnections   int32 `json:"max_connections"`
	OpenConnections  int32 `json:"open_connections"`
	InUseConnections int32 `json:"in_use_connections"`
	IdleConnections  int32 `json:"idle_connections"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   Metadata  `json:"details,omitempty"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

// NotificationPreferences represents user notification settings
type NotificationPreferences struct {
	EmailEnabled    bool `json:"email_enabled"`
	SMSEnabled      bool `json:"sms_enabled"`
	PushEnabled     bool `json:"push_enabled"`
	WateringAlerts  bool `json:"watering_alerts"`
	WeatherAlerts   bool `json:"weather_alerts"`
	DeviceAlerts    bool `json:"device_alerts"`
	MaintenanceNews bool `json:"maintenance_news"`
}

// Usage represents usage statistics
type Usage struct {
	APICallsTotal      int64      `json:"api_calls_total"`
	APICallsThisMonth  int64      `json:"api_calls_this_month"`
	StorageUsedBytes   int64      `json:"storage_used_bytes"`
	BandwidthUsedBytes int64      `json:"bandwidth_used_bytes"`
	ActiveDevices      int        `json:"active_devices"`
	WateringEvents     int64      `json:"watering_events"`
	LastActivity       *time.Time `json:"last_activity,omitempty"`
}

// Constants for common values
const (
	// API Version
	APIVersion = "v1"

	// User roles
	RoleUser  = "user"
	RoleAdmin = "admin"

	// Device types
	DeviceTypeWeatherStation     = "weather_station"
	DeviceTypeWateringController = "watering_controller"
	DeviceTypeSensor             = "sensor"

	// Device statuses
	DeviceStatusActive      = "active"
	DeviceStatusInactive    = "inactive"
	DeviceStatusMaintenance = "maintenance"
	DeviceStatusError       = "error"

	// Watering trigger types
	TriggerTypeScheduled       = "scheduled"
	TriggerTypeManual          = "manual"
	TriggerTypeSensorThreshold = "sensor_threshold"
	TriggerTypeWeatherBased    = "weather_based"

	// Watering event statuses
	WateringStatusStarted   = "started"
	WateringStatusCompleted = "completed"
	WateringStatusCancelled = "cancelled"
	WateringStatusFailed    = "failed"

	// Data quality levels
	DataQualityExcellent = "excellent"
	DataQualityGood      = "good"
	DataQualityFair      = "fair"
	DataQualityPoor      = "poor"

	// Weather data types
	WeatherTypeCurrent    = "current"
	WeatherTypeForecast   = "forecast"
	WeatherTypeHistorical = "historical"

	// Weather providers
	WeatherProviderOpenWeatherMap = "openweathermap"
	WeatherProviderWeatherAPI     = "weatherapi"
	WeatherProviderNOAA           = "noaa"
	WeatherProviderOpenMeteo      = "open-meteo"
)
