package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	gardenapi "github.com/ryanmcafee/smart-garden-bot/operator/api/v1"
	"github.com/ryanmcafee/smart-garden-bot/operator/internal/weather"
	"github.com/ryanmcafee/smart-garden-bot/operator/internal/device"
	"github.com/ryanmcafee/smart-garden-bot/operator/internal/scheduler"
)

// GardenControllerReconciler reconciles a GardenController object
type GardenControllerReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ReconcileInterval time.Duration
	WeatherService    weather.Service
	DeviceManager     device.Manager
	WateringScheduler scheduler.Scheduler
}

const (
	// Finalizer name
	gardenControllerFinalizer = "garden.smartbot.io/finalizer"

	// Condition types
	ConditionReady           = "Ready"
	ConditionWeatherUpdated  = "WeatherUpdated"
	ConditionDeviceConnected = "DeviceConnected"
	ConditionSensorsHealthy  = "SensorsHealthy"

	// Phases
	PhasePending     = "Pending"
	PhaseRunning     = "Running"
	PhaseError       = "Error"
	PhaseMaintenance = "Maintenance"
)

//+kubebuilder:rbac:groups=garden.smartbot.io,resources=gardencontrollers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=garden.smartbot.io,resources=gardencontrollers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=garden.smartbot.io,resources=gardencontrollers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *GardenControllerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the GardenController instance
	gardenController := &gardenapi.GardenController{}
	err := r.Get(ctx, req.NamespacedName, gardenController)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("GardenController resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get GardenController")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if gardenController.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, gardenController)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(gardenController, gardenControllerFinalizer) {
		controllerutil.AddFinalizer(gardenController, gardenControllerFinalizer)
		if err := r.Update(ctx, gardenController); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize status if needed
	if gardenController.Status.Phase == "" {
		gardenController.Status.Phase = PhasePending
		gardenController.Status.ObservedGeneration = gardenController.Generation
		if err := r.Status().Update(ctx, gardenController); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle maintenance mode
	if gardenController.Spec.MaintenanceMode {
		return r.handleMaintenanceMode(ctx, gardenController)
	}

	// Reconcile the garden controller
	result, err := r.reconcileNormal(ctx, gardenController)
	if err != nil {
		logger.Error(err, "Failed to reconcile GardenController")
		r.updatePhase(ctx, gardenController, PhaseError)
		return result, err
	}

	return result, nil
}

func (r *GardenControllerReconciler) handleDeletion(ctx context.Context, gc *gardenapi.GardenController) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling GardenController deletion")

	// Cleanup resources
	if err := r.cleanup(ctx, gc); err != nil {
		logger.Error(err, "Failed to cleanup resources")
		return ctrl.Result{}, err
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(gc, gardenControllerFinalizer)
	if err := r.Update(ctx, gc); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("GardenController deletion completed")
	return ctrl.Result{}, nil
}

func (r *GardenControllerReconciler) handleMaintenanceMode(ctx context.Context, gc *gardenapi.GardenController) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("GardenController in maintenance mode")

	// Stop all watering activities
	for _, zone := range gc.Spec.Zones {
		if err := r.DeviceManager.StopWatering(ctx, gc.Spec.DeviceConfig, zone.ID); err != nil {
			logger.Error(err, "Failed to stop watering for zone", "zoneId", zone.ID)
		}
	}

	// Update phase and status
	gc.Status.Phase = PhaseMaintenance
	r.setCondition(gc, ConditionReady, metav1.ConditionFalse, "MaintenanceMode", "Garden controller is in maintenance mode")

	if err := r.Status().Update(ctx, gc); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue after interval to check if maintenance mode is disabled
	return ctrl.Result{RequeueAfter: r.ReconcileInterval}, nil
}

