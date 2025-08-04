# Smart Garden Bot - Security Testing & Monitoring Plan

## Executive Summary

This document outlines a comprehensive security testing and monitoring strategy for the Smart Garden Bot platform. The plan encompasses automated security scanning, penetration testing, continuous security monitoring, threat detection, and incident response capabilities designed to maintain a robust security posture throughout the development lifecycle and production operations.

## Security Testing Strategy

### Static Application Security Testing (SAST)

#### Code Analysis Tools Configuration
```yaml
# GitHub Actions SAST Pipeline
name: Security Analysis
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * 1'  # Weekly scan

jobs:
  sast-analysis:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    # Semgrep Security Scanning
    - name: Run Semgrep
      uses: returntocorp/semgrep-action@v1
      with:
        config: >
          p/security-audit
          p/secrets
          p/owasp-top-ten
          p/golang
          p/typescript
        generateSarif: "1"
    
    # Upload results to GitHub Security
    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: semgrep.sarif
    
    # GoSec for Go-specific security issues
    - name: Run GoSec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-fmt sarif -out gosec-results.sarif ./...'
    
    # NodeJS Security Audit
    - name: NPM Security Audit
      run: |
        cd web-app
        npm audit --audit-level moderate
        npm audit --json > npm-audit.json
    
    # Secrets Detection
    - name: TruffleHog Secrets Scan
      uses: trufflesecurity/trufflehog@main
      with:
        path: ./
        base: main
        head: HEAD
        extra_args: --debug --only-verified
```

#### Code Quality Gates
```yaml
# SonarQube Quality Gate Configuration
sonarQube:
  qualityGate:
    conditions:
      - metric: "security_rating"
        operator: "GREATER_THAN"
        error: "1"  # A rating only
      - metric: "reliability_rating"
        operator: "GREATER_THAN"
        error: "1"
      - metric: "sqale_rating"
        operator: "GREATER_THAN"
        error: "1"
      - metric: "coverage"
        operator: "LESS_THAN"
        error: "80"
      - metric: "duplicated_lines_density"
        operator: "GREATER_THAN"
        error: "3"
      - metric: "vulnerabilities"
        operator: "GREATER_THAN"
        error: "0"
      - metric: "security_hotspots"
        operator: "GREATER_THAN"
        error: "0"
```

### Dynamic Application Security Testing (DAST)

#### OWASP ZAP Integration
```yaml
# ZAP Security Testing Pipeline
zapTesting:
  baseline:
    image: "owasp/zap2docker-stable"
    script: |
      zap-baseline.py \
        -t https://staging-api.smart-garden-bot.com \
        -g gen.conf \
        -r zap-baseline-report.html \
        -J zap-baseline-report.json \
        -w zap-baseline-report.md
  
  fullScan:
    image: "owasp/zap2docker-stable"
    script: |
      zap-full-scan.py \
        -t https://staging-api.smart-garden-bot.com \
        -g gen.conf \
        -r zap-full-report.html \
        -J zap-full-report.json \
        -a \
        -j \
        -m 10 \
        -z "-config authentication.method=httpAuthentication"
  
  apiScan:
    image: "owasp/zap2docker-stable"
    script: |
      zap-api-scan.py \
        -t https://staging-api.smart-garden-bot.com/openapi.json \
        -f openapi \
        -r zap-api-report.html \
        -J zap-api-report.json \
        -a \
        -x zap-api-report.xml
```

