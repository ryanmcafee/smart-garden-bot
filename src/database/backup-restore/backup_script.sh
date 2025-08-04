#!/bin/bash

# =============================================
# SMART GARDEN BOT DATABASE BACKUP SCRIPT
# =============================================
# This script creates comprehensive backups of the Smart Garden Bot database
# with support for full backups, schema-only backups, and data-only backups.
# It also includes cleanup of old backups and compression.

set -euo pipefail

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-smart_garden_bot}"
DB_USER="${DB_USER:-postgres}"
PGPASSWORD="${PGPASSWORD:-}"

# Backup configuration
BACKUP_DIR="${BACKUP_DIR:-/opt/backups/smart-garden-bot}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
COMPRESS="${COMPRESS:-true}"
BACKUP_TYPE="${BACKUP_TYPE:-full}" # full, schema-only, data-only
VERBOSE="${VERBOSE:-false}"

# S3 configuration (optional)
S3_BUCKET="${S3_BUCKET:-}"
S3_PREFIX="${S3_PREFIX:-backups/smart-garden-bot}"
UPLOAD_TO_S3="${UPLOAD_TO_S3:-false}"

# Timestamp for backup files
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
DATE=$(date +"%Y-%m-%d")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        ERROR)
            echo -e "${RED}[ERROR] ${timestamp}: ${message}${NC}" >&2
            ;;
        WARN)
            echo -e "${YELLOW}[WARN] ${timestamp}: ${message}${NC}" >&2
            ;;
        INFO)
            echo -e "${GREEN}[INFO] ${timestamp}: ${message}${NC}"
            ;;
        DEBUG)
            if [[ "$VERBOSE" == "true" ]]; then
                echo -e "${BLUE}[DEBUG] ${timestamp}: ${message}${NC}"
            fi
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    log INFO "Checking prerequisites..."
    
    # Check if pg_dump is available
    if ! command -v pg_dump &> /dev/null; then
        log ERROR "pg_dump not found. Please install PostgreSQL client tools."
        exit 1
    fi
    
    # Check if gzip is available (for compression)
    if [[ "$COMPRESS" == "true" ]] && ! command -v gzip &> /dev/null; then
        log WARN "gzip not found. Compression disabled."
        COMPRESS="false"
    fi
    
    # Check if aws cli is available (for S3 upload)
    if [[ "$UPLOAD_TO_S3" == "true" ]] && ! command -v aws &> /dev/null; then
        log WARN "AWS CLI not found. S3 upload disabled."
        UPLOAD_TO_S3="false"
    fi
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    
    log INFO "Prerequisites check completed."
}

# Test database connection
test_connection() {
    log INFO "Testing database connection..."
    
    if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" &> /dev/null; then
        log ERROR "Cannot connect to database. Please check connection parameters."
        exit 1
    fi
    
    log INFO "Database connection successful."
}

# Get database statistics
get_db_stats() {
    log INFO "Gathering database statistics..."
    
    local stats_query="
    SELECT 
        schemaname,
        tablename,
        n_tup_ins as inserts,
        n_tup_upd as updates,
        n_tup_del as deletes,
        n_live_tup as live_tuples,
        n_dead_tup as dead_tuples,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as table_size
    FROM pg_stat_user_tables 
    WHERE schemaname IN ('auth', 'gardens', 'sensors', 'billing', 'analytics')
    ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
    "
    
    local stats_file="${BACKUP_DIR}/db_stats_${TIMESTAMP}.txt"
    
    {
        echo "# Smart Garden Bot Database Statistics"
        echo "# Generated: $(date)"
        echo "# Database: ${DB_NAME}"
        echo "# Host: ${DB_HOST}:${DB_PORT}"
        echo ""
        
        echo "## Table Statistics"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$stats_query"
        
        echo ""
        echo "## Database Size"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT 
            pg_database.datname,
            pg_size_pretty(pg_database_size(pg_database.datname)) AS size
        FROM pg_database
        WHERE datname = '${DB_NAME}';
        "
        
        echo ""
        echo "## Schema Sizes"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT 
            schemaname,
            pg_size_pretty(sum(pg_total_relation_size(schemaname||'.'||tablename))) as schema_size
        FROM pg_stat_user_tables 
        WHERE schemaname IN ('auth', 'gardens', 'sensors', 'billing', 'analytics')
        GROUP BY schemaname
        ORDER BY sum(pg_total_relation_size(schemaname||'.'||tablename)) DESC;
        "
        
    } > "$stats_file"
    
    log INFO "Database statistics saved to: $stats_file"
}

