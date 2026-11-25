// End-to-end tests for Smart Garden Bot Dashboard
// These tests simulate real user interactions with the application

describe('Dashboard E2E Tests', () => {
  beforeEach(() => {
    // Mock Auth0 authentication
    cy.intercept('GET', '/api/auth/me', {
      statusCode: 200,
      body: {
        sub: 'auth0|test123',
        name: 'Test User',
        email: 'test@example.com',
        email_verified: true,
      },
    });

    // Mock API endpoints
    cy.intercept('GET', '/api/v1/gardens', {
      statusCode: 200,
      body: [
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
      ],
    }).as('getGardens');

    cy.intercept('GET', '/api/v1/weather/current*', {
      statusCode: 200,
      body: {
        location: 'San Francisco, CA',
        temperature: 22.5,
        humidity: 65.0,
        pressure: 1013.25,
        wind_speed: 5.2,
        wind_direction: 180,
        rainfall: 0.0,
        timestamp: '2024-01-01T12:00:00Z',
      },
    }).as('getCurrentWeather');

    cy.intercept('GET', '/api/v1/sensors/readings*', {
      statusCode: 200,
      body: [
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
        {
          id: '2',
          device_id: 'sensor-001',
          timestamp: '2024-01-01T11:00:00Z',
          temperature: 22.0,
          humidity: 70.0,
          soil_moisture: 42.0,
          soil_temperature: 20.5,
          light_intensity: 23000.0,
        },
      ],
    }).as('getSensorReadings');

    cy.intercept('GET', '/api/v1/watering/schedules*', {
      statusCode: 200,
      body: [
        {
          id: '1',
          zone_id: 'zone-1',
          schedule: '0 6 * * *',
          duration: 20,
          active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ],
    }).as('getWateringSchedules');

    // Visit the dashboard
    cy.visit('/dashboard');
  });

  it('should display the dashboard with user greeting', () => {
    cy.contains('Dashboard').should('be.visible');
    cy.contains('Welcome back, Test User!').should('be.visible');
    
    // Verify API calls were made
    cy.wait('@getGardens');
    cy.wait('@getCurrentWeather');
    cy.wait('@getSensorReadings');
    cy.wait('@getWateringSchedules');
  });

  it('should display garden statistics cards', () => {
    cy.wait('@getGardens');
    
    // Check for statistics cards
    cy.contains('Active Gardens').should('be.visible');
    cy.contains('Current Temperature').should('be.visible');
    cy.contains('Soil Moisture').should('be.visible');
    cy.contains('Next Watering').should('be.visible');
    
    // Verify garden count
    cy.get('[data-testid="garden-count"]').should('contain', '1');
    
    // Verify temperature display
    cy.get('[data-testid="current-temperature"]').should('contain', '23°C');
  });

  it('should navigate between dashboard tabs', () => {
    cy.wait('@getGardens');
    
    // Check initial tab is Overview
    cy.get('[role="tab"][aria-selected="true"]').should('contain', 'Overview');
    
    // Click on Sensor Data tab
    cy.get('[role="tab"]').contains('Sensor Data').click();
    cy.get('[role="tabpanel"]').should('be.visible');
    cy.get('[role="tab"][aria-selected="true"]').should('contain', 'Sensor Data');
    
    // Click on Watering tab
    cy.get('[role="tab"]').contains('Watering').click();
    cy.get('[role="tabpanel"]').should('be.visible');
    cy.get('[role="tab"][aria-selected="true"]').should('contain', 'Watering');
    
    // Click on Weather tab
    cy.get('[role="tab"]').contains('Weather').click();
    cy.get('[role="tabpanel"]').should('be.visible');
    cy.get('[role="tab"][aria-selected="true"]').should('contain', 'Weather');
    
    // Navigate back to Overview
    cy.get('[role="tab"]').contains('Overview').click();
    cy.get('[role="tab"][aria-selected="true"]').should('contain', 'Overview');
  });

  it('should display sensor data in the sensor tab', () => {
    cy.wait('@getSensorReadings');
    
    // Navigate to Sensor Data tab
    cy.get('[role="tab"]').contains('Sensor Data').click();
    
    // Check for sensor readings
    cy.get('[data-testid="sensor-reading"]').should('have.length.at.least', 1);
    
    // Verify sensor data values
    cy.contains('23.0°C').should('be.visible'); // Temperature
    cy.contains('68.0%').should('be.visible');  // Humidity
    cy.contains('45.0%').should('be.visible');  // Soil moisture
  });

  it('should display watering schedules in the watering tab', () => {
    cy.wait('@getWateringSchedules');
    
    // Navigate to Watering tab
    cy.get('[role="tab"]').contains('Watering').click();
    
    // Check for watering schedules
    cy.get('[data-testid="watering-schedule"]').should('have.length.at.least', 1);
    
    // Verify schedule details
    cy.contains('Daily at 6:00 AM').should('be.visible'); // Cron: 0 6 * * *
    cy.contains('20 minutes').should('be.visible');       // Duration
    cy.contains('Active').should('be.visible');           // Status
  });

  it('should display weather information in the weather tab', () => {
    cy.wait('@getCurrentWeather');
    
    // Navigate to Weather tab
    cy.get('[role="tab"]').contains('Weather').click();
    
    // Check for weather information
    cy.contains('San Francisco, CA').should('be.visible');
    cy.contains('22.5°C').should('be.visible');
    cy.contains('65.0%').should('be.visible'); // Humidity
    cy.contains('5.2 km/h').should('be.visible'); // Wind speed
  });

  it('should handle responsive design on mobile devices', () => {
    // Test mobile viewport
    cy.viewport('iphone-x');
    
    cy.wait('@getGardens');
    
    // Check that dashboard is still functional on mobile
    cy.contains('Dashboard').should('be.visible');
    cy.contains('Welcome back, Test User!').should('be.visible');
    
    // Verify tabs work on mobile
    cy.get('[role="tab"]').contains('Sensor Data').click();
    cy.get('[role="tabpanel"]').should('be.visible');
    
    // Check that cards stack properly on mobile
    cy.get('[data-testid="stats-card"]').should('be.visible');
  });

  it('should handle error states gracefully', () => {
    // Mock API error
    cy.intercept('GET', '/api/v1/gardens', {
      statusCode: 500,
      body: { error: 'Internal server error' },
    }).as('getGardensError');
    
    cy.visit('/dashboard');
    cy.wait('@getGardensError');
    
    // Should display error message
    cy.contains('Failed to load gardens').should('be.visible');
  });

  it('should refresh data when refresh button is clicked', () => {
    cy.wait('@getGardens');
    
    // Click refresh button (assuming it exists)
    cy.get('[data-testid="refresh-button"]').click();
    
    // Verify API calls are made again
    cy.wait('@getGardens');
    cy.wait('@getCurrentWeather');
    cy.wait('@getSensorReadings');
  });

  it('should handle manual watering trigger', () => {
    // Mock manual watering endpoint
    cy.intercept('POST', '/api/v1/watering/manual/*', {
      statusCode: 200,
      body: { success: true, message: 'Manual watering started' },
    }).as('manualWatering');
    
    cy.wait('@getGardens');
    
    // Navigate to Watering tab
    cy.get('[role="tab"]').contains('Watering').click();
    
    // Click manual watering button
    cy.get('[data-testid="manual-watering-button"]').first().click();
    
    // Confirm in modal
    cy.get('[data-testid="confirm-manual-watering"]').click();
    
    cy.wait('@manualWatering');
    
    // Should show success message
    cy.contains('Manual watering started').should('be.visible');
  });

  it('should display real-time updates', () => {
    cy.wait('@getGardens');
    
    // Simulate real-time sensor update
    cy.intercept('GET', '/api/v1/sensors/readings*', {
      statusCode: 200,
      body: [
        {
          id: '3',
          device_id: 'sensor-001',
          timestamp: new Date().toISOString(),
          temperature: 24.5, // Updated temperature
          humidity: 66.0,
          soil_moisture: 48.0, // Updated moisture
          soil_temperature: 21.5,
          light_intensity: 26000.0,
        },
      ],
    }).as('getUpdatedReadings');
    
    // Navigate to Sensor Data tab
    cy.get('[role="tab"]').contains('Sensor Data').click();
    
    // Trigger refresh (this would normally happen via WebSocket)
    cy.get('[data-testid="refresh-button"]').click();
    
    cy.wait('@getUpdatedReadings');
    
    // Check for updated values
    cy.contains('24.5°C').should('be.visible');
    cy.contains('48.0%').should('be.visible');
  });

  it('should handle keyboard navigation', () => {
    cy.wait('@getGardens');
    
    // Test tab navigation with keyboard
    cy.get('[role="tab"]').first().focus();
    cy.focused().should('contain', 'Overview');
    
    // Use arrow keys to navigate tabs
    cy.focused().type('{rightarrow}');
    cy.focused().should('contain', 'Sensor Data');
    
    cy.focused().type('{rightarrow}');
    cy.focused().should('contain', 'Watering');
    
    // Test Enter key to activate tab
    cy.focused().type('{enter}');
    cy.get('[role="tab"][aria-selected="true"]').should('contain', 'Watering');
  });

  it('should maintain accessibility standards', () => {
    cy.wait('@getGardens');
    
    // Check for proper heading hierarchy
    cy.get('h1').should('contain', 'Dashboard');
    
    // Check for proper ARIA labels
    cy.get('[role="tablist"]').should('exist');
    cy.get('[role="tab"]').should('have.attr', 'aria-selected');
    cy.get('[role="tabpanel"]').should('exist');
    
    // Check for proper focus management
    cy.get('[role="tab"]').first().focus();
    cy.focused().should('be.visible');
    
    // Check for alternative text on images (if any)
    cy.get('img').each(($img) => {
      cy.wrap($img).should('have.attr', 'alt');
    });
    
    // Check color contrast (this would require additional tools in a real scenario)
    // cy.checkA11y(); // If using cypress-axe plugin
  });

  it('should handle garden selection', () => {
    // Mock multiple gardens
    cy.intercept('GET', '/api/v1/gardens', {
      statusCode: 200,
      body: [
        {
          id: '1',
          name: 'Backyard Garden',
          location: 'San Francisco, CA',
          zones: [],
        },
        {
          id: '2',
          name: 'Front Yard Garden',
          location: 'San Francisco, CA',
          zones: [],
        },
      ],
    }).as('getMultipleGardens');
    
    cy.visit('/dashboard');
    cy.wait('@getMultipleGardens');
    
    // Check if garden selector is available
    cy.get('[data-testid="garden-selector"]').should('be.visible');
    
    // Select different garden
    cy.get('[data-testid="garden-selector"]').click();
    cy.contains('Front Yard Garden').click();
    
    // Verify the selection changed
    cy.get('[data-testid="selected-garden"]').should('contain', 'Front Yard Garden');
  });

  it('should persist user preferences', () => {
    cy.wait('@getGardens');
    
    // Change to Sensor Data tab
    cy.get('[role="tab"]').contains('Sensor Data').click();
    
    // Reload the page
    cy.reload();
    
    cy.wait('@getGardens');
    
    // Should remember the selected tab (if implemented)
    // This would depend on your state persistence implementation
    cy.get('[role="tab"][aria-selected="true"]').should('exist');
  });
});

// Performance tests
describe('Dashboard Performance', () => {
  it('should load within acceptable time limits', () => {
    const start = Date.now();
    
    cy.visit('/dashboard');
    
    cy.wait('@getGardens').then(() => {
      const loadTime = Date.now() - start;
      expect(loadTime).to.be.lessThan(3000); // Should load within 3 seconds
    });
  });

  it('should handle large datasets efficiently', () => {
    // Mock large sensor dataset
    const largeSensorData = Array.from({ length: 1000 }, (_, i) => ({
      id: `sensor-${i}`,
      device_id: 'sensor-001',
      timestamp: new Date(Date.now() - i * 60000).toISOString(),
      temperature: 20 + Math.random() * 10,
      humidity: 50 + Math.random() * 30,
      soil_moisture: 30 + Math.random() * 40,
    }));
    
    cy.intercept('GET', '/api/v1/sensors/readings*', {
      statusCode: 200,
      body: largeSensorData,
    }).as('getLargeSensorData');
    
    cy.visit('/dashboard');
    cy.wait('@getLargeSensorData');
    
    // Navigate to Sensor Data tab
    cy.get('[role="tab"]').contains('Sensor Data').click();
    
    // Should still be responsive with large dataset
    cy.get('[data-testid="sensor-chart"]').should('be.visible', { timeout: 5000 });
  });
});