#### Custom Security Test Suite
```python
#!/usr/bin/env python3
"""
Custom security testing suite for Smart Garden Bot API
"""

import requests
import json
import time
from urllib.parse import urljoin

class SecurityTestSuite:
    def __init__(self, base_url, auth_token):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {auth_token}',
            'User-Agent': 'SecurityTest/1.0'
        })
    
    def test_authentication_bypass(self):
        """Test for authentication bypass vulnerabilities"""
        bypass_attempts = [
            {'Authorization': 'Bearer invalid_token'},
            {'Authorization': 'Bearer '},
            {'Authorization': ''},
            {},  # No auth header
        ]
        
        results = []
        for headers in bypass_attempts:
            response = requests.get(
                urljoin(self.base_url, '/api/v1/gardens'),
                headers=headers
            )
            results.append({
                'headers': headers,
                'status_code': response.status_code,
                'vulnerable': response.status_code == 200
            })
        
        return results
    
    def test_sql_injection(self):
        """Test for SQL injection vulnerabilities"""
        payloads = [
            "' OR '1'='1",
            "'; DROP TABLE users; --",
            "' UNION SELECT * FROM auth.users --",
            "1' AND SLEEP(5) --",
        ]
        
        results = []
        for payload in payloads:
            start_time = time.time()
            response = self.session.get(
                urljoin(self.base_url, '/api/v1/gardens'),
                params={'search': payload}
            )
            response_time = time.time() - start_time
            
            results.append({
                'payload': payload,
                'status_code': response.status_code,
                'response_time': response_time,
                'vulnerable': (
                    response.status_code == 200 and 
                    ('error' not in response.text.lower() or response_time > 4)
                )
            })
        
        return results
    
    def test_xss_vulnerabilities(self):
        """Test for Cross-Site Scripting vulnerabilities"""
        payloads = [
            "<script>alert('XSS')</script>",
            "javascript:alert('XSS')",
            "<img src=x onerror=alert('XSS')>",
            "';alert('XSS');//",
        ]
        
        results = []
        for payload in payloads:
            response = self.session.post(
                urljoin(self.base_url, '/api/v1/gardens'),
                json={'name': payload, 'location': 'Test Location'}
            )
            
            # Check if payload is reflected without sanitization
            if response.status_code == 201:
                garden_response = self.session.get(
                    urljoin(self.base_url, f'/api/v1/gardens/{response.json()["id"]}')
                )
                results.append({
                    'payload': payload,
                    'vulnerable': payload in garden_response.text
                })
        
        return results
    
    def test_idor_vulnerabilities(self):
        """Test for Insecure Direct Object Reference vulnerabilities"""
        # Create a test garden
        garden_response = self.session.post(
            urljoin(self.base_url, '/api/v1/gardens'),
            json={'name': 'IDOR Test Garden', 'location': 'Test'}
        )
        
        if garden_response.status_code != 201:
            return {'error': 'Could not create test garden'}
        
        garden_id = garden_response.json()['id']
        
        # Test accessing other users' resources
        idor_tests = []
        for test_id in range(1, 10):  # Test first 10 IDs
            if test_id == garden_id:
                continue
                
            response = self.session.get(
                urljoin(self.base_url, f'/api/v1/gardens/{test_id}')
            )
            
            idor_tests.append({
                'tested_id': test_id,
                'status_code': response.status_code,
                'vulnerable': response.status_code == 200
            })
        
        return idor_tests
    
    def run_all_tests(self):
        """Run all security tests"""
        return {
            'timestamp': time.time(),
            'authentication_bypass': self.test_authentication_bypass(),
            'sql_injection': self.test_sql_injection(),
            'xss_vulnerabilities': self.test_xss_vulnerabilities(),
            'idor_vulnerabilities': self.test_idor_vulnerabilities()
        }

if __name__ == "__main__":
    suite = SecurityTestSuite(
        base_url="https://staging-api.smart-garden-bot.com",
        auth_token="test_token_here"
    )
    
    results = suite.run_all_tests()
    
    with open('security_test_results.json', 'w') as f:
        json.dump(results, f, indent=2)
    
    print("Security testing completed. Results saved to security_test_results.json")
```

### Dependency Security Scanning

#### Automated Dependency Auditing
```yaml
# Dependency Security Scanning
dependencyScanning:
  go:
    tool: "nancy"
    config: |
      nancy sleuth --path go.sum
      govulncheck ./...
  
  nodejs:
    tool: "npm-audit"
    config: |
      npm audit --audit-level=moderate
      yarn audit --level moderate
  
  docker:
    tool: "snyk"
    config: |
      snyk container test smart-garden-bot:latest
      snyk code test .
      snyk iac test kubernetes/
  
  infrastructure:
    tool: "checkov"
    config: |
      checkov -d kubernetes/ --framework kubernetes
      checkov -d terraform/ --framework terraform
```

