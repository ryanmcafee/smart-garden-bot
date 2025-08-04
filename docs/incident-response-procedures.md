# Smart Garden Bot - Incident Response Procedures

## Executive Summary

This document establishes comprehensive incident response procedures for the Smart Garden Bot platform, providing structured processes for detecting, responding to, and recovering from security incidents. The procedures are designed to minimize impact, preserve evidence, ensure regulatory compliance, and enable rapid recovery while maintaining transparency with stakeholders.

## Incident Response Team Structure

### Core Response Team
```yaml
incidentResponseTeam:
  incidentCommander:
    role: "Overall incident coordination and decision making"
    primary: "CISO"
    backup: "Security Team Lead"
    responsibilities:
      - Declare incident severity levels
      - Coordinate response activities
      - Authorize emergency actions
      - Communicate with executive leadership
    
  securityAnalyst:
    role: "Technical investigation and analysis"
    primary: "Senior Security Analyst"
    backup: "Security Engineer"
    responsibilities:
      - Analyze security events and logs
      - Perform forensic investigation
      - Identify attack vectors and scope
      - Coordinate with external security services
    
  systemsEngineer:
    role: "Infrastructure and system response"
    primary: "Senior DevOps Engineer"
    backup: "Platform Engineer"
    responsibilities:
      - Implement containment measures
      - Restore affected systems
      - Monitor system health during response
      - Coordinate with cloud providers
    
  communicationsLead:
    role: "Internal and external communications"
    primary: "Director of Communications"
    backup: "Legal Counsel"
    responsibilities:
      - Manage stakeholder communications
      - Draft external notifications
      - Coordinate with media if required
      - Ensure regulatory notification compliance
    
  legalCounsel:
    role: "Legal and regulatory compliance"
    primary: "General Counsel"
    backup: "Compliance Officer"
    responsibilities:
      - Advise on legal implications
      - Coordinate regulatory notifications
      - Manage law enforcement interactions
      - Review public statements
```

### Extended Response Team
```yaml
extendedTeam:
  businessContinuity:
    role: "Business operations continuity"
    lead: "COO"
    responsibilities:
      - Assess business impact
      - Coordinate alternative operations
      - Manage customer service response
      - Plan recovery operations
  
  humanResources:
    role: "Personnel and internal coordination"
    lead: "HR Director"
    responsibilities:
      - Coordinate internal communications
      - Manage employee concerns
      - Handle insider threat investigations
      - Support affected personnel
  
  customerSuccess:
    role: "Customer communication and support"
    lead: "Customer Success Manager"
    responsibilities:
      - Manage customer inquiries
      - Coordinate customer notifications
      - Provide service status updates
      - Document customer impact
```

## Incident Classification Framework

### Severity Levels
```yaml
severityLevels:
  critical:
    description: "Severe impact on confidentiality, integrity, or availability"
    src:
      - "Confirmed data breach with PII exposure"
      - "Complete system compromise"
      - "Ransomware infection"
      - "Payment system compromise"
    responseTime: "15 minutes"
    resolutionTarget: "4 hours"
    notificationRequired: ["CISO", "CEO", "Legal", "Board"]
    
  high:
    description: "Significant security incident with potential for major impact"
    src:
      - "Unauthorized access to sensitive systems"
      - "Successful privilege escalation"
      - "DDoS attack affecting service availability"
      - "Malware detection on production systems"
    responseTime: "30 minutes"
    resolutionTarget: "24 hours"
    notificationRequired: ["CISO", "Security Team", "Engineering"]
    
  medium:
    description: "Security incident with limited impact"
    src:
      - "Failed attack attempts"
      - "Policy violations"
      - "Suspicious user activity"
      - "Minor data exposure"
    responseTime: "1 hour"
    resolutionTarget: "72 hours"
    notificationRequired: ["Security Team", "Engineering"]
    
  low:
    description: "Security event requiring investigation"
    src:
      - "Anomalous login patterns"
      - "Minor policy deviations"
      - "Automated security alerts"
      - "Routine security findings"
    responseTime: "2 hours"
    resolutionTarget: "1 week"
    notificationRequired: ["Security Team"]
```

### Incident Categories
```yaml
incidentCategories:
  dataBreach:
    description: "Unauthorized access, disclosure, or theft of sensitive data"
    subTypes:
      - "PII exposure"
      - "Payment card data breach"
      - "Customer data theft"
      - "Intellectual property theft"
    regulatory: ["GDPR", "PCI DSS", "State breach laws"]
    
  systemCompromise:
    description: "Unauthorized access to or control of systems"
    subTypes:
      - "Server compromise"
      - "Database breach"
      - "Administrative account takeover"
      - "Supply chain compromise"
    regulatory: ["SOC 2", "ISO 27001"]
    
  serviceDisruption:
    description: "Intentional disruption of service availability"
    subTypes:
      - "DDoS attacks"
      - "Resource exhaustion"
      - "Service degradation"
      - "Infrastructure failure"
    regulatory: ["SOC 2"]
    
  maliciousCode:
    description: "Detection of malware, viruses, or malicious scripts"
    subTypes:
      - "Ransomware"
      - "Trojans"
      - "Cryptocurrency miners"
      - "Backdoors"
    regulatory: ["Various depending on impact"]
    
  insiderThreat:
    description: "Malicious or negligent actions by internal personnel"
    subTypes:
      - "Data theft by employee"
      - "Privilege abuse"
      - "Negligent data exposure"
      - "Sabotage"
    regulatory: ["GDPR", "SOC 2"]
```

## Incident Response Phases

