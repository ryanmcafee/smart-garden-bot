package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/database"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/models"
)

// GardenService handles garden-related operations
type GardenService struct {
	db *database.DB
}

// NewGardenService creates a new garden service
func NewGardenService(db *database.DB) *GardenService {
	return &GardenService{db: db}
}

// ===== GARDEN OPERATIONS =====

// ListGardens retrieves all gardens for a user
func (s *GardenService) ListGardens(ctx context.Context, userID uuid.UUID) ([]models.Garden, error) {
	query := `
		SELECT id, user_id, name, description, location, latitude, longitude, timezone, area_sqm, created_at, updated_at, metadata
		FROM gardens.gardens
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	
	var gardens []models.Garden
	err := s.db.SelectContext(ctx, &gardens, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list gardens: %w", err)
	}
	
	return gardens, nil
}

// GetGarden retrieves a specific garden
func (s *GardenService) GetGarden(ctx context.Context, userID, gardenID uuid.UUID) (*models.Garden, error) {
	query := `
		SELECT id, user_id, name, description, location, latitude, longitude, timezone, area_sqm, created_at, updated_at, metadata
		FROM gardens.gardens
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	
	var garden models.Garden
	err := s.db.GetContext(ctx, &garden, query, gardenID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("garden not found")
		}
		return nil, fmt.Errorf("failed to get garden: %w", err)
	}
	
	return &garden, nil
}

// CreateGarden creates a new garden
func (s *GardenService) CreateGarden(ctx context.Context, userID uuid.UUID, req *models.CreateGardenRequest) (*models.Garden, error) {
	garden := &models.Garden{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Timezone:    req.Timezone,
		AreaSqm:     req.AreaSqm,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}
	
	query := `
		INSERT INTO gardens.gardens (id, user_id, name, description, location, latitude, longitude, timezone, area_sqm, created_at, updated_at, metadata)
		VALUES (:id, :user_id, :name, :description, :location, :latitude, :longitude, :timezone, :area_sqm, :created_at, :updated_at, :metadata)
	`
	
	_, err := s.db.NamedExecContext(ctx, query, garden)
	if err != nil {
		return nil, fmt.Errorf("failed to create garden: %w", err)
	}
	
	return garden, nil
}

// UpdateGarden updates an existing garden
func (s *GardenService) UpdateGarden(ctx context.Context, userID, gardenID uuid.UUID, req *models.UpdateGardenRequest) (*models.Garden, error) {
	// Build dynamic query based on provided fields
	setParts := []string{}
	args := map[string]interface{}{
		"garden_id":  gardenID,
		"user_id":    userID,
		"updated_at": time.Now(),
	}
	
	if req.Name != nil {
		setParts = append(setParts, "name = :name")
		args["name"] = *req.Name
	}
	
	if req.Description != nil {
		setParts = append(setParts, "description = :description")
		args["description"] = *req.Description
	}
	
	if req.Location != nil {
		setParts = append(setParts, "location = :location")
		args["location"] = *req.Location
	}
	
	if req.Latitude != nil {
		setParts = append(setParts, "latitude = :latitude")
		args["latitude"] = *req.Latitude
	}
	
	if req.Longitude != nil {
		setParts = append(setParts, "longitude = :longitude")
		args["longitude"] = *req.Longitude
	}
	
	if req.Timezone != nil {
		setParts = append(setParts, "timezone = :timezone")
		args["timezone"] = *req.Timezone
	}
	
	if req.AreaSqm != nil {
		setParts = append(setParts, "area_sqm = :area_sqm")
		args["area_sqm"] = *req.AreaSqm
	}
	
	if req.Metadata != nil {
		setParts = append(setParts, "metadata = :metadata")
		args["metadata"] = req.Metadata
	}
	
	if len(setParts) == 0 {
		return s.GetGarden(ctx, userID, gardenID) // No changes, return existing garden
	}
	
	// Always update updated_at
	setParts = append(setParts, "updated_at = :updated_at")
	
	query := fmt.Sprintf(`
		UPDATE gardens.gardens 
		SET %s
		WHERE id = :garden_id AND user_id = :user_id AND deleted_at IS NULL
	`, strings.Join(setParts, ", "))
	
	result, err := s.db.NamedExecContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to update garden: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		return nil, fmt.Errorf("garden not found")
	}
	
	return s.GetGarden(ctx, userID, gardenID)
}

// DeleteGarden soft deletes a garden
func (s *GardenService) DeleteGarden(ctx context.Context, userID, gardenID uuid.UUID) error {
	query := `
		UPDATE gardens.gardens 
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`
	
	result, err := s.db.ExecContext(ctx, query, time.Now(), gardenID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete garden: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("garden not found")
	}
	
	return nil
}