# Create full backup
create_full_backup() {
    local backup_file="${BACKUP_DIR}/smart_garden_bot_full_${TIMESTAMP}.sql"
    local final_file="$backup_file"
    
    log INFO "Creating full database backup..."
    log DEBUG "Backup file: $backup_file"
    
    # Create the backup
    pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --verbose \
        --format=plain \
        --no-password \
        --file="$backup_file"
    
    # Compress if requested
    if [[ "$COMPRESS" == "true" ]]; then
        log INFO "Compressing backup..."
        gzip "$backup_file"
        final_file="${backup_file}.gz"
    fi
    
    # Get file size
    local size=$(du -h "$final_file" | cut -f1)
    log INFO "Full backup completed. Size: $size"
    log INFO "Backup location: $final_file"
    
    echo "$final_file"
}

# Create schema-only backup
create_schema_backup() {
    local backup_file="${BACKUP_DIR}/smart_garden_bot_schema_${TIMESTAMP}.sql"
    local final_file="$backup_file"
    
    log INFO "Creating schema-only backup..."
    log DEBUG "Backup file: $backup_file"
    
    # Create the backup
    pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --verbose \
        --format=plain \
        --no-password \
        --schema-only \
        --file="$backup_file"
    
    # Compress if requested
    if [[ "$COMPRESS" == "true" ]]; then
        log INFO "Compressing backup..."
        gzip "$backup_file"
        final_file="${backup_file}.gz"
    fi
    
    local size=$(du -h "$final_file" | cut -f1)
    log INFO "Schema backup completed. Size: $size"
    log INFO "Backup location: $final_file"
    
    echo "$final_file"
}

# Create data-only backup
create_data_backup() {
    local backup_file="${BACKUP_DIR}/smart_garden_bot_data_${TIMESTAMP}.sql"
    local final_file="$backup_file"
    
    log INFO "Creating data-only backup..."
    log DEBUG "Backup file: $backup_file"
    
    # Create the backup
    pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --verbose \
        --format=plain \
        --no-password \
        --data-only \
        --file="$backup_file"
    
    # Compress if requested
    if [[ "$COMPRESS" == "true" ]]; then
        log INFO "Compressing backup..."
        gzip "$backup_file"
        final_file="${backup_file}.gz"
    fi
    
    local size=$(du -h "$final_file" | cut -f1)
    log INFO "Data backup completed. Size: $size"
    log INFO "Backup location: $final_file"
    
    echo "$final_file"
}

# Create custom format backup (for large databases)
create_custom_backup() {
    local backup_file="${BACKUP_DIR}/smart_garden_bot_custom_${TIMESTAMP}.dump"
    
    log INFO "Creating custom format backup..."
    log DEBUG "Backup file: $backup_file"
    
    # Create the backup
    pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --verbose \
        --format=custom \
        --compress=9 \
        --no-password \
        --file="$backup_file"
    
    local size=$(du -h "$backup_file" | cut -f1)
    log INFO "Custom format backup completed. Size: $size"
    log INFO "Backup location: $backup_file"
    
    echo "$backup_file"
}

# Create incremental backup (using pg_basebackup for WAL-based backups)
create_incremental_backup() {
    local backup_dir="${BACKUP_DIR}/incremental_${TIMESTAMP}"
    
    log INFO "Creating incremental backup..."
    log DEBUG "Backup directory: $backup_dir"
    
    mkdir -p "$backup_dir"
    
    # Create base backup
    pg_basebackup \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -D "$backup_dir" \
        --format=tar \
        --gzip \
        --progress \
        --verbose
    
    local size=$(du -sh "$backup_dir" | cut -f1)
    log INFO "Incremental backup completed. Size: $size"
    log INFO "Backup location: $backup_dir"
    
    echo "$backup_dir"
}

# Upload backup to S3
upload_to_s3() {
    local backup_file="$1"
    local s3_key="${S3_PREFIX}/$(date +%Y)/$(date +%m)/$(basename "$backup_file")"
    
    log INFO "Uploading backup to S3..."
    log DEBUG "S3 location: s3://${S3_BUCKET}/${s3_key}"
    
    if aws s3 cp "$backup_file" "s3://${S3_BUCKET}/${s3_key}" --no-progress; then
        log INFO "Upload to S3 completed successfully."
        log INFO "S3 location: s3://${S3_BUCKET}/${s3_key}"
    else
        log ERROR "Failed to upload backup to S3."
        return 1
    fi
}

