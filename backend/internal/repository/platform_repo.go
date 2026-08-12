package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type platformRepository struct {
	db *sql.DB
}

func NewPlatformRepository(db *sql.DB) service.PlatformRepository {
	return newPlatformRepository(db)
}

func newPlatformRepository(db *sql.DB) *platformRepository {
	return &platformRepository{db: db}
}

func (r *platformRepository) HasAccountsByPlatformID(ctx context.Context, platformID int64) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM accounts WHERE platform_id = $1)`, platformID,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check platform accounts: %w", err)
	}
	return exists, nil
}

type platformReferenceCounts struct {
	accounts, apiKeys, usageLogs                              int64
	promptAuditJobs, promptAuditEvents, contentModerationLogs int64
	schedulerOutbox, opsErrorLogs, opsSystemMetrics           int64
	opsMetricsHourly, opsMetricsDaily, opsAlertSilences       int64
	opsAlertRules, opsAlertEvents                             int64
	contentModerationConfig, promptAuditConfig                int64
}

func (c platformReferenceCounts) total() int64 {
	return c.accounts + c.apiKeys + c.usageLogs + c.promptAuditJobs + c.promptAuditEvents +
		c.contentModerationLogs + c.schedulerOutbox + c.opsErrorLogs + c.opsSystemMetrics +
		c.opsMetricsHourly + c.opsMetricsDaily + c.opsAlertSilences + c.opsAlertRules +
		c.opsAlertEvents + c.contentModerationConfig + c.promptAuditConfig
}

func (c platformReferenceCounts) metadata() map[string]string {
	return map[string]string{
		"accounts":   strconv.FormatInt(c.accounts, 10),
		"api_keys":   strconv.FormatInt(c.apiKeys, 10),
		"usage_logs": strconv.FormatInt(c.usageLogs, 10),
		"audits":     strconv.FormatInt(c.promptAuditJobs+c.promptAuditEvents+c.contentModerationLogs, 10),
		"ops":        strconv.FormatInt(c.schedulerOutbox+c.opsErrorLogs+c.opsSystemMetrics+c.opsMetricsHourly+c.opsMetricsDaily+c.opsAlertSilences+c.opsAlertRules+c.opsAlertEvents, 10),
		"configs":    strconv.FormatInt(c.contentModerationConfig+c.promptAuditConfig, 10),
	}
}

// DeleteUnused serializes deletion with platform-owned writes and rejects any
// platform that is still referenced by current or historical data.
func (r *platformRepository) DeleteUnused(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin platform delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var lockedID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM platforms WHERE id = $1 FOR UPDATE`, id).Scan(&lockedID); err != nil {
		if err == sql.ErrNoRows {
			return service.ErrPlatformNotFound
		}
		return fmt.Errorf("lock platform: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `LOCK TABLE scheduler_outbox, ops_error_logs, ops_system_metrics,
		ops_metrics_hourly, ops_metrics_daily, ops_alert_silences, ops_alert_rules,
		ops_alert_events, settings IN SHARE MODE`); err != nil {
		return fmt.Errorf("lock platform reference tables: %w", err)
	}

	var counts platformReferenceCounts
	err = tx.QueryRowContext(ctx, platformReferenceCountSQL, id).Scan(
		&counts.accounts, &counts.apiKeys, &counts.usageLogs,
		&counts.promptAuditJobs, &counts.promptAuditEvents, &counts.contentModerationLogs,
		&counts.schedulerOutbox, &counts.opsErrorLogs, &counts.opsSystemMetrics,
		&counts.opsMetricsHourly, &counts.opsMetricsDaily, &counts.opsAlertSilences,
		&counts.opsAlertRules, &counts.opsAlertEvents,
		&counts.contentModerationConfig, &counts.promptAuditConfig,
	)
	if err != nil {
		return fmt.Errorf("count platform references: %w", err)
	}
	if counts.total() > 0 {
		return service.ErrPlatformInUse.WithMetadata(counts.metadata())
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM platforms WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete platform: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read platform delete result: %w", err)
	}
	if rows == 0 {
		return service.ErrPlatformNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit platform delete: %w", err)
	}
	return nil
}

const platformReferenceCountSQL = `SELECT
	(SELECT COUNT(*) FROM accounts WHERE platform_id = $1) AS accounts,
	(SELECT COUNT(*) FROM api_key_platforms WHERE platform_id = $1) AS api_keys,
	(SELECT COUNT(*) FROM usage_logs WHERE platform_id = $1) AS usage_logs,
	(SELECT COUNT(*) FROM prompt_audit_jobs WHERE platform_id = $1) AS prompt_audit_jobs,
	(SELECT COUNT(*) FROM prompt_audit_events WHERE platform_id = $1) AS prompt_audit_events,
	(SELECT COUNT(*) FROM content_moderation_logs WHERE platform_id = $1) AS content_moderation_logs,
	(SELECT COUNT(*) FROM scheduler_outbox WHERE platform_id = $1) AS scheduler_outbox,
	(SELECT COUNT(*) FROM ops_error_logs WHERE platform_id = $1) AS ops_error_logs,
	(SELECT COUNT(*) FROM ops_system_metrics WHERE platform_id = $1) AS ops_system_metrics,
	(SELECT COUNT(*) FROM ops_metrics_hourly WHERE platform_id = $1) AS ops_metrics_hourly,
	(SELECT COUNT(*) FROM ops_metrics_daily WHERE platform_id = $1) AS ops_metrics_daily,
	(SELECT COUNT(*) FROM ops_alert_silences WHERE platform_id = $1) AS ops_alert_silences,
	(SELECT COUNT(*) FROM ops_alert_rules WHERE COALESCE(filters, '{}'::jsonb) @> jsonb_build_object('platform_id', $1)) AS ops_alert_rules,
	(SELECT COUNT(*) FROM ops_alert_events WHERE COALESCE(dimensions, '{}'::jsonb) @> jsonb_build_object('platform_id', $1)) AS ops_alert_events,
	(SELECT COUNT(*) FROM settings WHERE key = 'content_moderation_config'
		AND COALESCE((value::jsonb)->>'all_platforms', 'true') = 'false'
		AND COALESCE(value::jsonb->'platform_ids', '[]'::jsonb) @> jsonb_build_array($1)) AS content_moderation_config,
	(SELECT COUNT(*) FROM settings WHERE key = 'prompt_audit_config'
		AND COALESCE((value::jsonb)->>'all_platforms', 'true') = 'false'
		AND COALESCE(value::jsonb->'platform_ids', '[]'::jsonb) @> jsonb_build_array($1)) AS prompt_audit_config`

// Create persists the platform and all of its model rules in one transaction.
// A partial account-pool configuration must never become schedulable.
func (r *platformRepository) Create(ctx context.Context, platform *service.Platform) error {
	if platform == nil {
		return fmt.Errorf("platform is nil")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin platform transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	capabilities, err := marshalEndpointCapabilities(platform.EndpointCapabilities)
	if err != nil {
		return err
	}
	var platformID int64
	var createdAt, updatedAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`INSERT INTO platforms (code, name, account_platform, status, endpoint_capabilities)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		platform.Code, platform.Name, platform.AccountPlatform, platform.Status, capabilities,
	).Scan(&platformID, &createdAt, &updatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrPlatformExists
		}
		return fmt.Errorf("insert platform: %w", err)
	}

	createdRules := make([]service.PlatformModelRule, len(platform.ModelRules))
	for index := range platform.ModelRules {
		rule, err := insertPlatformModelRule(ctx, tx, platformID, platform.Code, platform.ModelRules[index])
		if err != nil {
			return err
		}
		createdRules[index] = rule
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit platform transaction: %w", err)
	}

	platform.ID = platformID
	if createdAt.Valid {
		platform.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		platform.UpdatedAt = updatedAt.Time
	}
	platform.ModelRules = createdRules
	return nil
}

