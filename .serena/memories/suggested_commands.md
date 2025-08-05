# Smart Garden Bot - Suggested Commands

## Development Commands
- `task up` - Start all services (Creates Kind cluster and starts Tilt)
- `task down` - Stop all services (Stops Tilt and deletes Kind cluster)
- `task dev` - Development mode (Starts Tilt with streaming logs)
- `task init` - Initialize environment (Create cluster and install CloudNativePG)

## Testing Commands
- `task test` - Run all tests (API + Frontend)
- `task test:api` - Run API tests (Go) in src/api directory
- `task test:frontend` - Run frontend tests (Node.js) in src/frontend directory
- `task test:local` - Test local development setup using Deno script

## Build Commands
- `task build` - Production build (Build all components)
- `task build:api` - Build API Docker image
- `task build:frontend` - Build frontend Docker image

## Utility Commands
- `task clean` - Clean up local development environment
- `task logs` - View application logs
- `task status` - Check status of all services

## Frontend Commands (in src/frontend/)
- `npm run dev` - Start Next.js development server
- `npm run build` - Build for production
- `npm run lint` - Run ESLint
- `npm run type-check` - Run TypeScript compiler checks
- `npm test` - Run Jest tests
- `npm run test:coverage` - Run tests with coverage

## API Commands (in src/api/)
- `go run main.go` - Start API server
- `go test ./...` - Run Go tests
- `go mod download` - Download dependencies

## Kubernetes Commands
- `kubectl get namespaces` - Check namespaces
- `kubectl logs -n smart-garden-bot-api -l app=api --tail=100 -f` - View API logs
- `kind get clusters` - List Kind clusters
- `tilt get uiresource` - Check Tilt resources