### Phase 1: Preparation
```bash
#!/bin/bash
# Incident Response Preparation Script

# Validate incident response readiness
validate_incident_readiness() {
    echo "=== Incident Response Readiness Check ==="
    
    # Check communication channels
    echo "Checking communication channels..."
    curl -s -X POST $SLACK_WEBHOOK_URL -d '{"text":"IR readiness test"}' || echo "❌ Slack webhook failed"
    
    # Check incident tracking system
    curl -s -H "Authorization: Bearer $IR_API_TOKEN" $IR_SYSTEM_URL/health || echo "❌ Incident tracking system unavailable"
    
    # Check forensics tools
    command -v tcpdump >/dev/null 2>&1 || echo "❌ tcpdump not available"
    command -v volatility >/dev/null 2>&1 || echo "❌ volatility not available"
    
    # Check backup systems
    kubectl get pods -n backup-system || echo "❌ Backup system unavailable"
    
    # Check contact information
    echo "Verifying contact information..."
    for contact in "${EMERGENCY_CONTACTS[@]}"; do
        echo "📞 $contact"
    done
    
    echo "=== Readiness check complete ==="
}

# Pre-position incident response tools
prepare_ir_environment() {
    # Create incident workspace
    mkdir -p /var/lib/incident-response/{logs,evidence,reports,tools}
    
    # Download latest forensics tools
    cd /var/lib/incident-response/tools
    wget -q https://github.com/volatilityfoundation/volatility/archive/master.zip
    wget -q https://raw.githubusercontent.com/sans-dfir/sift-files/master/tools.txt
    
    # Prepare evidence collection scripts
    cat > /var/lib/incident-response/tools/collect-evidence.sh << 'EOF'
#!/bin/bash
INCIDENT_ID=$1
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
EVIDENCE_DIR="/var/lib/incident-response/evidence/$INCIDENT_ID-$TIMESTAMP"

mkdir -p $EVIDENCE_DIR

# System information
uname -a > $EVIDENCE_DIR/system-info.txt
df -h > $EVIDENCE_DIR/disk-usage.txt
free -m > $EVIDENCE_DIR/memory-usage.txt
ps aux > $EVIDENCE_DIR/running-processes.txt
netstat -tulpn > $EVIDENCE_DIR/network-connections.txt
lsof > $EVIDENCE_DIR/open-files.txt

# Log collection
journalctl --since "1 hour ago" > $EVIDENCE_DIR/system-logs.txt
dmesg > $EVIDENCE_DIR/kernel-logs.txt

# Kubernetes logs
kubectl get events --all-namespaces > $EVIDENCE_DIR/k8s-events.txt
kubectl logs -n smart-garden-bot --all-containers=true > $EVIDENCE_DIR/app-logs.txt

# Network capture (5 minutes)
timeout 300 tcpdump -i any -w $EVIDENCE_DIR/network-capture.pcap

# Create evidence hash
find $EVIDENCE_DIR -type f -exec sha256sum {} \; > $EVIDENCE_DIR/evidence-hashes.txt

echo "Evidence collected in: $EVIDENCE_DIR"
EOF
    
    chmod +x /var/lib/incident-response/tools/collect-evidence.sh
}

# Emergency contact validation
validate_emergency_contacts() {
    EMERGENCY_CONTACTS=(
        "CISO:+1-555-0101:ciso@smart-garden-bot.com"
        "CEO:+1-555-0102:ceo@smart-garden-bot.com"
        "Legal:+1-555-0103:legal@smart-garden-bot.com"
        "Security Team:+1-555-0104:security@smart-garden-bot.com"
    )
    
    for contact in "${EMERGENCY_CONTACTS[@]}"; do
        name=$(echo $contact | cut -d: -f1)
        phone=$(echo $contact | cut -d: -f2)
        email=$(echo $contact | cut -d: -f3)
        
        echo "Testing contact: $name"
        # Send test notification (disabled in prod)
        # aws sns publish --phone-number $phone --message "IR test message"
    done
}

# Run preparation checks
main() {
    validate_incident_readiness
    prepare_ir_environment
    validate_emergency_contacts
    
    echo "Incident response preparation complete"
}

main "$@"
```

