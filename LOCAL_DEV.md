# Local Development with Kind + Tilt

This setup provides a complete local Kubernetes development environment using Kind (Kubernetes in Docker) and Tilt for hot-reloading and orchestration.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Tilt](https://docs.tilt.dev/install.html)
- [Go](https://golang.org/doc/install) (for API development)
- [Node.js](https://nodejs.org/) (for frontend development)

## Quick Start

1. **Start the complete environment:**
   ```bash
   deno task up
   # or manually:
   # kind create cluster --config=kind-config.yaml
   # tilt up
   ```

2. **Access services:**
   - API: http://localhost:8080
   - Frontend: http://localhost:3000
   - Database: localhost:5432 (user: smartgarden, password: localdev123)
   - Development Dashboard: http://localhost:9090

3. **Run database migrations (first time):**
   ```bash
   tilt trigger db-migrate
   tilt trigger db-seed
   ```

## Architecture

The local environment consists of:

- **Kind Cluster**: 3-node Kubernetes cluster (1 control-plane, 2 workers)
- **Namespace Separation**:
  - `smart-garden-bot-api`: API server and frontend
  - `smart-garden-bot-db`: PostgreSQL database
  - `cnpg-system`: CloudNativePG operator
- **PostgreSQL**: Managed by CloudNativePG operator
- **Hot Reloading**: Tilt watches for changes and rebuilds automatically

## Development Workflow

### Making Changes

1. **API Changes**: Edit files in `src/api/` - Tilt will rebuild and redeploy automatically
2. **Frontend Changes**: Edit files in `src/frontend/` - Tilt will rebuild and reload
3. **Database Changes**: Add migrations to `src/database/migrations/` then run `tilt trigger db-migrate`

### Testing

```bash
# Run API tests
tilt trigger api-test

# Run frontend tests  
tilt trigger frontend-test
```

### Debugging

```bash
# View API logs
tilt trigger logs-api

# View database logs
tilt trigger logs-postgres

# Access database directly
kubectl exec -n smart-garden-bot-db -it deployment/smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden
```

## Configuration

### Environment Variables

API configuration is managed through Kubernetes secrets in `k8s/local/api-deployment.yaml`:

- `DB_HOST`: PostgreSQL service DNS name
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `JWT_SECRET`: JWT signing secret
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

### Database Connection

When running in Kubernetes, the API automatically detects the environment and uses:
- Host: `postgres-service.smart-garden-bot-db.svc.cluster.local`
- Port: `5432`
- Database: `smartgarden`
- User: `smartgarden`

## Cleanup

```bash
# Stop everything and cleanup
deno task down

# or manually:
# tilt down
# kind delete cluster

# Remove persistent data
tilt trigger cleanup
```

## Troubleshooting

### Common Issues

1. **Cluster creation fails**:
   ```bash
   # Check Docker is running
   docker ps
   
   # Delete existing cluster
   kind delete cluster
   kind create cluster --config=kind-config.yaml
   ```

2. **PostgreSQL not starting**:
   ```bash
   # Check CNPG operator is running
   kubectl get pods -n cnpg-system
   
   # Check cluster status
   kubectl get cluster -n smart-garden-bot-db
   ```

3. **API connection issues**:
   ```bash
   # Check API pod logs
   kubectl logs -n smart-garden-bot-api -l app=api
   
   # Check database connectivity
   kubectl exec -n smart-garden-bot-api -it deployment/api -- nslookup postgres-service.smart-garden-bot-db.svc.cluster.local
   ```

### Performance Tuning

- **Resource Limits**: Adjust in deployment manifests based on your machine specs
- **Build Cache**: Tilt uses Docker layer caching for faster rebuilds
- **Live Updates**: Go and Node.js changes are synced without full rebuilds

## Advanced Usage

### Custom Configuration

1. **Local Registry**: Tilt sets up a local Docker registry at `localhost:5001` for faster image pushes
2. **Multi-arch**: Kind cluster supports both AMD64 and ARM64 development
3. **Extensions**: Add Tilt extensions for additional functionality

### Integration with IDEs

- **VS Code**: Use Kubernetes extension for cluster management
- **IntelliJ**: Configure remote debugging for Go API
- **Database Tools**: Connect to `localhost:5432` for database operations

## Next Steps

- Add Redis for caching
- Set up distributed tracing with Jaeger
- Add Prometheus monitoring
- Configure Istio service mesh
- Set up CI/CD pipelines