# Smart Garden Bot - Security Architecture

## Executive Summary

This document outlines the comprehensive security architecture for the Smart Garden Bot platform, a cloud-native SaaS solution for automated garden irrigation management. The security framework addresses authentication, authorization, data protection, infrastructure security, and compliance requirements while maintaining usability and performance.

## Security Principles

### 1. Defense in Depth
- Multiple layers of security controls
- No single point of failure in security architecture
- Comprehensive monitoring and detection capabilities

### 2. Zero Trust Architecture
- Never trust, always verify
- Least privilege access principles
- Continuous verification of all entities

### 3. Privacy by Design
- Data minimization principles
- User consent management
- Transparent data handling practices

### 4. Security as Code
- Infrastructure as Code security practices
- Automated security testing and compliance
- Version-controlled security configurations

## Authentication & Authorization Architecture

### OpenID Connect Implementation

#### Primary Identity Provider: Auth0
```yaml
# Auth0 Configuration
auth0:
  domain: "smart-garden-bot.auth0.com"
  clientId: "${AUTH0_CLIENT_ID}"
  clientSecret: "${AUTH0_CLIENT_SECRET}"
  audience: "https://api.smart-garden-bot.com"
  scope: "openid profile email garden:read garden:write"
  algorithms: ["RS256"]
```

#### Fallback Identity Provider: KeyCloak
```yaml
# KeyCloak Configuration (Development/Self-hosted)
keycloak:
  realm: "smart-garden-bot"
  serverUrl: "https://keycloak.smart-garden-bot.com"
  clientId: "web-app"
  publicClient: false
  confidentialPort: 0
```

### JWT Token Handling

#### Token Structure
```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "key-id-123"
  },
  "payload": {
    "iss": "https://smart-garden-bot.auth0.com/",
    "sub": "auth0|507f1f77bcf86cd799439011",
    "aud": "https://api.smart-garden-bot.com",
    "exp": 1642684800,
    "iat": 1642681200,
    "scope": "garden:read garden:write sensor:read",
    "tenant_id": "tenant_12345",
    "roles": ["user", "garden_owner"]
  }
}
```

#### Token Validation Strategy
```go
// JWT Validation Middleware
type JWTValidator struct {
    jwksClient *jwks.Client
    issuer     string
    audience   string
}

func (v *JWTValidator) ValidateToken(tokenString string) (*jwt.Token, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        
        // Get key from JWKS endpoint
        kid := token.Header["kid"].(string)
        key, err := v.jwksClient.GetKey(kid)
        if err != nil {
            return nil, err
        }
        
        return key, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // Validate claims
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if !claims.VerifyIssuer(v.issuer, true) {
            return nil, errors.New("invalid issuer")
        }
        if !claims.VerifyAudience(v.audience, true) {
            return nil, errors.New("invalid audience")
        }
        if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
            return nil, errors.New("token expired")
        }
    }
    
    return token, nil
}
```

### API Key Management

#### Machine-to-Machine Authentication
```go
// API Key Structure
type APIKey struct {
    ID          string    `json:"id" db:"id"`
    UserID      string    `json:"user_id" db:"user_id"`
    Name        string    `json:"name" db:"name"`
    KeyHash     string    `json:"-" db:"key_hash"`
    Permissions []string  `json:"permissions" db:"permissions"`
    LastUsed    time.Time `json:"last_used" db:"last_used"`
    ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// API Key Generation
func GenerateAPIKey() (string, string, error) {
    // Generate random key
    keyBytes := make([]byte, 32)
    if _, err := rand.Read(keyBytes); err != nil {
        return "", "", err
    }
    
    key := "sgb_" + base64.URLEncoding.EncodeToString(keyBytes)
    
    // Hash key for storage
    hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
    if err != nil {
        return "", "", err
    }
    
    return key, string(hash), nil
}
```