### Phase 2: Detection and Analysis
```python
#!/usr/bin/env python3
"""
Incident detection and initial analysis system
"""

import json
import time
import hashlib
from datetime import datetime, timedelta
from dataclasses import dataclass, asdict
from typing import List, Dict, Optional

@dataclass
class SecurityEvent:
    id: str
    timestamp: datetime
    source: str
    event_type: str
    severity: str
    description: str
    raw_data: dict
    indicators: List[str]
    affected_systems: List[str]
    confidence_score: float

@dataclass
class IncidentAnalysis:
    incident_id: str
    classification: str
    severity: str
    confidence: float
    timeline: List[dict]
    indicators_of_compromise: List[str]
    attack_vector: str
    impact_assessment: str
    recommended_actions: List[str]

class IncidentAnalyzer:
    def __init__(self):
        self.ioc_database = self.load_ioc_database()
        self.attack_patterns = self.load_attack_patterns()
        
    def load_ioc_database(self):
        """Load indicators of compromise database"""
        return {
            'malicious_ips': [
                '192.0.2.1',    # Example bad IP
                '203.0.113.1',  # Example bad IP
            ],
            'malicious_domains': [
                'evil.com',
                'malware-c2.net'
            ],
            'malicious_hashes': [
                'a1b2c3d4e5f6...',  # Known malware hash
            ],
            'suspicious_patterns': [
                r'\.exe\s+>', # Executable redirection
                r'powershell.*-enc', # Encoded PowerShell
                r'wget.*\|\s*sh', # Download and execute
            ]
        }
    
    def load_attack_patterns(self):
        """Load MITRE ATT&CK patterns"""
        return {
            'credential_dumping': {
                'techniques': ['T1003', 'T1081'],
                'indicators': ['lsass.exe', 'procdump', 'mimikatz'],
                'severity': 'HIGH'
            },
            'lateral_movement': {
                'techniques': ['T1021', 'T1077'],
                'indicators': ['net use', 'psexec', 'wmic'],
                'severity': 'HIGH'
            },
            'data_exfiltration': {
                'techniques': ['T1041', 'T1048'],
                'indicators': ['large_data_transfer', 'uncommon_protocols'],
                'severity': 'CRITICAL'
            }
        }
    
    def analyze_event(self, event: SecurityEvent) -> Optional[IncidentAnalysis]:
        """Perform initial analysis of security event"""
        
        # Check against IOC database
        ioc_matches = self.check_iocs(event)
        
        # Pattern matching against attack techniques
        attack_patterns = self.match_attack_patterns(event)
        
        # Calculate confidence score
        confidence = self.calculate_confidence(event, ioc_matches, attack_patterns)
        
        if confidence < 0.5:
            return None  # Low confidence, likely false positive
        
        # Classify incident
        classification = self.classify_incident(event, attack_patterns)
        
        # Assess impact
        impact = self.assess_impact(event, classification)
        
        # Generate recommendations
        recommendations = self.generate_recommendations(classification, impact)
        
        return IncidentAnalysis(
            incident_id=self.generate_incident_id(event),
            classification=classification,
            severity=event.severity,
            confidence=confidence,
            timeline=[{
                'timestamp': event.timestamp.isoformat(),
                'event': 'Initial detection',
                'details': event.description
            }],
            indicators_of_compromise=ioc_matches,
            attack_vector=self.determine_attack_vector(event),
            impact_assessment=impact,
            recommended_actions=recommendations
        )
    
    def check_iocs(self, event: SecurityEvent) -> List[str]:
        """Check event against indicators of compromise"""
        matches = []
        raw_data_str = json.dumps(event.raw_data).lower()
        
        # Check IPs
        for ip in self.ioc_database['malicious_ips']:
            if ip in raw_data_str:
                matches.append(f"malicious_ip:{ip}")
        
        # Check domains
        for domain in self.ioc_database['malicious_domains']:
            if domain in raw_data_str:
                matches.append(f"malicious_domain:{domain}")
        
        # Check file hashes
        for hash_val in self.ioc_database['malicious_hashes']:
            if hash_val in raw_data_str:
                matches.append(f"malicious_hash:{hash_val}")
        
        return matches
    
    def match_attack_patterns(self, event: SecurityEvent) -> List[str]:
        """Match event against known attack patterns"""
        matches = []
        
        for pattern_name, pattern_data in self.attack_patterns.items():
            for indicator in pattern_data['indicators']:
                if indicator.lower() in json.dumps(event.raw_data).lower():
                    matches.append(pattern_name)
                    break
        
        return matches
    
    def calculate_confidence(self, event: SecurityEvent, ioc_matches: List[str], attack_patterns: List[str]) -> float:
        """Calculate confidence score for incident classification"""
        base_confidence = 0.3
        
        # Increase confidence for IOC matches
        base_confidence += len(ioc_matches) * 0.2
        
        # Increase confidence for attack pattern matches
        base_confidence += len(attack_patterns) * 0.15
        
        # Adjust based on event source reliability
        source_reliability = {
            'siem': 0.8,
            'ids': 0.7,
            'antivirus': 0.9,
            'user_report': 0.4,
            'automated_scan': 0.6
        }
        
        reliability = source_reliability.get(event.source, 0.5)
        base_confidence *= reliability
        
        return min(base_confidence, 1.0)
    
    def classify_incident(self, event: SecurityEvent, attack_patterns: List[str]) -> str:
        """Classify the type of security incident"""
        
        if 'data_exfiltration' in attack_patterns:
            return 'DATA_BREACH'
        elif 'credential_dumping' in attack_patterns or 'lateral_movement' in attack_patterns:
            return 'SYSTEM_COMPROMISE'
        elif event.event_type in ['ddos', 'dos']:
            return 'SERVICE_DISRUPTION'
        elif 'malware' in event.event_type.lower():
            return 'MALICIOUS_CODE'
        elif 'insider' in event.description.lower():
            return 'INSIDER_THREAT'
        else:
            return 'SECURITY_VIOLATION'
    
    def assess_impact(self, event: SecurityEvent, classification: str) -> str:
        """Assess the potential impact of the incident"""
        
        impact_matrix = {
            'DATA_BREACH': 'HIGH - Potential regulatory violations and customer impact',
            'SYSTEM_COMPROMISE': 'HIGH - Potential for lateral movement and data theft',
            'SERVICE_DISRUPTION': 'MEDIUM - Service availability impact',
            'MALICIOUS_CODE': 'MEDIUM - Potential system compromise and data theft',
            'INSIDER_THREAT': 'HIGH - Privileged access abuse potential',
            'SECURITY_VIOLATION': 'LOW - Policy violation requiring investigation'
        }
        
        base_impact = impact_matrix.get(classification, 'MEDIUM - Unknown impact')
        
        # Adjust based on affected systems
        critical_systems = ['payment', 'user_data', 'authentication']
        for system in event.affected_systems:
            if any(critical in system.lower() for critical in critical_systems):
                base_impact = base_impact.replace('LOW', 'MEDIUM').replace('MEDIUM', 'HIGH')
                break
        
        return base_impact
    
    def determine_attack_vector(self, event: SecurityEvent) -> str:
        """Determine the likely attack vector"""
        
        vectors = {
            'web_application': ['sql_injection', 'xss', 'csrf'],
            'network': ['port_scan', 'ddos', 'man_in_middle'],
            'email': ['phishing', 'malicious_attachment'],
            'insider': ['privilege_abuse', 'data_theft'],
            'supply_chain': ['compromised_dependency', 'backdoor']
        }
        
        event_lower = event.description.lower()
        
        for vector, indicators in vectors.items():
            if any(indicator in event_lower for indicator in indicators):
                return vector
        
        return 'unknown'
    
    def generate_recommendations(self, classification: str, impact: str) -> List[str]:
        """Generate immediate response recommendations"""
        
        recommendations = {
            'DATA_BREACH': [
                'Immediately isolate affected systems',
                'Preserve forensic evidence',
                'Notify legal team for regulatory compliance',
                'Prepare customer/regulatory notifications',
                'Engage external forensics team if needed'
            ],
            'SYSTEM_COMPROMISE': [
                'Isolate compromised systems',
                'Reset all potentially compromised credentials',
                'Scan for additional compromised systems',
                'Review and update access controls',
                'Monitor for lateral movement'
            ],
            'SERVICE_DISRUPTION': [
                'Implement DDoS mitigation measures',
                'Scale infrastructure if needed',
                'Block malicious traffic sources',
                'Communicate with customers about service status',
                'Prepare failover procedures'
            ],
            'MALICIOUS_CODE': [
                'Quarantine infected systems',
                'Run full antivirus scans',
                'Check for persistence mechanisms',
                'Review backup integrity',
                'Update security signatures'
            ]
        }
        
        return recommendations.get(classification, [
            'Investigate the incident thoroughly',
            'Document all findings',
            'Implement appropriate containment measures',
            'Monitor for related activity'
        ])
    
    def generate_incident_id(self, event: SecurityEvent) -> str:
        """Generate unique incident ID"""
        timestamp = datetime.now().strftime('%Y%m%d%H%M%S')
        event_hash = hashlib.md5(
            f"{event.timestamp}{event.source}{event.event_type}".encode()
        ).hexdigest()[:8]
        
        return f"INC-{timestamp}-{event_hash.upper()}"

# Example usage and testing
if __name__ == "__main__":
    analyzer = IncidentAnalyzer()
    
    # Simulate a security event
    test_event = SecurityEvent(
        id="evt_001",
        timestamp=datetime.now(),
        source="siem",
        event_type="unauthorized_access",
        severity="HIGH",
        description="Multiple failed login attempts followed by successful login from new location",
        raw_data={
            "source_ip": "192.0.2.1",
            "user_id": "user_12345",
            "failed_attempts": 15,
            "success_location": "Russia",
            "previous_locations": ["United States", "Canada"]
        },
        indicators=["brute_force", "geographical_anomaly"],
        affected_systems=["authentication_service"],
        confidence_score=0.8
    )
    
    analysis = analyzer.analyze_event(test_event)
    
    if analysis:
        print("Incident Analysis Results:")
        print(json.dumps(asdict(analysis), indent=2, default=str))
    else:
        print("Event did not meet threshold for incident classification")
```

