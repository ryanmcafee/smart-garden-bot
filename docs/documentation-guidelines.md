# Smart Garden Bot - Documentation Guidelines and Templates

## Overview

This document provides comprehensive guidelines and templates for creating and maintaining documentation across the Smart Garden Bot project. Consistent, high-quality documentation is essential for team productivity, user adoption, and long-term maintainability.

## Documentation Standards

### Writing Style Guide

**Tone and Voice**:
- Clear, concise, and professional
- Active voice preferred over passive voice
- Use simple, straightforward language
- Avoid jargon unless necessary (define when used)
- Write for the target audience's expertise level

**Formatting Standards**:
- Use Markdown for all documentation
- Follow consistent heading hierarchy (H1 > H2 > H3 > H4)
- Use code blocks with language specification
- Include table of contents for documents > 500 words
- Use bullet points and numbered lists for clarity

**Language Guidelines**:
```markdown
✅ Good src:
- "Create a new garden by clicking the Add Garden button"
- "The API returns a 200 status code on success"
- "Configure your weather station using the setup wizard"

❌ Avoid:
- "Gardens can be created by users through the interface"
- "Success will be indicated by the API"
- "Weather station configuration is possible via the wizard"
```

### Document Structure

**Standard Document Template**:
```markdown
# Document Title

## Overview
Brief description of the document's purpose and scope.

## Prerequisites
- Required knowledge, tools, or setup
- Links to prerequisite documentation

## Main Content
Organized into logical sections with clear headings.

## src
Practical src with code snippets and explanations.

## Troubleshooting
Common issues and their solutions.

## Related Documentation
Links to relevant documents and resources.

## Changelog
Document version history with dates and changes.
```

## API Documentation Standards

### OpenAPI 3.1 Specification

**API Documentation Structure**:
```yaml
openapi: 3.1.0
info:
  title: Smart Garden Bot API
  version: 1.0.0
  description: |
    REST API for the Smart Garden Bot platform providing garden management,
    sensor monitoring, and watering control capabilities.
    
    ## Authentication
    All endpoints require authentication via Bearer token or API key.
    
    ## Rate Limiting
    API requests are limited to 1000 requests per hour per user.
    
    ## Error Handling
    All errors follow RFC 7807 Problem Details standard.
  contact:
    name: Smart Garden Bot Team
    email: api-support@smartgardenbot.com
    url: https://docs.smartgardenbot.com
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT

servers:
  - url: https://api.smartgardenbot.com/v1
    description: Production server
  - url: https://staging-api.smartgardenbot.com/v1
    description: Staging server
```

**Endpoint Documentation Template**:
```yaml
paths:
  /gardens:
    post:
      summary: Create a new garden
      description: |
        Creates a new garden configuration for the authenticated user.
        The garden will be initialized with default settings and can be
        customized through subsequent API calls.
      operationId: createGarden
      tags:
        - Gardens
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateGardenRequest'
            src:
              basic_garden:
                summary: Basic garden setup
                value:
                  name: "My Vegetable Garden"
                  location:
                    latitude: 37.7749
                    longitude: -122.4194
                  timezone: "America/Los_Angeles"
      responses:
        '201':
          description: Garden created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Garden'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '429':
          $ref: '#/components/responses/RateLimited'
```

**Schema Documentation Best Practices**:
```yaml
components:
  schemas:
    Garden:
      type: object
      required:
        - id
        - name
        - location
        - created_at
      properties:
        id:
          type: string
          format: uuid
          description: Unique identifier for the garden
          example: "123e4567-e89b-12d3-a456-426614174000"
        name:
          type: string
          minLength: 1
          maxLength: 100
          description: Human-readable name for the garden
          example: "My Vegetable Garden"
        location:
          $ref: '#/components/schemas/Location'
        zones:
          type: array
          description: List of watering zones in the garden
          items:
            $ref: '#/components/schemas/Zone'
        created_at:
          type: string
          format: date-time
          description: Timestamp when the garden was created
          example: "2024-01-15T10:30:00Z"
      description: |
        Represents a garden configuration including location, zones,
        and associated devices. Each garden belongs to a single user
        and can contain multiple watering zones.
```

### API Documentation Generation

**Automated Documentation Pipeline**:
```yaml
# .github/workflows/docs.yml
name: Generate API Documentation

on:
  push:
    paths:
      - 'api/openapi.yaml'
      - 'docs/**'

jobs:
  generate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Generate OpenAPI docs
        uses: redocly/redoc-cli-github-action@v0.1.0
        with:
          args: 'build api/openapi.yaml --output docs/api/index.html'
      
      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs
```

