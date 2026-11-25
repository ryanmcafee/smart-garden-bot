package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/database"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
)

// SensorService handles sensor data and watering operations
type SensorService struct {
	db *database.DB
}

// NewSensorService creates a new sensor service
func NewSensorService(db *database.DB) *SensorService {
	return &SensorService{db: db}
}

// ===== SENSOR READING OPERATIONS =====

// CreateSensorReading creates a new sensor reading
func (s *SensorService) CreateSensorReading(ctx context.Context, req *models.CreateSensorReadingRequest) (*models.SensorReading, error) {
	timestamp := time.Now()
	if req.Timestamp != nil {
		timestamp = *req.Timestamp
	}
	
	dataQuality := models.DataQualityGood
	if req.DataQuality != nil {
		dataQuality = *req.DataQuality
	}
	
	reading := &models.SensorReading{
		ID:                  uuid.New(),
		DeviceID:            req.DeviceID,
		ZoneID:              req.ZoneID,
		Timestamp:           timestamp,
		Temperature:         req.Temperature,
		Humidity:            req.Humidity,
		SoilMoisture:        req.SoilMoisture,
		SoilTemperature:     req.SoilTemperature,
		PHLevel:             req.PHLevel,
		LightIntensity:      req.LightIntensity,
		UVIndex:             req.UVIndex,
		AtmosphericPressure: req.AtmosphericPressure,
		WindSpeed:           req.WindSpeed,
		WindDirection:       req.WindDirection,
		Rainfall:            req.Rainfall,
		BatteryVoltage:      req.BatteryVoltage,
		SolarVoltage:        req.SolarVoltage,
		SignalStrength:      req.SignalStrength,
		DataQuality:         dataQuality,
		AdditionalData:      req.AdditionalData,
	}
	
	query := `
		INSERT INTO sensors.readings (
			id, device_id, zone_id, timestamp, temperature, humidity, soil_moisture, 
			soil_temperature, ph_level, light_intensity, uv_index, atmospheric_pressure,
			wind_speed, wind_direction, rainfall, battery_voltage, solar_voltage,
			signal_strength, data_quality, additional_data
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
		RETURNING id, timestamp
	`
	
	err := s.db.Conn().QueryRowContext(ctx, query,
		reading.ID,
		reading.DeviceID,
		reading.ZoneID,
		reading.Timestamp,
		reading.Temperature,
		reading.Humidity,
		reading.SoilMoisture,
		reading.SoilTemperature,
		reading.PHLevel,
		reading.LightIntensity,
		reading.UVIndex,
		reading.AtmosphericPressure,
		reading.WindSpeed,
		reading.WindDirection,
		reading.Rainfall,
		reading.BatteryVoltage,
		reading.SolarVoltage,
		reading.SignalStrength,
		reading.DataQuality,
		reading.AdditionalData,
	).Scan(&reading.ID, &reading.Timestamp)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create sensor reading: %w", err)
	}
	
	return reading, nil
}