### Phase 3: Containment
```bash
#!/bin/bash
# Incident Containment Procedures

INCIDENT_ID=$1
INCIDENT_TYPE=$2
AFFECTED_SYSTEMS=$3

log_action() {
    echo "[$(date)] CONTAINMENT: $1" | tee -a /var/log/incident-$INCIDENT_ID.log
}

# Short-term containment actions
short_term_containment() {
    log_action "Initiating short-term containment for $INCIDENT_TYPE"
    
    case $INCIDENT_TYPE in
        "DATA_BREACH")
            log_action "Implementing data breach containment"
            
            # Isolate affected database servers
            kubectl patch deployment postgresql -p '{"spec":{"replicas":0}}'
            
            # Block external access to sensitive endpoints
            kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: emergency-isolation-$INCIDENT_ID
  namespace: smart-garden-bot
spec:
  podSelector:
    matchLabels:
      app: api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443  # Only allow HTTPS outbound for alerts
EOF
            
            # Preserve evidence
            /var/lib/incident-response/tools/collect-evidence.sh $INCIDENT_ID
            ;;
            
        "SYSTEM_COMPROMISE")
            log_action "Implementing system compromise containment"
            
            # Isolate affected pods
            for system in $(echo $AFFECTED_SYSTEMS | tr ',' ' '); do
                kubectl scale deployment $system --replicas=0
                log_action "Scaled down $system deployment"
            done
            
            # Reset all API keys
            psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "UPDATE auth.api_keys SET status='suspended', suspended_reason='security_incident', suspended_at=NOW();"
            
            # Force password reset for all admin users
            psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "UPDATE auth.users SET password_reset_required=true WHERE role IN ('admin', 'super_admin');"
            ;;
            
        "SERVICE_DISRUPTION")
            log_action "Implementing service disruption containment"
            
            # Enable DDoS protection
            curl -X PATCH "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/settings/security_level" \
                 -H "Authorization: Bearer $CLOUDFLARE_TOKEN" \
                 -H "Content-Type: application/json" \
                 --data '{"value":"under_attack"}'
            
            # Scale up infrastructure
            kubectl scale deployment api --replicas=10
            kubectl scale deployment web-app --replicas=5
            
            # Implement rate limiting
            kubectl apply -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: rate-limit-$INCIDENT_ID
  namespace: smart-garden-bot
spec:
  configPatches:
  - applyTo: HTTP_FILTER
    match:
      context: SIDECAR_INBOUND
    patch:
      operation: INSERT_BEFORE
      value:
        name: envoy.filters.http.local_ratelimit
        typed_config:
          "@type": type.googleapis.com/udpa.type.v1.TypedStruct
          type_url: type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
          value:
            stat_prefix: rate_limiter
            token_bucket:
              max_tokens: 10
              tokens_per_fill: 10
              fill_interval: 60s
            filter_enabled:
              runtime_key: rate_limit_enabled
              default_value:
                numerator: 100
                denominator: HUNDRED
            filter_enforced:
              runtime_key: rate_limit_enforced
              default_value:
                numerator: 100
                denominator: HUNDRED
EOF
            ;;
            
        "MALICIOUS_CODE")
            log_action "Implementing malware containment"
            
            # Quarantine affected containers
            for system in $(echo $AFFECTED_SYSTEMS | tr ',' ' '); do
                # Create backup of current state
                kubectl get deployment $system -o yaml > /var/lib/incident-response/evidence/$INCIDENT_ID-$system-backup.yaml
                
                # Replace with clean image
                kubectl set image deployment/$system $system=smart-garden-bot/$system:clean-$(date +%Y%m%d)
                
                log_action "Replaced $system with clean image"
            done
            
            # Update all container images to latest clean versions
            kubectl patch deployment api -p '{"spec":{"template":{"spec":{"containers":[{"name":"api","image":"smart-garden-bot/api:clean-latest"}]}}}}'
            ;;
    esac
    
    log_action "Short-term containment completed"
}

# Long-term containment actions
long_term_containment() {
    log_action "Implementing long-term containment measures"
    
    # Enhanced monitoring
    kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: enhanced-monitoring-$INCIDENT_ID
  namespace: smart-garden-bot
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
    - job_name: 'security-metrics'
      scrape_interval: 5s
      static_configs:
      - targets: ['localhost:9090']
      metrics_path: /metrics
      params:
        'match[]':
        - '{__name__=~"security_.*"}'
        - '{__name__=~"auth_.*"}'
        - '{__name__=~"http_requests_total"}'
EOF
    
    # Deploy security monitoring pod
    kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: security-monitor-$INCIDENT_ID
  namespace: smart-garden-bot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: security-monitor
  template:
    metadata:
      labels:
        app: security-monitor
    spec:
      containers:
      - name: monitor
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
      volumes:
      - name: config
        configMap:
          name: enhanced-monitoring-$INCIDENT_ID
EOF
    
    # Implement additional security controls
    case $INCIDENT_TYPE in
        "DATA_BREACH"|"SYSTEM_COMPROMISE")
            # Enable audit logging for all API calls
            kubectl patch deployment api -p '{"spec":{"template":{"spec":{"containers":[{"name":"api","env":[{"name":"AUDIT_LEVEL","value":"DEBUG"}]}]}}}}'
            
            # Require MFA for all admin operations
            psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "UPDATE auth.users SET mfa_required=true WHERE role IN ('admin', 'super_admin');"
            
            # Implement session timeout
            redis-cli CONFIG SET timeout 900  # 15 minutes
            ;;
    esac
    
    log_action "Long-term containment completed"
}

# Evidence preservation
preserve_evidence() {
    log_action "Preserving forensic evidence"
    
    EVIDENCE_DIR="/var/lib/incident-response/evidence/$INCIDENT_ID"
    mkdir -p $EVIDENCE_DIR
    
    # Collect system state
    kubectl get all -n smart-garden-bot -o yaml > $EVIDENCE_DIR/k8s-state.yaml
    kubectl describe pods -n smart-garden-bot > $EVIDENCE_DIR/pod-descriptions.txt
    kubectl get events -n smart-garden-bot --sort-by='.lastTimestamp' > $EVIDENCE_DIR/k8s-events.txt
    
    # Collect logs
    kubectl logs -n smart-garden-bot -l app=api --previous > $EVIDENCE_DIR/api-logs.txt
    kubectl logs -n smart-garden-bot -l app=web-app --previous > $EVIDENCE_DIR/web-app-logs.txt
    
    # Database snapshot
    pg_dump -h $DB_HOST -U $DB_USER $DB_NAME | gzip > $EVIDENCE_DIR/database-snapshot.sql.gz
    
    # Network capture
    kubectl exec -n smart-garden-bot deployment/api -- tcpdump -i any -w - | gzip > $EVIDENCE_DIR/network-capture.pcap.gz &
    TCPDUMP_PID=$!
    sleep 300  # Capture for 5 minutes
    kill $TCPDUMP_PID
    
    # Create evidence manifest
    cat > $EVIDENCE_DIR/evidence-manifest.txt << EOF
Incident ID: $INCIDENT_ID
Collection Time: $(date)
Collector: $(whoami)@$(hostname)
Evidence Items:
$(find $EVIDENCE_DIR -type f -exec ls -la {} \;)

Evidence Hashes:
$(find $EVIDENCE_DIR -type f -exec sha256sum {} \;)
EOF
    
    # Sign evidence manifest
    gpg --clearsign $EVIDENCE_DIR/evidence-manifest.txt
    
    log_action "Evidence preservation completed"
}

# Main containment execution
main() {
    if [ $# -lt 3 ]; then
        echo "Usage: $0 <incident_id> <incident_type> <affected_systems>"
        exit 1
    fi
    
    log_action "Starting containment procedures for incident $INCIDENT_ID"
    
    # Preserve evidence before any changes
    preserve_evidence
    
    # Execute containment phases
    short_term_containment
    long_term_containment
    
    # Notify incident commander
    curl -X POST -H 'Content-type: application/json' \
         --data "{\"text\":\"🔒 Containment completed for incident $INCIDENT_ID\n**Type:** $INCIDENT_TYPE\n**Systems:** $AFFECTED_SYSTEMS\n**Evidence:** Preserved in /var/lib/incident-response/evidence/$INCIDENT_ID\"}" \
         $SLACK_WEBHOOK_URL
    
    log_action "Containment procedures completed"
}

main "$@"
```

