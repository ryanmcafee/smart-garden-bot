import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import type { Garden, Zone, SensorReading } from '@/lib/api-client';

interface GardenState {
  // Selected garden and zone
  selectedGarden: Garden | null;
  selectedZone: Zone | null;
  
  // Real-time data
  latestSensorReadings: Record<string, SensorReading>;
  wateringStatus: Record<string, boolean>; // zoneId -> isWatering
  
  // UI state
  isLoading: boolean;
  error: string | null;
  
  // Actions
  setSelectedGarden: (garden: Garden | null) => void;
  setSelectedZone: (zone: Zone | null) => void;
  updateSensorReading: (reading: SensorReading) => void;
  setWateringStatus: (zoneId: string, isWatering: boolean) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  clearError: () => void;
}

export const useGardenStore = create<GardenState>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state
        selectedGarden: null,
        selectedZone: null,
        latestSensorReadings: {},
        wateringStatus: {},
        isLoading: false,
        error: null,

        // Actions
        setSelectedGarden: (garden) => {
          set({ selectedGarden: garden, selectedZone: null });
        },

        setSelectedZone: (zone) => {
          set({ selectedZone: zone });
        },

        updateSensorReading: (reading) => {
          set((state) => ({
            latestSensorReadings: {
              ...state.latestSensorReadings,
              [reading.device_id]: reading,
            },
          }));
        },

        setWateringStatus: (zoneId, isWatering) => {
          set((state) => ({
            wateringStatus: {
              ...state.wateringStatus,
              [zoneId]: isWatering,
            },
          }));
        },

        setLoading: (loading) => {
          set({ isLoading: loading });
        },

        setError: (error) => {
          set({ error });
        },

        clearError: () => {
          set({ error: null });
        },
      }),
      {
        name: 'garden-store',
        // Only persist selected garden/zone, not real-time data
        partialize: (state) => ({
          selectedGarden: state.selectedGarden,
          selectedZone: state.selectedZone,
        }),
      }
    ),
    {
      name: 'garden-store',
    }
  )
);

// Notification store for managing alerts and messages
interface NotificationState {
  notifications: Notification[];
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp'>) => void;
  removeNotification: (id: string) => void;
  clearNotifications: () => void;
}

interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: Date;
  autoClose?: boolean;
  duration?: number; // in milliseconds
}

export const useNotificationStore = create<NotificationState>()(
  devtools((set, get) => ({
    notifications: [],

    addNotification: (notification) => {
      const id = crypto.randomUUID();
      const newNotification: Notification = {
        ...notification,
        id,
        timestamp: new Date(),
        autoClose: notification.autoClose ?? true,
        duration: notification.duration ?? 5000,
      };

      set((state) => ({
        notifications: [...state.notifications, newNotification],
      }));

      // Auto-remove notification if autoClose is enabled
      if (newNotification.autoClose) {
        setTimeout(() => {
          get().removeNotification(id);
        }, newNotification.duration);
      }
    },

    removeNotification: (id) => {
      set((state) => ({
        notifications: state.notifications.filter((n) => n.id !== id),
      }));
    },

    clearNotifications: () => {
      set({ notifications: [] });
    },
  }))
);

// Settings store for user preferences
interface SettingsState {
  theme: 'light' | 'dark' | 'system';
  temperatureUnit: 'celsius' | 'fahrenheit';
  timeFormat: '12h' | '24h';
  notifications: {
    enabled: boolean;
    email: boolean;
    push: boolean;
    sms: boolean;
  };
  autoRefresh: {
    enabled: boolean;
    interval: number; // in seconds
  };
  
  // Actions
  updateTheme: (theme: 'light' | 'dark' | 'system') => void;
  updateTemperatureUnit: (unit: 'celsius' | 'fahrenheit') => void;
  updateTimeFormat: (format: '12h' | '24h') => void;
  updateNotificationSettings: (settings: Partial<SettingsState['notifications']>) => void;
  updateAutoRefreshSettings: (settings: Partial<SettingsState['autoRefresh']>) => void;
}

export const useSettingsStore = create<SettingsState>()(
  devtools(
    persist(
      (set) => ({
        // Initial state
        theme: 'system',
        temperatureUnit: 'celsius',
        timeFormat: '24h',
        notifications: {
          enabled: true,
          email: true,
          push: true,
          sms: false,
        },
        autoRefresh: {
          enabled: true,
          interval: 30, // 30 seconds
        },

        // Actions
        updateTheme: (theme) => {
          set({ theme });
        },

        updateTemperatureUnit: (temperatureUnit) => {
          set({ temperatureUnit });
        },

        updateTimeFormat: (timeFormat) => {
          set({ timeFormat });
        },

        updateNotificationSettings: (notificationSettings) => {
          set((state) => ({
            notifications: {
              ...state.notifications,
              ...notificationSettings,
            },
          }));
        },

        updateAutoRefreshSettings: (autoRefreshSettings) => {
          set((state) => ({
            autoRefresh: {
              ...state.autoRefresh,
              ...autoRefreshSettings,
            },
          }));
        },
      }),
      {
        name: 'settings-store',
      }
    ),
    {
      name: 'settings-store',
    }
  )
);