# Clean up old backups
cleanup_old_backups() {
    log INFO "Cleaning up backups older than $RETENTION_DAYS days..."
    
    # Local cleanup
    local deleted_count=0
    while IFS= read -r -d '' file; do
        log DEBUG "Deleting old backup: $file"
        rm "$file"
        ((deleted_count++))
    done < <(find "$BACKUP_DIR" -name "smart_garden_bot_*" -type f -mtime +${RETENTION_DAYS} -print0)
    
    if [[ $deleted_count -gt 0 ]]; then
        log INFO "Deleted $deleted_count old backup files."
    else
        log INFO "No old backup files to delete."
    fi
    
    # S3 cleanup (if enabled)
    if [[ "$UPLOAD_TO_S3" == "true" && -n "$S3_BUCKET" ]]; then
        log INFO "Cleaning up old S3 backups..."
        
        local cutoff_date=$(date -d "${RETENTION_DAYS} days ago" +%Y-%m-%d)
        
        # List and delete old S3 objects
        aws s3 ls "s3://${S3_BUCKET}/${S3_PREFIX}/" --recursive | \
        awk '{print $1, $2, $4}' | \
        while read -r date time key; do
            if [[ "$date" < "$cutoff_date" ]]; then
                log DEBUG "Deleting old S3 backup: $key"
                aws s3 rm "s3://${S3_BUCKET}/${key}"
            fi
        done
    fi
}

# Verify backup integrity
verify_backup() {
    local backup_file="$1"
    
    log INFO "Verifying backup integrity..."
    
    if [[ "$backup_file" == *.gz ]]; then
        # Test gzip integrity
        if gzip -t "$backup_file"; then
            log INFO "Backup compression integrity verified."
        else
            log ERROR "Backup compression integrity check failed."
            return 1
        fi
        
        # Test SQL syntax by attempting to parse
        if gzip -dc "$backup_file" | head -n 100 | grep -q "PostgreSQL database dump"; then
            log INFO "Backup SQL header verified."
        else
            log ERROR "Backup SQL header verification failed."
            return 1
        fi
    elif [[ "$backup_file" == *.dump ]]; then
        # Verify custom format backup
        if pg_restore --list "$backup_file" > /dev/null 2>&1; then
            log INFO "Custom format backup integrity verified."
        else
            log ERROR "Custom format backup integrity check failed."
            return 1
        fi
    else
        # Test plain SQL file
        if head -n 100 "$backup_file" | grep -q "PostgreSQL database dump"; then
            log INFO "Backup SQL header verified."
        else
            log ERROR "Backup SQL header verification failed."
            return 1
        fi
    fi
    
    log INFO "Backup verification completed successfully."
}