#### API Key Validation Middleware
```go
func APIKeyAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        // Extract API key
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(401, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }
        
        apiKey := parts[1]
        if !strings.HasPrefix(apiKey, "sgb_") {
            c.JSON(401, gin.H{"error": "Invalid API key format"})
            c.Abort()
            return
        }
        
        // Validate API key
        key, err := validateAPIKey(apiKey)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }
        
        // Set user context
        c.Set("user_id", key.UserID)
        c.Set("permissions", key.Permissions)
        c.Next()
    }
}
```

### Role-Based Access Control (RBAC)

#### Role Definitions
```yaml
roles:
  admin:
    description: "Platform administrator"
    permissions:
      - "platform:*"
      - "user:*"
      - "garden:*"
      - "billing:*"
      - "analytics:*"
  
  user:
    description: "Standard garden owner"
    permissions:
      - "garden:read"
      - "garden:write"
      - "sensor:read"
      - "watering:read"
      - "watering:write"
      - "profile:read"
      - "profile:write"
  
  device:
    description: "IoT device authentication"
    permissions:
      - "sensor:write"
      - "device:heartbeat"
  
  readonly:
    description: "Read-only access for monitoring"
    permissions:
      - "garden:read"
      - "sensor:read"
      - "watering:read"
```

#### Permission Validation
```go
type Permission struct {
    Resource string `json:"resource"`
    Action   string `json:"action"`
}

func (p Permission) String() string {
    return fmt.Sprintf("%s:%s", p.Resource, p.Action)
}

func HasPermission(userPermissions []string, required Permission) bool {
    requiredPerm := required.String()
    
    for _, perm := range userPermissions {
        // Check exact match
        if perm == requiredPerm {
            return true
        }
        
        // Check wildcard permissions
        if strings.HasSuffix(perm, ":*") {
            resource := strings.TrimSuffix(perm, ":*")
            if required.Resource == resource {
                return true
            }
        }
        
        // Check global admin
        if perm == "platform:*" {
            return true
        }
    }
    
    return false
}
```

### Multi-Tenant Security Isolation

#### Tenant Context Isolation
```go
type TenantContext struct {
    TenantID   string `json:"tenant_id"`
    UserID     string `json:"user_id"`
    Roles      []string `json:"roles"`
    GardenIDs  []string `json:"garden_ids"`
}

// Database Row-Level Security
func ApplyTenantFilters(query *gorm.DB, ctx *TenantContext) *gorm.DB {
    // Apply tenant isolation at database level
    return query.Where("tenant_id = ?", ctx.TenantID)
}

// Middleware for tenant context
func TenantContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // Load tenant context
        ctx, err := loadTenantContext(userID)
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to load tenant context"})
            c.Abort()
            return
        }
        
        c.Set("tenant_context", ctx)
        c.Next()
    }
}
```

## Data Security Architecture

### Encryption at Rest

#### Database Encryption
```yaml
# PostgreSQL Encryption Configuration
postgresql:
  ssl: true
  sslmode: require
  encryption:
    tde: true  # Transparent Data Encryption
    algorithm: AES-256
    key_management: vault
  backup_encryption: true
```

#### Application-Level Encryption
```go
// Sensitive Data Encryption
type EncryptionService struct {
    key []byte
}

func NewEncryptionService(key string) *EncryptionService {
    keyBytes, _ := hex.DecodeString(key)
    return &EncryptionService{key: keyBytes}
}

func (e *EncryptionService) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *EncryptionService) Decrypt(ciphertext string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }
    
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}
```

### Encryption in Transit

#### TLS Configuration
```yaml
# Ingress TLS Configuration
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: smart-garden-bot-ingress
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/ssl-protocols: "TLSv1.2 TLSv1.3"
    nginx.ingress.kubernetes.io/ssl-ciphers: "ECDHE-RSA-AES128-GCM-SHA256,ECDHE-RSA-AES256-GCM-SHA384"
spec:
  tls:
  - hosts:
    - api.smart-garden-bot.com
    - app.smart-garden-bot.com
    secretName: smart-garden-bot-tls
  rules:
  - host: api.smart-garden-bot.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api
            port:
              number: 8080
```