#### Vulnerability Management Process
```python
#!/usr/bin/env python3
"""
Vulnerability management and prioritization system
"""

from enum import Enum
from dataclasses import dataclass
from typing import List, Dict
import json

class VulnerabilitySeverity(Enum):
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    INFO = "info"

@dataclass
class Vulnerability:
    id: str
    title: str
    severity: VulnerabilitySeverity
    cvss_score: float
    affected_component: str
    description: str
    remediation: str
    exploitable: bool
    public_exploit: bool
    
class VulnerabilityManager:
    def __init__(self):
        self.sla_hours = {
            VulnerabilitySeverity.CRITICAL: 24,
            VulnerabilitySeverity.HIGH: 72,
            VulnerabilitySeverity.MEDIUM: 168,  # 1 week
            VulnerabilitySeverity.LOW: 720,    # 30 days
        }
    
    def prioritize_vulnerabilities(self, vulnerabilities: List[Vulnerability]) -> List[Vulnerability]:
        """Prioritize vulnerabilities based on severity and exploitability"""
        def priority_score(vuln):
            base_score = vuln.cvss_score
            
            # Increase priority for exploitable vulnerabilities
            if vuln.exploitable:
                base_score += 2
            
            # Increase priority for public exploits
            if vuln.public_exploit:
                base_score += 3
            
            # Component-based adjustments
            if vuln.affected_component in ['authentication', 'authorization', 'payment']:
                base_score += 1
            
            return base_score
        
        return sorted(vulnerabilities, key=priority_score, reverse=True)
    
    def generate_remediation_plan(self, vulnerabilities: List[Vulnerability]) -> Dict:
        """Generate a remediation plan with timelines"""
        prioritized = self.prioritize_vulnerabilities(vulnerabilities)
        
        plan = {
            'immediate_action': [],  # Critical/High with exploits
            'short_term': [],        # High severity
            'medium_term': [],       # Medium severity
            'long_term': []          # Low severity
        }
        
        for vuln in prioritized:
            if vuln.severity == VulnerabilitySeverity.CRITICAL or \
               (vuln.severity == VulnerabilitySeverity.HIGH and vuln.public_exploit):
                plan['immediate_action'].append(vuln)
            elif vuln.severity == VulnerabilitySeverity.HIGH:
                plan['short_term'].append(vuln)
            elif vuln.severity == VulnerabilitySeverity.MEDIUM:
                plan['medium_term'].append(vuln)
            else:
                plan['long_term'].append(vuln)
        
        return plan
```

## Continuous Security Monitoring

### Security Information and Event Management (SIEM)

#### ELK Stack Configuration
```yaml
# Elasticsearch configuration for security logs
elasticsearch:
  cluster.name: "smart-garden-bot-security"
  network.host: "0.0.0.0"
  discovery.type: "single-node"
  xpack.security.enabled: true
  xpack.security.transport.ssl.enabled: true
  
# Logstash pipeline for security events
logstash:
  pipeline:
    - pipeline.id: "security-events"
      config.string: |
        input {
          beats {
            port => 5044
            ssl => true
            ssl_certificate => "/etc/logstash/certs/logstash.crt"
            ssl_key => "/etc/logstash/certs/logstash.key"
          }
        }
        
        filter {
          if [fields][log_type] == "security" {
            grok {
              match => { 
                "message" => "%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{GREEDYDATA:message}" 
              }
            }
            
            # Parse authentication events
            if "authentication" in [message] {
              grok {
                match => { 
                  "message" => "user_id=%{UUID:user_id} action=%{WORD:action} result=%{WORD:result} ip=%{IP:client_ip}" 
                }
              }
              
              mutate {
                add_tag => ["authentication"]
              }
            }
            
            # Parse authorization events
            if "authorization" in [message] {
              grok {
                match => { 
                  "message" => "user_id=%{UUID:user_id} resource=%{URIPATH:resource} permission=%{WORD:permission} result=%{WORD:result}" 
                }
              }
              
              mutate {
                add_tag => ["authorization"]
              }
            }
            
            # Enrich with GeoIP data
            if [client_ip] {
              geoip {
                source => "client_ip"
                target => "geoip"
              }
            }
          }
        }
        
        output {
          elasticsearch {
            hosts => ["elasticsearch:9200"]
            index => "security-logs-%{+YYYY.MM.dd}"
          }
        }

# Kibana security dashboards
kibana:
  dashboards:
    - name: "Authentication Monitoring"
      panels:
        - failed_login_attempts
        - geographic_login_distribution
        - account_lockouts
        - suspicious_login_patterns
    
    - name: "API Security"
      panels:
        - api_error_rates
        - rate_limiting_triggers
        - suspicious_request_patterns
        - sql_injection_attempts
```