### Phase 4: Eradication and Recovery
```python
#!/usr/bin/env python3
"""
Incident eradication and recovery procedures
"""

import subprocess
import json
import time
from datetime import datetime
from typing import List, Dict
import psycopg2
import redis

class IncidentRecovery:
    def __init__(self, incident_id: str, incident_type: str):
        self.incident_id = incident_id
        self.incident_type = incident_type
        self.recovery_log = []
        
    def log_action(self, action: str, status: str = "INFO"):
        """Log recovery actions"""
        log_entry = {
            'timestamp': datetime.now().isoformat(),
            'action': action,
            'status': status,
            'incident_id': self.incident_id
        }
        self.recovery_log.append(log_entry)
        print(f"[{status}] {action}")
        
        # Write to log file
        with open(f'/var/log/incident-{self.incident_id}-recovery.log', 'a') as f:
            f.write(f"{log_entry['timestamp']} [{status}] {action}\n")
    
    def eradicate_threats(self):
        """Remove threats and vulnerabilities"""
        self.log_action("Starting threat eradication phase")
        
        if self.incident_type == "MALICIOUS_CODE":
            self.eradicate_malware()
        elif self.incident_type == "SYSTEM_COMPROMISE":
            self.remove_unauthorized_access()
        elif self.incident_type == "DATA_BREACH":
            self.secure_data_access()
        elif self.incident_type == "SERVICE_DISRUPTION":
            self.remove_attack_vectors()
        
        self.log_action("Threat eradication completed")
    
    def eradicate_malware(self):
        """Remove malware and restore clean state"""
        self.log_action("Eradicating malware")
        
        # Deploy clean container images
        clean_images = {
            'api': 'smart-garden-bot/api:clean-latest',
            'web-app': 'smart-garden-bot/web-app:clean-latest',
            'kubernetes-operator': 'smart-garden-bot/k8s-operator:clean-latest'
        }
        
        for deployment, image in clean_images.items():
            try:
                subprocess.run([
                    'kubectl', 'set', 'image', f'deployment/{deployment}',
                    f'{deployment}={image}'
                ], check=True)
                self.log_action(f"Updated {deployment} with clean image")
            except subprocess.CalledProcessError as e:
                self.log_action(f"Failed to update {deployment}: {e}", "ERROR")
        
        # Clear infected data volumes
        self.log_action("Clearing potentially infected data volumes")
        subprocess.run(['kubectl', 'delete', 'pvc', '-l', 'infected=true'], check=False)
        
        # Update security signatures
        self.log_action("Updating security signatures")
        # This would typically integrate with your security tools
        
    def remove_unauthorized_access(self):
        """Remove unauthorized access and backdoors"""
        self.log_action("Removing unauthorized access")
        
        # Reset all user passwords
        try:
            conn = psycopg2.connect(
                host=os.environ.get('DB_HOST'),
                database=os.environ.get('DB_NAME'),
                user=os.environ.get('DB_USER'),
                password=os.environ.get('DB_PASSWORD')
            )
            cursor = conn.cursor()
            
            # Force password reset for all users
            cursor.execute("""
                UPDATE auth.users 
                SET password_reset_required = true,
                    session_invalidated_at = NOW()
                WHERE active = true
            """)
            
            # Disable all API keys
            cursor.execute("""
                UPDATE auth.api_keys 
                SET status = 'suspended',
                    suspended_reason = 'security_incident',
                    suspended_at = NOW()
                WHERE status = 'active'
            """)
            
            conn.commit()
            self.log_action("Reset all user credentials")
            
        except Exception as e:
            self.log_action(f"Failed to reset credentials: {e}", "ERROR")
        finally:
            if conn:
                conn.close()
        
        # Clear all active sessions
        try:
            r = redis.Redis(host='redis-service', port=6379, db=0)
            r.flushdb()  # Clear all session data
            self.log_action("Cleared all active sessions")
        except Exception as e:
            self.log_action(f"Failed to clear sessions: {e}", "ERROR")
        
        # Remove any backdoor accounts
        self.remove_unauthorized_accounts()
        
    def remove_unauthorized_accounts(self):
        """Identify and remove unauthorized user accounts"""
        self.log_action("Scanning for unauthorized accounts")
        
        # This would typically involve comparing against HR systems
        # For now, we'll check for recently created admin accounts
        try:
            conn = psycopg2.connect(
                host=os.environ.get('DB_HOST'),
                database=os.environ.get('DB_NAME'),
                user=os.environ.get('DB_USER'),
                password=os.environ.get('DB_PASSWORD')
            )
            cursor = conn.cursor()
            
            # Find suspicious admin accounts created in last 7 days
            cursor.execute("""
                SELECT id, email, created_at, role 
                FROM auth.users 
                WHERE role IN ('admin', 'super_admin')
                AND created_at > NOW() - INTERVAL '7 days'
                ORDER BY created_at DESC
            """)
            
            suspicious_accounts = cursor.fetchall()
            
            for account in suspicious_accounts:
                user_id, email, created_at, role = account
                self.log_action(f"Found suspicious admin account: {email} (created: {created_at})")
                
                # This would typically require manual review
                # For automation, we could disable rather than delete
                cursor.execute("""
                    UPDATE auth.users 
                    SET active = false,
                        disabled_reason = 'security_review_required'
                    WHERE id = %s
                """, (user_id,))
                
                self.log_action(f"Disabled suspicious account: {email}")
            
            conn.commit()
            
        except Exception as e:
            self.log_action(f"Failed to scan for unauthorized accounts: {e}", "ERROR")
        finally:
            if conn:
                conn.close()
    
    def secure_data_access(self):
        """Implement additional data security measures"""
        self.log_action("Implementing enhanced data security")
        
        # Enable database audit logging
        try:
            conn = psycopg2.connect(
                host=os.environ.get('DB_HOST'),
                database=os.environ.get('DB_NAME'),
                user=os.environ.get('DB_USER'),
                password=os.environ.get('DB_PASSWORD')
            )
            cursor = conn.cursor()
            
            # Enable audit logging for sensitive tables
            sensitive_tables = ['auth.users', 'billing.customers', 'gardens.gardens']
            
            for table in sensitive_tables:
                cursor.execute(f"""
                    CREATE OR REPLACE FUNCTION audit_{table.replace('.', '_')}_changes()
                    RETURNS TRIGGER AS $$
                    BEGIN
                        INSERT INTO audit.data_changes (
                            table_name, operation, old_data, new_data, changed_by, changed_at
                        ) VALUES (
                            '{table}', TG_OP, 
                            CASE WHEN TG_OP = 'DELETE' THEN row_to_json(OLD) ELSE NULL END,
                            CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN row_to_json(NEW) ELSE NULL END,
                            current_user, NOW()
                        );
                        RETURN COALESCE(NEW, OLD);
                    END;
                    $$ LANGUAGE plpgsql;
                    
                    CREATE TRIGGER {table.replace('.', '_')}_audit_trigger
                    AFTER INSERT OR UPDATE OR DELETE ON {table}
                    FOR EACH ROW EXECUTE FUNCTION audit_{table.replace('.', '_')}_changes();
                """)
            
            conn.commit()
            self.log_action("Enabled enhanced database auditing")
            
        except Exception as e:
            self.log_action(f"Failed to enable database auditing: {e}", "ERROR")
        finally:
            if conn:
                conn.close()
    
    def remove_attack_vectors(self):
        """Remove attack vectors used in service disruption"""
        self.log_action("Removing attack vectors")
        
        # Block malicious IPs at multiple layers
        malicious_ips = self.get_malicious_ips()
        
        for ip in malicious_ips:
            # Block at WAF level
            self.block_ip_at_waf(ip)
            
            # Block at Kubernetes level
            self.block_ip_at_k8s(ip)
        
        # Update rate limiting rules
        self.update_rate_limiting()
        
    def get_malicious_ips(self) -> List[str]:
        """Extract malicious IPs from incident logs"""
        # This would typically parse incident logs and threat intelligence
        # For now, return example IPs
        return ['192.0.2.1', '203.0.113.1']
    
    def block_ip_at_waf(self, ip: str):
        """Block IP at WAF level"""
        try:
            # Example Cloudflare API call
            subprocess.run([
                'curl', '-X', 'POST',
                f'https://api.cloudflare.com/client/v4/zones/{os.environ.get("ZONE_ID")}/firewall/access_rules/rules',
                '-H', f'Authorization: Bearer {os.environ.get("CLOUDFLARE_TOKEN")}',
                '-H', 'Content-Type: application/json',
                '--data', f'{{"mode":"block","configuration":{{"target":"ip","value":"{ip}"}},"notes":"Blocked by incident response"}}'
            ], check=True)
            self.log_action(f"Blocked {ip} at WAF level")
        except Exception as e:
            self.log_action(f"Failed to block {ip} at WAF: {e}", "ERROR")
    
    def block_ip_at_k8s(self, ip: str):
        """Block IP at Kubernetes level"""
        network_policy = f'''
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: block-{ip.replace(".", "-")}-{self.incident_id}
  namespace: smart-garden-bot
spec:
  podSelector: {{}}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - ipBlock:
        cidr: 0.0.0.0/0
        except:
        - {ip}/32
'''
        
        try:
            with open(f'/tmp/block-{ip.replace(".", "-")}.yaml', 'w') as f:
                f.write(network_policy)
            
            subprocess.run(['kubectl', 'apply', '-f', f'/tmp/block-{ip.replace(".", "-")}.yaml'], check=True)
            self.log_action(f"Blocked {ip} at Kubernetes level")
        except Exception as e:
            self.log_action(f"Failed to block {ip} at K8s: {e}", "ERROR")
    
    def update_rate_limiting(self):
        """Update rate limiting configuration"""
        self.log_action("Updating rate limiting configuration")
        
        # This would update your rate limiting configuration
        # Example: Update Istio rate limiting
        rate_limit_config = '''
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: enhanced-rate-limit
  namespace: smart-garden-bot
spec:
  configPatches:
  - applyTo: HTTP_FILTER
    match:
      context: SIDECAR_INBOUND
    patch:
      operation: INSERT_BEFORE
      value:
        name: envoy.filters.http.local_ratelimit
        typed_config:
          "@type": type.googleapis.com/udpa.type.v1.TypedStruct
          type_url: type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
          value:
            stat_prefix: enhanced_rate_limiter
            token_bucket:
              max_tokens: 5
              tokens_per_fill: 5
              fill_interval: 60s
            filter_enabled:
              default_value:
                numerator: 100
                denominator: HUNDRED
'''
        
        try:
            with open('/tmp/enhanced-rate-limit.yaml', 'w') as f:
                f.write(rate_limit_config)
            
            subprocess.run(['kubectl', 'apply', '-f', '/tmp/enhanced-rate-limit.yaml'], check=True)
            self.log_action("Updated rate limiting configuration")
        except Exception as e:
            self.log_action(f"Failed to update rate limiting: {e}", "ERROR")
    
    def restore_services(self):
        """Restore services to normal operation"""
        self.log_action("Starting service restoration")
        
        # Restore from clean backups
        self.restore_from_backup()
        
        # Restart services in correct order
        self.restart_services()
        
        # Verify service health
        self.verify_service_health()
        
        # Re-enable normal operations
        self.enable_normal_operations()
        
        self.log_action("Service restoration completed")
    
    def restore_from_backup(self):
        """Restore data from clean backups"""
        self.log_action("Restoring from clean backups")
        
        # This would typically involve:
        # 1. Identifying the last clean backup before incident
        # 2. Restoring database from backup
        # 3. Restoring file systems from backup
        # 4. Validating backup integrity
        
        try:
            # Example database restore
            backup_date = "2024-01-15"  # This would be determined automatically
            subprocess.run([
                'pg_restore', '-h', os.environ.get('DB_HOST'),
                '-U', os.environ.get('DB_USER'),
                '-d', os.environ.get('DB_NAME'),
                f'/backups/database-{backup_date}.sql'
            ], check=True)
            
            self.log_action(f"Restored database from backup dated {backup_date}")
            
        except Exception as e:
            self.log_action(f"Failed to restore from backup: {e}", "ERROR")
    
    def restart_services(self):
        """Restart services in proper order"""
        service_order = ['postgresql', 'redis', 'api', 'web-app', 'kubernetes-operator']
        
        for service in service_order:
            try:
                subprocess.run(['kubectl', 'rollout', 'restart', f'deployment/{service}'], check=True)
                
                # Wait for rollout to complete
                subprocess.run([
                    'kubectl', 'rollout', 'status', f'deployment/{service}',
                    '--timeout=300s'
                ], check=True)
                
                self.log_action(f"Restarted {service} successfully")
                
            except Exception as e:
                self.log_action(f"Failed to restart {service}: {e}", "ERROR")
    
    def verify_service_health(self):
        """Verify all services are healthy"""
        self.log_action("Verifying service health")
        
        health_checks = [
            ('API Server', 'http://api:8080/health'),
            ('Web App', 'http://web-app:3000/health'),
            ('Database', 'postgresql://postgres:5432/smart_garden_bot')
        ]
        
        for service_name, endpoint in health_checks:
            try:
                if endpoint.startswith('http'):
                    subprocess.run(['curl', '-f', endpoint], check=True, capture_output=True)
                elif endpoint.startswith('postgresql'):
                    subprocess.run(['pg_isready', '-h', 'postgresql', '-p', '5432'], check=True)
                
                self.log_action(f"{service_name} health check passed")
                
            except Exception as e:
                self.log_action(f"{service_name} health check failed: {e}", "ERROR")
    
    def enable_normal_operations(self):
        """Re-enable normal operations after recovery"""
        self.log_action("Enabling normal operations")
        
        # Remove emergency restrictions
        subprocess.run([
            'kubectl', 'delete', 'networkpolicy',
            f'emergency-isolation-{self.incident_id}'
        ], check=False)
        
        # Restore normal rate limiting
        subprocess.run([
            'kubectl', 'delete', 'envoyfilter',
            f'rate-limit-{self.incident_id}'
        ], check=False)
        
        # Re-enable user registrations
        try:
            conn = psycopg2.connect(
                host=os.environ.get('DB_HOST'),
                database=os.environ.get('DB_NAME'),
                user=os.environ.get('DB_USER'),
                password=os.environ.get('DB_PASSWORD')
            )
            cursor = conn.cursor()
            
            cursor.execute("UPDATE system_settings SET value = 'true' WHERE key = 'registration_enabled'")
            conn.commit()
            
            self.log_action("Re-enabled user registrations")
            
        except Exception as e:
            self.log_action(f"Failed to re-enable registrations: {e}", "ERROR")
        finally:
            if conn:
                conn.close()
    
    def generate_recovery_report(self):
        """Generate recovery report"""
        report = {
            'incident_id': self.incident_id,
            'incident_type': self.incident_type,
            'recovery_start': self.recovery_log[0]['timestamp'] if self.recovery_log else None,
            'recovery_end': datetime.now().isoformat(),
            'actions_taken': len(self.recovery_log),
            'errors_encountered': len([log for log in self.recovery_log if log['status'] == 'ERROR']),
            'recovery_log': self.recovery_log
        }
        
        # Save report
        with open(f'/var/lib/incident-response/reports/{self.incident_id}-recovery-report.json', 'w') as f:
            json.dump(report, f, indent=2)
        
        self.log_action("Recovery report generated")
        return report

# Example usage
if __name__ == "__main__":
    import sys
    import os
    
    if len(sys.argv) != 3:
        print("Usage: python3 incident_recovery.py <incident_id> <incident_type>")
        sys.exit(1)
    
    incident_id = sys.argv[1]
    incident_type = sys.argv[2]
    
    recovery = IncidentRecovery(incident_id, incident_type)
    
    try:
        recovery.eradicate_threats()
        recovery.restore_services()
        report = recovery.generate_recovery_report()
        
        print(f"Recovery completed successfully. Report saved to recovery report file.")
        
    except Exception as e:
        recovery.log_action(f"Recovery failed: {e}", "CRITICAL")
        sys.exit(1)
```