### Database Security

#### PostgreSQL Security Configuration
```sql
-- Enable Row Level Security
ALTER TABLE gardens.gardens ENABLE ROW LEVEL SECURITY;
ALTER TABLE gardens.zones ENABLE ROW LEVEL SECURITY;
ALTER TABLE sensors.readings ENABLE ROW LEVEL SECURITY;

-- Create RLS Policies
CREATE POLICY tenant_isolation_gardens ON gardens.gardens
    FOR ALL TO application_role
    USING (user_id = current_setting('app.current_user_id')::uuid);

CREATE POLICY tenant_isolation_zones ON gardens.zones
    FOR ALL TO application_role
    USING (garden_id IN (
        SELECT id FROM gardens.gardens 
        WHERE user_id = current_setting('app.current_user_id')::uuid
    ));

CREATE POLICY tenant_isolation_sensor_readings ON sensors.readings
    FOR ALL TO application_role
    USING (device_id IN (
        SELECT d.id FROM gardens.devices d
        JOIN gardens.gardens g ON d.garden_id = g.id
        WHERE g.user_id = current_setting('app.current_user_id')::uuid
    ));

-- Create dedicated database roles
CREATE ROLE application_role NOINHERIT;
CREATE ROLE readonly_role NOINHERIT;
CREATE ROLE migration_role NOINHERIT;

-- Grant appropriate permissions
GRANT CONNECT ON DATABASE smart_garden_bot TO application_role;
GRANT USAGE ON SCHEMA gardens, sensors, auth, billing TO application_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA gardens, sensors TO application_role;
GRANT SELECT ON ALL TABLES IN SCHEMA auth, billing TO application_role;
```

### Input Validation and Sanitization

#### API Input Validation
```go
// Request validation using gin-validator
type CreateGardenRequest struct {
    Name        string  `json:"name" binding:"required,min=1,max=100" validate:"sanitized"`
    Location    string  `json:"location" binding:"required,max=200" validate:"sanitized"`
    Timezone    string  `json:"timezone" binding:"required" validate:"timezone"`
    Latitude    float64 `json:"latitude" binding:"required,min=-90,max=90"`
    Longitude   float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

// Custom sanitization validator
func init() {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("sanitized", validateSanitized)
        v.RegisterValidation("timezone", validateTimezone)
    }
}

func validateSanitized(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    sanitized := html.EscapeString(strings.TrimSpace(value))
    return value == sanitized
}

func validateTimezone(fl validator.FieldLevel) bool {
    _, err := time.LoadLocation(fl.Field().String())
    return err == nil
}
```

#### SQL Injection Prevention
```go
// Parameterized queries using GORM
func (r *GardenRepository) GetGardensByUser(userID string) ([]Garden, error) {
    var gardens []Garden
    err := r.db.Where("user_id = ?", userID).Find(&gardens).Error
    return gardens, err
}

// Raw query protection
func (r *GardenRepository) GetGardenWithCustomQuery(userID string, filters map[string]interface{}) ([]Garden, error) {
    var gardens []Garden
    query := r.db.Model(&Garden{}).Where("user_id = ?", userID)
    
    // Validate and sanitize filter keys
    allowedFilters := map[string]bool{
        "name": true,
        "location": true,
        "created_at": true,
    }
    
    for key, value := range filters {
        if !allowedFilters[key] {
            return nil, fmt.Errorf("invalid filter key: %s", key)
        }
        query = query.Where(fmt.Sprintf("%s = ?", key), value)
    }
    
    err := query.Find(&gardens).Error
    return gardens, err
}
```

#### XSS Prevention
```go
// Content Security Policy headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // CSP header
        csp := "default-src 'self'; " +
               "script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; " +
               "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
               "font-src 'self' https://fonts.gstatic.com; " +
               "img-src 'self' data: https:; " +
               "connect-src 'self' https://api.smart-garden-bot.com wss://api.smart-garden-bot.com"
               
        c.Header("Content-Security-Policy", csp)
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        c.Next()
    }
}

// HTML sanitization for user content
import "github.com/microcosm-cc/bluemonday"

func SanitizeHTML(input string) string {
    p := bluemonday.StrictPolicy()
    return p.Sanitize(input)
}
```