func (r *GardenControllerReconciler) reconcileNormal(ctx context.Context, gc *gardenapi.GardenController) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Update weather data
	if err := r.updateWeatherData(ctx, gc); err != nil {
		logger.Error(err, "Failed to update weather data")
		r.setCondition(gc, ConditionWeatherUpdated, metav1.ConditionFalse, "WeatherUpdateFailed", err.Error())
	} else {
		r.setCondition(gc, ConditionWeatherUpdated, metav1.ConditionTrue, "WeatherUpdateSucceeded", "Weather data updated successfully")
	}

	// Check device connectivity
	if err := r.checkDeviceConnectivity(ctx, gc); err != nil {
		logger.Error(err, "Device connectivity check failed")
		r.setCondition(gc, ConditionDeviceConnected, metav1.ConditionFalse, "DeviceConnectivityFailed", err.Error())
	} else {
		r.setCondition(gc, ConditionDeviceConnected, metav1.ConditionTrue, "DeviceConnected", "Device is connected and responsive")
	}

	// Update sensor statuses
	if err := r.updateSensorStatuses(ctx, gc); err != nil {
		logger.Error(err, "Failed to update sensor statuses")
		r.setCondition(gc, ConditionSensorsHealthy, metav1.ConditionFalse, "SensorUpdateFailed", err.Error())
	} else {
		r.setCondition(gc, ConditionSensorsHealthy, metav1.ConditionTrue, "SensorsHealthy", "All sensors are healthy")
	}

	// Process watering schedules
	if gc.Spec.AutoWatering {
		if err := r.processWateringSchedules(ctx, gc); err != nil {
			logger.Error(err, "Failed to process watering schedules")
		}
	}

	// Update zone statuses
	r.updateZoneStatuses(ctx, gc)

	// Update overall status
	r.updateOverallStatus(ctx, gc)

	// Update status
	if err := r.Status().Update(ctx, gc); err != nil {
		return ctrl.Result{}, err
	}

	// Schedule next reconciliation
	return ctrl.Result{RequeueAfter: r.ReconcileInterval}, nil
}

func (r *GardenControllerReconciler) updateWeatherData(ctx context.Context, gc *gardenapi.GardenController) error {
	logger := log.FromContext(ctx)

	// Skip if weather was updated recently
	if gc.Status.LastWeatherUpdate != nil {
		timeSinceUpdate := time.Since(gc.Status.LastWeatherUpdate.Time)
		updateInterval, _ := time.ParseDuration(gc.Spec.WeatherIntegration.UpdateInterval)
		if updateInterval == 0 {
			updateInterval = 30 * time.Minute
		}
		if timeSinceUpdate < updateInterval {
			return nil
		}
	}

	// Fetch weather data
	weatherData, err := r.WeatherService.GetCurrentWeather(ctx, gc.Spec.WeatherIntegration)
	if err != nil {
		return fmt.Errorf("failed to fetch weather data: %w", err)
	}

	// Update status
	now := metav1.NewTime(time.Now())
	gc.Status.LastWeatherUpdate = &now
	gc.Status.WeatherConditions = &gardenapi.WeatherConditions{
		Temperature:   weatherData.Temperature,
		Humidity:      weatherData.Humidity,
		Pressure:      weatherData.Pressure,
		WindSpeed:     weatherData.WindSpeed,
		RainfallToday: weatherData.RainfallToday,
		ForecastRain:  weatherData.ForecastRain,
		LastUpdated:   now,
	}

	logger.Info("Weather data updated", "temperature", weatherData.Temperature, "humidity", weatherData.Humidity)
	return nil
}

func (r *GardenControllerReconciler) checkDeviceConnectivity(ctx context.Context, gc *gardenapi.GardenController) error {
	return r.DeviceManager.CheckConnectivity(ctx, gc.Spec.DeviceConfig)
}

func (r *GardenControllerReconciler) updateSensorStatuses(ctx context.Context, gc *gardenapi.GardenController) error {
	logger := log.FromContext(ctx)

	var sensorStatuses []gardenapi.SensorStatus

	// Collect all sensor IDs from zones
	sensorIDs := make(map[string]bool)
	for _, zone := range gc.Spec.Zones {
		for _, sensorID := range zone.SensorIDs {
			sensorIDs[sensorID] = true
		}
	}

	// Check status of each sensor
	for sensorID := range sensorIDs {
		status, err := r.DeviceManager.GetSensorStatus(ctx, gc.Spec.DeviceConfig, sensorID)
		if err != nil {
			logger.Error(err, "Failed to get sensor status", "sensorId", sensorID)
			sensorStatuses = append(sensorStatuses, gardenapi.SensorStatus{
				SensorID:  sensorID,
				Connected: false,
				Error:     err.Error(),
			})
			continue
		}

		sensorStatus := gardenapi.SensorStatus{
			SensorID:       sensorID,
			Connected:      status.Connected,
			BatteryLevel:   status.BatteryLevel,
			SignalStrength: status.SignalStrength,
		}

		if status.LastReading != nil {
			lastReading := metav1.NewTime(*status.LastReading)
			sensorStatus.LastReading = &lastReading
		}

		sensorStatuses = append(sensorStatuses, sensorStatus)
	}

	gc.Status.SensorStatuses = sensorStatuses
	return nil
}

