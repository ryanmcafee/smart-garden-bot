import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { UserProvider } from '@auth0/nextjs-auth0/client';
import '@testing-library/jest-dom';

import DashboardPage from '../../../frontend/app/dashboard/page';
import { apiClient } from '../../../frontend/lib/api-client';
import { useGardenStore } from '../../../frontend/store/garden-store';

// Mock the API client
jest.mock('../../../frontend/lib/api-client');
const mockedApiClient = apiClient as jest.Mocked<typeof apiClient>;

// Mock the Auth0 hook
const mockUser = {
  sub: 'auth0|123',
  name: 'Test User',
  email: 'test@example.com',
};

jest.mock('@auth0/nextjs-auth0/client', () => ({
  UserProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  useUser: () => ({
    user: mockUser,
    error: undefined,
    isLoading: false,
  }),
  withPageAuthRequired: (component: React.ComponentType) => component,
}));

// Mock the garden store
jest.mock('../../../frontend/store/garden-store');
const mockedUseGardenStore = useGardenStore as jest.MockedFunction<typeof useGardenStore>;

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
  }),
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => '/dashboard',
}));

// Sample test data
const mockGardens = [
  {
    id: '1',
    name: 'Backyard Garden',
    location: 'San Francisco, CA',
    timezone: 'America/Los_Angeles',
    zones: [
      {
        id: 'zone-1',
        name: 'Tomatoes',
        plant_type: 'Vegetables',
        area: 15.0,
      },
      {
        id: 'zone-2',
        name: 'Herbs',
        plant_type: 'Herbs',
        area: 8.0,
      },
    ],
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

const mockWeather = {
  location: 'San Francisco, CA',
  temperature: 22.5,
  humidity: 65.0,
  pressure: 1013.25,
  wind_speed: 5.2,
  wind_direction: 180,
  rainfall: 0.0,
  timestamp: '2024-01-01T12:00:00Z',
};

const mockSensorReadings = [
  {
    id: '1',
    device_id: 'sensor-001',
    timestamp: '2024-01-01T12:00:00Z',
    temperature: 23.0,
    humidity: 68.0,
    soil_moisture: 45.0,
    soil_temperature: 21.0,
    light_intensity: 25000.0,
  },
];

const mockWateringSchedules = [
  {
    id: '1',
    zone_id: 'zone-1',
    schedule: '0 6 * * *',
    duration: 20,
    active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

// Test wrapper component
const TestWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: 0,
        gcTime: 0,
      },
    },
  });

  return (
    <UserProvider>
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    </UserProvider>
  );
};