## Infrastructure Security

### Kubernetes Security Best Practices

#### Pod Security Standards
```yaml
# Pod Security Policy
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: smart-garden-bot-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

#### Network Policies
```yaml
# API Server Network Policy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-netpol
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
          name: ingress-nginx
    - podSelector:
        matchLabels:
          app: web-app
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgresql
    ports:
    - protocol: TCP
      port: 5432
  - to: []  # Allow external API calls (weather, auth)
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 80

---
# Database Network Policy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: postgresql-netpol
  namespace: smart-garden-bot
spec:
  podSelector:
    matchLabels:
      app: postgresql
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: api
    - podSelector:
        matchLabels:
          app: kubernetes-operator
    ports:
    - protocol: TCP
      port: 5432
```

### Container Security

#### Secure Dockerfile Practices
```dockerfile
# Multi-stage build for API server
FROM golang:1.21-alpine AS builder

# Create non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM scratch

# Copy CA certificates for HTTPS calls
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy user account
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary
COPY --from=builder /app/main /main

# Use non-root user
USER appuser

EXPOSE 8080

ENTRYPOINT ["/main"]
```

#### Container Scanning Configuration
```yaml
# Trivy container scanning
apiVersion: v1
kind: ConfigMap
metadata:
  name: trivy-config
data:
  trivy.yaml: |
    security:
      vuln-type: ["os", "library"]
      severity: ["HIGH", "CRITICAL"]
      ignore-unfixed: true
    report:
      format: "json"
      output: "/tmp/trivy-report.json"
```

### Secrets Management

#### Sealed Secrets Configuration
```yaml
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: app-secrets
  namespace: smart-garden-bot
spec:
  encryptedData:
    database-password: AgBy1...  # Encrypted with cluster key
    auth0-client-secret: AgAz2...
    stripe-secret-key: AgBx3...
    encryption-key: AgCy4...
  template:
    metadata:
      name: app-secrets
      namespace: smart-garden-bot
    type: Opaque
```

#### External Secrets Integration
```yaml
# External Secrets Operator configuration
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-secret-store
  namespace: smart-garden-bot
spec:
  provider:
    vault:
      server: "https://vault.smart-garden-bot.com"
      path: "kv"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "smart-garden-bot"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-secrets
  namespace: smart-garden-bot
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-secret-store
    kind: SecretStore
  target:
    name: app-secrets
    creationPolicy: Owner
  data:
  - secretKey: database-password
    remoteRef:
      key: smart-garden-bot/database
      property: password
  - secretKey: auth0-client-secret
    remoteRef:
      key: smart-garden-bot/auth0
      property: client-secret
```

### Certificate Management

#### cert-manager Configuration
```yaml
# Let's Encrypt ClusterIssuer
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@smart-garden-bot.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
    - dns01:
        cloudflare:
          email: admin@smart-garden-bot.com
          apiTokenSecretRef:
            name: cloudflare-api-token
            key: api-token

---
# Certificate for internal services
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: internal-tls
  namespace: smart-garden-bot
spec:
  secretName: internal-tls-secret
  issuerRef:
    name: ca-issuer
    kind: ClusterIssuer
  dnsNames:
  - api.smart-garden-bot.svc.cluster.local
  - postgresql.smart-garden-bot.svc.cluster.local
```

## Third-Party Security Integration

### Weather API Security
```go
// Weather API client with security measures
type WeatherClient struct {
    apiKey     string
    httpClient *http.Client
    rateLimiter *rate.Limiter
}

func NewWeatherClient(apiKey string) *WeatherClient {
    return &WeatherClient{
        apiKey: apiKey,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    MinVersion: tls.VersionTLS12,
                },
            },
        },
        rateLimiter: rate.NewLimiter(rate.Every(time.Second), 10), // 10 requests per second
    }
}