#### Filebeat Security Configuration
```yaml
# Filebeat configuration for security log shipping
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/smart-garden-bot/security.log
    - /var/log/smart-garden-bot/audit.log
    - /var/log/nginx/access.log
    - /var/log/nginx/error.log
  fields:
    log_type: security
  fields_under_root: true
  multiline.pattern: '^\d{4}-\d{2}-\d{2}'
  multiline.negate: true
  multiline.match: after

processors:
- add_host_metadata:
    when.not.contains.tags: forwarded
- add_docker_metadata: ~
- add_kubernetes_metadata: ~

output.logstash:
  hosts: ["logstash:5044"]
  ssl.enabled: true
  ssl.certificate: "/etc/filebeat/certs/filebeat.crt"
  ssl.key: "/etc/filebeat/certs/filebeat.key"
  ssl.certificate_authorities: ["/etc/filebeat/certs/ca.crt"]

logging.level: info
logging.to_files: true
logging.files:
  path: /var/log/filebeat
  name: filebeat
  keepfiles: 7
  permissions: 0600
```

### Real-time Threat Detection

#### Anomaly Detection Rules
```python
#!/usr/bin/env python3
"""
Real-time security anomaly detection system
"""

import json
import redis
import time
from collections import defaultdict, deque
from datetime import datetime, timedelta

class SecurityAnomalyDetector:
    def __init__(self, redis_client):
        self.redis = redis_client
        self.thresholds = {
            'failed_logins_per_hour': 10,
            'api_errors_per_minute': 50,
            'unusual_location_login': True,
            'off_hours_admin_access': True,
            'bulk_data_access': 1000,  # Records per minute
            'rapid_api_key_creation': 5,  # Per hour
        }
        
        # Sliding window counters
        self.windows = {
            'failed_logins': deque(maxlen=3600),  # 1 hour window
            'api_errors': deque(maxlen=60),       # 1 minute window
            'data_access': deque(maxlen=60),      # 1 minute window
        }
    
    def detect_brute_force_attack(self, event):
        """Detect brute force authentication attempts"""
        if event.get('action') != 'login_failed':
            return None
        
        user_id = event.get('user_id')
        ip_address = event.get('client_ip')
        timestamp = datetime.now()
        
        # Track failed attempts per user
        user_key = f"failed_logins:user:{user_id}"
        user_failures = self.redis.incr(user_key)
        self.redis.expire(user_key, 3600)  # 1 hour expiry
        
        # Track failed attempts per IP
        ip_key = f"failed_logins:ip:{ip_address}"
        ip_failures = self.redis.incr(ip_key)
        self.redis.expire(ip_key, 3600)
        
        if user_failures >= 5 or ip_failures >= 10:
            return {
                'type': 'brute_force_attack',
                'severity': 'HIGH',
                'user_id': user_id,
                'client_ip': ip_address,
                'user_failures': user_failures,
                'ip_failures': ip_failures,
                'timestamp': timestamp.isoformat(),
                'action_required': 'account_lockout'
            }
        
        return None
    
    def detect_geographical_anomaly(self, event):
        """Detect login from unusual geographical location"""
        if event.get('action') != 'login_success':
            return None
        
        user_id = event.get('user_id')
        country = event.get('geoip', {}).get('country_name')
        
        if not country:
            return None
        
        # Get user's historical locations
        location_key = f"user_locations:{user_id}"
        historical_locations = self.redis.smembers(location_key)
        
        if country.encode() not in historical_locations and len(historical_locations) > 0:
            # Store new location
            self.redis.sadd(location_key, country)
            self.redis.expire(location_key, 86400 * 365)  # 1 year
            
            return {
                'type': 'geographical_anomaly',
                'severity': 'MEDIUM',
                'user_id': user_id,
                'new_country': country,
                'historical_countries': [loc.decode() for loc in historical_locations],
                'timestamp': datetime.now().isoformat(),
                'action_required': 'user_notification'
            }
        
        # Add to historical locations if first login
        if len(historical_locations) == 0:
            self.redis.sadd(location_key, country)
            self.redis.expire(location_key, 86400 * 365)
        
        return None
    
    def detect_data_exfiltration(self, event):
        """Detect potential data exfiltration attempts"""
        if event.get('action') != 'data_access':
            return None
        
        user_id = event.get('user_id')
        records_accessed = event.get('records_count', 0)
        timestamp = datetime.now()
        
        # Track data access volume
        access_key = f"data_access:{user_id}"
        current_volume = self.redis.get(access_key) or 0
        current_volume = int(current_volume) + records_accessed
        
        self.redis.setex(access_key, 60, current_volume)  # 1 minute window
        
        if current_volume > self.thresholds['bulk_data_access']:
            return {
                'type': 'potential_data_exfiltration',
                'severity': 'CRITICAL',
                'user_id': user_id,
                'records_accessed': current_volume,
                'threshold': self.thresholds['bulk_data_access'],
                'timestamp': timestamp.isoformat(),
                'action_required': 'immediate_investigation'
            }
        
        return None
    
    def detect_privilege_escalation(self, event):
        """Detect potential privilege escalation attempts"""
        if event.get('action') != 'authorization_denied':
            return None
        
        user_id = event.get('user_id')
        resource = event.get('resource')
        permission = event.get('permission')
        
        # Track authorization failures
        failure_key = f"auth_failures:{user_id}"
        failures = self.redis.incr(failure_key)
        self.redis.expire(failure_key, 3600)  # 1 hour window
        
        # Check for admin resource access attempts
        admin_resources = ['/admin', '/users', '/system', '/billing']
        is_admin_resource = any(admin_res in resource for admin_res in admin_resources)
        
        if failures >= 5 and is_admin_resource:
            return {
                'type': 'privilege_escalation_attempt',
                'severity': 'HIGH',
                'user_id': user_id,
                'attempted_resource': resource,
                'attempted_permission': permission,
                'failure_count': failures,
                'timestamp': datetime.now().isoformat(),
                'action_required': 'security_review'
            }
        
        return None
    
    def process_security_event(self, event):
        """Process incoming security event and detect anomalies"""
        anomalies = []
        
        # Run all detection rules
        detectors = [
            self.detect_brute_force_attack,
            self.detect_geographical_anomaly,
            self.detect_data_exfiltration,
            self.detect_privilege_escalation
        ]
        
        for detector in detectors:
            try:
                anomaly = detector(event)
                if anomaly:
                    anomalies.append(anomaly)
            except Exception as e:
                print(f"Error in detector {detector.__name__}: {e}")
        
        return anomalies

# Example usage
if __name__ == "__main__":
    redis_client = redis.Redis(host='localhost', port=6379, db=0)
    detector = SecurityAnomalyDetector(redis_client)
    
    # Simulate security events
    events = [
        {
            'action': 'login_failed',
            'user_id': 'user_123',
            'client_ip': '192.168.1.100',
            'timestamp': datetime.now().isoformat()
        },
        {
            'action': 'login_success',
            'user_id': 'user_456',
            'client_ip': '203.0.113.50',
            'geoip': {'country_name': 'Russia'},
            'timestamp': datetime.now().isoformat()
        }
    ]
    
    for event in events:
        anomalies = detector.process_security_event(event)
        for anomaly in anomalies:
            print(f"SECURITY ANOMALY DETECTED: {json.dumps(anomaly, indent=2)}")
```

