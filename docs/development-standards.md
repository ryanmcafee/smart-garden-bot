# Smart Garden Bot - Development Standards

## Overview

This document defines the development standards, coding practices, and workflows for the Smart Garden Bot project. These standards ensure code quality, maintainability, and effective team collaboration across all components of the platform.

## Code Style and Formatting Guidelines

### Go (API Server & Kubernetes Operator)

**Language Version**: Go 1.21+

**Formatting Standards**:
- Use `gofmt` and `goimports` for automatic formatting
- Line length: 120 characters maximum
- Use meaningful variable and function names
- Follow Go naming conventions (PascalCase for exported, camelCase for unexported)

**Code Organization**:
```go
// Package structure
pkg/
├── api/           # API handlers and routes
├── config/        # Configuration management
├── models/        # Data models and structs
├── services/      # Business logic
├── controllers/   # Kubernetes controllers
├── clients/       # External API clients
└── utils/         # Utility functions

// File naming: use snake_case for files
// user_service.go, weather_client.go
```

**Linting Configuration**:
```yaml
# .golangci.yml
linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 2
    min-occurrences: 2
  misspell:
    locale: US
  lll:
    line-length: 120

linters:
  enable:
    - gofmt
    - goimports
    - golint
    - govet
    - gosec
    - ineffassign
    - misspell
    - gocyclo
```

**Error Handling**:
```go
// Always handle errors explicitly
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use structured logging with context
log.WithFields(log.Fields{
    "user_id": userID,
    "operation": "create_garden",
}).Error("Failed to create garden", err)
```

### TypeScript/Next.js (Web Application)

**Language Version**: TypeScript 5.0+, Next.js 15+

**Formatting Standards**:
- Use Prettier with ESLint for consistent formatting
- Semicolons: required
- Quotes: double quotes for strings
- Line length: 100 characters maximum
- Indentation: 2 spaces

**ESLint Configuration**:
```json
{
  "extends": [
    "next/core-web-vitals",
    "@typescript-eslint/recommended",
    "prettier"
  ],
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/explicit-function-return-type": "warn",
    "prefer-const": "error",
    "no-var": "error"
  }
}
```

**Project Structure**:
```
src/
├── app/              # Next.js App Router pages
├── components/       # Reusable UI components
│   ├── ui/          # Base UI components
│   └── features/    # Feature-specific components
├── lib/             # Utility functions and configurations
├── hooks/           # Custom React hooks
├── types/           # TypeScript type definitions
├── services/        # API client functions
└── styles/          # Global styles and Tailwind config
```

**TypeScript Best Practices**:
```typescript
// Use explicit types for function parameters and returns
interface CreateGardenRequest {
  name: string;
  location: string;
  timezone: string;
}

export async function createGarden(
  request: CreateGardenRequest
): Promise<Garden> {
  // Implementation
}

// Use enums for constants
enum WateringStatus {
  ACTIVE = "active",
  IDLE = "idle",
  ERROR = "error",
}

// Use proper error handling
try {
  const garden = await createGarden(request);
  return garden;
} catch (error) {
  logger.error("Failed to create garden", { error, request });
  throw new Error("Garden creation failed");
}
```

### Database Standards

**Schema Design**:
- Use snake_case for table and column names
- Include created_at and updated_at timestamps on all tables
- Use UUIDs for primary keys where appropriate
- Implement proper foreign key constraints

**Migration Standards**:
```sql
-- migrations/001_create_gardens_table.up.sql
CREATE TABLE gardens.gardens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    location JSONB NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gardens_user_id ON gardens.gardens(user_id);
CREATE INDEX idx_gardens_location ON gardens.gardens USING GIN(location);
```

## Git Workflow and Branching Strategy

### Branch Naming Convention

```
feature/    # New features
bugfix/     # Bug fixes
hotfix/     # Production hotfixes
release/    # Release preparation
docs/       # Documentation updates
refactor/   # Code refactoring
test/       # Test-related changes

src:
feature/user-authentication
bugfix/sensor-data-validation
hotfix/weather-api-timeout
docs/api-documentation-update
```

### Git Workflow

**Main Branches**:
- `main`: Production-ready code
- `develop`: Integration branch for features