func (w *WeatherClient) GetForecast(location string) (*Forecast, error) {
    // Rate limiting
    if err := w.rateLimiter.Wait(context.Background()); err != nil {
        return nil, err
    }
    
    // Input validation
    if location == "" || len(location) > 100 {
        return nil, errors.New("invalid location parameter")
    }
    
    // Sanitize location for URL
    location = url.QueryEscape(strings.TrimSpace(location))
    
    url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s", 
                       location, w.apiKey)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    // Add security headers
    req.Header.Set("User-Agent", "SmartGardenBot/1.0")
    req.Header.Set("Accept", "application/json")
    
    resp, err := w.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // Validate response
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("weather API error: %d", resp.StatusCode)
    }
    
    var forecast Forecast
    if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
        return nil, err
    }
    
    return &forecast, nil
}
```

### IoT Device Security
```go
// Ecowitt device integration with security
type EcowittClient struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

func (e *EcowittClient) AuthenticateDevice(deviceID, secret string) error {
    // Validate device credentials
    hash := sha256.Sum256([]byte(deviceID + secret + e.apiKey))
    signature := hex.EncodeToString(hash[:])
    
    payload := map[string]string{
        "device_id": deviceID,
        "signature": signature,
        "timestamp": strconv.FormatInt(time.Now().Unix(), 10),
    }
    
    // Send authentication request
    jsonPayload, _ := json.Marshal(payload)
    resp, err := e.httpClient.Post(
        e.baseURL+"/auth", 
        "application/json",
        bytes.NewBuffer(jsonPayload),
    )
    
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("device authentication failed: %d", resp.StatusCode)
    }
    
    return nil
}