// ===== ZONE OPERATIONS =====

// ListZones retrieves all zones for a garden
func (s *GardenService) ListZones(ctx context.Context, gardenID uuid.UUID) ([]models.Zone, error) {
	query := `
		SELECT id, garden_id, name, description, plant_type, area_sqm, soil_type, sun_exposure, created_at, updated_at, metadata
		FROM gardens.zones
		WHERE garden_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	
	var zones []models.Zone
	err := s.db.SelectContext(ctx, &zones, query, gardenID)
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}
	
	return zones, nil
}

// CreateZone creates a new zone
func (s *GardenService) CreateZone(ctx context.Context, gardenID uuid.UUID, req *models.CreateZoneRequest) (*models.Zone, error) {
	zone := &models.Zone{
		ID:          uuid.New(),
		GardenID:    gardenID,
		Name:        req.Name,
		Description: req.Description,
		PlantType:   req.PlantType,
		AreaSqm:     req.AreaSqm,
		SoilType:    req.SoilType,
		SunExposure: req.SunExposure,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}
	
	query := `
		INSERT INTO gardens.zones (id, garden_id, name, description, plant_type, area_sqm, soil_type, sun_exposure, created_at, updated_at, metadata)
		VALUES (:id, :garden_id, :name, :description, :plant_type, :area_sqm, :soil_type, :sun_exposure, :created_at, :updated_at, :metadata)
	`
	
	_, err := s.db.NamedExecContext(ctx, query, zone)
	if err != nil {
		return nil, fmt.Errorf("failed to create zone: %w", err)
	}
	
	return zone, nil
}

// UpdateZone updates an existing zone
func (s *GardenService) UpdateZone(ctx context.Context, gardenID, zoneID uuid.UUID, req *models.UpdateZoneRequest) (*models.Zone, error) {
	// Build dynamic query based on provided fields
	setParts := []string{}
	args := map[string]interface{}{
		"zone_id":    zoneID,
		"garden_id":  gardenID,
		"updated_at": time.Now(),
	}
	
	if req.Name != nil {
		setParts = append(setParts, "name = :name")
		args["name"] = *req.Name
	}
	
	if req.Description != nil {
		setParts = append(setParts, "description = :description")
		args["description"] = *req.Description
	}
	
	if req.PlantType != nil {
		setParts = append(setParts, "plant_type = :plant_type")
		args["plant_type"] = *req.PlantType
	}
	
	if req.AreaSqm != nil {
		setParts = append(setParts, "area_sqm = :area_sqm")
		args["area_sqm"] = *req.AreaSqm
	}
	
	if req.SoilType != nil {
		setParts = append(setParts, "soil_type = :soil_type")
		args["soil_type"] = *req.SoilType
	}
	
	if req.SunExposure != nil {
		setParts = append(setParts, "sun_exposure = :sun_exposure")
		args["sun_exposure"] = *req.SunExposure
	}
	
	if req.Metadata != nil {
		setParts = append(setParts, "metadata = :metadata")
		args["metadata"] = req.Metadata
	}
	
	if len(setParts) == 0 {
		return s.GetZone(ctx, gardenID, zoneID) // No changes, return existing zone
	}
	
	// Always update updated_at
	setParts = append(setParts, "updated_at = :updated_at")
	
	query := fmt.Sprintf(`
		UPDATE gardens.zones 
		SET %s
		WHERE id = :zone_id AND garden_id = :garden_id AND deleted_at IS NULL
	`, strings.Join(setParts, ", "))
	
	result, err := s.db.NamedExecContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to update zone: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		return nil, fmt.Errorf("zone not found")
	}
	
	return s.GetZone(ctx, gardenID, zoneID)
}

// GetZone retrieves a specific zone
func (s *GardenService) GetZone(ctx context.Context, gardenID, zoneID uuid.UUID) (*models.Zone, error) {
	query := `
		SELECT id, garden_id, name, description, plant_type, area_sqm, soil_type, sun_exposure, created_at, updated_at, metadata
		FROM gardens.zones
		WHERE id = $1 AND garden_id = $2 AND deleted_at IS NULL
	`
	
	var zone models.Zone
	err := s.db.GetContext(ctx, &zone, query, zoneID, gardenID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("zone not found")
		}
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}
	
	return &zone, nil
}

// DeleteZone soft deletes a zone
func (s *GardenService) DeleteZone(ctx context.Context, gardenID, zoneID uuid.UUID) error {
	query := `
		UPDATE gardens.zones 
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND garden_id = $3 AND deleted_at IS NULL
	`
	
	result, err := s.db.ExecContext(ctx, query, time.Now(), zoneID, gardenID)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("zone not found")
	}
	
	return nil
}