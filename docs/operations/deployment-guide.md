# Smart Garden Bot Deployment Guide

## Overview
This guide provides comprehensive instructions for deploying the Smart Garden Bot platform using GitOps principles with ArgoCD and Kubernetes.

## Prerequisites

### Infrastructure Requirements
- **Kubernetes Cluster**: v1.25+ with 3+ nodes
- **Storage**: CSI-compatible storage class (gp3 recommended)
- **Load Balancer**: AWS ALB, NGINX, or cloud provider LB
- **DNS**: Route53 or equivalent with wildcard support
- **Backup Storage**: S3-compatible object storage

### Tools Required
```bash
# Install required CLI tools
# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Helm
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update && sudo apt-get install helm

# ArgoCD CLI
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
```

### Access Requirements
- Kubernetes cluster admin access
- GitHub repository access
- Container registry access (GHCR)
- AWS credentials (for S3 backups)
- Domain DNS management access

## Deployment Phases

## Phase 1: Cluster Preparation

### 1.1 Namespace Creation
```bash
# Create primary namespaces
kubectl apply -f k8s/base/namespace.yaml

# Verify namespaces
kubectl get namespaces | grep smart-garden-bot
```

### 1.2 Install ArgoCD
```bash
# Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for ArgoCD to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s

# Get initial admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port forward to access UI (optional)
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

### 1.3 Configure ArgoCD
```bash
# Login to ArgoCD
argocd login localhost:8080 --username admin --password [PASSWORD_FROM_ABOVE] --insecure

# Add GitHub repository
argocd repo add https://github.com/ryanmcafee/smart-garden-bot/smart-garden-bot.git

# Add Helm repositories
argocd repo add https://charts.bitnami.com/bitnami --type helm --name bitnami
argocd repo add https://charts.jetstack.io --type helm --name jetstack
argocd repo add https://prometheus-community.github.io/helm-charts --type helm --name prometheus-community
argocd repo add https://grafana.github.io/helm-charts --type helm --name grafana
```

## Phase 2: Infrastructure Components

### 2.1 Create AppProject
```bash
# Apply Smart Garden Bot project
kubectl apply -f gitops/projects/smart-garden-bot.yaml

# Verify project creation
argocd proj list
```

### 2.2 Deploy Infrastructure Applications
```bash
# Deploy cert-manager
kubectl apply -f gitops/applications/infrastructure.yaml

# Wait for cert-manager
kubectl wait --for=condition=ready pod -l app=cert-manager -n cert-manager --timeout=300s

# Deploy NGINX ingress controller
kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=controller -n ingress-nginx --timeout=300s

# Deploy CloudNativePG operator
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cloudnative-pg -n cnpg-system --timeout=300s
```

### 2.3 Configure TLS Certificates
```bash
# Apply certificate issuers
kubectl apply -f k8s/base/certificates.yaml

# Verify ClusterIssuers
kubectl get clusterissuer

# Check certificate creation
kubectl get certificate -A
```

## Phase 3: Secrets Management

### 3.1 Create Core Secrets
```bash
# Database credentials
kubectl create secret generic smart-garden-bot-postgres-auth \
  --from-literal=username=smartgarden \
  --from-literal=password="$(openssl rand -base64 32)" \
  -n smart-garden-bot

# Auth0 secrets
kubectl create secret generic smart-garden-bot-auth-secret \
  --from-literal=auth0-client-secret="YOUR_AUTH0_CLIENT_SECRET" \
  --from-literal=nextauth-secret="$(openssl rand -base64 32)" \
  -n smart-garden-bot

# Weather API secrets
kubectl create secret generic smart-garden-bot-weather-secret \
  --from-literal=openweather-api-key="YOUR_OPENWEATHER_API_KEY" \
  --from-literal=weatherapi-key="YOUR_WEATHERAPI_KEY" \
  -n smart-garden-bot

# Stripe secrets
kubectl create secret generic smart-garden-bot-stripe-secret \
  --from-literal=stripe-secret-key="YOUR_STRIPE_SECRET_KEY" \
  --from-literal=stripe-webhook-secret="YOUR_STRIPE_WEBHOOK_SECRET" \
  --from-literal=stripe-publishable-key="YOUR_STRIPE_PUBLISHABLE_KEY" \
  -n smart-garden-bot

# Backup credentials
kubectl create secret generic postgres-backup-credentials \
  --from-literal=ACCESS_KEY_ID="YOUR_AWS_ACCESS_KEY" \
  --from-literal=SECRET_ACCESS_KEY="YOUR_AWS_SECRET_KEY" \
  -n smart-garden-bot
```

### 3.2 Container Registry Access
```bash
# Create image pull secret for GHCR
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN \
  --docker-email=YOUR_EMAIL \
  -n smart-garden-bot

kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN \
  --docker-email=YOUR_EMAIL \
  -n smart-garden-bot-system
```

## Phase 4: Database Deployment

### 4.1 Deploy PostgreSQL Cluster
```bash
# Apply Custom Resource Definitions
kubectl apply -f k8s/crds/garden-controller-crd.yaml