**Feature Development**:
```bash
# Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/garden-management

# Work on feature, commit regularly
git add .
git commit -m "feat: add garden creation endpoint

- Implement POST /api/v1/gardens endpoint
- Add validation for garden configuration
- Include proper error handling and logging"

# Push and create pull request
git push origin feature/garden-management
```

### Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**src**:
```
feat(api): add weather forecast integration

- Integrate OpenWeatherMap API client
- Add weather data caching with Redis
- Implement fallback to Weather.gov API

Closes #123

fix(operator): resolve sensor data parsing error

The sensor timestamp parsing was failing for UTC timestamps.
Updated parser to handle ISO 8601 format correctly.

Fixes #456

docs: update API documentation for authentication

- Add OAuth2 flow documentation
- Include API key authentication src
- Update endpoint descriptions
```

## Code Review Process and Checklists

### Code Review Requirements

**Mandatory Reviews**:
- All code changes require at least 1 reviewer approval
- Security-sensitive changes require 2 reviewer approvals
- Database migrations require DBA review
- Breaking API changes require architecture team review

### Code Review Checklist

**Functionality**:
- [ ] Code works as intended and meets requirements
- [ ] Edge cases are handled appropriately
- [ ] Error handling is comprehensive and meaningful
- [ ] No obvious bugs or logic errors

**Code Quality**:
- [ ] Code follows established style guidelines
- [ ] Functions and variables have meaningful names
- [ ] Code is well-commented and self-documenting
- [ ] No code duplication or unnecessary complexity

**Security**:
- [ ] Input validation is implemented
- [ ] No sensitive data in logs or responses
- [ ] Authentication and authorization checks are present
- [ ] SQL injection and XSS vulnerabilities addressed

**Performance**:
- [ ] No obvious performance bottlenecks
- [ ] Database queries are optimized
- [ ] Appropriate caching strategies implemented
- [ ] Resource usage is reasonable

**Testing**:
- [ ] Unit tests cover new functionality
- [ ] Integration tests verify component interactions
- [ ] Test coverage meets minimum requirements (80%)
- [ ] Tests are meaningful and maintainable

**Documentation**:
- [ ] API documentation is updated
- [ ] Code comments explain complex logic
- [ ] README files are current
- [ ] Breaking changes are documented

### Pull Request Template

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
Describe the tests you ran to verify your changes:
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes

## Additional Notes
Any additional information that reviewers should know.
```

## Version Management and Tagging

### Semantic Versioning

Follow [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH

MAJOR: Breaking changes
MINOR: New features (backward compatible)
PATCH: Bug fixes (backward compatible)
```

### Release Process

**Pre-release Versions**:
```
1.0.0-alpha.1    # Alpha release
1.0.0-beta.1     # Beta release
1.0.0-rc.1       # Release candidate
```

**Tagging Convention**:
```bash
# Create annotated tags for releases
git tag -a v1.2.0 -m "Release version 1.2.0

Features:
- Garden zone management
- Advanced sensor monitoring
- Weather API integration improvements

Bug fixes:
- Fixed sensor data validation
- Resolved authentication timeout issues"

git push origin v1.2.0
```

### Component Versioning

**API Versioning**:
- Use URL path versioning: `/api/v1/`, `/api/v2/`
- Maintain backward compatibility for at least 2 major versions
- Deprecation notices with 6-month minimum notice period

**Database Migrations**:
- Sequential numbering: `001_initial_schema.sql`, `002_add_gardens_table.sql`
- Include both up and down migrations
- Test migrations on production-like data

## Dependency Management Policies

### Go Dependencies

**Dependency Management**:
```go
// go.mod
module github.com/ryanmcafee/smart-garden-bot/api

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.10.9
    // Pin specific versions for stability
)
```

**Security Scanning**:
```bash
# Regular security scans
go list -json -m all | nancy sleuth
govulncheck ./...
```

### Node.js Dependencies

**Package Management**:
```json
{
  "engines": {
    "node": ">=18.0.0",
    "deno": ">=1.40.0"
  },
  "packageManager": "pnpm@8.0.0"
}
```

**Security Policies**:
- Regular `npm audit` or `pnpm audit` runs
- Automated dependency updates with Dependabot
- Security vulnerability disclosure process

### Container Images

**Base Image Standards**:
```dockerfile
# Use official, minimal base images
FROM node:18-alpine AS builder
FROM gcr.io/distroless/static-debian11 AS runtime

# Multi-stage builds for smaller images
# Pin specific image tags, avoid 'latest'
```

