# Smart Garden Bot - Task Completion Guidelines

## When a Task is Completed

### Code Quality Checks
1. **Linting**: Run appropriate linters for the modified code
   - Go: `golangci-lint`, `gofmt`, `go vet`
   - TypeScript: `npm run lint` (ESLint)
   - SQL: `sqlfluff` for SQL linting

2. **Type Checking**: Ensure type safety
   - Go: `go build` to check compilation
   - TypeScript: `npm run type-check` or `tsc --noEmit`

3. **Formatting**: Apply consistent code formatting
   - Go: `gofmt` (automatically applied by most editors)
   - TypeScript: Prettier (via ESLint config)

### Testing Requirements
1. **Unit Tests**: Write and run unit tests for new functionality
   - Go: `go test ./...` (table-driven tests preferred)
   - TypeScript: `npm test` (Jest with Testing Library)

2. **Integration Tests**: Run integration tests if applicable
   - API: Use testcontainers for database integration
   - Frontend: Mock Service Worker for API integration

3. **Test Coverage**: Aim for 95%+ coverage for Go code

### Build Verification
1. **Local Builds**: Ensure components build successfully
   - API: `go build` or Docker build
   - Frontend: `npm run build`

2. **Docker Images**: Build and test Docker images if modified
   - `task build:api` or `task build:frontend`

### Development Environment
1. **Local Testing**: Verify changes work in local development
   - `task up` to start local environment
   - `task test:local` to run development setup tests
   - Manual testing of modified functionality

2. **Dependencies**: Update dependencies if needed
   - Go: `go mod tidy`
   - Node.js: Update package.json and run `npm install`

### Documentation
1. **Code Comments**: Add comments for complex logic (when appropriate)
2. **API Documentation**: Update OpenAPI specs if API changes
3. **README Updates**: Update relevant documentation if needed

### Git Workflow
1. **Commit Messages**: Use semantic commit format
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation
   - `refactor:` for code refactoring
   - `test:` for test additions/changes

2. **Pre-commit Hooks**: Run pre-commit checks before committing
3. **Branch Management**: Use feature branches, never commit directly to main

### Production Readiness
1. **Security**: Check for security vulnerabilities
2. **Performance**: Consider performance implications
3. **Monitoring**: Ensure proper logging and metrics are in place
4. **Configuration**: Verify environment-specific configuration