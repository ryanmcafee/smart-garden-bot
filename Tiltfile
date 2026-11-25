# Tiltfile for Smart Garden Bot Local Development
# This file orchestrates the entire local development environment

# Load extensions
load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://restart_process', 'docker_build_with_restart')

# =============================================================================
# Configuration
# =============================================================================

# Set up local registry (optional but recommended for faster builds)
local_resource(
  'registry',
  'docker run -d --restart=always -p 5001:5000 --name kind-registry registry:2 || true',
  readiness_probe=probe(
    exec=exec_action(['docker', 'exec', 'kind-registry', 'registry', '--version']),
    initial_delay_secs=2,
    timeout_secs=10
  )
)

# Connect registry to kind network
local_resource(
  'registry-connect',
  'docker network connect kind kind-registry || true',
  resource_deps=['registry']
)

# =============================================================================
# Infrastructure Setup
# =============================================================================

# Apply namespaces first
k8s_yaml('k8s/local/namespace.yaml')

# Install CloudNativePG operator
local_resource(
  'cnpg-operator',
  'kubectl apply --server-side -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.24/releases/cnpg-1.24.1.yaml && sleep 10',
  resource_deps=['registry-connect']
)

# Deploy PostgreSQL cluster
k8s_yaml('k8s/local/postgres-cluster.yaml')

# Wait for postgres cluster to be ready and set up port forwarding
local_resource(
  'postgres-cluster',
  'kubectl wait --for=condition=Ready cluster/smart-garden-bot-postgres -n smart-garden-bot-db --timeout=300s && kubectl port-forward -n smart-garden-bot-db svc/postgres-service 5432:5432 > /dev/null 2>&1 &',
  resource_deps=['cnpg-operator'],
  labels=["database"]
)

# =============================================================================
# Application Services
# =============================================================================

# Build and deploy API
docker_build(
  'kind-registry:5000/smart-garden-bot/api:local',
  'src/api',
  dockerfile='src/api/Dockerfile',
  only=['./'],
  live_update=[
    sync('src/api', '/app'),
    run('go build -o /app/main /app/main.go', trigger=['**/*.go'])
  ]
)

k8s_yaml('k8s/local/api-deployment.yaml')
k8s_resource(
  'api',
  resource_deps=['postgres-cluster'],
  port_forwards='8080:8080',
  labels=["backend"]
)

# Build and deploy Frontend (if needed)
docker_build(
  'kind-registry:5000/smart-garden-bot/frontend:local',
  'src/frontend',
  dockerfile='src/frontend/Dockerfile',
  only=['./'],
  live_update=[
    sync('src/frontend', '/app'),
    run('npm run build', trigger=['**/*.tsx', '**/*.ts', '**/*.json'])
  ]
)

# Frontend deployment
k8s_yaml(blob("""
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: smart-garden-bot-api
  labels:
    app: frontend
    component: frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
        component: frontend
    spec:
      containers:
      - name: frontend
        image: kind-registry:5000/smart-garden-bot/frontend:local
        ports:
        - containerPort: 3000
          name: http
        env:
        - name: NEXT_PUBLIC_API_URL
          value: "http://localhost:8080"
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: smart-garden-bot-api
  labels:
    app: frontend
spec:
  selector:
    app: frontend
  ports:
  - name: http
    port: 3000
    targetPort: http
  type: ClusterIP
"""))

k8s_resource(
  'frontend',
  resource_deps=['api'],
  port_forwards='3000:3000',
  labels=["frontend"]
)

# =============================================================================
# Development Tools and Utilities
# =============================================================================

# Database migration runner
local_resource(
  'db-migrate',
  'sleep 10 && kubectl exec -n smart-garden-bot-db deployment/smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -f /dev/stdin < src/database/migrations/001_initial_schema.up.sql',
  resource_deps=['postgres-cluster'],
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL
)

# Database seed data
local_resource(
  'db-seed',
  'sleep 5 && kubectl exec -n smart-garden-bot-db deployment/smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -f /dev/stdin < src/database/migrations/002_seed_data.up.sql',
  resource_deps=['db-migrate'],
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL
)