**Image Scanning**:
- Scan all images for vulnerabilities before deployment
- Use tools like Trivy or Snyk for container scanning
- Maintain updated base images

## Testing Standards

### Test Coverage Requirements

**Minimum Coverage**:
- Unit tests: 80% code coverage
- Integration tests: Critical paths covered
- End-to-end tests: User workflows covered

**Test Categories**:
```go
// Unit tests
func TestCreateGarden(t *testing.T) {
    // Test business logic in isolation
}

// Integration tests
func TestGardenAPIEndpoint(t *testing.T) {
    // Test component interactions
}

// Contract tests
func TestWeatherAPIContract(t *testing.T) {
    // Test external API contracts
}
```

### Testing Framework Standards

**Go Testing**:
```go
// Use testify for assertions
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    
    // Act
    user, err := service.CreateUser(ctx, request)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
}
```

**TypeScript Testing**:
```typescript
// Use Jest and React Testing Library
import { render, screen, fireEvent } from '@testing-library/react';
import { GardenForm } from './GardenForm';

describe('GardenForm', () => {
  it('should submit valid garden data', async () => {
    const onSubmit = jest.fn();
    render(<GardenForm onSubmit={onSubmit} />);
    
    fireEvent.change(screen.getByLabelText('Garden Name'), {
      target: { value: 'My Garden' }
    });
    
    fireEvent.click(screen.getByText('Create Garden'));
    
    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        name: 'My Garden'
      });
    });
  });
});
```

## Quality Gates

### Continuous Integration Checks

**Required Checks**:
- [ ] Linting passes (ESLint, golangci-lint)
- [ ] Unit tests pass with 80%+ coverage
- [ ] Integration tests pass
- [ ] Security scans pass
- [ ] Build succeeds without warnings
- [ ] Documentation builds successfully

**Pre-merge Requirements**:
- All CI checks must pass
- At least one code review approval
- No merge conflicts with target branch
- Branch is up-to-date with target branch

### Performance Standards

**API Performance**:
- 95th percentile response time < 200ms
- 99th percentile response time < 500ms
- Database queries < 100ms for 99% of operations
- Memory usage < 512MB per service instance

**Frontend Performance**:
- First Contentful Paint < 1.5s
- Largest Contentful Paint < 2.5s
- Cumulative Layout Shift < 0.1
- First Input Delay < 100ms

## Development Environment Standards

### Local Development Setup

**Required Tools**:
- Go 1.21+
- Node.js 18+ or Deno 1.40+
- Docker and Docker Compose
- Kubernetes (minikube or kind)
- PostgreSQL 15+
- Redis 7+

**Development Configuration**:
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: smart_garden_bot
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: devpass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

**IDE Configuration**:
```json
// .vscode/settings.json
{
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "typescript.preferences.importModuleSpecifier": "relative",
  "eslint.autoFixOnSave": true
}
```

## Security Standards

### Secure Coding Practices

**Input Validation**:
```go
// Always validate and sanitize input
func CreateGarden(req CreateGardenRequest) error {
    if err := validator.Validate(req); err != nil {
        return fmt.Errorf("invalid request: %w", err)
    }
    
    // Sanitize input
    req.Name = html.EscapeString(strings.TrimSpace(req.Name))
    return nil
}
```

**Authentication & Authorization**:
```go
// Use middleware for authentication
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        
        // Validate JWT token
        claims, err := validateJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user", claims)
        c.Next()
    }
}
```

**Secrets Management**:
- Use environment variables for configuration
- Store sensitive data in Kubernetes secrets
- Rotate API keys and passwords regularly
- Never commit secrets to version control

### Security Scanning

**Static Analysis**:
```bash
# Go security scanning
gosec ./...
go list -json -m all | nancy sleuth

# JavaScript security scanning
npm audit
snyk test
```

**Container Security**:
```bash
# Scan container images
trivy image smart-garden-bot:latest
docker scan smart-garden-bot:latest
```

## Conclusion

These development standards provide a comprehensive framework for maintaining code quality, security, and team productivity. All team members are expected to follow these guidelines, and they should be regularly reviewed and updated as the project evolves.

Regular training sessions and code review discussions will help ensure these standards are understood and consistently applied across the development team.