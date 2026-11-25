package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GardenControllerSpec defines the desired state of GardenController
type GardenControllerSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// UserID identifies the owner of this garden
	// +kubebuilder:validation:Required
	UserID string `json:"userId"`

	// GardenID from the API database
	// +kubebuilder:validation:Required
	GardenID string `json:"gardenId"`

	// Location defines the garden's geographic location
	// +kubebuilder:validation:Required
	Location GardenLocation `json:"location"`

	// Zones defines the watering zones in this garden
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Zones []Zone `json:"zones"`

	// WeatherIntegration configures weather data sources
	// +kubebuilder:validation:Required
	WeatherIntegration WeatherConfig `json:"weatherIntegration"`

	// DeviceConfig specifies the IoT device configuration
	// +kubebuilder:validation:Required
	DeviceConfig DeviceConfiguration `json:"deviceConfig"`

	// AutoWatering enables automatic watering based on conditions
	// +kubebuilder:default=true
	AutoWatering bool `json:"autoWatering,omitempty"`

	// MaintenanceMode disables all watering when enabled
	// +kubebuilder:default=false
	MaintenanceMode bool `json:"maintenanceMode,omitempty"`

	// NotificationConfig defines how to send alerts
	NotificationConfig *NotificationConfig `json:"notificationConfig,omitempty"`
}

// GardenLocation defines geographic coordinates and timezone
type GardenLocation struct {
	// Name of the location (e.g., "Backyard Garden")
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Latitude in decimal degrees
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=-90
	// +kubebuilder:validation:Maximum=90
	Latitude float64 `json:"latitude"`

	// Longitude in decimal degrees
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=-180
	// +kubebuilder:validation:Maximum=180
	Longitude float64 `json:"longitude"`

	// Timezone in IANA format (e.g., "America/New_York")
	// +kubebuilder:validation:Required
	Timezone string `json:"timezone"`
}

// Zone represents a watering zone with its configuration
type Zone struct {
	// ID uniquely identifies this zone
	// +kubebuilder:validation:Required
	ID string `json:"id"`

	// Name is a human-readable name for the zone
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// PlantType describes what's planted in this zone
	PlantType string `json:"plantType,omitempty"`

	// SensorIDs lists the sensors monitoring this zone
	SensorIDs []string `json:"sensorIds,omitempty"`

	// Schedule defines when to water this zone (cron format)
	// +kubebuilder:validation:Required
	Schedule string `json:"schedule"`

	// Duration specifies how long to water (in minutes)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=120
	Duration int32 `json:"duration"`

	// MoisureThreshold is the soil moisture level below which watering is triggered
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MoistureThreshold *int32 `json:"moistureThreshold,omitempty"`

	// Enabled controls whether this zone is active
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
}

// WeatherConfig defines weather integration settings
type WeatherConfig struct {
	// Provider specifies the weather data provider
	// +kubebuilder:validation:Enum=openweathermap;weatherapi;noaa
	// +kubebuilder:default=openweathermap
	Provider string `json:"provider"`

	// Location for weather data (city name or coordinates)
	// +kubebuilder:validation:Required
	Location string `json:"location"`

	// APIKeySecret references a secret containing the API key
	APIKeySecret *SecretReference `json:"apiKeySecret,omitempty"`

	// UpdateInterval specifies how often to fetch weather data
	// +kubebuilder:default="30m"
	UpdateInterval string `json:"updateInterval,omitempty"`

	// RainThreshold defines the rainfall amount (mm) that skips watering
	// +kubebuilder:default=5.0
	RainThreshold float64 `json:"rainThreshold,omitempty"`
}

// DeviceConfiguration specifies IoT device settings
type DeviceConfiguration struct {
	// Type of the IoT device
	// +kubebuilder:validation:Enum=ecowitt;custom
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// Endpoint for device communication
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// CredentialsSecret references credentials for device authentication
	CredentialsSecret *SecretReference `json:"credentialsSecret,omitempty"`

	// Timeout for device communication
	// +kubebuilder:default="30s"
	Timeout string `json:"timeout,omitempty"`

	// RetryAttempts for failed device operations
	// +kubebuilder:default=3
	RetryAttempts int32 `json:"retryAttempts,omitempty"`
}

// NotificationConfig defines notification settings
type NotificationConfig struct {
	// Email notification settings
	Email *EmailNotification `json:"email,omitempty"`

	// Webhook notification settings
	Webhook *WebhookNotification `json:"webhook,omitempty"`

	// Slack notification settings
	Slack *SlackNotification `json:"slack,omitempty"`
}

// EmailNotification defines email notification settings
type EmailNotification struct {
	// Enabled controls whether email notifications are sent
	Enabled bool `json:"enabled"`

	// Recipients list of email addresses
	Recipients []string `json:"recipients,omitempty"`

	// SMTPSecret references SMTP configuration secret
	SMTPSecret *SecretReference `json:"smtpSecret,omitempty"`
}