### Infrastructure Monitoring

#### Kubernetes Security Monitoring
```yaml
# Falco rules for Kubernetes security monitoring
apiVersion: v1
kind: ConfigMap
metadata:
  name: falco-rules
  namespace: smart-garden-bot
data:
  k8s_audit_rules.yaml: |
    - rule: Detect crypto miners
      desc: Detect cryptocurrency miners based on common process names
      condition: >
        spawned_process and 
        (proc.name in (crypto_miners) or 
         proc.cmdline contains "xmrig" or 
         proc.cmdline contains "stratum")
      output: Crypto miner detected (user=%user.name command=%proc.cmdline)
      priority: CRITICAL
    
    - rule: Detect privilege escalation
      desc: Detect attempts to escalate privileges
      condition: >
        spawned_process and 
        (proc.name in (su, sudo, doas) or 
         proc.cmdline contains "chmod +s")
      output: Privilege escalation attempt (user=%user.name command=%proc.cmdline)
      priority: HIGH
    
    - rule: Detect sensitive file access
      desc: Detect access to sensitive files
      condition: >
        open_read and 
        fd.name in (/etc/passwd, /etc/shadow, /etc/sudoers, 
                     /root/.ssh/id_rsa, /home/*/.ssh/id_rsa)
      output: Sensitive file accessed (user=%user.name file=%fd.name)
      priority: HIGH
    
    - rule: Detect network connections from containers
      desc: Detect unexpected network connections from containers
      condition: >
        inbound_outbound and 
        container and 
        not proc.name in (curl, wget, node, go, postgres)
      output: Unexpected network connection (container=%container.name dest=%fd.rip)
      priority: MEDIUM

---
# Falco deployment
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: falco
  namespace: smart-garden-bot
spec:
  selector:
    matchLabels:
      app: falco
  template:
    metadata:
      labels:
        app: falco
    spec:
      serviceAccount: falco
      hostNetwork: true
      hostPID: true
      containers:
      - name: falco
        image: falcosecurity/falco:latest
        securityContext:
          privileged: true
        volumeMounts:
        - name: proc
          mountPath: /host/proc
          readOnly: true
        - name: boot
          mountPath: /host/boot
          readOnly: true
        - name: lib-modules
          mountPath: /host/lib/modules
          readOnly: true
        - name: dev
          mountPath: /host/dev
        - name: falco-rules
          mountPath: /etc/falco/rules.d
      volumes:
      - name: proc
        hostPath:
          path: /proc
      - name: boot
        hostPath:
          path: /boot
      - name: lib-modules
        hostPath:
          path: /lib/modules
      - name: dev
        hostPath:
          path: /dev
      - name: falco-rules
        configMap:
          name: falco-rules
```