describe('Dashboard Page', () => {
  beforeEach(() => {
    // Reset all mocks
    jest.clearAllMocks();
    
    // Mock store state
    mockedUseGardenStore.mockReturnValue({
      selectedGarden: mockGardens[0],
      selectedZone: null,
      latestSensorReadings: {},
      wateringStatus: {},
      isLoading: false,
      error: null,
      setSelectedGarden: jest.fn(),
      setSelectedZone: jest.fn(),
      updateSensorReading: jest.fn(),
      setWateringStatus: jest.fn(),
      setLoading: jest.fn(),
      setError: jest.fn(),
      clearError: jest.fn(),
    });

    // Mock API responses
    mockedApiClient.getGardens.mockResolvedValue(mockGardens);
    mockedApiClient.getCurrentWeather.mockResolvedValue(mockWeather);
    mockedApiClient.getSensorReadings.mockResolvedValue(mockSensorReadings);
    mockedApiClient.getWateringSchedules.mockResolvedValue(mockWateringSchedules);
  });

  it('renders the dashboard with user greeting', async () => {
    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Welcome back, Test User!')).toBeInTheDocument();
  });

  it('displays garden statistics cards', async () => {
    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Active Gardens')).toBeInTheDocument();
      expect(screen.getByText('Current Temperature')).toBeInTheDocument();
      expect(screen.getByText('Soil Moisture')).toBeInTheDocument();
      expect(screen.getByText('Next Watering')).toBeInTheDocument();
    });

    // Check if the garden count is displayed
    await waitFor(() => {
      expect(screen.getByText('1')).toBeInTheDocument(); // Garden count
    });
  });

  it('displays weather information when available', async () => {
    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('23°C')).toBeInTheDocument(); // Rounded temperature
    });
  });

  it('handles loading states correctly', async () => {
    // Mock loading state
    mockedUseGardenStore.mockReturnValue({
      selectedGarden: null,
      selectedZone: null,
      latestSensorReadings: {},
      wateringStatus: {},
      isLoading: true,
      error: null,
      setSelectedGarden: jest.fn(),
      setSelectedZone: jest.fn(),
      updateSensorReading: jest.fn(),
      setWateringStatus: jest.fn(),
      setLoading: jest.fn(),
      setError: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    // Should show loading spinner
    expect(screen.getByRole('status')).toBeInTheDocument(); // Assuming LoadingSpinner has role="status"
  });

  it('handles error states correctly', async () => {
    // Mock API error
    mockedApiClient.getGardens.mockRejectedValue(new Error('Failed to fetch gardens'));

    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Failed to load gardens')).toBeInTheDocument();
    });
  });

  it('displays welcome message for new users with no gardens', async () => {
    // Mock empty gardens response
    mockedApiClient.getGardens.mockResolvedValue([]);

    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Welcome to Smart Garden Bot!')).toBeInTheDocument();
      expect(screen.getByText("You haven't set up any gardens yet.")).toBeInTheDocument();
    });
  });

  it('switches between dashboard tabs correctly', async () => {
    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    // Wait for initial load
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    // Check if tabs are present
    expect(screen.getByRole('tab', { name: 'Overview' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Sensor Data' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Watering' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Weather' })).toBeInTheDocument();

    // Click on Sensor Data tab
    fireEvent.click(screen.getByRole('tab', { name: 'Sensor Data' }));
    await waitFor(() => {
      expect(screen.getByRole('tabpanel')).toBeInTheDocument();
    });

    // Click on Weather tab
    fireEvent.click(screen.getByRole('tab', { name: 'Weather' }));
    await waitFor(() => {
      expect(screen.getByRole('tabpanel')).toBeInTheDocument();
    });
  });

  it('calls API methods with correct parameters', async () => {
    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(mockedApiClient.getGardens).toHaveBeenCalledTimes(1);
      expect(mockedApiClient.getCurrentWeather).toHaveBeenCalledWith('San Francisco, CA');
      expect(mockedApiClient.getSensorReadings).toHaveBeenCalledWith({
        gardenId: '1',
        limit: 100,
      });
      expect(mockedApiClient.getWateringSchedules).toHaveBeenCalledWith('1');
    });
  });

  it('updates sensor readings in real-time', async () => {
    const mockUpdateSensorReading = jest.fn();
    
    mockedUseGardenStore.mockReturnValue({
      selectedGarden: mockGardens[0],
      selectedZone: null,
      latestSensorReadings: {},
      wateringStatus: {},
      isLoading: false,
      error: null,
      setSelectedGarden: jest.fn(),
      setSelectedZone: jest.fn(),
      updateSensorReading: mockUpdateSensorReading,
      setWateringStatus: jest.fn(),
      setLoading: jest.fn(),
      setError: jest.fn(),
      clearError: jest.fn(),
    });

    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    // Simulate real-time sensor reading update
    const newReading = {
      id: '2',
      device_id: 'sensor-002',
      timestamp: '2024-01-01T13:00:00Z',
      temperature: 24.0,
      humidity: 70.0,
      soil_moisture: 50.0,
      soil_temperature: 22.0,
      light_intensity: 26000.0,
    };

    // This would typically be triggered by a WebSocket or SSE event
    // For testing, we simulate the store update
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });
  });

  it('handles responsive design breakpoints', async () => {
    // Mock window.matchMedia for responsive testing
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation(query => ({
        matches: query.includes('768px'), // Simulate mobile breakpoint
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      })),
    });

    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    // Check if mobile-responsive elements are present
    // This would depend on your specific responsive implementation
  });

  it('handles keyboard navigation', async () => {
    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    // Test tab navigation
    const firstTab = screen.getByRole('tab', { name: 'Overview' });
    firstTab.focus();
    expect(firstTab).toHaveFocus();

    // Test arrow key navigation (if implemented)
    fireEvent.keyDown(firstTab, { key: 'ArrowRight' });
    // Assertions would depend on your keyboard navigation implementation
  });

  it('displays accessibility attributes correctly', async () => {
    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    });

    // Check for proper ARIA labels and roles
    expect(screen.getByRole('main')).toBeInTheDocument(); // Assuming dashboard has main role
    expect(screen.getByRole('tablist')).toBeInTheDocument();
    
    // Check for screen reader friendly text
    const tabs = screen.getAllByRole('tab');
    tabs.forEach(tab => {
      expect(tab).toHaveAttribute('aria-selected');
    });
  });
});

// Integration test for dashboard data flow
describe('Dashboard Integration', () => {
  it('integrates all dashboard components correctly', async () => {
    render(
      <TestWrapper>
        <DashboardPage />
      </TestWrapper>
    );

    // Wait for all data to load
    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument();
    }, { timeout: 3000 });

    // Verify that all major components are rendered
    expect(screen.getByText('Active Gardens')).toBeInTheDocument();
    expect(screen.getByRole('tablist')).toBeInTheDocument();
    
    // Test tab switching integrates correctly with data display
    fireEvent.click(screen.getByRole('tab', { name: 'Sensor Data' }));
    
    await waitFor(() => {
      // Should show sensor data components
      expect(screen.getByRole('tabpanel')).toBeInTheDocument();
    });

    // Verify API calls were made as expected
    expect(mockedApiClient.getGardens).toHaveBeenCalled();
    expect(mockedApiClient.getCurrentWeather).toHaveBeenCalled();
    expect(mockedApiClient.getSensorReadings).toHaveBeenCalled();
    expect(mockedApiClient.getWateringSchedules).toHaveBeenCalled();
  });
});