func (r *GardenControllerReconciler) processWateringSchedules(ctx context.Context, gc *gardenapi.GardenController) error {
	logger := log.FromContext(ctx)

	for _, zone := range gc.Spec.Zones {
		if !zone.Enabled {
			continue
		}

		// Check if watering is needed
		shouldWater, reason, err := r.WateringScheduler.ShouldWater(ctx, gc, zone)
		if err != nil {
			logger.Error(err, "Failed to determine if watering is needed", "zoneId", zone.ID)
			continue
		}

		if shouldWater {
			logger.Info("Starting watering", "zoneId", zone.ID, "reason", reason)
			
			// Start watering
			err := r.DeviceManager.StartWatering(ctx, gc.Spec.DeviceConfig, zone.ID, time.Duration(zone.Duration)*time.Minute)
			if err != nil {
				logger.Error(err, "Failed to start watering", "zoneId", zone.ID)
				continue
			}

			// Record watering event
			wateringEvent := &gardenapi.WateringEvent{
				ZoneID:    zone.ID,
				StartTime: metav1.NewTime(time.Now()),
				Duration:  fmt.Sprintf("%dm", zone.Duration),
				Reason:    reason,
				Status:    "Started",
			}
			gc.Status.LastWateringEvent = wateringEvent

			logger.Info("Watering started successfully", "zoneId", zone.ID)
		}
	}

	return nil
}

func (r *GardenControllerReconciler) updateZoneStatuses(ctx context.Context, gc *gardenapi.GardenController) {
	var zoneStatuses []gardenapi.ZoneStatus

	for _, zone := range gc.Spec.Zones {
		status := gardenapi.ZoneStatus{
			ZoneID: zone.ID,
			Health: "Healthy",
		}

		// Check if zone is currently watering
		isWatering, err := r.DeviceManager.IsWatering(ctx, gc.Spec.DeviceConfig, zone.ID)
		if err == nil {
			status.IsWatering = isWatering
		}

		// Get soil moisture if sensors are available
		if len(zone.SensorIDs) > 0 {
			// Use first sensor for now (could average multiple sensors)
			sensorID := zone.SensorIDs[0]
			reading, err := r.DeviceManager.GetLatestSensorReading(ctx, gc.Spec.DeviceConfig, sensorID)
			if err == nil && reading != nil {
				status.SoilMoisture = &reading.SoilMoisture
			}
		}

		// Calculate next watering time
		if zone.Enabled {
			nextWatering, err := r.WateringScheduler.GetNextWateringTime(zone.Schedule)
			if err == nil {
				next := metav1.NewTime(nextWatering)
				status.NextWatering = &next
			}
		}

		zoneStatuses = append(zoneStatuses, status)
	}

	gc.Status.ZoneStatuses = zoneStatuses
}

func (r *GardenControllerReconciler) updateOverallStatus(ctx context.Context, gc *gardenapi.GardenController) {
	// Determine overall phase
	hasErrors := false
	for _, condition := range gc.Status.Conditions {
		if condition.Status == metav1.ConditionFalse {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		gc.Status.Phase = PhaseError
		r.setCondition(gc, ConditionReady, metav1.ConditionFalse, "HasErrors", "One or more components have errors")
	} else {
		gc.Status.Phase = PhaseRunning
		r.setCondition(gc, ConditionReady, metav1.ConditionTrue, "AllComponentsHealthy", "All components are operating normally")
	}

	gc.Status.ObservedGeneration = gc.Generation
}

func (r *GardenControllerReconciler) updatePhase(ctx context.Context, gc *gardenapi.GardenController, phase string) {
	gc.Status.Phase = phase
	r.Status().Update(ctx, gc)
}

func (r *GardenControllerReconciler) setCondition(gc *gardenapi.GardenController, conditionType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	// Find and update existing condition or append new one
	found := false
	for i, existingCondition := range gc.Status.Conditions {
		if existingCondition.Type == conditionType {
			if existingCondition.Status != status || existingCondition.Reason != reason {
				gc.Status.Conditions[i] = condition
			}
			found = true
			break
		}
	}

	if !found {
		gc.Status.Conditions = append(gc.Status.Conditions, condition)
	}
}

func (r *GardenControllerReconciler) cleanup(ctx context.Context, gc *gardenapi.GardenController) error {
	logger := log.FromContext(ctx)

	// Stop all watering activities
	for _, zone := range gc.Spec.Zones {
		if err := r.DeviceManager.StopWatering(ctx, gc.Spec.DeviceConfig, zone.ID); err != nil {
			logger.Error(err, "Failed to stop watering during cleanup", "zoneId", zone.ID)
		}
	}

	// Additional cleanup tasks can be added here
	logger.Info("Cleanup completed for GardenController")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GardenControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gardenapi.GardenController{}).
		WithOptions(ctrl.Options{
			MaxConcurrentReconciles: 1,
		}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}