## Architecture Decision Records (ADRs)

### ADR Template

```markdown
# ADR-XXX: [Short title of solved problem and solution]

## Status
[Proposed | Accepted | Deprecated | Superseded by ADR-YYY]

## Context
What is the issue that we're seeing that is motivating this decision or change?
Include the technological, business, and project context.

## Decision
What is the change that we're proposing or have agreed to implement?

## Consequences
What becomes easier or more difficult to do and any risks introduced by this change?

### Positive Consequences
- [e.g., improvement of quality attribute satisfaction, follow-up decisions required, ...]

### Negative Consequences
- [e.g., compromising quality attribute, follow-up decisions required, ...]

## Alternatives Considered
What other alternatives were considered? Why were they not chosen?

## Implementation Notes
Any specific implementation details, timeline, or migration strategy.

## References
- [Link to discussion]
- [Relevant documentation]
- [Related ADRs]

---
Date: YYYY-MM-DD
Authors: [List of authors]
Reviewers: [List of reviewers]
```

### ADR src

**ADR-001: Database Selection**:
```markdown
# ADR-001: PostgreSQL as Primary Database

## Status
Accepted

## Context
The Smart Garden Bot requires a reliable database system to store user data,
garden configurations, sensor readings, and billing information. The system
needs to support:
- ACID transactions for financial data
- Time-series data for sensor readings
- Vector storage for AI/ML features
- Multi-tenant data isolation

## Decision
We will use PostgreSQL 15+ as our primary database with the following extensions:
- pgvector for vector similarity search
- TimescaleDB for time-series optimization
- Multi-schema design for logical separation

## Consequences

### Positive Consequences
- Strong consistency guarantees for financial data
- Rich ecosystem of tools and extensions
- Excellent performance for both OLTP and analytical workloads
- Built-in JSON support for flexible schema requirements
- Strong community support and documentation

### Negative Consequences
- Higher operational complexity compared to managed NoSQL solutions
- Need for database expertise on the team
- Potential scaling challenges at extreme scale (though not expected)

## Alternatives Considered
- **MongoDB**: Rejected due to lack of ACID guarantees for financial data
- **MySQL**: Rejected due to limited JSON support and extension ecosystem
- **Amazon DynamoDB**: Rejected due to vendor lock-in and limited query capabilities

## Implementation Notes
- Use CloudNativePG operator for Kubernetes deployment
- Implement connection pooling with PgBouncer
- Set up read replicas for query load distribution
- Partition time-series tables by month

Date: 2024-01-15
Authors: [Architecture Team]
Reviewers: [CTO, Senior Engineers]
```

## Code Documentation Standards

### Inline Code Comments

**Go Documentation Standards**:
```go
// Package weather provides integration with multiple weather API providers
// using a circuit breaker pattern for resilience.
package weather

// WeatherProvider defines the interface for weather data providers.
// Implementations should handle rate limiting and error recovery internally.
type WeatherProvider interface {
    // GetForecast retrieves weather forecast data for the specified location.
    // Returns forecast data for up to 10 days or an error if the request fails.
    //
    // The location parameter should be in format "latitude,longitude" or
    // a valid location identifier supported by the provider.
    GetForecast(ctx context.Context, location string) (*Forecast, error)
    
    // GetCurrentConditions retrieves current weather conditions.
    // Returns current weather data or an error if unavailable.
    GetCurrentConditions(ctx context.Context, location string) (*Conditions, error)
}

// OpenWeatherMapProvider implements WeatherProvider using OpenWeatherMap API.
// It includes automatic retry logic and rate limiting to respect API quotas.
type OpenWeatherMapProvider struct {
    apiKey     string
    httpClient *http.Client
    rateLimiter *rate.Limiter
}

// NewOpenWeatherMapProvider creates a new OpenWeatherMap provider instance.
// The apiKey parameter must be a valid OpenWeatherMap API key.
// Returns a configured provider ready for use.
func NewOpenWeatherMapProvider(apiKey string) *OpenWeatherMapProvider {
    return &OpenWeatherMapProvider{
        apiKey:      apiKey,
        httpClient:  &http.Client{Timeout: 30 * time.Second},
        rateLimiter: rate.NewLimiter(rate.Every(time.Second), 60), // 60 requests per second
    }
}
```