// Update replaces the complete rule set in the same transaction as the
// platform fields. Leaving stale rules behind could route a model to the wrong
// account pool after an administrator edits the mapping.
func (r *platformRepository) Update(ctx context.Context, platform *service.Platform) error {
	if platform == nil || platform.ID <= 0 {
		return fmt.Errorf("platform id is required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin platform transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	capabilities, err := marshalEndpointCapabilities(platform.EndpointCapabilities)
	if err != nil {
		return err
	}
	updated, err := tx.ExecContext(ctx,
		`UPDATE platforms
		 SET code = $1, name = $2, account_platform = $3, status = $4, endpoint_capabilities = $5, updated_at = NOW()
		 WHERE id = $6`,
		platform.Code, platform.Name, platform.AccountPlatform, platform.Status, capabilities, platform.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrPlatformExists
		}
		return fmt.Errorf("update platform: %w", err)
	}
	rows, err := updated.RowsAffected()
	if err != nil {
		return fmt.Errorf("read platform update result: %w", err)
	}
	if rows == 0 {
		return service.ErrPlatformNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM platform_model_rules WHERE platform_id = $1`, platform.ID); err != nil {
		return fmt.Errorf("delete platform model rules: %w", err)
	}

	updatedRules := make([]service.PlatformModelRule, len(platform.ModelRules))
	for index := range platform.ModelRules {
		rule, err := insertPlatformModelRule(ctx, tx, platform.ID, platform.Code, platform.ModelRules[index])
		if err != nil {
			return err
		}
		updatedRules[index] = rule
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit platform transaction: %w", err)
	}
	platform.ModelRules = updatedRules
	return nil
}

func (r *platformRepository) GetByID(ctx context.Context, id int64) (*service.Platform, error) {
	platform := &service.Platform{}
	var capabilities []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT id, code, name, account_platform, status, endpoint_capabilities, created_at, updated_at
		 FROM platforms WHERE id = $1`, id,
	).Scan(
		&platform.ID,
		&platform.Code,
		&platform.Name,
		&platform.AccountPlatform,
		&platform.Status,
		&capabilities,
		&platform.CreatedAt,
		&platform.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, service.ErrPlatformNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get platform: %w", err)
	}
	platform.EndpointCapabilities, err = decodeEndpointCapabilities(capabilities)
	if err != nil {
		return nil, err
	}
	rules, err := r.listPlatformRules(ctx, platform.ID, platform.Code, platform.AccountPlatform, platform.EndpointCapabilities)
	if err != nil {
		return nil, err
	}
	platform.ModelRules = rules
	return platform, nil
}

// List returns all platform configurations, including disabled pools and
// rules, for administration. Runtime model resolution intentionally uses the
// narrower ListModelRules query instead.
func (r *platformRepository) List(ctx context.Context) ([]service.Platform, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, code, name, account_platform, status, endpoint_capabilities, created_at, updated_at
		 FROM platforms ORDER BY code ASC, id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list platforms: %w", err)
	}
	defer func() { _ = rows.Close() }()

	platforms := make([]service.Platform, 0)
	for rows.Next() {
		var platform service.Platform
		var capabilities []byte
		if err := rows.Scan(
			&platform.ID,
			&platform.Code,
			&platform.Name,
			&platform.AccountPlatform,
			&platform.Status,
			&capabilities,
			&platform.CreatedAt,
			&platform.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan platform: %w", err)
		}
		platform.EndpointCapabilities, err = decodeEndpointCapabilities(capabilities)
		if err != nil {
			return nil, err
		}
		rules, err := r.listPlatformRules(ctx, platform.ID, platform.Code, platform.AccountPlatform, platform.EndpointCapabilities)
		if err != nil {
			return nil, err
		}
		platform.ModelRules = rules
		platforms = append(platforms, platform)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platforms: %w", err)
	}
	return platforms, nil
}

