package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/config"
)

type WeatherService struct {
	config         config.WeatherConfig
	client         *http.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

type WeatherData struct {
	Location    string    `json:"location"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Pressure    float64   `json:"pressure"`
	WindSpeed   float64   `json:"wind_speed"`
	WindDir     float64   `json:"wind_direction"`
	Rainfall    float64   `json:"rainfall"`
	UV          float64   `json:"uv_index"`
	Timestamp   time.Time `json:"timestamp"`
}

type ForecastData struct {
	Location string         `json:"location"`
	Days     []DayForecast  `json:"days"`
}

type DayForecast struct {
	Date           time.Time `json:"date"`
	TempMax        float64   `json:"temp_max"`
	TempMin        float64   `json:"temp_min"`
	Humidity       float64   `json:"humidity"`
	PrecipChance   float64   `json:"precip_chance"`
	PrecipAmount   float64   `json:"precip_amount"`
	WindSpeed      float64   `json:"wind_speed"`
	Description    string    `json:"description"`
}

type OpenWeatherMapResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity float64 `json:"humidity"`
		Pressure float64 `json:"pressure"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Name string `json:"name"`
}

type OpenWeatherMapForecastResponse struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			TempMax  float64 `json:"temp_max"`
			TempMin  float64 `json:"temp_min"`
			Humidity float64 `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
		Pop  float64 `json:"pop"`
		Rain struct {
			ThreeH float64 `json:"3h"`
		} `json:"rain"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	} `json:"list"`
	City struct {
		Name string `json:"name"`
	} `json:"city"`
}

func NewWeatherService(cfg config.WeatherConfig) *WeatherService {
	// Configure circuit breaker
	settings := gobreaker.Settings{
		Name:        "weather-api",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker '%s' changed from '%s' to '%s'\n", name, from, to)
		},
	}

	return &WeatherService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		circuitBreaker: gobreaker.NewCircuitBreaker(settings),
	}
}

func (ws *WeatherService) GetCurrentWeather(ctx context.Context, location string) (*WeatherData, error) {
	result, err := ws.circuitBreaker.Execute(func() (interface{}, error) {
		return ws.fetchCurrentWeather(ctx, location)
	})

	if err != nil {
		return nil, fmt.Errorf("weather service error: %w", err)
	}

	return result.(*WeatherData), nil
}

func (ws *WeatherService) GetForecast(ctx context.Context, location string, days int) (*ForecastData, error) {
	result, err := ws.circuitBreaker.Execute(func() (interface{}, error) {
		return ws.fetchForecast(ctx, location, days)
	})

	if err != nil {
		return nil, fmt.Errorf("weather forecast error: %w", err)
	}

	return result.(*ForecastData), nil
}

func (ws *WeatherService) fetchCurrentWeather(ctx context.Context, location string) (*WeatherData, error) {
	switch ws.config.DefaultProvider {
	case "openweathermap":
		return ws.fetchOpenWeatherMapCurrent(ctx, location)
	case "weatherapi":
		return ws.fetchWeatherAPICurrent(ctx, location)
	default:
		return nil, fmt.Errorf("unsupported weather provider: %s", ws.config.DefaultProvider)
	}
}

