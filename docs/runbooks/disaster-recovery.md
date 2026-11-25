# Disaster Recovery Plan

## Overview
This document outlines the comprehensive disaster recovery procedures for the Smart Garden Bot platform, covering complete infrastructure failure scenarios.

## Recovery Time Objectives (RTO) and Recovery Point Objectives (RPO)

| Component | RTO | RPO | Priority |
|-----------|-----|-----|----------|
| API Server | 15 minutes | 0 | Critical |
| Web Application | 10 minutes | 0 | Critical |
| Database | 30 minutes | 15 minutes | Critical |
| Operator | 20 minutes | 0 | High |
| Monitoring | 45 minutes | 60 minutes | Medium |

## Disaster Scenarios

### Scenario 1: Complete Kubernetes Cluster Failure

**Symptoms**: Entire cluster unreachable, all services down

**Pre-requisites**:
- New Kubernetes cluster provisioned
- kubectl access configured
- ArgoCD installed and configured
- DNS pointing to new cluster

**Recovery Steps**:

#### Step 1: Cluster Setup (0-10 minutes)
```bash
# Verify cluster access
kubectl cluster-info

# Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for ArgoCD to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s
```

#### Step 2: Infrastructure Components (10-20 minutes)
```bash
# Apply infrastructure applications
kubectl apply -f gitops/applications/infrastructure.yaml

# Wait for cert-manager
kubectl wait --for=condition=ready pod -l app=cert-manager -n cert-manager --timeout=300s

# Wait for nginx-ingress
kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=controller -n ingress-nginx --timeout=300s

# Wait for CloudNativePG operator
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cloudnative-pg -n cnpg-system --timeout=300s
```

#### Step 3: Database Recovery (20-50 minutes)
```bash
# Create database cluster with backup restoration
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: smart-garden-bot-postgres
  namespace: smart-garden-bot
spec:
  instances: 3
  
  bootstrap:
    recovery:
      backup:
        name: smart-garden-bot-postgres-backup-latest
  
  storage:
    size: 100Gi
    storageClass: gp3
  
  postgresql:
    parameters:
      max_connections: "200"
      shared_buffers: "256MB"
      
  backup:
    retentionPolicy: "30d"
    barmanObjectStore:
      destinationPath: "s3://smart-garden-bot-backups/postgres"
      s3Credentials:
        accessKeyId:
          name: postgres-backup-credentials
          key: ACCESS_KEY_ID
        secretAccessKey:
          name: postgres-backup-credentials
          key: SECRET_ACCESS_KEY
EOF

# Wait for database recovery
kubectl wait --for=condition=Ready cluster -n smart-garden-bot smart-garden-bot-postgres --timeout=1800s

# Verify database integrity
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "SELECT count(*) FROM auth.users;"
```

#### Step 4: Application Deployment (50-65 minutes)
```bash
# Deploy production application
kubectl apply -f gitops/applications/smart-garden-bot-production.yaml

# Wait for API server
kubectl wait --for=condition=available deployment -l app.kubernetes.io/component=api -n smart-garden-bot --timeout=600s

# Wait for web application
kubectl wait --for=condition=available deployment -l app.kubernetes.io/component=web -n smart-garden-bot --timeout=600s

# Wait for operator
kubectl wait --for=condition=available deployment -l app.kubernetes.io/component=operator -n smart-garden-bot-system --timeout=600s
```

#### Step 5: Monitoring Stack (65-90 minutes)
```bash
# Deploy monitoring stack
kubectl apply -f gitops/applications/monitoring.yaml

# Wait for Prometheus
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n monitoring --timeout=600s

# Wait for Grafana
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=grafana -n monitoring --timeout=600s
```

### Scenario 2: Multi-Region Cloud Provider Outage

**Symptoms**: Entire AWS region unavailable, DNS resolution failing

**Recovery Steps**:

