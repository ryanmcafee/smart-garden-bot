// Removed Auth0 dependency

export interface Garden {
  id: string;
  name: string;
  location: string;
  timezone: string;
  zones: Zone[];
  created_at: string;
  updated_at: string;
}

export interface Zone {
  id: string;
  name: string;
  plant_type: string;
  area: number;
  watering_schedule?: string;
}

export interface SensorReading {
  id: string;
  device_id: string;
  timestamp: string;
  temperature: number;
  humidity: number;
  soil_moisture: number;
  solar_radiation?: number;
  rainfall?: number;
  wind_speed?: number;
}

export interface WeatherData {
  location: string;
  temperature: number;
  humidity: number;
  pressure: number;
  wind_speed: number;
  wind_direction: number;
  rainfall: number;
  uv_index?: number;
  timestamp: string;
}

export interface WateringSchedule {
  id: string;
  zone_id: string;
  schedule: string;
  duration: number;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface APIError {
  error: string;
  message: string;
  code: number;
  request_id?: string;
}

class APIClient {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1';
  }

  private async getAuthHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    // TODO: Add authentication when needed
    // For now, no authentication is used

    return headers;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers = await this.getAuthHeaders();
    
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers: {
        ...headers,
        ...options.headers,
      },
    });

    if (!response.ok) {
      const errorData: APIError = await response.json().catch(() => ({
        error: 'Request failed',
        message: `HTTP ${response.status}: ${response.statusText}`,
        code: response.status,
      }));
      
      const error = new Error(errorData.message) as Error & { status: number; data: APIError };
      error.status = response.status;
      error.data = errorData;
      throw error;
    }

    return response.json();
  }

  // User methods
  async getProfile(): Promise<User> {
    return this.request<User>('/users/profile');
  }

  async updateProfile(data: Partial<User>): Promise<User> {
    return this.request<User>('/users/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  // Garden methods
  async getGardens(): Promise<Garden[]> {
    return this.request<Garden[]>('/gardens');
  }

  async getGarden(id: string): Promise<Garden> {
    return this.request<Garden>(`/gardens/${id}`);
  }

  async createGarden(data: Omit<Garden, 'id' | 'created_at' | 'updated_at' | 'zones'>): Promise<Garden> {
    return this.request<Garden>('/gardens', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateGarden(id: string, data: Partial<Garden>): Promise<Garden> {
    return this.request<Garden>(`/gardens/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteGarden(id: string): Promise<void> {
    return this.request<void>(`/gardens/${id}`, {
      method: 'DELETE',
    });
  }

  // Zone methods
  async getZones(gardenId: string): Promise<Zone[]> {
    return this.request<Zone[]>(`/gardens/${gardenId}/zones`);
  }

  async createZone(gardenId: string, data: Omit<Zone, 'id'>): Promise<Zone> {
    return this.request<Zone>(`/gardens/${gardenId}/zones`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateZone(gardenId: string, zoneId: string, data: Partial<Zone>): Promise<Zone> {
    return this.request<Zone>(`/gardens/${gardenId}/zones/${zoneId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteZone(gardenId: string, zoneId: string): Promise<void> {
    return this.request<void>(`/gardens/${gardenId}/zones/${zoneId}`, {
      method: 'DELETE',
    });
  }

  // Sensor methods
  async getSensorReadings(params?: {
    gardenId?: string;
    deviceId?: string;
    startTime?: string;
    endTime?: string;
    limit?: number;
  }): Promise<SensorReading[]> {
    const searchParams = new URLSearchParams();
    if (params?.gardenId) searchParams.set('garden_id', params.gardenId);
    if (params?.deviceId) searchParams.set('device_id', params.deviceId);
    if (params?.startTime) searchParams.set('start_time', params.startTime);
    if (params?.endTime) searchParams.set('end_time', params.endTime);
    if (params?.limit) searchParams.set('limit', params.limit.toString());

    const query = searchParams.toString();
    return this.request<SensorReading[]>(`/sensors/readings${query ? `?${query}` : ''}`);
  }

  async createSensorReading(data: Omit<SensorReading, 'id'>): Promise<SensorReading> {
    return this.request<SensorReading>('/sensors/readings', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getSensorDevices(): Promise<any[]> {
    return this.request<any[]>('/sensors/devices');
  }

  // Weather methods
  async getCurrentWeather(location: string): Promise<WeatherData> {
    return this.request<WeatherData>(`/weather/current?location=${encodeURIComponent(location)}`);
  }

  async getWeatherForecast(location: string, days?: number): Promise<any> {
    const params = new URLSearchParams({ location });
    if (days) params.set('days', days.toString());
    return this.request<any>(`/weather/forecast?${params.toString()}`);
  }

  // Watering methods
  async getWateringSchedules(gardenId: string): Promise<WateringSchedule[]> {
    return this.request<WateringSchedule[]>(`/watering/schedules?garden_id=${gardenId}`);
  }

  async createWateringSchedule(data: Omit<WateringSchedule, 'id' | 'created_at' | 'updated_at'>): Promise<WateringSchedule> {
    return this.request<WateringSchedule>('/watering/schedules', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateWateringSchedule(id: string, data: Partial<WateringSchedule>): Promise<WateringSchedule> {
    return this.request<WateringSchedule>(`/watering/schedules/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteWateringSchedule(id: string): Promise<void> {
    return this.request<void>(`/watering/schedules/${id}`, {
      method: 'DELETE',
    });
  }

  async triggerManualWatering(zoneId: string, duration: number): Promise<void> {
    return this.request<void>(`/watering/manual/${zoneId}`, {
      method: 'POST',
      body: JSON.stringify({ duration }),
    });
  }

  // Real-time data with Server-Sent Events
  createEventSource(endpoint: string): EventSource {
    return new EventSource(`${this.baseURL}${endpoint}`);
  }

  // WebSocket connection for real-time updates
  createWebSocket(endpoint: string): WebSocket {
    const wsURL = this.baseURL.replace('http', 'ws').replace('https', 'wss');
    return new WebSocket(`${wsURL}${endpoint}`);
  }
}

export const apiClient = new APIClient();