#### Network Security Monitoring
```yaml
# Network monitoring with Istio service mesh
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: smart-garden-bot-authz
  namespace: smart-garden-bot
spec:
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/smart-garden-bot/sa/api"]
    to:
    - operation:
        methods: ["GET", "POST", "PUT", "DELETE"]
    when:
    - key: source.ip
      notValues: ["0.0.0.0/0"]  # Block traffic from unknown sources
  
  - from:
    - source:
        namespaces: ["ingress-nginx"]
    to:
    - operation:
        methods: ["GET", "POST"]
        paths: ["/api/v1/*"]

---
# Istio telemetry for security monitoring
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: security-metrics
  namespace: smart-garden-bot
spec:
  metrics:
  - providers:
    - name: prometheus
  - overrides:
    - match:
        metric: ALL_METRICS
      tagOverrides:
        source_workload:
          value: "%{source_workload}"
        destination_service:
          value: "%{destination_service_name}"
        response_code:
          value: "%{response_code}"
        request_protocol:
          value: "%{request_protocol}"
```

## Incident Response Procedures

### Security Incident Classification
```python
#!/usr/bin/env python3
"""
Security incident classification and response automation
"""

from enum import Enum
from dataclasses import dataclass
from datetime import datetime, timedelta
import json

class IncidentSeverity(Enum):
    CRITICAL = "critical"    # Data breach, system compromise
    HIGH = "high"           # Unauthorized access, service disruption
    MEDIUM = "medium"       # Security policy violation, failed attack
    LOW = "low"            # Suspicious activity, policy deviation

class IncidentCategory(Enum):
    DATA_BREACH = "data_breach"
    UNAUTHORIZED_ACCESS = "unauthorized_access"
    MALWARE = "malware"
    DOS_ATTACK = "dos_attack"
    INSIDER_THREAT = "insider_threat"
    POLICY_VIOLATION = "policy_violation"
    SYSTEM_COMPROMISE = "system_compromise"

@dataclass
class SecurityIncident:
    id: str
    title: str
    description: str
    severity: IncidentSeverity
    category: IncidentCategory
    detected_at: datetime
    source: str
    affected_systems: list
    affected_users: list
    estimated_impact: str
    containment_status: str = "open"
    assigned_to: str = None

class IncidentResponseManager:
    def __init__(self):
        self.escalation_matrix = {
            IncidentSeverity.CRITICAL: {
                'notification_time': 15,  # minutes
                'response_time': 60,      # minutes
                'resolution_time': 4,     # hours
                'stakeholders': ['ciso', 'ceo', 'legal', 'pr']
            },
            IncidentSeverity.HIGH: {
                'notification_time': 30,
                'response_time': 120,
                'resolution_time': 24,
                'stakeholders': ['ciso', 'security_team', 'engineering']
            },
            IncidentSeverity.MEDIUM: {
                'notification_time': 60,
                'response_time': 240,
                'resolution_time': 72,
                'stakeholders': ['security_team', 'engineering']
            },
            IncidentSeverity.LOW: {
                'notification_time': 120,
                'response_time': 480,
                'resolution_time': 168,
                'stakeholders': ['security_team']
            }
        }
    
    def classify_incident(self, event_data):
        """Automatically classify security incidents"""
        incident_type = event_data.get('type', '').lower()
        severity = IncidentSeverity.LOW
        category = IncidentCategory.POLICY_VIOLATION
        
        # Classification rules
        if 'data_breach' in incident_type or 'exfiltration' in incident_type:
            severity = IncidentSeverity.CRITICAL
            category = IncidentCategory.DATA_BREACH
        elif 'unauthorized_access' in incident_type or 'privilege_escalation' in incident_type:
            severity = IncidentSeverity.HIGH
            category = IncidentCategory.UNAUTHORIZED_ACCESS
        elif 'brute_force' in incident_type or 'dos' in incident_type:
            severity = IncidentSeverity.HIGH
            category = IncidentCategory.DOS_ATTACK
        elif 'malware' in incident_type or 'crypto_miner' in incident_type:
            severity = IncidentSeverity.CRITICAL
            category = IncidentCategory.MALWARE
        elif 'geographical_anomaly' in incident_type:
            severity = IncidentSeverity.MEDIUM
            category = IncidentCategory.POLICY_VIOLATION
        
        return SecurityIncident(
            id=f"INC-{datetime.now().strftime('%Y%m%d-%H%M%S')}",
            title=event_data.get('title', f"Security Incident: {incident_type}"),
            description=event_data.get('description', ''),
            severity=severity,
            category=category,
            detected_at=datetime.now(),
            source=event_data.get('source', 'automated_detection'),
            affected_systems=event_data.get('affected_systems', []),
            affected_users=event_data.get('affected_users', []),
            estimated_impact=self.estimate_impact(severity, event_data)
        )
    
    def estimate_impact(self, severity, event_data):
        """Estimate incident impact"""
        if severity == IncidentSeverity.CRITICAL:
            return "High - Potential data breach or system compromise"
        elif severity == IncidentSeverity.HIGH:
            return "Medium - Service disruption or unauthorized access"
        elif severity == IncidentSeverity.MEDIUM:
            return "Low - Security policy violation"
        else:
            return "Minimal - Suspicious activity detected"
    
    def initiate_response(self, incident: SecurityIncident):
        """Initiate incident response procedures"""
        response_plan = {
            'incident_id': incident.id,
            'severity': incident.severity.value,
            'immediate_actions': [],
            'investigation_steps': [],
            'communication_plan': [],
            'containment_actions': []
        }
        
        # Define immediate actions based on incident type
        if incident.category == IncidentCategory.DATA_BREACH:
            response_plan['immediate_actions'] = [
                'Isolate affected systems',
                'Preserve evidence',
                'Notify legal team',
                'Prepare breach notification',
                'Contact law enforcement if required'
            ]
        
        elif incident.category == IncidentCategory.UNAUTHORIZED_ACCESS:
            response_plan['immediate_actions'] = [
                'Disable compromised accounts',
                'Reset affected passwords',
                'Review access logs',
                'Check for lateral movement',
                'Update access controls'
            ]
        
        elif incident.category == IncidentCategory.DOS_ATTACK:
            response_plan['immediate_actions'] = [
                'Implement rate limiting',
                'Block malicious IPs',
                'Scale infrastructure if needed',
                'Monitor service availability',
                'Prepare customer communication'
            ]
        
        # Add investigation steps
        response_plan['investigation_steps'] = [
            'Collect and preserve logs',
            'Interview relevant personnel',
            'Analyze attack vectors',
            'Assess damage scope',
            'Document timeline of events'
        ]
        
        # Communication plan
        stakeholders = self.escalation_matrix[incident.severity]['stakeholders']
        response_plan['communication_plan'] = [
            f'Notify {stakeholder}' for stakeholder in stakeholders
        ]
        
        return response_plan
```