#### Step 1: DNS Failover (0-5 minutes)
```bash
# Update Route53 records to point to DR region
aws route53 change-resource-record-sets --hosted-zone-id Z123456789 --change-batch '{
  "Changes": [{
    "Action": "UPSERT",
    "ResourceRecordSet": {
      "Name": "api.smartgardenbot.com",
      "Type": "A",
      "AliasTarget": {
        "DNSName": "dr-lb.us-west-2.elb.amazonaws.com",
        "EvaluateTargetHealth": false,
        "HostedZoneId": "Z1D633PJN98FT9"
      }
    }
  }]
}'
```

#### Step 2: DR Cluster Activation (5-15 minutes)
```bash
# Switch kubectl context to DR cluster
kubectl config use-context dr-cluster

# Activate standby applications
kubectl patch application smart-garden-bot-production -n argocd -p '{"spec":{"syncPolicy":{"automated":{"prune":true,"selfHeal":true}}}}'

# Scale up from standby mode
kubectl scale deployment smart-garden-bot-api -n smart-garden-bot --replicas=5
kubectl scale deployment smart-garden-bot-web -n smart-garden-bot --replicas=3
```

#### Step 3: Database Sync Verification (15-25 minutes)
```bash
# Check database replication lag (if cross-region replication was enabled)
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT 
  application_name,
  client_addr,
  state,
  sync_state,
  pg_wal_lsn_diff(pg_current_wal_lsn(), sent_lsn) as send_lag,
  pg_wal_lsn_diff(sent_lsn, flush_lsn) as receive_lag,
  pg_wal_lsn_diff(flush_lsn, replay_lsn) as replay_lag
FROM pg_stat_replication;
"

# If significant lag, restore from latest backup
```

### Scenario 3: Data Corruption/Security Breach

**Symptoms**: Unauthorized data access, data integrity compromised

**Immediate Response** (0-10 minutes):
```bash
# Isolate affected systems
kubectl scale deployment smart-garden-bot-api -n smart-garden-bot --replicas=0

# Block external traffic
kubectl patch ingress smart-garden-bot -n smart-garden-bot -p '{"spec":{"rules":[]}}'

# Preserve evidence
kubectl get events --all-namespaces --sort-by='.lastTimestamp' > incident-events.log
kubectl logs -n smart-garden-bot --all-containers=true --since=24h > incident-logs.log
```

**Investigation** (10-60 minutes):
```bash
# Check for unauthorized access
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT 
  datname,
  usename,
  application_name,
  client_addr,
  backend_start,
  query_start,
  state,
  query
FROM pg_stat_activity 
WHERE usename != 'smartgarden' 
  OR client_addr NOT LIKE '10.%'
ORDER BY backend_start DESC;
"

# Audit recent database changes
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del, last_vacuum, last_analyze
FROM pg_stat_user_tables 
WHERE n_tup_ins > 0 OR n_tup_upd > 0 OR n_tup_del > 0
ORDER BY greatest(n_tup_ins, n_tup_upd, n_tup_del) DESC;
"
```

**Recovery** (60-120 minutes):
```bash
# Point-in-time recovery to known good state
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: smart-garden-bot-postgres-clean
  namespace: smart-garden-bot-recovery
spec:
  instances: 1
  
  bootstrap:
    recovery:
      backup:
        name: smart-garden-bot-postgres-backup-clean
      recoveryTarget:
        targetTime: "2024-01-15 10:00:00"  # Before breach
        targetTimezone: "UTC"
EOF

# Export clean data
kubectl exec -n smart-garden-bot-recovery smart-garden-bot-postgres-clean-1 -- pg_dump -U smartgarden smartgarden > clean_data.sql

# Reset all passwords and API keys
kubectl create secret generic smart-garden-bot-auth-secret-new -n smart-garden-bot \
  --from-literal=auth0-client-secret="$(openssl rand -base64 32)" \
  --from-literal=nextauth-secret="$(openssl rand -base64 32)"
```

## Recovery Verification Checklist

### Application Health
- [ ] API server responding to health checks
- [ ] Web application loading correctly
- [ ] Operator processing garden controllers
- [ ] Database accepting connections
- [ ] All pods in Running state
- [ ] No error logs in recent 15 minutes

### Data Integrity
- [ ] User authentication working
- [ ] Garden configurations preserved
- [ ] Sensor data retrievable
- [ ] Billing information intact
- [ ] Recent backups successful

