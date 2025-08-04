# Smart Garden Bot - System Architecture

## Executive Summary

The Smart Garden Bot is a cloud-native, microservices-based SaaS platform that automates garden irrigation through intelligent weather-based decision making and IoT sensor integration. This architecture supports scalable multi-tenant operations with enterprise-grade security, monitoring, and deployment practices.

## System Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Users & External Systems                 │
├─────────────────────────────────────────────────────────────────┤
│  Web Browser    │    Mobile App    │    Weather APIs    │  IoT   │
│     (SPA)       │      (PWA)       │    (External)     │ Devices│
└─────────────────────────────────────────────────────────────────┘
                                    │
                            ┌───────▼───────┐
                            │   Load Balancer│
                            │   (Ingress)    │
                            └───────┬───────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                          │                           │
        ▼                          ▼                           ▼
┌─────────────┐            ┌─────────────┐            ┌─────────────┐
│    Web      │            │    API      │            │ Kubernetes  │
│ Application │            │  Gateway    │◄──────────►│  Operator   │
│  (Next.js)  │            │  (Go API)   │            │    (Go)     │
└─────────────┘            └─────────────┘            └─────────────┘
                                    │                           │
                                    ▼                           ▼
                           ┌─────────────┐            ┌─────────────┐
                           │ PostgreSQL  │            │ IoT Control │
                           │  Database   │            │   Layer     │
                           │ (CloudNativePG)│         │ (Ecowitt)   │
                           └─────────────┘            └─────────────┘
```

### Core Components

1. **Web Application** - Next.js/TypeScript frontend with mobile-responsive design
2. **REST API** - Go-based microservice with OpenAPI 3.1 specification
3. **Kubernetes Operator** - Go-based controller for IoT device management
4. **Database Layer** - PostgreSQL with pgvector for AI/ML capabilities
5. **Authentication Service** - OpenID Connect integration (Auth0/KeyCloak)
6. **External Integrations** - Weather APIs, IoT devices, and payment processing

## Component Architecture

### 1. Web Application Layer

**Technology Stack:**
- **Framework**: Next.js 15+ with App Router
- **Runtime**: Deno for enhanced security and TypeScript support
- **Language**: TypeScript for type safety
- **Styling**: Tailwind CSS for responsive design
- **State Management**: React Query for server state, Zustand for client state

**Key Features:**
- Server-side rendering (SSR) for performance
- Progressive Web App (PWA) capabilities
- Mobile-responsive design
- Real-time dashboard updates via WebSockets/Server-Sent Events

**Architecture Pattern**: Jamstack with API-first design

### 2. API Gateway Layer

**Technology Stack:**
- **Language**: Go 1.21+
- **Framework**: Gin or Fiber for HTTP routing
- **Documentation**: OpenAPI 3.1 with Swagger UI
- **Monitoring**: Prometheus metrics integration

**Endpoints Architecture:**
```
/api/v1/
├── /auth/          # Authentication endpoints
├── /users/         # User management
├── /gardens/       # Garden configuration
├── /sensors/       # Sensor data
├── /watering/      # Watering schedules
├── /weather/       # Weather integration
├── /billing/       # Stripe integration
├── /health         # Health checks
├── /metrics        # Prometheus metrics
├── /openapi.json   # OpenAPI specification
└── /swagger-ui     # API documentation
```

**Quality Attributes:**
- **Availability**: 99.9% uptime with health checks
- **Performance**: <200ms response times for 95th percentile
- **Security**: OAuth2/OpenID Connect with API keys
- **Observability**: Structured logging, metrics, and tracing

### 3. Kubernetes Operator

**Technology Stack:**
- **Language**: Go with controller-runtime framework
- **Pattern**: Kubernetes Controller pattern with Custom Resources
- **Deployment**: Helm charts with GitOps workflow

**Custom Resources:**
```yaml
apiVersion: garden.smartbot.io/v1
kind: GardenController
metadata:
  name: backyard-garden