# API tests
local_resource(
  'api-test',
  'cd src/api && go test ./...',
  deps=['src/api'],
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
  labels=["testing"]
)

# Frontend tests
local_resource(
  'frontend-test',
  'cd src/frontend && npm test -- --passWithNoTests',
  deps=['src/frontend'],
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
  labels=["testing"]
)

# =============================================================================
# Monitoring and Observability (Optional)
# =============================================================================

# Simple monitoring dashboard
k8s_yaml(blob("""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: monitoring-dashboard
  namespace: smart-garden-bot-api
data:
  index.html: |
    <!DOCTYPE html>
    <html>
    <head>
        <title>Smart Garden Bot - Local Development</title>
        <style>
            body { font-family: Arial, sans-serif; margin: 20px; }
            .service { margin: 10px 0; padding: 10px; border: 1px solid #ccc; }
            .status { color: green; font-weight: bold; }
        </style>
    </head>
    <body>
        <h1>Smart Garden Bot - Local Development Environment</h1>
        <div class="service">
            <h3>API Service</h3>
            <p>URL: <a href="http://localhost:8080/health">http://localhost:8080/health</a></p>
            <p class="status">Running on Port 8080</p>
        </div>
        <div class="service">
            <h3>Frontend</h3>
            <p>URL: <a href="http://localhost:3000">http://localhost:3000</a></p>
            <p class="status">Running on Port 3000</p>
        </div>
        <div class="service">
            <h3>PostgreSQL Database</h3>
            <p>Connection: localhost:5432</p>
            <p>Database: smartgarden</p>
            <p>User: smartgarden</p>
            <p class="status">Available via port-forward</p>
        </div>
    </body>
    </html>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: monitoring-dashboard
  namespace: smart-garden-bot-api
spec:
  replicas: 1
  selector:
    matchLabels:
      app: monitoring-dashboard
  template:
    metadata:
      labels:
        app: monitoring-dashboard
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        volumeMounts:
        - name: dashboard-config
          mountPath: /usr/share/nginx/html
      volumes:
      - name: dashboard-config
        configMap:
          name: monitoring-dashboard
---
apiVersion: v1
kind: Service
metadata:
  name: monitoring-dashboard-service
  namespace: smart-garden-bot-api
spec:
  selector:
    app: monitoring-dashboard
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP
"""))

k8s_resource(
  'monitoring-dashboard',
  port_forwards='9090:80',
  labels=["monitoring"]
)

# =============================================================================
# Development Helpers
# =============================================================================

# Quick logs access
local_resource(
  'logs-api',
  'kubectl logs -n smart-garden-bot-api -l app=api --tail=100 -f',
  resource_deps=['api'],
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
  labels=["logs"]
)

local_resource(
  'logs-postgres',
  'kubectl logs -n smart-garden-bot-db -l cnpg.io/cluster=smart-garden-bot-postgres --tail=100 -f',
  resource_deps=['postgres-cluster'],
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
  labels=["logs"]
)

# =============================================================================
# Cleanup and Utilities
# =============================================================================

# Cleanup script
local_resource(
  'cleanup',
  '''
  kubectl delete namespace smart-garden-bot-api --ignore-not-found=true
  kubectl delete namespace smart-garden-bot-db --ignore-not-found=true
  docker stop kind-registry || true
  docker rm kind-registry || true
  ''',
  auto_init=False,
  trigger_mode=TRIGGER_MODE_MANUAL,
  labels=["cleanup"]
)

print("""
🌱 Smart Garden Bot Local Development Environment 🌱

Tilt is now managing your local Kubernetes development environment!

Key services:
• API: http://localhost:8080
• Frontend: http://localhost:3000  
• Database: localhost:5432 (postgres)
• Dashboard: http://localhost:9090

Manual commands available:
• tilt trigger db-migrate    - Run database migrations
• tilt trigger db-seed       - Seed database with test data
• tilt trigger api-test      - Run API tests
• tilt trigger frontend-test - Run frontend tests
• tilt trigger logs-api      - Stream API logs
• tilt trigger logs-postgres - Stream PostgreSQL logs
• tilt trigger cleanup       - Clean up all resources

Happy coding! 🚀
""")