### External Integrations
- [ ] Weather API connectivity
- [ ] Auth0 authentication
- [ ] Stripe payment processing
- [ ] Email notifications
- [ ] IoT device communication

### Performance
- [ ] API response times < 200ms
- [ ] Database query performance normal
- [ ] Memory usage within limits
- [ ] CPU usage stable
- [ ] Network latency acceptable

## Communication Plan

### Internal Communication
```bash
# Slack incident channel update template
🚨 DISASTER RECOVERY UPDATE - [TIMESTAMP]
Scenario: [SCENARIO_TYPE]
Status: [IN_PROGRESS/TESTING/COMPLETE]
ETA: [TIME_ESTIMATE]
Services Affected: [LIST]
Current Action: [CURRENT_STEP]
Next Update: [TIME]
```

### External Communication
- **Status Page**: Update immediately with incident details
- **Customer Email**: Send within 30 minutes if >1 hour downtime expected
- **Social Media**: Post updates every hour for extended outages
- **Press/Media**: Coordinate through marketing team if public attention

### Stakeholder Notifications
- **CEO/Leadership**: Immediate notification for >30 min outages
- **Customer Success**: Immediate for customer-facing issues
- **Engineering Team**: All hands mobilization
- **Legal/Compliance**: For security breaches or data issues

## Post-Recovery Procedures

### Immediate (0-2 hours)
1. **Service Verification**: Complete health check
2. **Performance Testing**: Load test to verify capacity
3. **Data Validation**: Spot check critical data integrity
4. **Monitoring Setup**: Ensure all alerts are functioning
5. **Team Communication**: All-clear notification

### Short-term (2-24 hours)
1. **Comprehensive Testing**: Full regression test suite
2. **User Communication**: Service restoration notification
3. **Backup Verification**: Ensure backups are current
4. **Security Scan**: Full security audit if breach suspected
5. **Performance Monitoring**: Watch for degradation

### Medium-term (1-7 days)
1. **Incident Report**: Detailed post-mortem
2. **Process Improvement**: Update procedures based on learnings
3. **Automation Enhancement**: Identify manual steps to automate
4. **Training Update**: Revise team training materials
5. **Business Continuity**: Review and update DR plan

## Testing and Validation

### Monthly DR Tests
```bash
# Automated DR test script
#!/bin/bash
echo "Starting monthly DR test..."

# Create test namespace
kubectl create namespace dr-test-$(date +%Y%m%d)

# Deploy minimal version of stack
helm install smart-garden-bot-dr-test ./helm/smart-garden-bot \
  --namespace dr-test-$(date +%Y%m%d) \
  --set postgresql.enabled=false \
  --set ingress.enabled=false

# Run health checks
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=smart-garden-bot -n dr-test-$(date +%Y%m%d) --timeout=300s

# Cleanup
kubectl delete namespace dr-test-$(date +%Y%m%d)

echo "DR test completed successfully"
```

### Quarterly Full DR Drill
- Complete cluster rebuild exercise
- Database restoration from backup
- Full application stack deployment
- End-to-end functionality testing
- Documentation and process review

## Key Contacts

### Immediate Response Team
- **Incident Commander**: @incident-commander
- **Database Expert**: @dba-lead
- **Platform Engineer**: @platform-lead
- **Security Lead**: @security-lead

### Extended Response Team
- **Engineering Manager**: @eng-manager
- **DevOps Lead**: @devops-lead
- **Customer Success**: @customer-success
- **Communications**: @comms-lead

### External Contacts
- **AWS Support**: Enterprise Support Case
- **DNS Provider**: Support ticket + phone
- **Monitoring Service**: Status verification
- **Legal Counsel**: For compliance issues

## Document Maintenance
- **Review Frequency**: Monthly
- **Update Triggers**: After incidents, technology changes, team changes
- **Approval Required**: Engineering Manager, Security Lead
- **Distribution**: All engineering team members, on-call rotation

## Revision History
- v1.0 - Initial disaster recovery plan
- v1.1 - Added security breach procedures
- v1.2 - Enhanced communication templates
- v1.3 - Added automated testing procedures