func insertPlatformModelRule(
	ctx context.Context,
	tx *sql.Tx,
	platformID int64,
	platformCode string,
	rule service.PlatformModelRule,
) (service.PlatformModelRule, error) {
	status := service.StatusDisabled
	if rule.Enabled {
		status = service.StatusActive
	}
	var createdAt, updatedAt sql.NullTime
	err := tx.QueryRowContext(ctx,
		`INSERT INTO platform_model_rules
		 (platform_id, model_pattern, upstream_model, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		platformID, rule.ModelPattern, rule.UpstreamModel, status,
	).Scan(&rule.ID, &createdAt, &updatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return service.PlatformModelRule{}, service.ErrPlatformModelRule
		}
		return service.PlatformModelRule{}, fmt.Errorf("insert platform model rule: %w", err)
	}
	rule.PlatformID = platformID
	rule.PlatformCode = platformCode
	if createdAt.Valid {
		rule.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		rule.UpdatedAt = updatedAt.Time
	}
	return rule, nil
}

// ListModelRules deliberately joins the platform status. A disabled account
// pool cannot claim a model name or be selected by the request path.
func (r *platformRepository) ListModelRules(ctx context.Context) ([]service.PlatformModelRule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT r.id, r.platform_id, p.code, p.account_platform, r.model_pattern, r.upstream_model,
		        p.endpoint_capabilities, r.created_at, r.updated_at
		 FROM platform_model_rules r
		 JOIN platforms p ON p.id = r.platform_id
		 WHERE p.status = $1 AND r.status = $2
		 ORDER BY r.platform_id ASC, r.id ASC`,
		service.StatusActive, service.StatusActive,
	)
	if err != nil {
		return nil, fmt.Errorf("query active platform model rules: %w", err)
	}
	defer func() { _ = rows.Close() }()

	rules := make([]service.PlatformModelRule, 0)
	for rows.Next() {
		var rule service.PlatformModelRule
		var capabilities []byte
		if err := rows.Scan(
			&rule.ID,
			&rule.PlatformID,
			&rule.PlatformCode,
			&rule.AccountPlatform,
			&rule.ModelPattern,
			&rule.UpstreamModel,
			&capabilities,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan active platform model rule: %w", err)
		}
		endpointCapabilities, err := decodeEndpointCapabilities(capabilities)
		if err != nil {
			return nil, err
		}
		rule.EndpointCapabilities = endpointCapabilities
		rule.Enabled = true
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active platform model rules: %w", err)
	}
	return rules, nil
}

