# Runbook: Database Recovery

## Overview
This runbook covers PostgreSQL database recovery procedures using CloudNativePG for the Smart Garden Bot platform.

## Alert Details
- **Alert Name**: PostgreSQLDown
- **Severity**: Critical
- **Impact**: Complete data unavailability, API server cannot function

## Quick Reference

### Database Information
- **Cluster Name**: `smart-garden-bot-postgres`
- **Namespace**: `smart-garden-bot`
- **Primary Service**: `smart-garden-bot-postgres-rw`
- **Read-Only Service**: `smart-garden-bot-postgres-ro`
- **Backup Location**: `s3://smart-garden-bot-backups/postgres`

## Immediate Response (First 2 minutes)

### 1. Assess Database Status
```bash
# Check cluster status
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres

# Check all database pods
kubectl get pods -n smart-garden-bot -l cnpg.io/cluster=smart-garden-bot-postgres

# Check recent events
kubectl get events -n smart-garden-bot --sort-by='.lastTimestamp' | grep postgres
```

### 2. Quick Connectivity Test
```bash
# Test primary connection
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- pg_isready -U smartgarden

# Check if read replicas are available
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-2 -- pg_isready -U smartgarden
```

## Diagnostic Procedures

### 3. Detailed Cluster Analysis
```bash
# Get detailed cluster information
kubectl describe cluster -n smart-garden-bot smart-garden-bot-postgres

# Check cluster conditions
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres -o jsonpath='{.status.conditions[*]}'

# Review cluster logs
kubectl logs -n smart-garden-bot -l cnpg.io/cluster=smart-garden-bot-postgres --tail=100
```

### 4. Storage and Resource Check
```bash
# Check PVC status
kubectl get pvc -n smart-garden-bot -l cnpg.io/cluster=smart-garden-bot-postgres

# Check node storage capacity
kubectl describe pv | grep -A5 -B5 smart-garden-bot-postgres

# Check resource usage
kubectl top pods -n smart-garden-bot -l cnpg.io/cluster=smart-garden-bot-postgres
```

### 5. Backup Status Verification
```bash
# List available backups
kubectl get backup -n smart-garden-bot

# Check backup details
kubectl describe backup -n smart-garden-bot -l cnpg.io/cluster=smart-garden-bot-postgres

# Check scheduled backup status
kubectl get schedulebackup -n smart-garden-bot
```

## Recovery Scenarios

## Scenario 1: Single Pod Failure

**Symptoms**: One database pod down, cluster still functional

**Resolution**:
```bash
# CloudNativePG should auto-recover, but if not:
kubectl delete pod -n smart-garden-bot smart-garden-bot-postgres-[failed-pod-number]

# Wait for pod recreation
kubectl wait --for=condition=Ready pod -l cnpg.io/cluster=smart-garden-bot-postgres -n smart-garden-bot --timeout=300s

# Verify cluster health
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres
```

## Scenario 2: Primary Database Failure

**Symptoms**: Primary pod down, automatic failover may occur

**Resolution**:
```bash
# Check if automatic failover occurred
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres -o jsonpath='{.status.currentPrimary}'

# If manual failover needed:
kubectl cnpg promote -n smart-garden-bot smart-garden-bot-postgres smart-garden-bot-postgres-2

# Verify new primary
kubectl get cluster -n smart-garden-bot smart-garden-bot-postgres -o jsonpath='{.status.currentPrimary}'
```

## Scenario 3: Complete Cluster Failure

**Symptoms**: All database pods down, data corruption possible

**Resolution**:
```bash
# Check if any pods can be recovered
kubectl get pods -n smart-garden-bot -l cnpg.io/cluster=smart-garden-bot-postgres

# If complete data loss, restore from backup
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: smart-garden-bot-postgres-recovery
  namespace: smart-garden-bot
spec:
  instances: 3
  
  bootstrap:
    recovery:
      backup:
        name: smart-garden-bot-postgres-backup-$(date +%Y%m%d)
      recoveryTarget:
        targetTime: "$(date -u -d '1 hour ago' +%Y-%m-%d\ %H:%M:%S)"
  
  storage:
    size: 100Gi
    storageClass: gp3
  
  postgresql:
    parameters:
      max_connections: "200"
      shared_buffers: "256MB"
EOF

# Wait for recovery completion
kubectl wait --for=condition=Ready cluster -n smart-garden-bot smart-garden-bot-postgres-recovery --timeout=1800s
```

## Scenario 4: Point-in-Time Recovery

**Symptoms**: Need to recover to specific point in time (data corruption, human error)

**Resolution**:
```bash
# Create recovery cluster with specific timestamp
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: smart-garden-bot-postgres-pitr
  namespace: smart-garden-bot
spec:
  instances: 1
  
  bootstrap:
    recovery:
      backup:
        name: smart-garden-bot-postgres-backup-latest
      recoveryTarget:
        targetTime: "2024-01-15 14:30:00"
        targetTimezone: "UTC"
  
  storage:
    size: 100Gi
    storageClass: gp3
EOF

# Export data once recovered
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-pitr-1 -- pg_dumpall -U postgres > recovery_dump.sql

# Import to main cluster after verification
```