// WebhookNotification defines webhook notification settings
type WebhookNotification struct {
	// Enabled controls whether webhook notifications are sent
	Enabled bool `json:"enabled"`

	// URL for the webhook endpoint
	URL string `json:"url,omitempty"`

	// Secret for webhook authentication
	Secret *SecretReference `json:"secret,omitempty"`
}

// SlackNotification defines Slack notification settings
type SlackNotification struct {
	// Enabled controls whether Slack notifications are sent
	Enabled bool `json:"enabled"`

	// WebhookURL for Slack integration
	WebhookURL string `json:"webhookUrl,omitempty"`

	// Channel to send notifications to
	Channel string `json:"channel,omitempty"`
}

// SecretReference references a Kubernetes secret
type SecretReference struct {
	// Name of the secret
	Name string `json:"name"`

	// Key within the secret
	Key string `json:"key"`
}

// GardenControllerStatus defines the observed state of GardenController
type GardenControllerStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Phase represents the current phase of the garden controller
	// +kubebuilder:validation:Enum=Pending;Running;Error;Maintenance
	Phase string `json:"phase,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastWeatherUpdate indicates when weather data was last fetched
	LastWeatherUpdate *metav1.Time `json:"lastWeatherUpdate,omitempty"`

	// LastWateringEvent records the most recent watering activity
	LastWateringEvent *WateringEvent `json:"lastWateringEvent,omitempty"`

	// ZoneStatuses provides status for each zone
	ZoneStatuses []ZoneStatus `json:"zoneStatuses,omitempty"`

	// SensorStatuses provides status for connected sensors
	SensorStatuses []SensorStatus `json:"sensorStatuses,omitempty"`

	// WeatherConditions contains current weather information
	WeatherConditions *WeatherConditions `json:"weatherConditions,omitempty"`

	// NextScheduledWatering indicates when the next watering will occur
	NextScheduledWatering *metav1.Time `json:"nextScheduledWatering,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed spec
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// WateringEvent represents a watering activity
type WateringEvent struct {
	// ZoneID identifies which zone was watered
	ZoneID string `json:"zoneId"`

	// StartTime when watering began
	StartTime metav1.Time `json:"startTime"`

	// EndTime when watering completed
	EndTime *metav1.Time `json:"endTime,omitempty"`

	// Duration of the watering session
	Duration string `json:"duration,omitempty"`

	// Reason why watering was triggered
	Reason string `json:"reason,omitempty"`

	// Status of the watering event
	// +kubebuilder:validation:Enum=Started;Completed;Failed;Cancelled
	Status string `json:"status"`

	// Error message if watering failed
	Error string `json:"error,omitempty"`
}

// ZoneStatus represents the status of a watering zone
type ZoneStatus struct {
	// ZoneID identifies the zone
	ZoneID string `json:"zoneId"`

	// IsWatering indicates if the zone is currently being watered
	IsWatering bool `json:"isWatering"`

	// LastWatered timestamp of the last watering
	LastWatered *metav1.Time `json:"lastWatered,omitempty"`

	// SoilMoisture percentage (0-100)
	SoilMoisture *float64 `json:"soilMoisture,omitempty"`

	// NextWatering scheduled time for next watering
	NextWatering *metav1.Time `json:"nextWatering,omitempty"`

	// Health status of the zone
	// +kubebuilder:validation:Enum=Healthy;Warning;Error
	Health string `json:"health,omitempty"`
}

// SensorStatus represents the status of a sensor
type SensorStatus struct {
	// SensorID identifies the sensor
	SensorID string `json:"sensorId"`

	// Connected indicates if the sensor is reachable
	Connected bool `json:"connected"`

	// LastReading timestamp of the last sensor reading
	LastReading *metav1.Time `json:"lastReading,omitempty"`

	// BatteryLevel percentage (0-100)
	BatteryLevel *float64 `json:"batteryLevel,omitempty"`

	// SignalStrength percentage (0-100)
	SignalStrength *float64 `json:"signalStrength,omitempty"`

	// Error message if sensor is malfunctioning
	Error string `json:"error,omitempty"`
}

// WeatherConditions represents current weather data
type WeatherConditions struct {
	// Temperature in Celsius
	Temperature float64 `json:"temperature"`

	// Humidity percentage
	Humidity float64 `json:"humidity"`

	// Pressure in hPa
	Pressure float64 `json:"pressure"`

	// WindSpeed in km/h
	WindSpeed float64 `json:"windSpeed"`

	// RainfallToday in mm
	RainfallToday float64 `json:"rainfallToday"`

	// ForecastRain expected rainfall in next 24h (mm)
	ForecastRain float64 `json:"forecastRain"`

	// LastUpdated timestamp when weather data was fetched
	LastUpdated metav1.Time `json:"lastUpdated"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=gc
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Zones",type=integer,JSONPath=`.spec.zones[*].id`
//+kubebuilder:printcolumn:name="Auto-Watering",type=boolean,JSONPath=`.spec.autoWatering`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// GardenController is the Schema for the gardencontrollers API
type GardenController struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GardenControllerSpec   `json:"spec,omitempty"`
	Status GardenControllerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GardenControllerList contains a list of GardenController
type GardenControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GardenController `json:"items"`
}