func (e *EcowittClient) ValidateWebhookSignature(payload []byte, signature string) bool {
    // HMAC signature validation
    mac := hmac.New(sha256.New, []byte(e.apiKey))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

### Stripe Integration Security
```go
// Stripe webhook signature verification
func VerifyStripeWebhook(payload []byte, signature string, secret string) error {
    return stripe.VerifyWebhookSignature(payload, signature, secret)
}

// PCI DSS compliant payment processing
func ProcessPayment(customerID string, amount int64) error {
    // Never store card details - use Stripe tokens/payment methods
    stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
    
    params := &stripe.PaymentIntentParams{
        Amount:   stripe.Int64(amount),
        Currency: stripe.String("usd"),
        Customer: stripe.String(customerID),
        Confirm:  stripe.Bool(true),
    }
    
    pi, err := paymentintent.New(params)
    if err != nil {
        // Log error without exposing sensitive details
        log.Printf("Payment processing failed for customer %s", customerID)
        return errors.New("payment processing failed")
    }
    
    // Audit log successful payment
    auditLog := AuditEvent{
        UserID:    customerID,
        Action:    "payment_processed",
        Amount:    amount,
        PaymentID: pi.ID,
        Timestamp: time.Now(),
    }
    
    if err := logAuditEvent(auditLog); err != nil {
        log.Printf("Failed to log audit event: %v", err)
    }
    
    return nil
}
```

## Security Monitoring and Logging

### Audit Logging
```go
// Audit event structure
type AuditEvent struct {
    ID        string                 `json:"id"`
    UserID    string                 `json:"user_id,omitempty"`
    TenantID  string                 `json:"tenant_id,omitempty"`
    Action    string                 `json:"action"`
    Resource  string                 `json:"resource,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    IPAddress string                 `json:"ip_address,omitempty"`
    UserAgent string                 `json:"user_agent,omitempty"`
    Success   bool                   `json:"success"`
    Error     string                 `json:"error,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}

// Audit logger middleware
func AuditLoggerMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        auditEvent := AuditEvent{
            ID:        generateID(),
            UserID:    param.Keys["user_id"].(string),
            Action:    fmt.Sprintf("%s %s", param.Method, param.Path),
            IPAddress: param.ClientIP,
            UserAgent: param.Request.UserAgent(),
            Success:   param.StatusCode < 400,
            Timestamp: param.TimeStamp,
        }
        
        if param.ErrorMessage != "" {
            auditEvent.Error = param.ErrorMessage
        }
        
        // Send to audit log system
        auditJSON, _ := json.Marshal(auditEvent)
        fmt.Println(string(auditJSON))
        
        return ""
    })
}
```

### Security Metrics
```go
// Prometheus security metrics
var (
    AuthenticationAttempts = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_attempts_total",
            Help: "Total number of authentication attempts",
        },
        []string{"method", "success"},
    )
    
    APIKeyUsage = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_key_usage_total",
            Help: "Total API key usage",
        },
        []string{"key_id", "endpoint"},
    )
    
    SecurityViolations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "security_violations_total",
            Help: "Total security violations detected",
        },
        []string{"type", "severity"},
    )
)

func init() {
    prometheus.MustRegister(AuthenticationAttempts)
    prometheus.MustRegister(APIKeyUsage)
    prometheus.MustRegister(SecurityViolations)
}
```

### Intrusion Detection
```go
// Rate limiting and suspicious activity detection
type SecurityMonitor struct {
    redis    *redis.Client
    alerts   chan SecurityAlert
}

type SecurityAlert struct {
    Type      string    `json:"type"`
    UserID    string    `json:"user_id"`
    IPAddress string    `json:"ip_address"`
    Details   string    `json:"details"`
    Severity  string    `json:"severity"`
    Timestamp time.Time `json:"timestamp"`
}

func (s *SecurityMonitor) CheckSuspiciousActivity(userID, ipAddress, action string) error {
    // Check rate limits
    key := fmt.Sprintf("rate_limit:%s:%s", userID, action)
    count, err := s.redis.Incr(context.Background(), key).Result()
    if err != nil {
        return err
    }
    
    if count == 1 {
        s.redis.Expire(context.Background(), key, time.Hour)
    }
    
    // Define thresholds
    thresholds := map[string]int64{
        "login_attempt":    10,
        "api_key_creation": 5,
        "password_reset":   3,
    }
    
    if threshold, exists := thresholds[action]; exists && count > threshold {
        alert := SecurityAlert{
            Type:      "rate_limit_exceeded",
            UserID:    userID,
            IPAddress: ipAddress,
            Details:   fmt.Sprintf("Action %s exceeded threshold: %d", action, count),
            Severity:  "HIGH",
            Timestamp: time.Now(),
        }
        
        s.alerts <- alert
        return errors.New("rate limit exceeded")
    }
    
    return nil
}
```

## Security Testing Strategy

### Automated Security Testing
```yaml
# GitHub Actions security testing
name: Security Scan
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        format: 'sarif'
        output: 'trivy-results.sarif'
    
    - name: Upload Trivy scan results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'
    
    - name: Run Semgrep security scan
      uses: returntocorp/semgrep-action@v1
      with:
        config: p/security-audit p/secrets
    
    - name: Run gosec security scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: './...'
```

### Penetration Testing
```bash
#!/bin/bash
# Automated penetration testing script

# OWASP ZAP automated security testing
docker run -t owasp/zap2docker-stable zap-baseline.py \
  -t https://api.smart-garden-bot.com \
  -r zap-report.html \
  -J zap-report.json

# Nuclei vulnerability scanning
nuclei -u https://api.smart-garden-bot.com \
  -t ssl,cors,xss,sqli \
  -o nuclei-results.txt

# API security testing with OWASP API Security Top 10
python3 api-security-test.py \
  --target https://api.smart-garden-bot.com \
  --auth-token $API_TOKEN \
  --output api-security-report.json
```

This comprehensive security architecture provides a robust foundation for the Smart Garden Bot platform, addressing all major security concerns while maintaining practical implementation guidelines. The next deliverable will focus on compliance requirements and checklists.