spec:
  zones:
    - name: vegetables
      sensors: ["soil-moisture-01", "temperature-01"]
      schedule: "0 6 * * *"
    - name: flowers
      sensors: ["soil-moisture-02"]
      schedule: "0 18 * * *"
  weatherIntegration:
    provider: "openweather"
    location: "12345"
  deviceConfig:
    type: "ecowitt"
    endpoint: "192.168.1.100"
```

**Controller Responsibilities:**
- Monitor weather forecasts and adjust watering schedules
- Process sensor data for irrigation decisions
- Interface with Ecowitt WittFlow controllers
- Maintain state reconciliation with desired garden configuration

### 4. Data Architecture

**Database Strategy: Single PostgreSQL with Multi-Schema Design**

```sql
-- Schema Organization
CREATE SCHEMA auth;      -- User authentication data
CREATE SCHEMA gardens;   -- Garden and device configuration
CREATE SCHEMA sensors;   -- Time-series sensor data
CREATE SCHEMA billing;   -- Subscription and payment data
CREATE SCHEMA analytics; -- ML/AI processing data
```

**Key Tables:**
```sql
-- Core user and garden entities
auth.users (id, email, created_at, updated_at)
gardens.gardens (id, user_id, name, location, timezone)
gardens.zones (id, garden_id, name, plant_type, area)
gardens.devices (id, garden_id, type, config, status)

-- Time-series data with partitioning
sensors.readings (
  id, device_id, timestamp, 
  temperature, humidity, soil_moisture, 
  solar_radiation, rainfall, wind_speed,
  PARTITION BY RANGE (timestamp)
)

-- AI/ML integration
analytics.embeddings (id, data_type, vector, metadata)
```

**Data Management:**
- **Partitioning**: Time-based partitioning for sensor data
- **Retention**: 1 year detailed data, 5 years aggregated
- **Backups**: Point-in-time recovery with 30-day retention
- **Extensions**: pgvector for similarity search, TimescaleDB for time-series optimization

## Security Architecture

### Authentication & Authorization

**Identity Provider Integration:**
```
┌─────────────┐    OIDC     ┌─────────────┐    JWT      ┌─────────────┐
│   Frontend  │◄───────────►│   Auth0/    │◄───────────►│  API Server │
│   (SPA)     │             │  KeyCloak   │             │   (Go API)  │
└─────────────┘             └─────────────┘             └─────────────┘
                                   │
                                   ▼
                            ┌─────────────┐
                            │ User Store  │
                            │(PostgreSQL) │
                            └─────────────┘
```

**Authorization Model:**
- **RBAC**: Role-based access control (Admin, User, Device)
- **Resource-based**: Users can only access their own gardens
- **API Keys**: Machine-to-machine authentication for operators

### Security Controls

1. **Transport Security**: TLS 1.3 for all communications
2. **Data Encryption**: AES-256 encryption at rest
3. **Secrets Management**: Kubernetes secrets with sealed-secrets
4. **Network Security**: Network policies and service mesh (Istio)
5. **Input Validation**: OpenAPI schema validation and sanitization

## Integration Architecture

### Weather API Integration

**Multi-Provider Strategy with Fallback:**
```go
type WeatherProvider interface {
    GetForecast(location string) (*Forecast, error)
    GetCurrentConditions(location string) (*Conditions, error)
}

// Primary: OpenWeatherMap
// Fallback: Weather.gov (NOAA) -> WeatherAPI -> Open-Meteo
```

**Circuit Breaker Pattern:**
- Automatic failover between providers
- Rate limiting and quota management
- Caching with Redis for API efficiency

### IoT Device Integration

**Ecowitt Integration Pattern:**
```
┌─────────────────┐    HTTP/MQTT    ┌─────────────────┐
│   Ecowitt       │◄───────────────►│  K8s Operator   │
│ Weather Station │                 │   (Go Client)   │
│   & WittFlow    │                 └─────────────────┘
└─────────────────┘                           │
                                              ▼
                                    ┌─────────────────┐
                                    │   PostgreSQL    │
                                    │ (Sensor Data)   │
                                    └─────────────────┘