// GetSensorReadings retrieves sensor readings with filtering
func (s *SensorService) GetSensorReadings(ctx context.Context, query *models.SensorReadingsQuery) ([]models.SensorReading, error) {
	// Build dynamic query
	baseQuery := `
		SELECT id, device_id, zone_id, timestamp, temperature, humidity, soil_moisture,
		       soil_temperature, ph_level, light_intensity, uv_index, atmospheric_pressure,
		       wind_speed, wind_direction, rainfall, battery_voltage, solar_voltage,
		       signal_strength, data_quality, additional_data
		FROM sensors.readings
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 1
	whereParts := []string{}
	
	if query.DeviceID != nil {
		whereParts = append(whereParts, fmt.Sprintf("device_id = $%d", argCount))
		args = append(args, *query.DeviceID)
		argCount++
	}
	
	if query.ZoneID != nil {
		whereParts = append(whereParts, fmt.Sprintf("zone_id = $%d", argCount))
		args = append(args, *query.ZoneID)
		argCount++
	}
	
	// Handle garden_id filter by joining with zones table
	if query.GardenID != nil {
		whereParts = append(whereParts, fmt.Sprintf(`
			zone_id IN (SELECT id FROM gardens.zones WHERE garden_id = $%d)
		`, argCount))
		args = append(args, *query.GardenID)
		argCount++
	}
	
	if query.StartTime != nil {
		whereParts = append(whereParts, fmt.Sprintf("timestamp >= $%d", argCount))
		args = append(args, *query.StartTime)
		argCount++
	}
	
	if query.EndTime != nil {
		whereParts = append(whereParts, fmt.Sprintf("timestamp <= $%d", argCount))
		args = append(args, *query.EndTime)
		argCount++
	}
	
	if len(whereParts) > 0 {
		baseQuery += " AND " + fmt.Sprintf("%s", whereParts[0])
		for i := 1; i < len(whereParts); i++ {
			baseQuery += " AND " + whereParts[i]
		}
	}
	
	// Add ordering
	orderBy := "timestamp"
	if query.OrderBy != "" {
		orderBy = query.OrderBy
	}
	orderDir := "DESC"
	if query.OrderDir != "" {
		orderDir = query.OrderDir
	}
	baseQuery += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)
	
	// Add pagination
	limit := 50
	if query.Limit > 0 {
		limit = query.Limit
	}
	offset := 0
	if query.Offset > 0 {
		offset = query.Offset
	}
	
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)
	
	rows, err := s.db.Conn().QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sensor readings: %w", err)
	}
	defer rows.Close()
	
	var readings []models.SensorReading
	for rows.Next() {
		var reading models.SensorReading
		err := rows.Scan(
			&reading.ID,
			&reading.DeviceID,
			&reading.ZoneID,
			&reading.Timestamp,
			&reading.Temperature,
			&reading.Humidity,
			&reading.SoilMoisture,
			&reading.SoilTemperature,
			&reading.PHLevel,
			&reading.LightIntensity,
			&reading.UVIndex,
			&reading.AtmosphericPressure,
			&reading.WindSpeed,
			&reading.WindDirection,
			&reading.Rainfall,
			&reading.BatteryVoltage,
			&reading.SolarVoltage,
			&reading.SignalStrength,
			&reading.DataQuality,
			&reading.AdditionalData,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sensor reading: %w", err)
		}
		readings = append(readings, reading)
	}
	
	return readings, nil
}

// GetTimeSeriesData retrieves time-series data for a specific metric
func (s *SensorService) GetTimeSeriesData(ctx context.Context, deviceID uuid.UUID, metric string, startTime, endTime time.Time) (*models.TimeSeriesResponse, error) {
	// Map metric names to column names
	columnMap := map[string]string{
		"temperature":         "temperature",
		"humidity":            "humidity",
		"soil_moisture":       "soil_moisture",
		"soil_temperature":    "soil_temperature",
		"ph_level":            "ph_level",
		"light_intensity":     "light_intensity",
		"uv_index":            "uv_index",
		"atmospheric_pressure": "atmospheric_pressure",
		"wind_speed":          "wind_speed",
		"rainfall":            "rainfall",
		"battery_voltage":     "battery_voltage",
		"solar_voltage":       "solar_voltage",
	}
	
	column, exists := columnMap[metric]
	if !exists {
		return nil, fmt.Errorf("invalid metric: %s", metric)
	}
	
	query := fmt.Sprintf(`
		SELECT timestamp, %s
		FROM sensors.readings
		WHERE device_id = $1 AND timestamp BETWEEN $2 AND $3 AND %s IS NOT NULL
		ORDER BY timestamp ASC
	`, column, column)
	
	rows, err := s.db.Conn().QueryContext(ctx, query, deviceID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get time series data: %w", err)
	}
	defer rows.Close()
	
	var dataPoints []models.TimeSeriesDataPoint
	var zoneID *uuid.UUID
	
	for rows.Next() {
		var timestamp time.Time
		var value float64
		err := rows.Scan(&timestamp, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan time series data: %w", err)
		}
		
		dataPoints = append(dataPoints, models.TimeSeriesDataPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	
	// Get zone ID from the first reading
	if len(dataPoints) > 0 {
		zoneQuery := `SELECT zone_id FROM sensors.readings WHERE device_id = $1 LIMIT 1`
		s.db.Conn().QueryRowContext(ctx, zoneQuery, deviceID).Scan(&zoneID)
	}
	
	// Determine unit based on metric
	unitMap := map[string]string{
		"temperature":         "°C",
		"humidity":            "%",
		"soil_moisture":       "%",
		"soil_temperature":    "°C",
		"ph_level":            "pH",
		"light_intensity":     "lux",
		"uv_index":            "",
		"atmospheric_pressure": "hPa",
		"wind_speed":          "km/h",
		"rainfall":            "mm",
		"battery_voltage":     "V",
		"solar_voltage":       "V",
	}
	
	return &models.TimeSeriesResponse{
		Metric:     metric,
		Unit:       unitMap[metric],
		DeviceID:   deviceID,
		ZoneID:     zoneID,
		StartTime:  startTime,
		EndTime:    endTime,
		DataPoints: dataPoints,
	}, nil
}

// GetLatestReadingForDevice gets the most recent reading for a device
func (s *SensorService) GetLatestReadingForDevice(ctx context.Context, deviceID uuid.UUID) (*models.SensorReading, error) {
	query := `
		SELECT id, device_id, zone_id, timestamp, temperature, humidity, soil_moisture,
		       soil_temperature, ph_level, light_intensity, uv_index, atmospheric_pressure,
		       wind_speed, wind_direction, rainfall, battery_voltage, solar_voltage,
		       signal_strength, data_quality, additional_data
		FROM sensors.readings
		WHERE device_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	
	var reading models.SensorReading
	err := s.db.Conn().QueryRowContext(ctx, query, deviceID).Scan(
		&reading.ID,
		&reading.DeviceID,
		&reading.ZoneID,
		&reading.Timestamp,
		&reading.Temperature,
		&reading.Humidity,
		&reading.SoilMoisture,
		&reading.SoilTemperature,
		&reading.PHLevel,
		&reading.LightIntensity,
		&reading.UVIndex,
		&reading.AtmosphericPressure,
		&reading.WindSpeed,
		&reading.WindDirection,
		&reading.Rainfall,
		&reading.BatteryVoltage,
		&reading.SolarVoltage,
		&reading.SignalStrength,
		&reading.DataQuality,
		&reading.AdditionalData,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get latest reading: %w", err)
	}
	
	return &reading, nil
}