# Deploy PostgreSQL cluster
kubectl apply -f k8s/base/postgresql-cluster.yaml

# Wait for database cluster to be ready
kubectl wait --for=condition=Ready cluster -n smart-garden-bot smart-garden-bot-postgres --timeout=600s

# Verify cluster status
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres

# Test database connectivity
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- pg_isready -U smartgarden
```

### 4.2 Initialize Database Schema
```bash
# Connect to database and verify schemas
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT schema_name FROM information_schema.schemata 
WHERE schema_name IN ('auth', 'gardens', 'sensors', 'billing', 'analytics');
"

# Verify extensions
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT name, default_version, installed_version 
FROM pg_available_extensions 
WHERE name IN ('uuid-ossp', 'pgcrypto', 'vector', 'timescaledb');
"
```

## Phase 5: Application Deployment

### 5.1 Deploy RBAC Configuration
```bash
# Apply RBAC for operator and API server
kubectl apply -f k8s/rbac/operator-rbac.yaml

# Verify service accounts
kubectl get serviceaccount -n smart-garden-bot
kubectl get serviceaccount -n smart-garden-bot-system
```

### 5.2 Deploy Application Stack
```bash
# For development environment
kubectl apply -f gitops/applications/smart-garden-bot-development.yaml

# For production environment (requires manual sync)
kubectl apply -f gitops/applications/smart-garden-bot-production.yaml

# Monitor deployment progress
argocd app list
argocd app get smart-garden-bot-production
```

### 5.3 Verify Application Health
```bash
# Check pod status
kubectl get pods -n smart-garden-bot
kubectl get pods -n smart-garden-bot-system

# Check service endpoints
kubectl get services -n smart-garden-bot

# Test health endpoints
kubectl exec -n smart-garden-bot deployment/smart-garden-bot-api -- curl -f http://localhost:8080/health
kubectl exec -n smart-garden-bot deployment/smart-garden-bot-api -- curl -f http://localhost:8080/ready
```

## Phase 6: Monitoring and Observability

### 6.1 Deploy Monitoring Stack
```bash
# Deploy Prometheus and Grafana
kubectl apply -f gitops/applications/monitoring.yaml

# Wait for monitoring components
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n monitoring --timeout=600s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=grafana -n monitoring --timeout=600s

# Apply monitoring configurations
kubectl apply -f k8s/base/monitoring.yaml
```

### 6.2 Configure Grafana
```bash
# Get Grafana admin password
kubectl get secret -n monitoring kube-prometheus-stack-grafana -o jsonpath="{.data.admin-password}" | base64 -d

# Port forward to access Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80

# Access Grafana at http://localhost:3000
# Username: admin
# Password: [from above command]
```

### 6.3 Verify Metrics Collection
```bash
# Check ServiceMonitors
kubectl get servicemonitor -n smart-garden-bot
kubectl get servicemonitor -n smart-garden-bot-system

# Verify Prometheus targets
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
# Access http://localhost:9090/targets
```

## Phase 7: Network and Security

### 7.1 Apply Network Policies
```bash
# Deploy network policies
kubectl apply -f k8s/base/network-policies.yaml

# Verify network policies
kubectl get networkpolicy -n smart-garden-bot
kubectl get networkpolicy -n smart-garden-bot-system
```

### 7.2 Configure Ingress
```bash
# Check ingress controller status
kubectl get pods -n ingress-nginx

# Verify ingress resources
kubectl get ingress -n smart-garden-bot

# Test external access
curl -k https://api.smartgardenbot.io/health
curl -k https://app.smartgardenbot.io
```

### 7.3 Optional: Deploy Istio Service Mesh
```bash
# Install Istio
curl -L https://istio.io/downloadIstio | sh -
export PATH=$PWD/istio-*/bin:$PATH
istioctl install --set values.defaultRevision=default

# Label namespace for injection
kubectl label namespace smart-garden-bot istio-injection=enabled

# Apply Istio configurations
kubectl apply -f k8s/base/istio-config.yaml

# Verify mesh status
istioctl proxy-status
```

## Phase 8: Validation and Testing

### 8.1 End-to-End Health Check
```bash
#!/bin/bash
# E2E health check script

echo "🔍 Starting end-to-end health check..."

# API health checks
echo "Testing API endpoints..."
curl -f https://api.smartgardenbot.io/health || exit 1
curl -f https://api.smartgardenbot.io/ready || exit 1
curl -f https://api.smartgardenbot.io/api/v1/openapi.json || exit 1

# Web application check
echo "Testing web application..."
curl -f https://app.smartgardenbot.io || exit 1

# Database connectivity
echo "Testing database connectivity..."
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- pg_isready -U smartgarden || exit 1

# Operator status
echo "Testing operator..."
kubectl get deployment -n smart-garden-bot-system smart-garden-bot-operator || exit 1

# Monitoring stack
echo "Testing monitoring..."
kubectl get pods -n monitoring -l app.kubernetes.io/name=prometheus | grep Running || exit 1