```

### Payment Integration

**Stripe Integration:**
- Webhook-based subscription management
- PCI DSS compliance through Stripe Elements
- Multi-tier pricing with usage-based billing

## Deployment Architecture

### Kubernetes Infrastructure

**Cluster Architecture:**
```yaml
# Production Cluster Layout
apiVersion: v1
kind: Namespace
metadata:
  name: smart-garden-bot
---
# Application Components
- Web Application (Deployment + Service + Ingress)
- API Server (Deployment + Service)
- Kubernetes Operator (Deployment + RBAC)
- PostgreSQL (CloudNativePG Cluster)
- Redis (Helm Chart)
- Monitoring Stack (Prometheus + Grafana)
```

**GitOps Deployment Pipeline:**
```
┌─────────────┐    Push     ┌─────────────┐   Sync    ┌─────────────┐
│   GitHub    │────────────►│   GitHub    │──────────►│  ArgoCD     │
│   Source    │             │   Actions   │           │(GitOps CD)  │
└─────────────┘             └─────────────┘           └─────────────┘
                                    │                         │
                                    ▼                         ▼
                            ┌─────────────┐           ┌─────────────┐
                            │ Container   │           │ Kubernetes  │
                            │ Registry    │           │   Cluster   │
                            └─────────────┘           └─────────────┘
```

**Helm Chart Structure:**
```
smart-garden-bot/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── web-app/
│   ├── api/
│   ├── operator/
│   ├── database/
│   └── monitoring/
└── charts/
    ├── postgresql/
    └── redis/
```

### Infrastructure as Code

**Technology Stack:**
- **Kubernetes**: Container orchestration
- **Helm**: Package management and templating
- **ArgoCD**: GitOps continuous deployment
- **CloudNativePG**: PostgreSQL operator
- **Prometheus**: Metrics and alerting
- **Grafana**: Visualization and dashboards

## Scalability & Performance

### Horizontal Scaling Strategy

**Application Tier:**
- **Web App**: CDN-cached static assets + auto-scaling pods
- **API Server**: Stateless design with load balancing
- **Database**: Read replicas for query optimization

**Performance Targets:**
- **Web Application**: First Contentful Paint < 1.5s
- **API Response**: 95th percentile < 200ms
- **Database**: Query performance < 100ms for 99% of operations

### Caching Strategy

**Multi-Layer Caching:**
1. **CDN**: Static assets and API responses (CloudFlare)
2. **Application**: Redis for session data and weather forecasts
3. **Database**: Query result caching with invalidation policies

## Monitoring & Observability

### Metrics Collection

**Prometheus Metrics:**
```yaml
# API Server Metrics
- http_requests_total{method, endpoint, status}
- http_request_duration_seconds{method, endpoint}
- active_users_total
- watering_events_total{zone, success}

# Business Metrics
- gardens_registered_total
- sensor_readings_per_second
- weather_api_requests_total{provider, status}
```

### Logging Strategy

**Structured Logging:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "api",
  "trace_id": "abc123",
  "user_id": "user-456",
  "event": "watering_scheduled",
  "garden_id": "garden-789",
  "zone": "vegetables"
}
```

### Alerting Rules

**Critical Alerts:**
- API server availability < 99.5%
- Database connection failures
- Weather API integration failures
- Watering system communication errors

## AI/ML Integration

### Plant Disease Detection

**Architecture:**
```
┌─────────────┐    Image     ┌─────────────┐   Vector    ┌─────────────┐
│   Mobile    │─────────────►│  ML Model   │────────────►│ PostgreSQL  │
│   Camera    │              │  (PyTorch)  │             │ (pgvector)  │
└─────────────┘              └─────────────┘             └─────────────┘
                                     │
                                     ▼
                             ┌─────────────┐
                             │ Disease DB  │
                             │(Embeddings) │
                             └─────────────┘
```