// ===== WATERING EVENT OPERATIONS =====

// CreateWateringEvent creates a new watering event
func (s *SensorService) CreateWateringEvent(ctx context.Context, req *models.CreateWateringEventRequest) (*models.WateringEvent, error) {
	event := &models.WateringEvent{
		ID:                     uuid.New(),
		ZoneID:                 req.ZoneID,
		ScheduleID:             req.ScheduleID,
		StartedAt:              time.Now(),
		PlannedDurationMinutes: req.PlannedDurationMinutes,
		TriggerType:            req.TriggerType,
		TriggerReason:          req.TriggerReason,
		SoilMoistureBefore:     req.SoilMoistureBefore,
		Temperature:            req.Temperature,
		WeatherConditions:      req.WeatherConditions,
		Status:                 models.WateringStatusStarted,
		Metadata:               req.Metadata,
	}
	
	query := `
		INSERT INTO sensors.watering_events (
			id, zone_id, schedule_id, started_at, planned_duration_minutes,
			trigger_type, trigger_reason, soil_moisture_before, temperature,
			weather_conditions, status, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		RETURNING id, started_at
	`
	
	err := s.db.Conn().QueryRowContext(ctx, query,
		event.ID,
		event.ZoneID,
		event.ScheduleID,
		event.StartedAt,
		event.PlannedDurationMinutes,
		event.TriggerType,
		event.TriggerReason,
		event.SoilMoistureBefore,
		event.Temperature,
		event.WeatherConditions,
		event.Status,
		event.Metadata,
	).Scan(&event.ID, &event.StartedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create watering event: %w", err)
	}
	
	return event, nil
}