**TypeScript Documentation Standards**:
```typescript
/**
 * Garden management service providing CRUD operations and validation.
 * 
 * This service handles all garden-related operations including creation,
 * updates, and validation of garden configurations. It integrates with
 * the backend API and manages local state synchronization.
 * 
 * @example
 * ```typescript
 * const gardenService = new GardenService(apiClient);
 * const garden = await gardenService.createGarden({
 *   name: "My Garden",
 *   location: { lat: 37.7749, lng: -122.4194 },
 *   timezone: "America/Los_Angeles"
 * });
 * ```
 */
export class GardenService {
  private apiClient: ApiClient;
  
  constructor(apiClient: ApiClient) {
    this.apiClient = apiClient;
  }
  
  /**
   * Creates a new garden with the provided configuration.
   * 
   * @param request - Garden creation parameters
   * @param request.name - Human-readable garden name (1-100 characters)
   * @param request.location - Geographic coordinates for the garden
   * @param request.timezone - IANA timezone identifier
   * @returns Promise resolving to the created garden
   * @throws {ValidationError} When request parameters are invalid
   * @throws {ApiError} When the API request fails
   * 
   * @example
   * ```typescript
   * const garden = await gardenService.createGarden({
   *   name: "Backyard Vegetables",
   *   location: { lat: 40.7128, lng: -74.0060 },
   *   timezone: "America/New_York"
   * });
   * console.log(`Created garden: ${garden.name}`);
   * ```
   */
  async createGarden(request: CreateGardenRequest): Promise<Garden> {
    // Validate input parameters
    if (!request.name || request.name.length === 0) {
      throw new ValidationError("Garden name is required");
    }
    
    try {
      const response = await this.apiClient.post('/gardens', request);
      return response.data;
    } catch (error) {
      throw new ApiError(`Failed to create garden: ${error.message}`);
    }
  }
}
```

### README Documentation

**Repository README Template**:
```markdown
# Smart Garden Bot

[![Build Status](https://github.com/ryanmcafee/smart-garden-bot/workflows/CI/badge.svg)](https://github.com/ryanmcafee/smart-garden-bot/actions)
[![Coverage Status](https://codecov.io/gh/smart-garden-bot/smart-garden-bot/branch/main/graph/badge.svg)](https://codecov.io/gh/smart-garden-bot/smart-garden-bot)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Intelligent irrigation management platform that automates garden watering based on weather forecasts and sensor data.

## Features

- 🌱 **Smart Watering**: Automated irrigation based on weather forecasts and soil conditions
- 📊 **Real-time Monitoring**: Live sensor data dashboard with historical trends
- 🌤️ **Weather Integration**: Multi-provider weather API integration with fallback
- 📱 **Mobile Responsive**: Progressive web app for desktop and mobile devices
- ☁️ **Cloud Native**: Kubernetes-based deployment with GitOps workflow
- 🔒 **Enterprise Security**: OAuth2/OIDC authentication with role-based access

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+ or Deno 1.40+
- Docker and Docker Compose
- Kubernetes cluster (minikube/kind for local development)

### Local Development

1. **Clone the repository**:
   ```bash
   git clone https://github.com/ryanmcafee/smart-garden-bot.git
   cd smart-garden-bot
   ```

2. **Start development services**:
   ```bash
   docker-compose -f docker-compose.dev.yml up -d
   ```

3. **Initialize the database**:
   ```bash
   make db-migrate
   ```

4. **Start the API server**:
   ```bash
   cd api
   go run main.go
   ```

5. **Start the web application**:
   ```bash
   cd web
   npm install
   npm run dev
   ```

6. **Access the application**:
   - Web app: http://localhost:3000
   - API docs: http://localhost:8080/swagger-ui
   - API health: http://localhost:8080/health

## Architecture

The Smart Garden Bot consists of three main components:

- **Web Application**: Next.js frontend with TypeScript and Tailwind CSS
- **API Server**: Go-based REST API with OpenAPI 3.1 specification
- **Kubernetes Operator**: Go controller for IoT device management

For detailed architecture information, see [Architecture Documentation](./architecture.md).

## Documentation

- [API Documentation](https://docs.smartgardenbot.com/api/)
- [Development Guide](./docs/development-standards.md)
- [Deployment Guide](./docs/deployment.md)
- [Contributing Guidelines](./CONTRIBUTING.md)

## Contributing

We welcome contributions! Please see our [Contributing Guide](./CONTRIBUTING.md) for details on:

- Code of Conduct
- Development workflow
- Pull request process
- Coding standards

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## Support

- 📖 [Documentation](https://docs.smartgardenbot.com)
- 🐛 [Issue Tracker](https://github.com/ryanmcafee/smart-garden-bot/issues)
- 💬 [Discussions](https://github.com/ryanmcafee/smart-garden-bot/discussions)
- 📧 [Email Support](mailto:support@smartgardenbot.com)
```

## User Guide Documentation

### User Guide Template