### Automated Response Actions
```bash
#!/bin/bash
# Automated incident response script

INCIDENT_ID=$1
INCIDENT_TYPE=$2
SEVERITY=$3

log_action() {
    echo "$(date): $1" >> /var/log/incident-response.log
}

block_ip_address() {
    local ip=$1
    log_action "Blocking IP address: $ip"
    
    # Add to iptables
    iptables -A INPUT -s $ip -j DROP
    
    # Add to fail2ban
    fail2ban-client set sshd banip $ip
    
    # Update WAF rules
    curl -X POST "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/firewall/access_rules/rules" \
         -H "Authorization: Bearer $CLOUDFLARE_TOKEN" \
         -H "Content-Type: application/json" \
         --data "{\"mode\":\"block\",\"configuration\":{\"target\":\"ip\",\"value\":\"$ip\"},\"notes\":\"Blocked by incident response: $INCIDENT_ID\"}"
}

disable_user_account() {
    local user_id=$1
    log_action "Disabling user account: $user_id"
    
    # Disable in database
    psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "UPDATE auth.users SET status='disabled', disabled_reason='security_incident', disabled_at=NOW() WHERE id='$user_id';"
    
    # Revoke all active sessions
    redis-cli DEL "user_session:$user_id:*"
    
    # Disable API keys
    psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "UPDATE auth.api_keys SET status='disabled' WHERE user_id='$user_id';"
}

isolate_container() {
    local container_name=$1
    log_action "Isolating container: $container_name"
    
    # Create network policy to isolate container
    kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: isolate-$container_name
  namespace: smart-garden-bot
spec:
  podSelector:
    matchLabels:
      app: $container_name
  policyTypes:
  - Ingress
  - Egress
  # No ingress or egress rules = deny all
EOF
}

collect_forensic_data() {
    local system=$1
    log_action "Collecting forensic data from: $system"
    
    # Create forensic data directory
    mkdir -p /var/lib/incident-response/$INCIDENT_ID
    
    # Collect system logs
    journalctl --since="1 hour ago" > /var/lib/incident-response/$INCIDENT_ID/system.log
    
    # Collect application logs
    kubectl logs -n smart-garden-bot -l app=$system --since=1h > /var/lib/incident-response/$INCIDENT_ID/app-$system.log
    
    # Collect network connections
    netstat -tulpn > /var/lib/incident-response/$INCIDENT_ID/network-connections.txt
    
    # Collect process list
    ps aux > /var/lib/incident-response/$INCIDENT_ID/processes.txt
    
    # Create evidence package
    tar -czf /var/lib/incident-response/$INCIDENT_ID-evidence.tar.gz /var/lib/incident-response/$INCIDENT_ID/
}

notify_stakeholders() {
    local message=$1
    log_action "Sending notifications: $message"
    
    # Send to Slack
    curl -X POST -H 'Content-type: application/json' \
         --data "{\"text\":\"🚨 Security Incident Alert\n**Incident ID:** $INCIDENT_ID\n**Severity:** $SEVERITY\n**Type:** $INCIDENT_TYPE\n**Message:** $message\"}" \
         $SLACK_WEBHOOK_URL
    
    # Send email alerts
    echo "$message" | mail -s "Security Incident: $INCIDENT_ID" security@smart-garden-bot.com
    
    # Update incident tracking system
    curl -X POST -H 'Content-Type: application/json' \
         --data "{\"incident_id\":\"$INCIDENT_ID\",\"status\":\"investigating\",\"message\":\"$message\"}" \
         https://incident-tracker.smart-garden-bot.com/api/incidents
}

# Main incident response logic
case $INCIDENT_TYPE in
    "brute_force_attack")
        IP_ADDRESS=$(grep "client_ip" /var/log/incident-$INCIDENT_ID.json | head -1 | jq -r '.client_ip')
        block_ip_address $IP_ADDRESS
        notify_stakeholders "Brute force attack detected and blocked from IP: $IP_ADDRESS"
        ;;
        
    "unauthorized_access")
        USER_ID=$(grep "user_id" /var/log/incident-$INCIDENT_ID.json | head -1 | jq -r '.user_id')
        disable_user_account $USER_ID
        collect_forensic_data "api"
        notify_stakeholders "Unauthorized access detected. User account $USER_ID has been disabled."
        ;;
        
    "data_exfiltration")
        USER_ID=$(grep "user_id" /var/log/incident-$INCIDENT_ID.json | head -1 | jq -r '.user_id')
        disable_user_account $USER_ID
        collect_forensic_data "api"
        isolate_container "api"
        notify_stakeholders "CRITICAL: Potential data exfiltration detected. System isolated and evidence collected."
        ;;
        
    "malware_detected")
        CONTAINER=$(grep "container" /var/log/incident-$INCIDENT_ID.json | head -1 | jq -r '.container_name')
        isolate_container $CONTAINER
        collect_forensic_data $CONTAINER
        notify_stakeholders "Malware detected in container $CONTAINER. Container isolated."
        ;;
        
    *)
        log_action "Unknown incident type: $INCIDENT_TYPE"
        collect_forensic_data "all"
        notify_stakeholders "Security incident detected. Type: $INCIDENT_TYPE. Investigation initiated."
        ;;
esac

log_action "Incident response completed for $INCIDENT_ID"
```

This comprehensive security testing and monitoring plan provides automated security scanning, continuous threat detection, and incident response capabilities to maintain a robust security posture for the Smart Garden Bot platform. The plan integrates with existing DevOps workflows and provides real-time visibility into security threats and vulnerabilities.