func (ws *WeatherService) fetchOpenWeatherMapCurrent(ctx context.Context, location string) (*WeatherData, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		location, ws.config.OpenWeatherMapAPIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var owmResp OpenWeatherMapResponse
	if err := json.NewDecoder(resp.Body).Decode(&owmResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &WeatherData{
		Location:    owmResp.Name,
		Temperature: owmResp.Main.Temp,
		Humidity:    owmResp.Main.Humidity,
		Pressure:    owmResp.Main.Pressure,
		WindSpeed:   owmResp.Wind.Speed,
		WindDir:     owmResp.Wind.Deg,
		Timestamp:   time.Now(),
	}, nil
}

func (ws *WeatherService) fetchForecast(ctx context.Context, location string, days int) (*ForecastData, error) {
	switch ws.config.DefaultProvider {
	case "openweathermap":
		return ws.fetchOpenWeatherMapForecast(ctx, location, days)
	case "weatherapi":
		return ws.fetchWeatherAPIForecast(ctx, location, days)
	default:
		return nil, fmt.Errorf("unsupported weather provider: %s", ws.config.DefaultProvider)
	}
}

func (ws *WeatherService) fetchOpenWeatherMapForecast(ctx context.Context, location string, days int) (*ForecastData, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric",
		location, ws.config.OpenWeatherMapAPIKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather forecast: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var owmResp OpenWeatherMapForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&owmResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Group forecast data by day
	dayMap := make(map[string]*DayForecast)
	for _, item := range owmResp.List {
		date := time.Unix(item.Dt, 0).Truncate(24 * time.Hour)
		dateKey := date.Format("2006-01-02")

		if dayForecast, exists := dayMap[dateKey]; exists {
			// Update with max/min temperatures
			if item.Main.TempMax > dayForecast.TempMax {
				dayForecast.TempMax = item.Main.TempMax
			}
			if item.Main.TempMin < dayForecast.TempMin {
				dayForecast.TempMin = item.Main.TempMin
			}
			// Average other values
			dayForecast.Humidity = (dayForecast.Humidity + item.Main.Humidity) / 2
			dayForecast.PrecipChance = (dayForecast.PrecipChance + item.Pop*100) / 2
			dayForecast.PrecipAmount += item.Rain.ThreeH
			dayForecast.WindSpeed = (dayForecast.WindSpeed + item.Wind.Speed) / 2
		} else {
			dayMap[dateKey] = &DayForecast{
				Date:         date,
				TempMax:      item.Main.TempMax,
				TempMin:      item.Main.TempMin,
				Humidity:     item.Main.Humidity,
				PrecipChance: item.Pop * 100,
				PrecipAmount: item.Rain.ThreeH,
				WindSpeed:    item.Wind.Speed,
				Description:  item.Weather[0].Description,
			}
		}
	}

	// Convert map to slice and limit to requested days
	var forecastDays []DayForecast
	for _, day := range dayMap {
		forecastDays = append(forecastDays, *day)
		if len(forecastDays) >= days {
			break
		}
	}

	return &ForecastData{
		Location: owmResp.City.Name,
		Days:     forecastDays,
	}, nil
}

func (ws *WeatherService) fetchWeatherAPICurrent(ctx context.Context, location string) (*WeatherData, error) {
	// Placeholder for WeatherAPI.com integration
	return nil, fmt.Errorf("WeatherAPI.com integration not implemented")
}

func (ws *WeatherService) fetchWeatherAPIForecast(ctx context.Context, location string, days int) (*ForecastData, error) {
	// Placeholder for WeatherAPI.com integration
	return nil, fmt.Errorf("WeatherAPI.com integration not implemented")
}

// GetWateringRecommendation analyzes weather data to provide watering recommendations
func (ws *WeatherService) GetWateringRecommendation(ctx context.Context, location string) (bool, string, error) {
	forecast, err := ws.GetForecast(ctx, location, 3)
	if err != nil {
		return true, "Unable to get weather data, proceed with normal watering", nil
	}

	// Simple logic: skip watering if rain is expected in next 24-48 hours
	for i, day := range forecast.Days {
		if i < 2 && day.PrecipChance > 70 && day.PrecipAmount > 5 {
			return false, fmt.Sprintf("Skip watering: %.0f%% chance of %.1fmm rain on %s",
				day.PrecipChance, day.PrecipAmount, day.Date.Format("Jan 2")), nil
		}
	}

	// Check for very high temperatures (need more water)
	current, err := ws.GetCurrentWeather(ctx, location)
	if err == nil && current.Temperature > 35 {
		return true, "High temperature detected, consider additional watering", nil
	}

	return true, "Normal watering conditions", nil
}