```markdown
# Smart Garden Bot User Guide

## Getting Started

### Account Setup

1. **Create Account**:
   - Visit [Smart Garden Bot](https://smartgardenbot.com)
   - Click "Sign Up" and complete registration
   - Verify your email address

2. **Initial Configuration**:
   - Set your location and timezone
   - Configure notification preferences
   - Add billing information (if using paid features)

### Garden Setup

#### Creating Your First Garden

1. **Add Garden**:
   ```
   Dashboard → Gardens → Add New Garden
   ```

2. **Garden Configuration**:
   - **Name**: Choose a descriptive name (e.g., "Backyard Vegetables")
   - **Location**: Enter your address or coordinates
   - **Timezone**: Select your local timezone
   - **Garden Type**: Choose from preset configurations

3. **Zone Configuration**:
   - Add watering zones for different plant types
   - Set watering schedules and duration
   - Configure soil moisture thresholds

#### Device Integration

**Weather Station Setup**:
1. Connect your Ecowitt weather station to Wi-Fi
2. In the app: Settings → Devices → Add Weather Station
3. Enter your weather station's MAC address
4. Test the connection and verify data flow

**Watering Controller Setup**:
1. Install Ecowitt WittFlow controller
2. Connect to your irrigation system
3. In the app: Settings → Devices → Add Watering Controller
4. Configure zone mappings and test watering

### Daily Usage

#### Monitoring Your Garden

**Dashboard Overview**:
- Current weather conditions and forecast
- Soil moisture levels across all zones
- Recent watering activity
- Plant health alerts and recommendations

**Historical Data**:
- View sensor data trends over time
- Compare weather forecasts to actual conditions
- Track watering efficiency and plant growth

#### Managing Watering Schedules

**Automatic Mode** (Recommended):
- System automatically adjusts watering based on:
  - Weather forecasts (skips watering before rain)
  - Soil moisture readings
  - Plant type requirements
  - Seasonal adjustments

**Manual Override**:
- Emergency watering: Instant zone activation
- Schedule adjustments: Temporary schedule changes
- Vacation mode: Extended dry periods

### Troubleshooting

#### Common Issues

**Weather Station Not Reporting**:
1. Check Wi-Fi connection
2. Verify device power and battery levels
3. Check app device status page
4. Contact support if issues persist

**Watering Not Activating**:
1. Verify controller power and connectivity
2. Check zone valve connections
3. Review watering schedule settings
4. Test manual watering activation

**Inaccurate Sensor Readings**:
1. Clean sensor probes of debris
2. Verify sensor placement in representative soil
3. Calibrate sensors if readings seem off
4. Replace sensors if consistently inaccurate

#### Getting Help

- **In-App Help**: Help → Support → Chat
- **Email Support**: support@smartgardenbot.com
- **Documentation**: docs.smartgardenbot.com
- **Community Forum**: community.smartgardenbot.com
```

## Operational Runbooks

### Deployment Runbook Template

```markdown
# Smart Garden Bot Deployment Runbook

## Pre-deployment Checklist

### Code Quality Gates
- [ ] All tests pass (unit, integration, e2e)
- [ ] Code coverage meets minimum threshold (80%)
- [ ] Security scans pass (no high/critical vulnerabilities)
- [ ] Performance benchmarks meet requirements
- [ ] Documentation is updated

### Infrastructure Readiness
- [ ] Kubernetes cluster is healthy
- [ ] Database migrations are tested
- [ ] External service dependencies are verified
- [ ] Monitoring and alerting are configured
- [ ] Backup systems are operational

### Release Preparation
- [ ] Release notes are prepared
- [ ] Feature flags are configured
- [ ] Rollback plan is documented
- [ ] Team notification channels are ready
- [ ] On-call engineer is identified

## Deployment Process

### 1. Pre-deployment Verification

```bash
# Verify cluster health
kubectl get nodes
kubectl top nodes

# Check application status
kubectl get pods -n smart-garden-bot
kubectl get services -n smart-garden-bot

# Verify database connectivity
kubectl exec -it postgres-primary-0 -- psql -U postgres -c "SELECT version();"
```

### 2. Database Migrations

```bash
# Run migrations in staging first
kubectl apply -f k8s/migrations/001-add-garden-zones.yaml

# Verify migration success
kubectl logs -f migration-job-001

# Apply to production after staging verification
kubectl apply -f k8s/migrations/001-add-garden-zones.yaml -n production
```

### 3. Application Deployment

```bash
# Deploy using ArgoCD (GitOps)
git tag v1.2.0
git push origin v1.2.0

# Or manual Helm deployment
helm upgrade smart-garden-bot ./helm/smart-garden-bot \
  --namespace smart-garden-bot \
  --values values.production.yaml