# Send notification (optional)
send_notification() {
    local status="$1"
    local backup_file="$2"
    local size="$3"
    
    # Webhook notification (if configured)
    if [[ -n "${WEBHOOK_URL:-}" ]]; then
        local payload="{
            \"text\": \"Smart Garden Bot Backup $status\",
            \"attachments\": [{
                \"color\": \"$([ "$status" = "SUCCESS" ] && echo "good" || echo "danger")\",
                \"fields\": [
                    {\"title\": \"Database\", \"value\": \"$DB_NAME\", \"short\": true},
                    {\"title\": \"Host\", \"value\": \"$DB_HOST:$DB_PORT\", \"short\": true},
                    {\"title\": \"Backup Type\", \"value\": \"$BACKUP_TYPE\", \"short\": true},
                    {\"title\": \"File Size\", \"value\": \"$size\", \"short\": true},
                    {\"title\": \"Timestamp\", \"value\": \"$(date)\", \"short\": false}
                ]
            }]
        }"
        
        curl -X POST -H 'Content-type: application/json' \
             --data "$payload" \
             "$WEBHOOK_URL" > /dev/null 2>&1
    fi
    
    # Email notification (if configured)
    if [[ -n "${EMAIL_TO:-}" ]] && command -v mail &> /dev/null; then
        local subject="Smart Garden Bot Backup $status - $DATE"
        local body="
Backup Status: $status
Database: $DB_NAME
Host: $DB_HOST:$DB_PORT
Backup Type: $BACKUP_TYPE
File: $backup_file
Size: $size
Timestamp: $(date)
        "
        
        echo "$body" | mail -s "$subject" "$EMAIL_TO"
    fi
}

# Main backup function
perform_backup() {
    local backup_file=""
    local size=""
    
    case "$BACKUP_TYPE" in
        "full")
            backup_file=$(create_full_backup)
            ;;
        "schema-only")
            backup_file=$(create_schema_backup)
            ;;
        "data-only")
            backup_file=$(create_data_backup)
            ;;
        "custom")
            backup_file=$(create_custom_backup)
            ;;
        "incremental")
            backup_file=$(create_incremental_backup)
            ;;
        *)
            log ERROR "Invalid backup type: $BACKUP_TYPE"
            exit 1
            ;;
    esac
    
    # Verify backup
    if ! verify_backup "$backup_file"; then
        log ERROR "Backup verification failed."
        send_notification "FAILED" "$backup_file" "N/A"
        exit 1
    fi
    
    # Get final size
    size=$(du -h "$backup_file" | cut -f1)
    
    # Upload to S3 if configured
    if [[ "$UPLOAD_TO_S3" == "true" ]]; then
        upload_to_s3 "$backup_file"
    fi
    
    # Send success notification
    send_notification "SUCCESS" "$backup_file" "$size"
    
    log INFO "Backup process completed successfully!"
    log INFO "Backup file: $backup_file"
    log INFO "Size: $size"
}

# Print usage
usage() {
    cat << EOF
Smart Garden Bot Database Backup Script

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -t, --type TYPE         Backup type: full, schema-only, data-only, custom, incremental (default: full)
    -d, --dir DIR           Backup directory (default: /opt/backups/smart-garden-bot)
    -r, --retention DAYS    Retention period in days (default: 30)
    -c, --compress          Compress backups (default: true)
    -s, --s3                Upload to S3 (default: false)
    -v, --verbose           Verbose output (default: false)
    --no-stats              Skip database statistics gathering
    --no-cleanup            Skip cleanup of old backups

Environment Variables:
    DB_HOST                 Database host (default: localhost)
    DB_PORT                 Database port (default: 5432)
    DB_NAME                 Database name (default: smart_garden_bot)
    DB_USER                 Database user (default: postgres)
    PGPASSWORD              Database password
    S3_BUCKET               S3 bucket for uploads
    S3_PREFIX               S3 prefix for backups (default: backups/smart-garden-bot)
    WEBHOOK_URL             Webhook URL for notifications
    EMAIL_TO                Email address for notifications

src:
    # Full backup with compression
    $0 --type full --compress

    # Schema-only backup
    $0 --type schema-only --dir /tmp/backups

    # Custom format backup with S3 upload
    $0 --type custom --s3

    # Incremental backup with verbose output
    $0 --type incremental --verbose
EOF
}

# Parse command line arguments
GATHER_STATS=true
CLEANUP_OLD=true

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -t|--type)
            BACKUP_TYPE="$2"
            shift 2
            ;;
        -d|--dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        -r|--retention)
            RETENTION_DAYS="$2"
            shift 2
            ;;
        -c|--compress)
            COMPRESS="true"
            shift
            ;;
        -s|--s3)
            UPLOAD_TO_S3="true"
            shift
            ;;
        -v|--verbose)
            VERBOSE="true"
            shift
            ;;
        --no-stats)
            GATHER_STATS=false
            shift
            ;;
        --no-cleanup)
            CLEANUP_OLD=false
            shift
            ;;
        *)
            log ERROR "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    log INFO "Starting Smart Garden Bot database backup..."
    log INFO "Backup type: $BACKUP_TYPE"
    log INFO "Target database: $DB_NAME@$DB_HOST:$DB_PORT"
    log INFO "Backup directory: $BACKUP_DIR"
    
    # Run backup process
    check_prerequisites
    test_connection
    
    if [[ "$GATHER_STATS" == "true" ]]; then
        get_db_stats
    fi
    
    perform_backup
    
    if [[ "$CLEANUP_OLD" == "true" ]]; then
        cleanup_old_backups
    fi
    
    log INFO "Backup process completed!"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi