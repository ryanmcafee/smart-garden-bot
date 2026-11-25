# Runbook: API Server Down

## Overview
This runbook provides step-by-step instructions for diagnosing and resolving API server outages.

## Alert Details
- **Alert Name**: ApiDown
- **Severity**: Critical
- **Threshold**: API server unreachable for 5+ minutes
- **Impact**: Complete service outage, users cannot access the application

## Initial Response (First 5 minutes)

### 1. Acknowledge Alert
```bash
# Acknowledge in PagerDuty/alerting system
# Post in #incidents Slack channel
```

### 2. Quick Health Check
```bash
# Check if API endpoints are responding
curl -f https://api.smartgardenbot.com/health
curl -f https://api.smartgardenbot.com/ready

# Check ArgoCD for deployment status
kubectl get applications -n argocd smart-garden-bot-production
```

### 3. Check Pod Status
```bash
# List API server pods
kubectl get pods -n smart-garden-bot -l app.kubernetes.io/component=api

# Check pod details
kubectl describe pods -n smart-garden-bot -l app.kubernetes.io/component=api

# Check recent events
kubectl get events -n smart-garden-bot --sort-by='.lastTimestamp' | grep api
```

## Diagnostic Steps

### 4. Check Resource Usage
```bash
# CPU and memory usage
kubectl top pods -n smart-garden-bot -l app.kubernetes.io/component=api

# Check HPA status
kubectl get hpa -n smart-garden-bot smart-garden-bot-api

# Check node resources
kubectl describe nodes
```

### 5. Review Logs
```bash
# API server logs
kubectl logs -n smart-garden-bot -l app.kubernetes.io/component=api --tail=100

# Check for crash loops
kubectl logs -n smart-garden-bot -l app.kubernetes.io/component=api --previous

# Check ingress logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/component=controller --tail=50 | grep smart-garden-bot
```

### 6. Database Connectivity
```bash
# Check PostgreSQL cluster status
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres

# Check database connections
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "SELECT count(*) FROM pg_stat_activity;"

# Test database connectivity from API pod
kubectl exec -n smart-garden-bot deployment/smart-garden-bot-api -- /bin/sh -c "pg_isready -h smart-garden-bot-postgres-rw -p 5432"
```

### 7. External Dependencies
```bash
# Check Auth0 connectivity
kubectl exec -n smart-garden-bot deployment/smart-garden-bot-api -- /bin/sh -c "curl -s -o /dev/null -w '%{http_code}' https://smartgarden.auth0.com/.well-known/jwks.json"

# Check weather API connectivity
kubectl exec -n smart-garden-bot deployment/smart-garden-bot-api -- /bin/sh -c "curl -s -o /dev/null -w '%{http_code}' https://api.openweathermap.org"
```

## Common Resolutions

### Scenario 1: Pod Crash Loop
**Symptoms**: Pods repeatedly restarting, CrashLoopBackOff status

**Resolution**:
```bash
# Check recent deployment
kubectl rollout history deployment/smart-garden-bot-api -n smart-garden-bot

# Rollback to previous version if recent deployment
kubectl rollout undo deployment/smart-garden-bot-api -n smart-garden-bot

# Wait for rollout to complete
kubectl rollout status deployment/smart-garden-bot-api -n smart-garden-bot
```

### Scenario 2: Resource Exhaustion
**Symptoms**: High CPU/memory usage, OOMKilled events

**Resolution**:
```bash
# Scale up replicas temporarily
kubectl scale deployment smart-garden-bot-api -n smart-garden-bot --replicas=10

# Update resource limits
kubectl patch deployment smart-garden-bot-api -n smart-garden-bot -p '{"spec":{"template":{"spec":{"containers":[{"name":"api","resources":{"limits":{"memory":"2Gi","cpu":"1000m"}}}]}}}}'
```

### Scenario 3: Database Connection Issues
**Symptoms**: Connection timeout errors, database unavailable

**Resolution**:
```bash
# Check PostgreSQL cluster
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres -o yaml

# Restart API pods to refresh connections
kubectl rollout restart deployment/smart-garden-bot-api -n smart-garden-bot

# If database is down, check backup restoration procedures
```

### Scenario 4: Configuration Issues
**Symptoms**: Missing environment variables, secret mounting errors

**Resolution**:
```bash
# Check secrets
kubectl get secrets -n smart-garden-bot

# Verify secret content (don't log actual values)
kubectl describe secret smart-garden-bot-auth-secret -n smart-garden-bot

# Check configmap
kubectl get configmap smart-garden-bot-config -n smart-garden-bot -o yaml
```

## Escalation Procedures

### When to Escalate
- Issue not resolved within 15 minutes
- Data corruption suspected
- Multiple systems affected
- Infrastructure-level problems

### Escalation Contacts
1. **Platform Team Lead**: @platform-lead
2. **Database Administrator**: @dba-oncall
3. **Infrastructure Team**: @infra-team
4. **Security Team**: @security-team (if security-related)

## Post-Incident Actions

### 1. Verify Service Recovery
```bash
# Comprehensive health check
curl -f https://api.smartgardenbot.com/health
curl -f https://api.smartgardenbot.com/ready
curl -f https://api.smartgardenbot.com/api/v1/openapi.json

# Check metrics are being collected
# View Grafana dashboard for API server metrics
```

### 2. Update Status Page
- Mark incident as resolved
- Post summary of issue and resolution
- Update ETA if any follow-up work needed

### 3. Document Root Cause
- Create detailed incident report
- Update this runbook if procedures changed
- Schedule post-mortem if significant outage

## Prevention Measures

### Monitoring Improvements
- Enhance alerting thresholds based on incident learnings
- Add synthetic monitoring for critical endpoints
- Monitor database connection pool usage

### Infrastructure Hardening
- Review resource limits and requests
- Implement circuit breakers for external dependencies
- Improve graceful shutdown handling

### Testing
- Regular disaster recovery drills
- Chaos engineering exercises
- Load testing to identify breaking points

## Related Runbooks
- [Database Recovery](database-recovery.md)
- [Rolling Back Deployments](rollback-deployments.md)  
- [Network Troubleshooting](network-troubleshooting.md)
- [Certificate Issues](certificate-issues.md)

## Revision History
- v1.0 - Initial version
- v1.1 - Added database connectivity checks
- v1.2 - Enhanced escalation procedures