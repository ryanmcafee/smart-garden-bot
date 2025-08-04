# Smart Garden Bot - Comprehensive Testing & Quality Assurance Strategy

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Testing Strategy Overview](#testing-strategy-overview)
3. [Unit Testing Strategy](#unit-testing-strategy)
4. [Integration Testing Strategy](#integration-testing-strategy)
5. [End-to-End Testing Strategy](#end-to-end-testing-strategy)
6. [Performance Testing Strategy](#performance-testing-strategy)
7. [Security Testing Strategy](#security-testing-strategy)
8. [IoT Testing Strategy](#iot-testing-strategy)
9. [Kubernetes Testing Strategy](#kubernetes-testing-strategy)
10. [Quality Framework](#quality-framework)
11. [CI/CD Testing Integration](#cicd-testing-integration)
12. [Test Data Management](#test-data-management)
13. [Implementation Roadmap](#implementation-roadmap)
14. [Appendices](#appendices)

## Executive Summary

This document outlines a comprehensive testing and quality assurance strategy for the Smart Garden Bot platform, a cloud-native SaaS solution built with Go, TypeScript/Next.js, and Kubernetes. The strategy addresses the unique challenges of testing IoT-integrated systems, microservices architecture, and real-time data processing while maintaining enterprise-grade quality standards.

**Key Objectives:**
- Ensure 99.9% system availability with comprehensive testing coverage
- Implement automated testing pipelines with quality gates
- Establish robust IoT device simulation and testing frameworks
- Provide end-to-end validation of weather-based irrigation decisions
- Maintain security and performance standards throughout the development lifecycle

## Testing Strategy Overview

### Testing Pyramid Architecture

```
                    ┌─────────────────┐
                    │   E2E Tests     │ ← 5% (Critical user journeys)
                    │   (Playwright)  │
                ┌───┴─────────────────┴───┐
                │  Integration Tests      │ ← 25% (API, DB, External services)
                │  (Testcontainers)       │
            ┌───┴─────────────────────────┴───┐
            │      Unit Tests                 │ ← 70% (Business logic, utilities)
            │   (Go: testify, TS: Jest)       │
            └─────────────────────────────────┘
```

### Testing Scope by Component

| Component | Unit | Integration | E2E | Performance | Security |
|-----------|------|-------------|-----|-------------|----------|
| **Go API Server** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Next.js Web App** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Kubernetes Operator** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Database Layer** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Weather APIs** | - | ✅ | ✅ | ✅ | ✅ |
| **IoT Devices** | ✅ | ✅ | ✅ | ✅ | ✅ |

## Unit Testing Strategy

### Go API Server Testing

**Framework:** Go standard `testing` package + `testify` for assertions

**Structure:**
```go
// Directory structure
api/
├── internal/
│   ├── handlers/
│   │   ├── user_handler.go
│   │   └── user_handler_test.go
│   ├── services/
│   │   ├── weather_service.go
│   │   └── weather_service_test.go
│   └── repositories/
│       ├── user_repository.go
│       └── user_repository_test.go
└── cmd/
    └── server/
        ├── main.go
        └── main_test.go
```

**Key Testing Patterns:**

1. **Repository Layer Testing:**
```go
func TestUserRepository_Create(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    repo := repositories.NewUserRepository(db)
    
    // Test cases
    tests := []struct {
        name    string
        user    *models.User
        wantErr bool
    }{
        {
            name: "valid user creation",
            user: &models.User{
                Email: "test@example.com",
                Name:  "Test User",
            },
            wantErr: false,
        },
        {
            name: "duplicate email",
            user: &models.User{
                Email: "duplicate@example.com",
                Name:  "Duplicate User",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.Create(context.Background(), tt.user)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotZero(t, tt.user.ID)
            }
        })
    }
}
```

2. **Service Layer Testing with Mocks:**
```go
func TestWeatherService_GetForecast(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockClient := mocks.NewMockWeatherClient(ctrl)
    service := services.NewWeatherService(mockClient)
    
    mockClient.EXPECT().
        GetWeatherData("12345").
        Return(&weatherapi.Response{
            Temperature: 25.0,
            Humidity:    60.0,
        }, nil)
    
    forecast, err := service.GetForecast("12345")
    assert.NoError(t, err)
    assert.Equal(t, 25.0, forecast.Temperature)
}
```

**Coverage Requirements:**
- Minimum: 80% statement coverage
- Target: 90% statement coverage
- Critical paths: 100% coverage (authentication, billing, watering logic)

**Testing Tools:**
- **gomock**: Mock generation for interfaces
- **testify**: Rich assertion library
- **dockertest**: Database testing with real PostgreSQL
- **httptest**: HTTP handler testing

### TypeScript/Next.js Testing

**Framework:** Jest + Testing Library + MSW (Mock Service Worker)

**Structure:**
```
web-app/
├── src/
│   ├── components/
│   │   ├── Dashboard/
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Dashboard.test.tsx
│   │   │   └── Dashboard.stories.tsx
│   ├── hooks/
│   │   ├── useWeatherData.ts
│   │   └── useWeatherData.test.ts
│   ├── services/
│   │   ├── api.ts
│   │   └── api.test.ts
│   └── utils/
│       ├── dateUtils.ts
│       └── dateUtils.test.ts
├── __tests__/
│   ├── __mocks__/
│   └── setup.ts
└── jest.config.js
```

**Key Testing Patterns:**

1. **Component Testing:**
```typescript
import { render, screen, waitFor } from '@testing-library/react';
import { Dashboard } from './Dashboard';
import { server } from '../../../__tests__/__mocks__/server';

describe('Dashboard', () => {
  beforeAll(() => server.listen());
  afterEach(() => server.resetHandlers());
  afterAll(() => server.close());

  it('displays garden overview correctly', async () => {
    render(<Dashboard gardenId="garden-123" />);
    
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    
    await waitFor(() => {
      expect(screen.getByText('Backyard Garden')).toBeInTheDocument();
    });
    
    expect(screen.getByText('3 active zones')).toBeInTheDocument();
    expect(screen.getByText('Last watered: 2 hours ago')).toBeInTheDocument();
  });
});
```

2. **Custom Hook Testing:**
```typescript
import { renderHook, waitFor } from '@testing-library/react';
import { useWeatherData } from './useWeatherData';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useWeatherData', () => {
  it('fetches weather data successfully', async () => {
    const { result } = renderHook(() => useWeatherData('12345'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toMatchObject({
      temperature: expect.any(Number),
      humidity: expect.any(Number),
      forecast: expect.any(Array),
    });
  });
});
```

**Coverage Requirements:**
- Minimum: 80% statement coverage
- Target: 90% statement coverage
- UI components: Focus on user interactions and edge cases

**Testing Tools:**
- **Jest**: Test runner and assertion library
- **@testing-library/react**: Component testing utilities
- **@testing-library/user-event**: User interaction simulation
- **MSW**: API mocking
- **Storybook**: Component documentation and visual testing

### Kubernetes Operator Testing

**Framework:** Go testing + controller-runtime testing utilities

**Structure:**
```go
operator/
├── controllers/
│   ├── gardencontroller_controller.go
│   ├── gardencontroller_controller_test.go
│   └── suite_test.go
├── internal/
│   ├── weather/
│   │   ├── client.go
│   │   └── client_test.go
│   └── irrigation/
│       ├── scheduler.go
│       └── scheduler_test.go
└── api/v1/
    ├── gardencontroller_types.go
    └── gardencontroller_types_test.go
```

**Key Testing Patterns:**

1. **Controller Testing:**
```go
func TestGardenControllerReconciler_Reconcile(t *testing.T) {
    scheme := runtime.NewScheme()
    utilruntime.Must(gardencontrollerv1.AddToScheme(scheme))
    
    client := fake.NewClientBuilder().WithScheme(scheme).Build()
    reconciler := &GardenControllerReconciler{
        Client: client,
        Scheme: scheme,
    }
    
    ctx := context.Background()
    gardenController := &gardencontrollerv1.GardenController{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-garden",
            Namespace: "default",
        },
        Spec: gardencontrollerv1.GardenControllerSpec{
            Zones: []gardencontrollerv1.Zone{
                {
                    Name:    "vegetables",
                    Schedule: "0 6 * * *",
                },
            },
        },
    }
    
    err := client.Create(ctx, gardenController)
    require.NoError(t, err)
    
    req := reconcile.Request{
        NamespacedName: types.NamespacedName{
            Name:      "test-garden",
            Namespace: "default",
        },
    }
    
    result, err := reconciler.Reconcile(ctx, req)
    assert.NoError(t, err)
    assert.False(t, result.Requeue)
}
```

**Coverage Requirements:**
- Minimum: 85% statement coverage
- Critical controller logic: 100% coverage

## Integration Testing Strategy

### API Integration Testing

**Framework:** Go testing + Testcontainers + Docker

**Architecture:**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Test Suite    │───►│  API Server     │───►│  PostgreSQL     │
│                 │    │  (Real)         │    │  (Container)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Mock Services  │
                       │  (Weather API,  │
                       │   Auth0, etc.)  │
                       └─────────────────┘
```

**Implementation Example:**
```go
func TestIntegration_UserGardenFlow(t *testing.T) {
    // Setup test environment
    ctx := context.Background()
    container := setupPostgresContainer(t, ctx)
    defer container.Terminate(ctx)
    
    // Setup API server
    config := &config.Config{
        DatabaseURL: container.ConnectionString(),
        JWTSecret:   "test-secret",
    }
    server := setupApi(t, config)
    defer server.Close()
    
    // Setup mock external services
    weatherMock := setupWeatherAPIMock()
    defer weatherMock.Close()
    
    client := &http.Client{Timeout: 5 * time.Second}
    baseURL := server.URL
    
    // Test complete user journey
    t.Run("complete user garden setup flow", func(t *testing.T) {
        // 1. User registration
        userToken := registerUser(t, client, baseURL, "test@example.com")
        
        // 2. Create garden
        gardenID := createGarden(t, client, baseURL, userToken, &Garden{
            Name:     "Test Garden",
            Location: "12345",
            Timezone: "America/New_York",
        })
        
        // 3. Add zones
        zoneID := createZone(t, client, baseURL, userToken, gardenID, &Zone{
            Name:      "Vegetables",
            PlantType: "tomatoes",
            Area:      100.0,
        })
        
        // 4. Setup watering schedule
        scheduleID := createWateringSchedule(t, client, baseURL, userToken, zoneID, &Schedule{
            CronExpression: "0 6 * * *",
            Duration:       30,
        })
        
        // 5. Verify weather integration
        forecast := getWeatherForecast(t, client, baseURL, userToken, gardenID)
        assert.NotEmpty(t, forecast.Days)
        
        // 6. Test watering decision logic
        decision := getWateringDecision(t, client, baseURL, userToken, zoneID)
        assert.NotNil(t, decision)
    })
}
```

### Database Integration Testing

**Strategy:** Use Testcontainers for real PostgreSQL instances

```go
func setupPostgresContainer(t *testing.T, ctx context.Context) *postgres.PostgreSQLContainer {
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).WithStartupTimeout(30*time.Second)),
    )
    require.NoError(t, err)
    return container
}
```

### External Service Integration Testing

**Weather API Testing:**
```go
type WeatherAPITestSuite struct {
    suite.Suite
    mockServer *httptest.Server
    client     *weather.Client
}

func (s *WeatherAPITestSuite) SetupTest() {
    s.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/current":
            response := &weather.CurrentResponse{
                Temperature: 25.5,
                Humidity:    65,
                Conditions:  "sunny",
            }
            json.NewEncoder(w).Encode(response)
        case "/forecast":
            response := &weather.ForecastResponse{
                Days: []weather.DayForecast{
                    {Date: "2024-01-15", High: 28, Low: 18, Precipitation: 0},
                    {Date: "2024-01-16", High: 30, Low: 20, Precipitation: 0.2},
                },
            }
            json.NewEncoder(w).Encode(response)
        }
    }))
    
    config := &weather.Config{
        BaseURL: s.mockServer.URL,
        APIKey:  "test-key",
    }
    s.client = weather.NewClient(config)
}
```

## End-to-End Testing Strategy

### Framework: Playwright

**Architecture:**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Playwright    │───►│   Next.js App   │───►│   API Server    │
│   Test Runner   │    │   (Real)        │    │   (Real)        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Browser       │    │   PostgreSQL    │
                       │   (Chromium)    │    │   (Container)   │
                       └─────────────────┘    └─────────────────┘
```

**Test Structure:**
```
e2e/
├── tests/
│   ├── auth/
│   │   ├── login.spec.ts
│   │   └── registration.spec.ts
│   ├── garden/
│   │   ├── garden-setup.spec.ts
│   │   ├── zone-management.spec.ts
│   │   └── watering-schedule.spec.ts
│   ├── dashboard/
│   │   ├── overview.spec.ts
│   │   └── sensor-data.spec.ts
│   └── billing/
│       └── subscription.spec.ts
├── fixtures/
│   ├── users.json
│   ├── gardens.json
│   └── sensor-data.json
├── page-objects/
│   ├── LoginPage.ts
│   ├── DashboardPage.ts
│   └── GardenSetupPage.ts
└── playwright.config.ts
```

**Critical User Journeys:**

1. **New User Onboarding:**
```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from '../page-objects/LoginPage';
import { DashboardPage } from '../page-objects/DashboardPage';
import { GardenSetupPage } from '../page-objects/GardenSetupPage';

test.describe('New User Onboarding', () => {
  test('complete garden setup flow', async ({ page }) => {
    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);
    const gardenSetupPage = new GardenSetupPage(page);

    // 1. User registration
    await page.goto('/');
    await loginPage.clickSignUp();
    await loginPage.fillRegistrationForm({
      email: 'newuser@example.com',
      password: 'SecurePass123!',
      confirmPassword: 'SecurePass123!',
    });
    await loginPage.submitRegistration();

    // 2. Email verification (mock)
    await loginPage.verifyEmail('verification-code-123');

    // 3. Initial garden setup
    await expect(page.locator('h1')).toContainText('Welcome to Smart Garden Bot');
    await page.click('button:has-text("Set Up Your First Garden")');

    // 4. Garden configuration
    await gardenSetupPage.fillGardenDetails({
      name: 'My Backyard Garden',
      location: '90210',
      timezone: 'America/Los_Angeles',
    });

    // 5. Add zones
    await gardenSetupPage.addZone({
      name: 'Vegetable Patch',
      plantType: 'Mixed Vegetables',
      area: 200,
    });

    // 6. Configure watering schedule
    await gardenSetupPage.setWateringSchedule({
      time: '06:00',
      duration: 30,
      frequency: 'daily',
    });

    // 7. Verify dashboard
    await expect(dashboardPage.getGardenName()).toContainText('My Backyard Garden');
    await expect(dashboardPage.getZoneCount()).toContainText('1 zone');
    await expect(dashboardPage.getNextWatering()).toContainText('Tomorrow at 6:00 AM');
  });
});
```

2. **Weather-Based Watering Decision:**
```typescript
test('weather integration affects watering decisions', async ({ page }) => {
  // Setup existing garden
  await setupTestGarden(page);
  
  // Mock weather API responses
  await page.route('**/api/weather/**', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        current: { temperature: 85, humidity: 40 },
        forecast: [
          { date: '2024-01-15', precipitation: 0, high: 88 },
          { date: '2024-01-16', precipitation: 0.8, high: 75 },
        ],
      }),
    });
  });

  // Navigate to watering schedule
  await page.goto('/dashboard/watering');
  
  // Verify watering recommendations
  await expect(page.locator('[data-testid="watering-recommendation"]'))
    .toContainText('Watering recommended today - no rain expected');
  
  await expect(page.locator('[data-testid="next-watering-skip"]'))
    .toContainText('Skip tomorrow - rain expected');
});
```

**E2E Testing Coverage:**
- Authentication flows (login, registration, password reset)
- Garden setup and configuration
- Zone management and watering schedules
- Sensor data visualization
- Weather integration and watering decisions
- Billing and subscription management
- Mobile responsive behavior
- Cross-browser compatibility (Chrome, Firefox, Safari)

## Performance Testing Strategy

### Load Testing Framework: K6

**Test Types:**
1. **Load Testing**: Normal expected traffic
2. **Stress Testing**: Peak traffic scenarios
3. **Spike Testing**: Sudden traffic increases
4. **Endurance Testing**: Extended periods
5. **Volume Testing**: Large amounts of data

**Architecture:**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│      K6         │───►│   API Server    │───►│   PostgreSQL    │
│  Load Generator │    │   (Target)      │    │   (Database)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Metrics       │    │   Monitoring    │    │   Resource      │
│  Collection     │    │   Dashboard     │    │   Monitoring    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**Key Test Scenarios:**

1. **API Load Testing:**
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

export let errorRate = new Rate('errors');

export let options = {
  stages: [
    { duration: '2m', target: 100 }, // Ramp up
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 200 }, // Ramp up to 200
    { duration: '5m', target: 200 }, // Stay at 200
    { duration: '2m', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    errors: ['rate<0.1'],             // Error rate under 10%
  },
};

export default function() {
  // Test authentication
  let loginResponse = http.post('http://api.smartgardenbot.com/auth/login', {
    email: 'loadtest@example.com',
    password: 'testpass123',
  });
  
  check(loginResponse, {
    'login status is 200': (r) => r.status === 200,
  }) || errorRate.add(1);
  
  let token = loginResponse.json('token');
  let headers = { Authorization: `Bearer ${token}` };
  
  // Test garden data retrieval
  let gardenResponse = http.get('http://api.smartgardenbot.com/api/v1/gardens', {
    headers: headers,
  });
  
  check(gardenResponse, {
    'garden status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  }) || errorRate.add(1);
  
  // Test sensor data endpoint
  let sensorResponse = http.get(
    'http://api.smartgardenbot.com/api/v1/sensors/garden-123/data?range=24h',
    { headers: headers }
  );
  
  check(sensorResponse, {
    'sensor data status is 200': (r) => r.status === 200,
    'sensor data not empty': (r) => r.json('data').length > 0,
  }) || errorRate.add(1);
  
  sleep(1);
}
```

2. **Database Performance Testing:**
```javascript
export function databaseStressTest() {
  let scenarios = {
    sensor_data_writes: {
      executor: 'constant-arrival-rate',
      rate: 1000, // 1000 requests per second
      timeUnit: '1s',
      duration: '10m',
      preAllocatedVUs: 50,
      exec: 'sensorDataWrite',
    },
    dashboard_reads: {
      executor: 'constant-vus',
      vus: 100,
      duration: '10m',
      exec: 'dashboardRead',
    },
  };
}
```

**Performance Targets:**
- **API Response Time**: 95th percentile < 200ms
- **Database Query Time**: 99th percentile < 100ms
- **Concurrent Users**: Support 1000+ concurrent users
- **Throughput**: Handle 10,000+ requests per minute
- **Resource Utilization**: CPU < 70%, Memory < 80%

### Frontend Performance Testing

**Tools:** Lighthouse CI, WebPageTest, Playwright

**Key Metrics:**
- **First Contentful Paint (FCP)**: < 1.5s
- **Largest Contentful Paint (LCP)**: < 2.5s
- **First Input Delay (FID)**: < 100ms
- **Cumulative Layout Shift (CLS)**: < 0.1

**Performance Test Suite:**
```typescript
import { test, expect } from '@playwright/test';

test.describe('Performance Tests', () => {
  test('dashboard loads within performance budget', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Measure performance metrics
    const performanceMetrics = await page.evaluate(() => {
      return JSON.parse(JSON.stringify(performance.getEntriesByType('navigation')[0]));
    });
    
    // Assert performance requirements
    expect(performanceMetrics.domContentLoadedEventEnd).toBeLessThan(1500);
    expect(performanceMetrics.loadEventEnd).toBeLessThan(3000);
  });

  test('sensor data visualization renders efficiently', async ({ page }) => {
    await page.goto('/dashboard/sensors');
    
    // Wait for chart to load
    await page.waitForSelector('[data-testid="sensor-chart"]');
    
    // Measure rendering performance
    const renderTime = await page.evaluate(() => {
      const startTime = performance.now();
      // Trigger chart re-render
      window.dispatchEvent(new Event('resize'));
      return performance.now() - startTime;
    });
    
    expect(renderTime).toBeLessThan(100); // Chart renders in <100ms
  });
});
```

## Security Testing Strategy

### Static Application Security Testing (SAST)

**Tools:**
- **Go**: `gosec`, `staticcheck`, `nancy` (dependency scanning)
- **TypeScript**: `eslint-plugin-security`, `audit-ci`
- **Docker**: `trivy`, `docker-bench-security`
- **Kubernetes**: `kube-score`, `polaris`

**Implementation:**
```yaml
# .github/workflows/security-scan.yml
name: Security Scan
on: [push, pull_request]

jobs:
  sast:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Gosec Security Scanner
        uses: securecodewarrior/github-action-gosec@master
        with:
          args: '-fmt sarif -out gosec.sarif ./...'
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          format: 'sarif'
          output: 'trivy.sarif'
      
      - name: Upload SARIF files
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: '.'
```

### Dynamic Application Security Testing (DAST)

**Tools:** OWASP ZAP, Burp Suite

**Test Coverage:**
- Authentication bypass attempts
- SQL injection testing
- Cross-site scripting (XSS) testing
- CSRF protection validation
- API security testing
- Input validation testing

**ZAP Integration:**
```yaml
security-test:
  runs-on: ubuntu-latest
  steps:
    - name: Checkout
      uses: actions/checkout@v3
    
    - name: Start application
      run: |
        docker-compose up -d
        sleep 30
    
    - name: ZAP Baseline Scan
      uses: zaproxy/action-baseline@v0.7.0
      with:
        target: 'http://localhost:3000'
        rules_file_name: '.zap/rules.tsv'
        cmd_options: '-a'
    
    - name: ZAP Full Scan
      uses: zaproxy/action-full-scan@v0.4.0
      with:
        target: 'http://localhost:3000'
        rules_file_name: '.zap/rules.tsv'
```

### API Security Testing

**Framework:** Custom Go testing + OWASP API Security Top 10

```go
func TestAPISecurityBaseline(t *testing.T) {
    client := &http.Client{Timeout: 5 * time.Second}
    baseURL := "http://localhost:8080"
    
    tests := []struct {
        name     string
        method   string
        endpoint string
        headers  map[string]string
        body     string
        wantCode int
    }{
        {
            name:     "unauthorized access blocked",
            method:   "GET",
            endpoint: "/api/v1/gardens",
            wantCode: 401,
        },
        {
            name:     "malformed JWT rejected",
            method:   "GET",
            endpoint: "/api/v1/gardens",
            headers:  map[string]string{"Authorization": "Bearer invalid.token.here"},
            wantCode: 401,
        },
        {
            name:     "SQL injection attempt blocked",
            method:   "GET",
            endpoint: "/api/v1/gardens?id=1'; DROP TABLE users; --",
            headers:  validAuthHeaders(),
            wantCode: 400,
        },
        {
            name:     "XSS payload sanitized",
            method:   "POST",
            endpoint: "/api/v1/gardens",
            headers:  validAuthHeaders(),
            body:     `{"name": "<script>alert('xss')</script>"}`,
            wantCode: 400,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req, err := http.NewRequest(tt.method, baseURL+tt.endpoint, strings.NewReader(tt.body))
            require.NoError(t, err)
            
            for k, v := range tt.headers {
                req.Header.Set(k, v)
            }
            
            resp, err := client.Do(req)
            require.NoError(t, err)
            defer resp.Body.Close()
            
            assert.Equal(t, tt.wantCode, resp.StatusCode)
        })
    }
}
```

### Penetration Testing

**Scope:**
- Web application security
- API endpoint security  
- Authentication mechanisms
- Authorization controls
- Input validation
- Session management
- Database security
- Infrastructure security

**Schedule:**
- Quarterly automated scans
- Annual third-party penetration testing
- Security assessment before major releases

## IoT Testing Strategy

### Device Simulation Framework

**Architecture:**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Test Suite    │───►│ Device Simulator│───►│  API Server     │
│                 │    │   (MockTCP)     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Sensor Data    │
                       │   Generator     │
                       └─────────────────┘
```

**Device Simulator Implementation:**
```go
type EcowittSimulator struct {
    server   *httptest.Server
    stations []WeatherStation
    sensors  []Sensor
}

type WeatherStation struct {
    ID          string
    Location    string
    LastUpdate  time.Time
    SensorData  map[string]interface{}
}

func NewEcowittSimulator() *EcowittSimulator {
    sim := &EcowittSimulator{
        stations: make([]WeatherStation, 0),
        sensors:  make([]Sensor, 0),
    }
    
    mux := http.NewServeMux()
    mux.HandleFunc("/data/report", sim.handleDataReport)
    mux.HandleFunc("/device/status", sim.handleDeviceStatus)
    mux.HandleFunc("/control/valve", sim.handleValveControl)
    
    sim.server = httptest.NewServer(mux)
    return sim
}

func (s *EcowittSimulator) AddWeatherStation(id, location string) {
    station := WeatherStation{
        ID:       id,
        Location: location,
        SensorData: map[string]interface{}{
            "temperature":    20.0 + rand.Float64()*20,
            "humidity":       40.0 + rand.Float64()*40,
            "soil_moisture":  30.0 + rand.Float64()*40,
            "rainfall":       0.0,
            "wind_speed":     0.0 + rand.Float64()*15,
            "solar_radiation": 0.0 + rand.Float64()*1000,
        },
        LastUpdate: time.Now(),
    }
    s.stations = append(s.stations, station)
}

func (s *EcowittSimulator) SimulateSensorReading(stationID string, sensorType string, value float64) {
    for i, station := range s.stations {
        if station.ID == stationID {
            s.stations[i].SensorData[sensorType] = value
            s.stations[i].LastUpdate = time.Now()
            
            // Send data to API
            s.sendSensorData(station)
            break
        }
    }
}

func (s *EcowittSimulator) handleDataReport(w http.ResponseWriter, r *http.Request) {
    // Parse incoming sensor data report
    // Validate data format
    // Store in simulator state
    // Respond with acknowledgment
}
```

### Sensor Data Validation Testing

```go
func TestSensorDataValidation(t *testing.T) {
    simulator := NewEcowittSimulator()
    defer simulator.Close()
    
    // Setup test garden with sensors
    testGarden := setupTestGarden(t)
    simulator.AddWeatherStation(testGarden.StationID, testGarden.Location)
    
    tests := []struct {
        name        string
        sensorType  string
        value       float64
        expectValid bool
    }{
        {"valid temperature", "temperature", 25.5, true},
        {"temperature too high", "temperature", 60.0, false},
        {"temperature too low", "temperature", -30.0, false},
        {"valid humidity", "humidity", 65.0, true},
        {"humidity out of range", "humidity", 150.0, false},
        {"valid soil moisture", "soil_moisture", 45.0, true},
        {"negative soil moisture", "soil_moisture", -5.0, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Simulate sensor reading
            simulator.SimulateSensorReading(testGarden.StationID, tt.sensorType, tt.value)
            
            // Wait for API processing
            time.Sleep(100 * time.Millisecond)
            
            // Verify data was processed correctly
            sensorData := getSensorData(t, testGarden.ID, tt.sensorType)
            
            if tt.expectValid {
                assert.Equal(t, tt.value, sensorData.Value)
                assert.True(t, sensorData.Valid)
            } else {
                assert.False(t, sensorData.Valid)
                assert.NotEmpty(t, sensorData.ValidationErrors)
            }
        })
    }
}
```

### Communication Protocol Testing

```go
func TestMQTTCommunication(t *testing.T) {
    // Setup MQTT broker
    broker := setupTestMQTTBroker(t)
    defer broker.Close()
    
    // Setup device simulator
    device := NewMQTTDevice("test-station-001", broker.URL())
    defer device.Close()
    
    // Setup API server with MQTT client
    api := setupApiWithMQTT(t, broker.URL())
    defer api.Close()
    
    t.Run("device publishes sensor data", func(t *testing.T) {
        sensorData := SensorData{
            StationID:   "test-station-001",
            Temperature: 24.5,
            Humidity:    62.0,
            Timestamp:   time.Now(),
        }
        
        err := device.PublishSensorData(sensorData)
        assert.NoError(t, err)
        
        // Verify API received and processed data
        time.Sleep(500 * time.Millisecond)
        
        received := getReceivedSensorData(t, api, "test-station-001")
        assert.Equal(t, sensorData.Temperature, received.Temperature)
        assert.Equal(t, sensorData.Humidity, received.Humidity)
    })
    
    t.Run("API sends watering commands", func(t *testing.T) {
        command := WateringCommand{
            StationID: "test-station-001",
            ZoneID:    "zone-001",
            Action:    "start",
            Duration:  1800, // 30 minutes
        }
        
        err := api.SendWateringCommand(command)
        assert.NoError(t, err)
        
        // Verify device received command
        receivedCommand := device.GetLastCommand()
        assert.Equal(t, command.Action, receivedCommand.Action)
        assert.Equal(t, command.Duration, receivedCommand.Duration)
    })
}
```

### Fault Tolerance Testing

```go
func TestIoTFaultTolerance(t *testing.T) {
    t.Run("network connectivity issues", func(t *testing.T) {
        // Simulate network interruption
        // Verify system handles gracefully
        // Test reconnection logic
    })
    
    t.Run("sensor malfunction", func(t *testing.T) {
        // Simulate sensor sending invalid data
        // Verify data validation and error handling
        // Test fallback mechanisms
    })
    
    t.Run("device power cycle", func(t *testing.T) {
        // Simulate device restart
        // Verify state recovery
        // Test message queue handling
    })
}
```

## Kubernetes Testing Strategy

### Operator Testing Framework

**Structure:**
```go
// operator/test/e2e/
├── e2e_test.go
├── utils/
│   ├── cluster.go
│   ├── resources.go
│   └── wait.go
├── testdata/
│   ├── garden-basic.yaml
│   ├── garden-multizone.yaml
│   └── weather-config.yaml
└── suite_test.go
```

**Controller Testing:**
```go
func TestGardenControllerE2E(t *testing.T) {
    ctx := context.Background()
    
    // Setup test cluster (kind/k3s)
    cluster := setupTestCluster(t)
    defer cluster.Cleanup()
    
    // Install operator
    err := installOperator(ctx, cluster.Client())
    require.NoError(t, err)
    
    t.Run("basic garden controller lifecycle", func(t *testing.T) {
        // Create GardenController resource
        garden := &gardencontrollerv1.GardenController{
            ObjectMeta: metav1.ObjectMeta{
                Name:      "test-garden",
                Namespace: "default",
            },
            Spec: gardencontrollerv1.GardenControllerSpec{
                Location: "90210",
                Zones: []gardencontrollerv1.Zone{
                    {
                        Name:     "vegetables",
                        Schedule: "0 6 * * *",
                        Duration: metav1.Duration{Duration: 30 * time.Minute},
                    },
                },
                WeatherProvider: gardencontrollerv1.WeatherProvider{
                    Type: "openweathermap",
                    Config: map[string]string{
                        "api_key": "test-key",
                    },
                },
            },
        }
        
        err := cluster.Client().Create(ctx, garden)
        require.NoError(t, err)
        
        // Wait for controller to reconcile
        err = waitForGardenControllerReady(ctx, cluster.Client(), "test-garden", "default")
        assert.NoError(t, err)
        
        // Verify status updates
        updatedGarden := &gardencontrollerv1.GardenController{}
        err = cluster.Client().Get(ctx, types.NamespacedName{
            Name: "test-garden", 
            Namespace: "default",
        }, updatedGarden)
        require.NoError(t, err)
        
        assert.Equal(t, "Ready", updatedGarden.Status.Phase)
        assert.NotEmpty(t, updatedGarden.Status.Conditions)
    })
    
    t.Run("weather-based watering decision", func(t *testing.T) {
        // Mock weather API
        weatherMock := setupWeatherMock(t)
        defer weatherMock.Close()
        
        // Update garden controller with mock endpoint
        garden := getGardenController(t, cluster.Client(), "test-garden")
        garden.Spec.WeatherProvider.Config["base_url"] = weatherMock.URL
        
        err := cluster.Client().Update(ctx, garden)
        require.NoError(t, err)
        
        // Trigger reconciliation
        triggerReconciliation(t, cluster.Client(), "test-garden")
        
        // Verify watering decision based on weather
        decision := getWateringDecision(t, cluster.Client(), "test-garden", "vegetables")
        assert.NotNil(t, decision)
        
        // Verify CronJob created/updated
        cronJob := getCronJob(t, cluster.Client(), "test-garden-vegetables")
        assert.NotNil(t, cronJob)
    })
}
```

### Deployment Validation Testing

```go
func TestHelmDeployment(t *testing.T) {
    // Setup test cluster
    cluster := setupTestCluster(t)
    defer cluster.Cleanup()
    
    // Deploy via Helm
    helmChart := "./charts/smart-garden-bot"
    values := map[string]interface{}{
        "image": map[string]interface{}{
            "tag": "test",
        },
        "postgresql": map[string]interface{}{
            "enabled": true,
        },
    }
    
    release, err := deployHelmChart(cluster, helmChart, "test-release", values)
    require.NoError(t, err)
    defer release.Uninstall()
    
    // Wait for all resources to be ready
    err = waitForDeploymentReady(cluster.Client(), "test-release-api", 5*time.Minute)
    assert.NoError(t, err)
    
    err = waitForDeploymentReady(cluster.Client(), "test-release-web", 5*time.Minute)
    assert.NoError(t, err)
    
    // Verify service mesh configuration
    virtualService := getVirtualService(cluster.Client(), "test-release")
    assert.NotNil(t, virtualService)
    
    // Test ingress connectivity
    ingressIP := getIngressIP(cluster.Client(), "test-release")
    resp, err := http.Get(fmt.Sprintf("http://%s/health", ingressIP))
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### Resource Constraints Testing

```go
func TestResourceLimits(t *testing.T) {
    cluster := setupTestCluster(t)
    defer cluster.Cleanup()
    
    // Deploy with resource constraints
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "resource-test",
            Namespace: "default",
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: int32Ptr(1),
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{"app": "resource-test"},
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{"app": "resource-test"},
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "api",
                            Image: "smartgardenbot/api:test",
                            Resources: corev1.ResourceRequirements{
                                Requests: corev1.ResourceList{
                                    corev1.ResourceMemory: resource.MustParse("128Mi"),
                                    corev1.ResourceCPU:    resource.MustParse("100m"),
                                },
                                Limits: corev1.ResourceList{
                                    corev1.ResourceMemory: resource.MustParse("256Mi"),
                                    corev1.ResourceCPU:    resource.MustParse("200m"),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    
    err := cluster.Client().Create(context.Background(), deployment)
    require.NoError(t, err)
    
    // Wait for deployment and monitor resource usage
    err = waitForDeploymentReady(cluster.Client(), "resource-test", 2*time.Minute)
    assert.NoError(t, err)
    
    // Verify resource consumption stays within limits
    metrics := getResourceMetrics(cluster, "resource-test")
    assert.True(t, metrics.MemoryUsage < 256*1024*1024) // 256MB
    assert.True(t, metrics.CPUUsage < 0.2)              // 200m cores
}
```

## Quality Framework

### Code Quality Metrics

**Go Code Quality:**
```yaml
# .golangci.yml
run:
  timeout: 5m
  modules-download-mode: readonly

linters-settings:
  gocyclo:
    min-complexity: 10
  goimports:
    local-prefixes: github.com/smartgardenbot
  golint:
    min-confidence: 0.8
  govet:
    check-shadowing: true
  misspell:
    locale: US

linters:
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - dupl
    - errcheck
    - gochecknoinits
    - goconst
    - gocyclo
    - gofmt
    - goimports
    - golint
    - gomnd
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - interfacer
    - lll
    - misspell
    - nakedret
    - rowserrcheck
    - scopelint
    - staticcheck
    - structcheck
    - stylecheck
    - testpackage
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gomnd
        - lll
```

**TypeScript Code Quality:**
```json
{
  "extends": [
    "@typescript-eslint/recommended",
    "@typescript-eslint/recommended-requiring-type-checking",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended",
    "plugin:jsx-a11y/recommended",
    "plugin:security/recommended"
  ],
  "parser": "@typescript-eslint/parser",
  "parserOptions": {
    "ecmaVersion": 2022,
    "sourceType": "module",
    "project": "./tsconfig.json"
  },
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/explicit-function-return-type": "warn",
    "@typescript-eslint/no-explicit-any": "error",
    "react/prop-types": "off",
    "react/react-in-jsx-scope": "off",
    "complexity": ["error", 10],
    "max-lines-per-function": ["error", 50],
    "max-depth": ["error", 4]
  }
}
```

### Code Coverage Standards

**Coverage Requirements:**
- **Unit Tests**: Minimum 80%, Target 90%
- **Integration Tests**: Minimum 70%, Target 85%
- **Critical Paths**: 100% coverage required
  - Authentication flows
  - Payment processing
  - Watering logic
  - Safety mechanisms

**Coverage Reporting:**
```yaml
# .github/workflows/coverage.yml
name: Code Coverage
on: [push, pull_request]

jobs:
  coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.21
      
      - name: Run Go tests with coverage
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: 18
      
      - name: Run TypeScript tests with coverage
        run: |
          cd web-app
          npm ci
          npm run test:coverage
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out,./web-app/coverage/lcov.info
          fail_ci_if_error: true
```

### Documentation Standards

**Go Documentation:**
```go
// Package weatherapi provides integration with external weather services.
// It supports multiple providers with automatic failover and rate limiting.
package weatherapi

// WeatherProvider defines the interface for weather service providers.
// Implementations must handle rate limiting and error recovery.
type WeatherProvider interface {
    // GetCurrentWeather retrieves current weather conditions for the specified location.
    // The location parameter should be a zip code, city name, or coordinates.
    // Returns an error if the service is unavailable or the location is invalid.
    GetCurrentWeather(ctx context.Context, location string) (*CurrentWeather, error)
    
    // GetForecast retrieves weather forecast for the specified location and duration.
    // Duration must be between 1 and 10 days.
    GetForecast(ctx context.Context, location string, days int) (*Forecast, error)
}
```

**TypeScript Documentation:**
```typescript
/**
 * Hook for managing garden sensor data with real-time updates
 * 
 * @param gardenId - Unique identifier for the garden
 * @param refreshInterval - Polling interval in milliseconds (default: 30000)
 * @returns Object containing sensor data, loading state, and error information
 * 
 * @example
 * ```tsx
 * function SensorDashboard({ gardenId }: { gardenId: string }) {
 *   const { data, isLoading, error } = useSensorData(gardenId, 60000);
 *   
 *   if (isLoading) return <Spinner />;
 *   if (error) return <ErrorMessage error={error} />;
 *   
 *   return <SensorChart data={data} />;
 * }
 * ```
 */
export function useSensorData(
  gardenId: string,
  refreshInterval: number = 30000
): UseSensorDataResult {
  // Implementation...
}
```

### Review Process

**Pull Request Requirements:**
1. **Automated Checks Pass**:
   - Linting and formatting
   - Unit test coverage ≥ 80%
   - Security scans pass
   - Build succeeds

2. **Code Review Checklist**:
   - [ ] Code follows established patterns
   - [ ] Tests cover new functionality
   - [ ] Documentation is updated
   - [ ] Performance impact considered
   - [ ] Security implications reviewed
   - [ ] Breaking changes documented

3. **Review Assignment Rules**:
   - Backend changes: 2 Go developers
   - Frontend changes: 2 TypeScript developers
   - Infrastructure changes: 1 DevOps engineer
   - Security-sensitive changes: 1 security reviewer

## CI/CD Testing Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

env:
  GO_VERSION: 1.21
  NODE_VERSION: 18

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: Run Go linting
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: ${{ env.NODE_VERSION }}
      
      - name: Run TypeScript linting
        run: |
          cd web-app
          npm ci
          npm run lint

  unit-tests:
    runs-on: ubuntu-latest
    needs: lint
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: Run Go tests
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: ${{ env.NODE_VERSION }}
      
      - name: Run TypeScript tests
        run: |
          cd web-app
          npm ci
          npm run test:coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out,./web-app/coverage/lcov.info

  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup test environment
        run: |
          docker-compose -f docker-compose.test.yml up -d
          sleep 30
      
      - name: Run integration tests
        run: |
          go test -v -tags=integration ./tests/integration/...
      
      - name: Cleanup
        if: always()
        run: docker-compose -f docker-compose.test.yml down

  security-scan:
    runs-on: ubuntu-latest
    needs: unit-tests
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Gosec
        uses: securecodewarrior/github-action-gosec@master
        with:
          args: '-fmt sarif -out gosec.sarif ./...'
      
      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          format: 'sarif'
          output: 'trivy.sarif'
      
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: '.'

  e2e-tests:
    runs-on: ubuntu-latest
    needs: integration-tests
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup full environment
        run: |
          docker-compose up -d
          sleep 60
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: ${{ env.NODE_VERSION }}
      
      - name: Install Playwright
        run: |
          cd e2e
          npm ci
          npx playwright install --with-deps
      
      - name: Run E2E tests
        run: |
          cd e2e
          npx playwright test
      
      - name: Upload test results
        uses: actions/upload-artifact@v3
        if: failure()
        with:
          name: playwright-report
          path: e2e/playwright-report/

  performance-tests:
    runs-on: ubuntu-latest
    needs: integration-tests
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup performance test environment
        run: |
          docker-compose -f docker-compose.perf.yml up -d
          sleep 60
      
      - name: Run K6 load tests
        uses: grafana/k6-action@v0.2.0
        with:
          filename: tests/performance/load-test.js
          flags: --out json=results.json
      
      - name: Upload performance results
        uses: actions/upload-artifact@v3
        with:
          name: k6-results
          path: results.json

  build-and-push:
    runs-on: ubuntu-latest
    needs: [security-scan, e2e-tests]
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Login to Container Registry
        uses: docker/login-action@v2
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Build and push API image
        uses: docker/build-push-action@v4
        with:
          context: ./api
          push: true
          tags: ghcr.io/${{ github.repository }}/api:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Build and push web app image
        uses: docker/build-push-action@v4
        with:
          context: ./web-app
          push: true
          tags: ghcr.io/${{ github.repository }}/web:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Build and push operator image
        uses: docker/build-push-action@v4
        with:
          context: ./operator
          push: true
          tags: ghcr.io/${{ github.repository }}/operator:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy-staging:
    runs-on: ubuntu-latest
    needs: build-and-push
    if: github.ref == 'refs/heads/main'
    environment: staging
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to staging
        run: |
          # Update Helm values with new image tags
          # Deploy to staging environment
          # Run smoke tests
          echo "Deploying to staging..."
```

### Quality Gates

**Gate Definitions:**
```yaml
# quality-gates.yml
gates:
  unit_tests:
    required: true
    minimum_coverage: 80
    max_duration: "10m"
  
  integration_tests:
    required: true
    max_duration: "20m"
  
  security_scan:
    required: true
    max_high_vulnerabilities: 0
    max_medium_vulnerabilities: 5
  
  performance_tests:
    required_for_main: true
    max_p95_response_time: "200ms"
    min_throughput: "1000rps"
  
  e2e_tests:
    required: true
    max_duration: "30m"
    min_success_rate: "99%"
```

### Test Environment Management

**Environment Configuration:**
```yaml
# docker-compose.test.yml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: smartgardenbot_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: testpass
    ports:
      - "5432:5432"
    volumes:
      - ./scripts/init-test-db.sql:/docker-entrypoint-initdb.d/init.sql
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
  
  api:
    build:
      context: ./api
      dockerfile: Dockerfile.test
    environment:
      DATABASE_URL: postgres://test:testpass@postgres:5432/smartgardenbot_test
      REDIS_URL: redis://redis:6379
      JWT_SECRET: test-secret
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
  
  web:
    build:
      context: ./web-app
      dockerfile: Dockerfile.test
    environment:
      NEXT_PUBLIC_API_URL: http://api:8080
    ports:
      - "3000:3000"
    depends_on:
      - api

  weather-mock:
    image: mockserver/mockserver:5.15.0
    ports:
      - "9999:1080"
    environment:
      MOCKSERVER_INITIALIZATION_JSON_PATH: /config/weather-mock.json
    volumes:
      - ./tests/mocks/weather-mock.json:/config/weather-mock.json
```

## Test Data Management

### Test Data Strategy

**Data Categories:**
1. **Static Test Data**: Predictable, version-controlled data
2. **Generated Test Data**: Dynamically created for each test run
3. **Anonymized Production Data**: Sanitized real-world data
4. **Synthetic Data**: AI-generated realistic test data

### Test Data Generation

**Go Test Data Factory:**
```go
package testdata

import (
    "time"
    "github.com/smartgardenbot/internal/models"
    "github.com/brianvoe/gofakeit/v6"
)

type Factory struct {
    faker *gofakeit.Faker
}

func NewFactory(seed int64) *Factory {
    faker := gofakeit.New(seed)
    return &Factory{faker: faker}
}

func (f *Factory) CreateUser() *models.User {
    return &models.User{
        Email:     f.faker.Email(),
        Name:      f.faker.Name(),
        CreatedAt: time.Now().Add(-time.Duration(f.faker.IntRange(1, 365)) * 24 * time.Hour),
        Location:  f.faker.Zip(),
        Timezone:  f.faker.TimeZone(),
    }
}

func (f *Factory) CreateGarden(userID string) *models.Garden {
    return &models.Garden{
        UserID:    userID,
        Name:      f.faker.Sentence(3),
        Location:  f.faker.Zip(),
        Timezone:  f.faker.TimeZone(),
        Area:      float64(f.faker.IntRange(50, 500)),
        CreatedAt: time.Now().Add(-time.Duration(f.faker.IntRange(1, 100)) * 24 * time.Hour),
    }
}

func (f *Factory) CreateSensorData(deviceID string, timestamp time.Time) *models.SensorReading {
    return &models.SensorReading{
        DeviceID:       deviceID,
        Timestamp:      timestamp,
        Temperature:    20.0 + f.faker.Float64Range(0, 25),
        Humidity:       40.0 + f.faker.Float64Range(0, 40),
        SoilMoisture:   30.0 + f.faker.Float64Range(0, 40),
        SolarRadiation: f.faker.Float64Range(0, 1000),
        Rainfall:       f.faker.Float64Range(0, 10),
        WindSpeed:      f.faker.Float64Range(0, 20),
    }
}

func (f *Factory) CreateWeatherForecast(location string, days int) *models.WeatherForecast {
    forecast := &models.WeatherForecast{
        Location:  location,
        UpdatedAt: time.Now(),
        Days:      make([]models.DayForecast, days),
    }
    
    baseDate := time.Now().Truncate(24 * time.Hour)
    for i := 0; i < days; i++ {
        forecast.Days[i] = models.DayForecast{
            Date:          baseDate.Add(time.Duration(i) * 24 * time.Hour),
            High:          f.faker.Float64Range(15, 35),
            Low:           f.faker.Float64Range(5, 20),
            Precipitation: f.faker.Float64Range(0, 2),
            Humidity:      f.faker.Float64Range(30, 90),
            WindSpeed:     f.faker.Float64Range(0, 25),
            Conditions:    f.faker.RandomString([]string{"sunny", "cloudy", "rainy", "stormy"}),
        }
    }
    
    return forecast
}
```

**TypeScript Test Data Factory:**
```typescript
import { faker } from '@faker-js/faker';

export class TestDataFactory {
  static createUser(overrides: Partial<User> = {}): User {
    return {
      id: faker.string.uuid(),
      email: faker.internet.email(),
      name: faker.person.fullName(),
      location: faker.location.zipCode(),
      timezone: faker.location.timeZone(),
      createdAt: faker.date.past().toISOString(),
      updatedAt: faker.date.recent().toISOString(),
      ...overrides,
    };
  }

  static createGarden(overrides: Partial<Garden> = {}): Garden {
    return {
      id: faker.string.uuid(),
      name: faker.lorem.words(3),
      location: faker.location.zipCode(),
      timezone: faker.location.timeZone(),
      area: faker.number.int({ min: 50, max: 500 }),
      zones: [],
      createdAt: faker.date.past().toISOString(),
      updatedAt: faker.date.recent().toISOString(),
      ...overrides,
    };
  }

  static createSensorData(overrides: Partial<SensorReading> = {}): SensorReading {
    return {
      id: faker.string.uuid(),
      deviceId: faker.string.uuid(),
      timestamp: faker.date.recent().toISOString(),
      temperature: faker.number.float({ min: 15, max: 40, precision: 0.1 }),
      humidity: faker.number.float({ min: 30, max: 90, precision: 0.1 }),
      soilMoisture: faker.number.float({ min: 20, max: 80, precision: 0.1 }),
      solarRadiation: faker.number.float({ min: 0, max: 1000, precision: 0.1 }),
      rainfall: faker.number.float({ min: 0, max: 10, precision: 0.1 }),
      windSpeed: faker.number.float({ min: 0, max: 25, precision: 0.1 }),
      ...overrides,
    };
  }

  static createWeatherForecast(days: number = 7): WeatherForecast {
    const forecast: WeatherForecast = {
      location: faker.location.zipCode(),
      updatedAt: faker.date.recent().toISOString(),
      days: [],
    };

    for (let i = 0; i < days; i++) {
      const date = new Date();
      date.setDate(date.getDate() + i);
      
      forecast.days.push({
        date: date.toISOString().split('T')[0],
        high: faker.number.float({ min: 20, max: 35, precision: 0.1 }),
        low: faker.number.float({ min: 5, max: 20, precision: 0.1 }),
        precipitation: faker.number.float({ min: 0, max: 2, precision: 0.1 }),
        humidity: faker.number.float({ min: 30, max: 90, precision: 0.1 }),
        windSpeed: faker.number.float({ min: 0, max: 25, precision: 0.1 }),
        conditions: faker.helpers.arrayElement([
          'sunny', 'cloudy', 'rainy', 'stormy', 'partly-cloudy'
        ]),
      });
    }

    return forecast;
  }
}
```

### Database Seeding

**Test Database Seeder:**
```go
package testdb

import (
    "context"
    "database/sql"
    "github.com/smartgardenbot/internal/testdata"
)

type Seeder struct {
    db      *sql.DB
    factory *testdata.Factory
}

func NewSeeder(db *sql.DB) *Seeder {
    return &Seeder{
        db:      db,
        factory: testdata.NewFactory(12345), // Fixed seed for reproducibility
    }
}

func (s *Seeder) SeedBasicData(ctx context.Context) error {
    // Create test users
    users := make([]*models.User, 10)
    for i := range users {
        users[i] = s.factory.CreateUser()
        if err := s.insertUser(ctx, users[i]); err != nil {
            return err
        }
    }
    
    // Create gardens for each user
    for _, user := range users {
        garden := s.factory.CreateGarden(user.ID)
        if err := s.insertGarden(ctx, garden); err != nil {
            return err
        }
        
        // Create zones for each garden
        for i := 0; i < 3; i++ {
            zone := s.factory.CreateZone(garden.ID)
            if err := s.insertZone(ctx, zone); err != nil {
                return err
            }
        }
    }
    
    return nil
}

func (s *Seeder) SeedSensorData(ctx context.Context, days int) error {
    gardens, err := s.getAllGardens(ctx)
    if err != nil {
        return err
    }
    
    for _, garden := range gardens {
        devices, err := s.getGardenDevices(ctx, garden.ID)
        if err != nil {
            return err
        }
        
        for _, device := range devices {
            // Generate hourly data for specified days
            startTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
            for h := 0; h < days*24; h++ {
                timestamp := startTime.Add(time.Duration(h) * time.Hour)
                reading := s.factory.CreateSensorData(device.ID, timestamp)
                if err := s.insertSensorReading(ctx, reading); err != nil {
                    return err
                }
            }
        }
    }
    
    return nil
}
```

### Test Data Cleanup

**Cleanup Strategy:**
```go
func TestWithCleanup(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Seed test data
    seeder := testdb.NewSeeder(db)
    ctx := context.Background()
    
    err := seeder.SeedBasicData(ctx)
    require.NoError(t, err)
    
    // Register cleanup function
    t.Cleanup(func() {
        cleanupTestData(db)
    })
    
    // Run actual test
    testFunction(t, db)
}

func cleanupTestData(db *sql.DB) {
    // Clean in reverse dependency order
    cleanupQueries := []string{
        "DELETE FROM sensor_readings WHERE device_id LIKE 'test-%'",
        "DELETE FROM watering_schedules WHERE zone_id IN (SELECT id FROM zones WHERE garden_id LIKE 'test-%')",
        "DELETE FROM zones WHERE garden_id LIKE 'test-%'",
        "DELETE FROM gardens WHERE id LIKE 'test-%'",
        "DELETE FROM users WHERE email LIKE '%@testdomain.com'",
    }
    
    for _, query := range cleanupQueries {
        if _, err := db.Exec(query); err != nil {
            // Log error but don't fail cleanup
            log.Printf("Cleanup error: %v", err)
        }
    }
}
```

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-2)
- [ ] Setup testing infrastructure
- [ ] Implement unit testing frameworks for Go and TypeScript
- [ ] Configure linting and code quality tools
- [ ] Setup basic CI/CD pipeline with quality gates
- [ ] Implement test data factories and database seeding

### Phase 2: Core Testing (Weeks 3-4)
- [ ] Complete unit test coverage for existing code
- [ ] Implement integration testing framework with Testcontainers
- [ ] Setup E2E testing with Playwright
- [ ] Implement security scanning in CI/CD pipeline
- [ ] Create mock services for external APIs

### Phase 3: Advanced Testing (Weeks 5-6)
- [ ] Implement IoT device simulation framework
- [ ] Setup performance testing with K6
- [ ] Create Kubernetes operator testing framework
- [ ] Implement comprehensive API security testing
- [ ] Setup test environment management

### Phase 4: Quality Assurance (Weeks 7-8)
- [ ] Establish code coverage enforcement
- [ ] Implement automated quality gates
- [ ] Setup monitoring and alerting for test results
- [ ] Create comprehensive test documentation
- [ ] Train team on testing practices and tools

### Phase 5: Optimization (Weeks 9-10)
- [ ] Optimize test execution performance
- [ ] Implement parallel test execution
- [ ] Setup test result analytics and reporting
- [ ] Establish regular penetration testing schedule
- [ ] Create test maintenance and update procedures

## Appendices

### Appendix A: Recommended Tools and Frameworks

**Testing Frameworks:**
- **Go**: `testing`, `testify`, `gomock`, `dockertest`
- **TypeScript**: Jest, Testing Library, MSW, Playwright
- **Integration**: Testcontainers, Docker Compose
- **Performance**: K6, Artillery, Lighthouse
- **Security**: OWASP ZAP, Gosec, ESLint Security, Trivy

**Quality Tools:**
- **Linting**: golangci-lint, ESLint, Prettier
- **Coverage**: Go cover, Istanbul/nyc
- **Security**: Snyk, WhiteSource, GitHub Security Advisories
- **Documentation**: godoc, TypeDoc, Storybook

**CI/CD Integration:**
- **GitHub Actions**: Comprehensive workflow src
- **Quality Gates**: SonarQube, CodeClimate
- **Container Scanning**: Trivy, Clair, Anchore
- **Monitoring**: Prometheus, Grafana, PagerDuty

### Appendix B: Testing Best Practices

**Unit Testing:**
- Write tests before code (TDD approach)
- Keep tests isolated and independent
- Use descriptive test names
- Mock external dependencies
- Test edge cases and error conditions

**Integration Testing:**
- Use real databases with containers
- Test complete workflows
- Validate API contracts
- Test error handling and recovery
- Verify data consistency

**E2E Testing:**
- Focus on critical user journeys
- Use page object pattern
- Implement proper wait strategies
- Test across multiple browsers
- Include mobile responsive testing

**Performance Testing:**
- Establish baseline metrics
- Test under realistic load
- Monitor resource utilization
- Test degradation scenarios
- Automate performance regression detection

**Security Testing:**
- Integrate security scanning in CI/CD
- Test authentication and authorization
- Validate input sanitization
- Test for common vulnerabilities
- Regular dependency updates

### Appendix C: Metrics and KPIs

**Testing Metrics:**
- Test coverage percentage
- Test execution time
- Test failure rate
- Defect detection rate
- Mean time to resolution

**Quality Metrics:**
- Code complexity scores
- Technical debt ratio
- Security vulnerability count
- Performance benchmark results
- Documentation coverage

**Process Metrics:**
- Pipeline success rate
- Deployment frequency
- Lead time for changes
- Mean time to recovery
- Change failure rate

This comprehensive testing and quality assurance strategy provides a solid foundation for ensuring the Smart Garden Bot platform meets enterprise-grade quality standards while maintaining rapid development velocity and reliable operations.