// UpdateWateringEvent updates an existing watering event
func (s *SensorService) UpdateWateringEvent(ctx context.Context, eventID uuid.UUID, req *models.UpdateWateringEventRequest) (*models.WateringEvent, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argCount := 1
	
	if req.ActualDurationMinutes != nil {
		setParts = append(setParts, fmt.Sprintf("actual_duration_minutes = $%d", argCount))
		args = append(args, *req.ActualDurationMinutes)
		argCount++
	}
	
	if req.WaterVolumeLiters != nil {
		setParts = append(setParts, fmt.Sprintf("water_volume_liters = $%d", argCount))
		args = append(args, *req.WaterVolumeLiters)
		argCount++
	}
	
	if req.SoilMoistureAfter != nil {
		setParts = append(setParts, fmt.Sprintf("soil_moisture_after = $%d", argCount))
		args = append(args, *req.SoilMoistureAfter)
		argCount++
	}
	
	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *req.Status)
		argCount++
		
		// Set ended_at if status is completed, cancelled, or failed
		if *req.Status == models.WateringStatusCompleted || 
		   *req.Status == models.WateringStatusCancelled || 
		   *req.Status == models.WateringStatusFailed {
			setParts = append(setParts, fmt.Sprintf("ended_at = $%d", argCount))
			args = append(args, time.Now())
			argCount++
		}
	}
	
	if req.ErrorMessage != nil {
		setParts = append(setParts, fmt.Sprintf("error_message = $%d", argCount))
		args = append(args, *req.ErrorMessage)
		argCount++
	}
	
	if req.Metadata != nil {
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argCount))
		args = append(args, req.Metadata)
		argCount++
	}
	
	if len(setParts) == 0 {
		return s.GetWateringEvent(ctx, eventID)
	}
	
	// Add WHERE clause
	args = append(args, eventID)
	whereClause := fmt.Sprintf("WHERE id = $%d", argCount)
	
	query := fmt.Sprintf(`
		UPDATE sensors.watering_events 
		SET %s
		%s
		RETURNING id, zone_id, schedule_id, started_at, ended_at, planned_duration_minutes,
		          actual_duration_minutes, water_volume_liters, trigger_type, trigger_reason,
		          soil_moisture_before, soil_moisture_after, temperature, weather_conditions,
		          status, error_message, metadata
	`, fmt.Sprintf("%s", setParts), whereClause)
	
	var event models.WateringEvent
	err := s.db.Conn().QueryRowContext(ctx, query, args...).Scan(
		&event.ID,
		&event.ZoneID,
		&event.ScheduleID,
		&event.StartedAt,
		&event.EndedAt,
		&event.PlannedDurationMinutes,
		&event.ActualDurationMinutes,
		&event.WaterVolumeLiters,
		&event.TriggerType,
		&event.TriggerReason,
		&event.SoilMoistureBefore,
		&event.SoilMoistureAfter,
		&event.Temperature,
		&event.WeatherConditions,
		&event.Status,
		&event.ErrorMessage,
		&event.Metadata,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update watering event: %w", err)
	}
	
	return &event, nil
}

// GetWateringEvent retrieves a watering event by ID
func (s *SensorService) GetWateringEvent(ctx context.Context, eventID uuid.UUID) (*models.WateringEvent, error) {
	query := `
		SELECT id, zone_id, schedule_id, started_at, ended_at, planned_duration_minutes,
		       actual_duration_minutes, water_volume_liters, trigger_type, trigger_reason,
		       soil_moisture_before, soil_moisture_after, temperature, weather_conditions,
		       status, error_message, metadata
		FROM sensors.watering_events
		WHERE id = $1
	`
	
	var event models.WateringEvent
	err := s.db.Conn().QueryRowContext(ctx, query, eventID).Scan(
		&event.ID,
		&event.ZoneID,
		&event.ScheduleID,
		&event.StartedAt,
		&event.EndedAt,
		&event.PlannedDurationMinutes,
		&event.ActualDurationMinutes,
		&event.WaterVolumeLiters,
		&event.TriggerType,
		&event.TriggerReason,
		&event.SoilMoistureBefore,
		&event.SoilMoistureAfter,
		&event.Temperature,
		&event.WeatherConditions,
		&event.Status,
		&event.ErrorMessage,
		&event.Metadata,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get watering event: %w", err)
	}
	
	return &event, nil
}