echo "✅ All health checks passed!"
```

### 8.2 Load Testing
```bash
# Install k6 for load testing
curl -L https://github.com/grafana/k6/releases/download/v0.45.0/k6-v0.45.0-linux-amd64.tar.gz | tar -xz
sudo mv k6-v0.45.0-linux-amd64/k6 /usr/local/bin/

# Basic load test
k6 run - <<EOF
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 10,
  duration: '30s',
};

export default function () {
  let response = http.get('https://api.smartgardenbot.io/health');
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
EOF
```

## Phase 9: Backup and Recovery Setup

### 9.1 Verify Backup Configuration
```bash
# Check backup configuration
kubectl get backup -n smart-garden-bot
kubectl get scheduledbackup -n smart-garden-bot

# Trigger manual backup
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Backup
metadata:
  name: smart-garden-bot-postgres-manual-$(date +%Y%m%d-%H%M%S)
  namespace: smart-garden-bot
spec:
  cluster:
    name: smart-garden-bot-postgres
EOF

# Monitor backup progress
kubectl get backup -n smart-garden-bot -w
```

### 9.2 Test Backup Restoration
```bash
# Create test restoration (in separate namespace)
kubectl create namespace backup-test

kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: test-restore
  namespace: backup-test
spec:
  instances: 1
  
  bootstrap:
    recovery:
      backup:
        name: smart-garden-bot-postgres-manual-$(date +%Y%m%d -d "1 hour ago")
  
  storage:
    size: 20Gi
    storageClass: gp3
EOF

# Wait and verify
kubectl wait --for=condition=Ready cluster -n backup-test test-restore --timeout=600s
kubectl exec -n backup-test test-restore-1 -- psql -U smartgarden -d smartgarden -c "SELECT count(*) FROM auth.users;"

# Cleanup test
kubectl delete namespace backup-test
```

## Troubleshooting Common Issues

### ArgoCD Application Out of Sync
```bash
# Check application status
argocd app get smart-garden-bot-production

# Force sync
argocd app sync smart-garden-bot-production

# Check for drift
argocd app diff smart-garden-bot-production
```

### Database Connection Issues
```bash
# Check cluster status
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres

# Verify service endpoints
kubectl get endpoints -n smart-garden-bot smart-garden-bot-postgres-rw

# Test connectivity from API pod
kubectl exec -n smart-garden-bot deployment/smart-garden-bot-api -- nc -zv smart-garden-bot-postgres-rw 5432
```

### Certificate Issues
```bash
# Check certificate status
kubectl get certificate -n smart-garden-bot

# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager

# Debug certificate request
kubectl describe certificaterequest -n smart-garden-bot
```

### Ingress Issues
```bash
# Check ingress controller logs  
kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller

# Verify ingress configuration
kubectl describe ingress -n smart-garden-bot smart-garden-bot

# Test from ingress controller
kubectl exec -n ingress-nginx deployment/ingress-nginx-controller -- curl -H "Host: api.smartgardenbot.io" http://smart-garden-bot-api.smart-garden-bot:8080/health
```

## Post-Deployment Checklist

### Security Verification
- [ ] All secrets properly configured
- [ ] Network policies active
- [ ] TLS certificates valid
- [ ] RBAC permissions minimal
- [ ] Container images signed
- [ ] Vulnerability scans passed

### Performance Validation
- [ ] Response times < 200ms
- [ ] Database queries optimized
- [ ] Resource limits appropriate
- [ ] Autoscaling configured
- [ ] Load balancing working

### Monitoring Setup
- [ ] All metrics collecting
- [ ] Dashboards configured
- [ ] Alerts triggered correctly
- [ ] Log aggregation working
- [ ] Distributed tracing active

### Backup Verification
- [ ] Scheduled backups running
- [ ] Backup restoration tested
- [ ] Retention policies set
- [ ] Cross-region replication
- [ ] Disaster recovery plan updated

### Documentation
- [ ] Deployment notes recorded
- [ ] Access credentials documented
- [ ] Monitoring runbooks updated
- [ ] Escalation procedures current
- [ ] Architecture diagrams updated

## Maintenance Tasks

### Daily
- Monitor application health
- Check backup status
- Review error logs
- Verify certificate expiry

### Weekly  
- Review resource usage
- Update security patches
- Test backup restoration
- Review performance metrics

### Monthly
- Update dependencies
- Review capacity planning
- Conduct security scan
- Update documentation

### Quarterly
- Disaster recovery drill
- Performance optimization
- Security audit
- Technology stack review

## Support Contacts

### Internal Team
- **Platform Team**: @platform-team
- **Database Administrator**: @dba-team  
- **Security Team**: @security-team
- **DevOps Team**: @devops-team

### External Support
- **AWS Support**: Enterprise Support Portal
- **GitHub Support**: Enterprise Support
- **DNS Provider**: Support Ticket System
- **Certificate Authority**: Let's Encrypt Community

## Conclusion

This deployment guide provides comprehensive instructions for deploying the Smart Garden Bot platform. Following these procedures ensures a secure, scalable, and maintainable deployment that follows cloud-native best practices.

For questions or issues not covered in this guide, please refer to the troubleshooting section or contact the platform team.