### Phase 5: Post-Incident Activities
```yaml
# Post-incident review template
postIncidentReview:
  metadata:
    incident_id: "INC-20240115-12345"
    incident_date: "2024-01-15T10:30:00Z"
    review_date: "2024-01-22T14:00:00Z"
    participants:
      - "CISO"
      - "Security Team Lead"
      - "DevOps Engineer"
      - "Product Manager"
      - "Legal Counsel"
  
  timeline:
    detection: "2024-01-15T10:30:00Z"
    response_initiated: "2024-01-15T10:45:00Z"
    containment_achieved: "2024-01-15T12:15:00Z"
    eradication_completed: "2024-01-16T08:00:00Z"
    recovery_completed: "2024-01-16T16:00:00Z"
    lessons_learned_session: "2024-01-22T14:00:00Z"
  
  impact_assessment:
    customers_affected: 1250
    data_compromised: "User email addresses and encrypted passwords"
    financial_impact: "$25,000 (estimated)"
    regulatory_notifications: ["State Attorney General", "Affected customers"]
    media_coverage: "Limited local coverage"
  
  root_cause_analysis:
    primary_cause: "Unpatched vulnerability in authentication service"
    contributing_factors:
      - "Delayed patch management process"
      - "Insufficient automated vulnerability scanning"
      - "Missing network segmentation"
    
  what_went_well:
    - "Incident detected within 30 minutes"
    - "Response team assembled quickly"
    - "Containment achieved within SLA"
    - "Customer communication was timely and transparent"
    - "Regulatory notifications completed on time"
  
  areas_for_improvement:
    - "Patch management process needs automation"
    - "Vulnerability scanning frequency should increase"
    - "Network segmentation should be implemented"
    - "Incident response training needs refresher"
    - "Evidence collection procedures need refinement"
  
  action_items:
    - task: "Implement automated patch management"
      owner: "DevOps Team"
      due_date: "2024-02-15"
      priority: "HIGH"
    
    - task: "Deploy network segmentation"
      owner: "Security Team"
      due_date: "2024-03-01"
      priority: "HIGH"
    
    - task: "Increase vulnerability scan frequency to daily"
      owner: "Security Team"
      due_date: "2024-01-30"
      priority: "MEDIUM"
    
    - task: "Conduct incident response tabletop exercise"
      owner: "CISO"
      due_date: "2024-02-28"
      priority: "MEDIUM"
  
  metrics:
    detection_time: "30 minutes"
    response_time: "15 minutes"
    containment_time: "1 hour 45 minutes"
    recovery_time: "29 hours 30 minutes"
    customer_notification_time: "2 hours"
    
  compliance_review:
    gdpr_compliance: "Compliant - notifications sent within 72 hours"
    pci_compliance: "Not applicable - no payment data involved"
    soc2_compliance: "Incident documented per SOC 2 requirements"
    
  communication_review:
    internal_communications: "Effective use of Slack for coordination"
    customer_communications: "Clear and timely notifications sent"
    regulatory_communications: "All required notifications completed"
    media_relations: "Minimal media attention, handled appropriately"
```