# Monitor rollout
kubectl rollout status deployment/api -n smart-garden-bot
kubectl rollout status deployment/web-app -n smart-garden-bot
```

### 4. Post-deployment Verification

```bash
# Health check endpoints
curl -f https://api.smartgardenbot.com/health
curl -f https://smartgardenbot.com/health

# Verify database connectivity
kubectl exec -it api-0 -- /app/health-check --database

# Check metrics and logs
kubectl logs -f deployment/api -n smart-garden-bot
```

## Rollback Procedures

### Database Rollback
```bash
# If migration issues occur
kubectl apply -f k8s/migrations/001-add-garden-zones.down.yaml

# Verify rollback
kubectl logs -f migration-rollback-job-001
```

### Application Rollback
```bash
# Helm rollback
helm rollback smart-garden-bot -n smart-garden-bot

# Or ArgoCD rollback
argocd app rollback smart-garden-bot --revision <previous-revision>

# Monitor rollback
kubectl rollout status deployment/api -n smart-garden-bot
```

## Incident Response

### Severity Levels

**P0 - Critical**:
- Complete service outage
- Data loss or corruption
- Security breach

**P1 - High**:
- Major feature unavailable
- Performance degradation >50%
- Authentication failures

**P2 - Medium**:
- Minor feature issues
- Performance degradation <50%
- Non-critical bugs

### Response Procedures

1. **Immediate Response** (0-15 minutes):
   - Acknowledge incident in PagerDuty
   - Create incident channel (#incident-YYYY-MM-DD-NNN)
   - Notify stakeholders based on severity

2. **Investigation** (15-60 minutes):
   - Gather relevant logs and metrics
   - Identify root cause
   - Implement immediate mitigation if possible

3. **Resolution** (varies):
   - Apply fixes or rollback if necessary
   - Verify resolution across all environments
   - Update incident channel with status

4. **Post-incident** (within 48 hours):
   - Conduct post-mortem meeting
   - Document lessons learned
   - Create follow-up tasks for prevention
```

## Knowledge Base Structure

### Knowledge Base Organization

```
docs/
├── api/                    # API documentation
│   ├── openapi.yaml       # OpenAPI specification
│   ├── endpoints/         # Detailed endpoint docs
│   └── src/          # API usage src
├── architecture/          # System architecture
│   ├── overview.md        # High-level architecture
│   ├── components/        # Component specifications
│   └── adrs/             # Architecture Decision Records
├── deployment/           # Deployment documentation
│   ├── local-dev.md      # Local development setup
│   ├── staging.md        # Staging deployment
│   ├── production.md     # Production deployment
│   └── runbooks/         # Operational runbooks
├── user-guides/          # End-user documentation
│   ├── getting-started.md
│   ├── garden-setup.md
│   ├── device-integration.md
│   └── troubleshooting.md
├── developer/            # Developer documentation
│   ├── development-standards.md
│   ├── testing-guidelines.md
│   ├── security-practices.md
│   └── contributing.md
└── templates/            # Documentation templates
    ├── adr-template.md
    ├── runbook-template.md
    └── user-guide-template.md
```

## Documentation Maintenance

### Review and Update Process

**Regular Reviews**:
- Monthly: User guides and tutorials
- Quarterly: API documentation and developer guides
- Per release: Architecture documentation and runbooks
- Annually: Complete documentation audit

**Update Triggers**:
- New feature releases
- API changes or deprecations
- Architecture modifications
- User feedback and support tickets
- Security updates and patches

**Quality Metrics**:
- Documentation coverage per component
- User engagement metrics (page views, time spent)
- Support ticket reduction after doc updates
- Developer onboarding time improvements

### Automated Documentation

**Documentation Pipeline**:
```yaml
# .github/workflows/docs-validation.yml
name: Documentation Validation

on:
  pull_request:
    paths:
      - 'docs/**'
      - '*.md'

jobs:
  validate-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Lint documentation
        uses: DavidAnson/markdownlint-cli2-action@v13
        with:
          globs: '**/*.md'
      
      - name: Check links
        uses: gaurav-nelson/github-action-markdown-link-check@v1
        with:
          use-quiet-mode: 'yes'
          use-verbose-mode: 'yes'
          config-file: '.markdown-link-check.json'
      
      - name: Validate OpenAPI
        uses: char0n/swagger-editor-validate@v1
        with:
          definition-file: api/openapi.yaml
```

This documentation guidelines document provides comprehensive templates and standards for maintaining high-quality documentation across the Smart Garden Bot project, ensuring consistency, accuracy, and usefulness for all stakeholders.