## Data Verification Procedures

### 5. Post-Recovery Validation
```bash
# Connect to database and run basic queries
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del 
FROM pg_stat_user_tables 
ORDER BY schemaname, tablename;
"

# Check data integrity
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT 
  'auth.users' as table_name, 
  count(*) as row_count 
FROM auth.users
UNION ALL
SELECT 
  'gardens.gardens' as table_name, 
  count(*) as row_count 
FROM gardens.gardens
UNION ALL
SELECT 
  'sensors.readings' as table_name, 
  count(*) as row_count 
FROM sensors.readings
WHERE timestamp > NOW() - INTERVAL '24 hours';
"

# Verify extensions
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT name, default_version, installed_version 
FROM pg_available_extensions 
WHERE name IN ('uuid-ossp', 'pgcrypto', 'vector', 'timescaledb');
"
```

### 6. Application Connectivity Test
```bash
# Restart API server to refresh connections
kubectl rollout restart deployment/smart-garden-bot-api -n smart-garden-bot

# Test API endpoints
curl -f https://api.smartgardenbot.com/health
curl -f https://api.smartgardenbot.com/ready

# Check API logs for database errors
kubectl logs -n smart-garden-bot -l app.kubernetes.io/component=api --tail=50 | grep -i error
```

## Backup and Maintenance

### 7. Immediate Backup After Recovery
```bash
# Trigger manual backup
kubectl apply -f - <<EOF
apiVersion: postgresql.cnpg.io/v1
kind: Backup
metadata:
  name: smart-garden-bot-postgres-post-recovery-$(date +%Y%m%d-%H%M%S)
  namespace: smart-garden-bot
spec:
  cluster:
    name: smart-garden-bot-postgres
EOF

# Monitor backup progress
kubectl get backup -n smart-garden-bot -w
```

### 8. Cleanup Old Resources
```bash
# Remove failed cluster if replaced
kubectl delete cluster -n smart-garden-bot smart-garden-bot-postgres-failed --wait=false

# Clean up old PVCs if needed (BE CAREFUL)
kubectl get pvc -n smart-garden-bot -l cnpg.io/cluster=smart-garden-bot-postgres-failed
# kubectl delete pvc -n smart-garden-bot [pvc-name] # Only after confirming data is safe
```

## Monitoring and Alerts

### 9. Update Monitoring
```bash
# Check if metrics are being collected
kubectl get servicemonitor -n smart-garden-bot

# Verify Prometheus targets
# Check Prometheus UI -> Targets -> postgresql
```

### 10. Test Alert Rules
```bash
# Temporarily trigger alert to verify it's working
kubectl scale deployment smart-garden-bot-postgres-exporter -n smart-garden-bot --replicas=0

# Wait for alert to fire, then restore
kubectl scale deployment smart-garden-bot-postgres-exporter -n smart-garden-bot --replicas=1
```

## Prevention and Hardening

### Regular Maintenance Tasks
```bash
# Weekly backup verification
kubectl get backup -n smart-garden-bot --sort-by=.metadata.creationTimestamp

# Monthly backup restore test (to separate namespace)
# Quarterly disaster recovery drill
```

### Performance Optimization
```bash
# Check slow queries
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "
SELECT query, mean_exec_time, calls, total_exec_time 
FROM pg_stat_statements 
ORDER BY mean_exec_time DESC 
LIMIT 10;
"

# Update statistics
kubectl exec -n smart-garden-bot smart-garden-bot-postgres-1 -- psql -U smartgarden -d smartgarden -c "ANALYZE;"
```

## Escalation and Communication

### When to Escalate
- Recovery taking longer than 30 minutes
- Data corruption detected
- Backup restoration failing
- Multiple recovery attempts unsuccessful

### Communication Templates

**Slack Incident Channel**:
```
🚨 DATABASE RECOVERY IN PROGRESS
Status: [IN_PROGRESS/COMPLETE/FAILED]
Start Time: [TIMESTAMP]
Expected Recovery: [TIMESTAMP]
Data Loss: [NONE/MINIMAL/SIGNIFICANT]
Actions Taken: [SUMMARY]
Next Steps: [ACTIONS]
```

**Status Page Update**:
```
We are experiencing database connectivity issues and are working to restore service. 
All user data is safely backed up. We expect to have service restored within [TIME].
Updates will be provided every 15 minutes.
```

## Related Procedures
- [API Server Recovery](api-recovery.md)
- [Backup Verification](backup-verification.md)
- [Performance Tuning](database-tuning.md)
- [Security Hardening](database-security.md)

## Contacts
- **DBA On-Call**: @dba-oncall
- **Platform Team**: @platform-team
- **Infrastructure**: @infra-team
- **Security**: @security-team

## Revision History
- v1.0 - Initial CloudNativePG procedures
- v1.1 - Added point-in-time recovery
- v1.2 - Enhanced verification procedures