# Smart Garden Bot - Code Style and Conventions

## General Conventions
- Use TypeScript for all scripting (no Bash/Python as per CLAUDE.md)
- Semantic commit messages (feat:, fix:, docs:, etc.)
- Modern CLI tools: `rg`, `fd`, `bat`, `eza`, `jq`, `yq`, `fzf`, `delta`
- Never use legacy tools: `grep`, `find`, `cat`, `ls`, `df`, `top`, `xxd`

## Go Conventions (API/Operator)
- Go version: 1.18+
- Module path: github.com/ryanmcafee/smart-garden-bot/api
- Framework: Gin for HTTP routing
- Database: sqlx for SQL operations, lib/pq for PostgreSQL
- Error handling: Structured error responses
- Testing: Use testify, table-driven tests, target 95%+ coverage
- Logging: Structured JSON logging with correlation IDs
- Metrics: Prometheus metrics with custom business metrics

## TypeScript/React Conventions (Frontend)
- TypeScript version: 5.7+
- React version: 18.3+
- Next.js version: 15.4+ with App Router
- Node.js version: 24+
- State Management: Zustand and React Query (@tanstack/react-query)
- Styling: Tailwind CSS with Radix UI components
- Forms: React Hook Form with Zod validation
- Testing: Jest with Testing Library, test naming should be descriptive
- File structure: App Router structure with components/, lib/, hooks/, store/

## Database Conventions
- PostgreSQL with pgvector extension
- Multi-schema design (auth, gardens, sensors, analytics)
- Time-series data with automatic partitioning
- CloudNativePG operator for Kubernetes deployment

## Container & Kubernetes
- Multi-stage Docker builds for all components
- Helm charts for deployment
- Namespace separation (smart-garden-bot-api, smart-garden-bot-db, cnpg-system)
- Health checks: liveness, readiness, startup probes
- Resource limits and requests defined

## Development Workflow
- Kind + Tilt for local development with hot-reloading
- GitHub Actions for CI/CD
- ArgoCD for GitOps continuous deployment
- Pre-commit hooks for code quality
- Integration tests using testcontainers (Go) and Pact