func (r *platformRepository) listPlatformRules(ctx context.Context, platformID int64, platformCode, accountPlatform string, platformCapabilities []string) ([]service.PlatformModelRule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, platform_id, model_pattern, upstream_model, status, created_at, updated_at
		 FROM platform_model_rules WHERE platform_id = $1 ORDER BY id ASC`, platformID,
	)
	if err != nil {
		return nil, fmt.Errorf("query platform model rules: %w", err)
	}
	defer func() { _ = rows.Close() }()

	rules := make([]service.PlatformModelRule, 0)
	for rows.Next() {
		var rule service.PlatformModelRule
		var status string
		if err := rows.Scan(
			&rule.ID,
			&rule.PlatformID,
			&rule.ModelPattern,
			&rule.UpstreamModel,
			&status,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan platform model rule: %w", err)
		}
		rule.PlatformCode = platformCode
		rule.AccountPlatform = accountPlatform
		rule.EndpointCapabilities = append([]string(nil), platformCapabilities...)
		rule.Enabled = status == service.StatusActive
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platform model rules: %w", err)
	}
	return rules, nil
}

func marshalEndpointCapabilities(capabilities []string) ([]byte, error) {
	if len(capabilities) == 0 {
		return []byte("[]"), nil
	}
	encoded, err := json.Marshal(capabilities)
	if err != nil {
		return nil, fmt.Errorf("encode endpoint capabilities: %w", err)
	}
	return encoded, nil
}

func decodeEndpointCapabilities(encoded []byte) ([]string, error) {
	if len(encoded) == 0 {
		return []string{}, nil
	}
	var capabilities []string
	if err := json.Unmarshal(encoded, &capabilities); err != nil {
		return nil, fmt.Errorf("decode platform endpoint capabilities: %w", err)
	}
	return capabilities, nil
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