// GetWateringEvents retrieves watering events for a zone or garden
func (s *SensorService) GetWateringEvents(ctx context.Context, zoneID *uuid.UUID, gardenID *uuid.UUID, limit, offset int) ([]models.WateringEvent, error) {
	baseQuery := `
		SELECT we.id, we.zone_id, we.schedule_id, we.started_at, we.ended_at, we.planned_duration_minutes,
		       we.actual_duration_minutes, we.water_volume_liters, we.trigger_type, we.trigger_reason,
		       we.soil_moisture_before, we.soil_moisture_after, we.temperature, we.weather_conditions,
		       we.status, we.error_message, we.metadata
		FROM sensors.watering_events we
	`
	
	args := []interface{}{}
	argCount := 1
	whereParts := []string{}
	
	if zoneID != nil {
		whereParts = append(whereParts, fmt.Sprintf("we.zone_id = $%d", argCount))
		args = append(args, *zoneID)
		argCount++
	} else if gardenID != nil {
		baseQuery += " JOIN gardens.zones z ON we.zone_id = z.id"
		whereParts = append(whereParts, fmt.Sprintf("z.garden_id = $%d", argCount))
		args = append(args, *gardenID)
		argCount++
	}
	
	if len(whereParts) > 0 {
		baseQuery += " WHERE " + fmt.Sprintf("%s", whereParts)
	}
	
	baseQuery += " ORDER BY we.started_at DESC"
	
	if limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
		argCount++
	}
	
	if offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}
	
	rows, err := s.db.Conn().QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get watering events: %w", err)
	}
	defer rows.Close()
	
	var events []models.WateringEvent
	for rows.Next() {
		var event models.WateringEvent
		err := rows.Scan(
			&event.ID,
			&event.ZoneID,
			&event.ScheduleID,
			&event.StartedAt,
			&event.EndedAt,
			&event.PlannedDurationMinutes,
			&event.ActualDurationMinutes,
			&event.WaterVolumeLiters,
			&event.TriggerType,
			&event.TriggerReason,
			&event.SoilMoistureBefore,
			&event.SoilMoistureAfter,
			&event.Temperature,
			&event.WeatherConditions,
			&event.Status,
			&event.ErrorMessage,
			&event.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan watering event: %w", err)
		}
		events = append(events, event)
	}
	
	return events, nil
}

// GetSensorStats returns statistics about sensor data
func (s *SensorService) GetSensorStats(ctx context.Context, gardenID *uuid.UUID) (map[string]interface{}, error) {
	baseQuery := `
		SELECT 
			COUNT(*) as total_readings,
			COUNT(DISTINCT device_id) as active_devices,
			COUNT(*) FILTER (WHERE timestamp > NOW() - INTERVAL '24 hours') as readings_24h,
			COUNT(*) FILTER (WHERE timestamp > NOW() - INTERVAL '7 days') as readings_7d
		FROM sensors.readings r
	`
	
	args := []interface{}{}
	if gardenID != nil {
		baseQuery += ` 
			JOIN gardens.zones z ON r.zone_id = z.id
			WHERE z.garden_id = $1
		`
		args = append(args, *gardenID)
	}
	
	var totalReadings, activeDevices, readings24h, readings7d int
	err := s.db.Conn().QueryRowContext(ctx, baseQuery, args...).Scan(
		&totalReadings, &activeDevices, &readings24h, &readings7d,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get sensor stats: %w", err)
	}
	
	stats := map[string]interface{}{
		"total_readings":  totalReadings,
		"active_devices":  activeDevices,
		"readings_24h":    readings24h,
		"readings_7d":     readings7d,
	}
	
	return stats, nil
}