## Regulatory Notification Procedures

### GDPR Data Breach Notification
```python
#!/usr/bin/env python3
"""
GDPR-compliant data breach notification system
"""

from datetime import datetime, timedelta
import json
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart

class GDPRNotificationManager:
    def __init__(self):
        self.notification_deadlines = {
            'supervisory_authority': timedelta(hours=72),
            'data_subjects': timedelta(hours=72),  # Without undue delay
            'dpo': timedelta(hours=24)
        }
    
    def assess_notification_requirement(self, incident_data):
        """Assess if GDPR notification is required"""
        
        # Factors that trigger notification requirement
        high_risk_factors = [
            'personal_data_involved',
            'sensitive_data_involved',
            'large_number_affected',
            'identity_theft_risk',
            'financial_loss_risk',
            'reputation_damage_risk',
            'discrimination_risk'
        ]
        
        risk_score = 0
        for factor in high_risk_factors:
            if incident_data.get(factor, False):
                risk_score += 1
        
        # Notification required if high risk or personal data involved
        notification_required = (
            incident_data.get('personal_data_involved', False) or
            risk_score >= 3
        )
        
        return {
            'notification_required': notification_required,
            'risk_score': risk_score,
            'deadline_supervisory_authority': datetime.now() + self.notification_deadlines['supervisory_authority'],
            'deadline_data_subjects': datetime.now() + self.notification_deadlines['data_subjects']
        }
    
    def notify_supervisory_authority(self, incident_data):
        """Send notification to supervisory authority"""
        
        notification = {
            'controller_details': {
                'name': 'Smart Garden Bot Inc.',
                'address': '123 Tech Street, San Francisco, CA 94105',
                'contact_person': 'Data Protection Officer',
                'email': 'dpo@smart-garden-bot.com',
                'phone': '+1-555-0100'
            },
            'breach_details': {
                'nature_of_breach': incident_data.get('incident_type'),
                'categories_of_data': incident_data.get('data_categories', []),
                'approximate_number_affected': incident_data.get('affected_users_count'),
                'likely_consequences': incident_data.get('impact_assessment'),
                'measures_taken': incident_data.get('containment_measures', [])
            },
            'timeline': {
                'breach_occurred': incident_data.get('incident_start_time'),
                'breach_discovered': incident_data.get('detection_time'),
                'notification_sent': datetime.now().isoformat()
            }
        }
        
        # Send to supervisory authority (this would be customized per jurisdiction)
        self.send_notification_email(
            to='breaches@ico.org.uk',  # Example: UK ICO
            subject=f'GDPR Data Breach Notification - {incident_data["incident_id"]}',
            content=json.dumps(notification, indent=2)
        )
        
        return notification
    
    def notify_data_subjects(self, incident_data, affected_users):
        """Send notification to affected data subjects"""
        
        notification_template = """
Dear Valued Customer,

We are writing to inform you of a security incident that may have affected your personal information stored in our Smart Garden Bot platform.

What Happened:
{incident_description}

Information Involved:
{data_categories}

What We Are Doing:
{response_measures}

What You Can Do:
{recommended_actions}

Contact Information:
If you have questions about this incident, please contact us at:
Email: security@smart-garden-bot.com
Phone: 1-800-GARDEN-1

We sincerely apologize for this incident and any inconvenience it may cause.

Sincerely,
Smart Garden Bot Security Team
        """
        
        for user in affected_users:
            personalized_notification = notification_template.format(
                incident_description=incident_data.get('user_friendly_description'),
                data_categories=', '.join(incident_data.get('data_categories', [])),
                response_measures='\n'.join(incident_data.get('response_measures', [])),
                recommended_actions='\n'.join(incident_data.get('user_recommendations', []))
            )
            
            self.send_notification_email(
                to=user['email'],
                subject='Important Security Notice - Smart Garden Bot',
                content=personalized_notification
            )
    
    def send_notification_email(self, to, subject, content):
        """Send notification email"""
        try:
            msg = MIMEMultipart()
            msg['From'] = 'security@smart-garden-bot.com'
            msg['To'] = to
            msg['Subject'] = subject
            
            msg.attach(MIMEText(content, 'plain'))
            
            # This would use your actual SMTP configuration
            server = smtplib.SMTP('smtp.smart-garden-bot.com', 587)
            server.starttls()
            server.login('security@smart-garden-bot.com', 'secure_password')
            text = msg.as_string()
            server.sendmail('security@smart-garden-bot.com', to, text)
            server.quit()
            
            print(f"Notification sent to {to}")
            
        except Exception as e:
            print(f"Failed to send notification to {to}: {e}")

# Example usage
if __name__ == "__main__":
    notification_manager = GDPRNotificationManager()
    
    # Example incident data
    incident_data = {
        'incident_id': 'INC-20240115-12345',
        'incident_type': 'Unauthorized access to user database',
        'personal_data_involved': True,
        'sensitive_data_involved': False,
        'large_number_affected': True,
        'affected_users_count': 1250,
        'data_categories': ['email addresses', 'encrypted passwords', 'garden locations'],
        'impact_assessment': 'Low risk of identity theft, potential spam emails',
        'containment_measures': ['Database access revoked', 'Passwords reset', 'Security patches applied'],
        'detection_time': '2024-01-15T10:30:00Z',
        'incident_start_time': '2024-01-15T09:00:00Z'
    }
    
    # Assess notification requirement
    assessment = notification_manager.assess_notification_requirement(incident_data)
    print(f"Notification assessment: {json.dumps(assessment, indent=2, default=str)}")
    
    if assessment['notification_required']:
        # Send notifications
        notification_manager.notify_supervisory_authority(incident_data)
        print("Supervisory authority notification sent")
```

This comprehensive incident response procedures document provides structured processes for handling security incidents from detection through post-incident activities. The procedures are designed to be practical, compliant with regulatory requirements, and adaptable to different types of security incidents that the Smart Garden Bot platform might face.