**Implementation Strategy:**
- **Model Serving**: TorchServe or TensorFlow Serving
- **Vector Storage**: pgvector for similarity search
- **Training Pipeline**: MLflow for model versioning

## Technology Decision Matrix

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Frontend** | Next.js + Deno | SSR performance, TypeScript support, security |
| **API** | Go + Gin | Performance, concurrency, strong typing |
| **Database** | PostgreSQL + pgvector | ACID compliance, vector support, mature ecosystem |
| **Container** | Docker + Kubernetes | Cloud-native, scalability, operational maturity |
| **Authentication** | Auth0/KeyCloak | OIDC compliance, enterprise features |
| **Monitoring** | Prometheus + Grafana | Industry standard, rich ecosystem |
| **Deployment** | Helm + ArgoCD | GitOps best practices, declarative configuration |

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)
- [x] Project setup and development environment
- [ ] Core API server with authentication
- [ ] Basic web application structure
- [ ] PostgreSQL database setup
- [ ] CI/CD pipeline foundation

### Phase 2: Core Features (Weeks 5-8)
- [ ] User registration and garden configuration
- [ ] Weather API integration
- [ ] Basic watering schedule management
- [ ] Sensor data collection endpoints
- [ ] Kubernetes operator framework

### Phase 3: IoT Integration (Weeks 9-12)
- [ ] Ecowitt device integration
- [ ] Real-time sensor data processing
- [ ] Automated watering decisions
- [ ] Mobile-responsive dashboard
- [ ] Basic monitoring and alerting

### Phase 4: Advanced Features (Weeks 13-16)
- [ ] AI/ML plant disease detection
- [ ] Advanced analytics dashboard
- [ ] Billing and subscription management
- [ ] Multi-zone garden support
- [ ] Performance optimization

### Phase 5: Production Ready (Weeks 17-20)
- [ ] Security hardening and penetration testing
- [ ] Load testing and performance tuning
- [ ] Documentation and user guides
- [ ] Production deployment and monitoring
- [ ] Beta user onboarding

## Architectural Decision Records (ADRs)

### ADR-001: Microservices vs Monolith
**Decision**: Start with a modular monolith, evolve to microservices
**Rationale**: Faster initial development, easier debugging, can split later based on usage patterns

### ADR-002: Database Strategy
**Decision**: Single PostgreSQL with multi-schema design
**Rationale**: ACID compliance for financial data, pgvector for AI features, operational simplicity

### ADR-003: Authentication Provider
**Decision**: Auth0 for production, KeyCloak for development
**Rationale**: Managed service reduces operational overhead, KeyCloak for local development

### ADR-004: Weather API Strategy
**Decision**: Multi-provider with circuit breaker pattern
**Rationale**: Resilience against single provider failures, cost optimization

### ADR-005: Container Orchestration
**Decision**: Kubernetes with Helm charts
**Rationale**: Industry standard, rich ecosystem, aligns with cloud-native goals

## Risk Assessment & Mitigation

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Weather API rate limits | High | Medium | Multi-provider strategy, caching |
| IoT device connectivity | High | Medium | Offline mode, retry mechanisms |
| Database performance | Medium | Low | Read replicas, query optimization |
| Third-party dependencies | Medium | Medium | Vendor evaluation, fallback options |

### Operational Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Security breaches | High | Low | Security audits, penetration testing |
| Scalability bottlenecks | Medium | Medium | Load testing, monitoring |
| Data loss | High | Low | Automated backups, disaster recovery |
| Service availability | High | Low | Multi-region deployment, health checks |

## Conclusion

This architecture provides a solid foundation for the Smart Garden Bot platform, balancing immediate functionality needs with long-term scalability and maintainability. The cloud-native, microservices-ready design supports the business goal of creating a scalable SaaS platform while providing users with reliable, intelligent garden automation capabilities.

The modular approach allows for incremental development and deployment, while the comprehensive monitoring and observability stack ensures operational excellence as the platform scales to serve multiple users and gardens.