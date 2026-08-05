package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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

	var platformID int64
	var createdAt, updatedAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`INSERT INTO platforms (code, name, account_platform, status, legacy_group_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		platform.Code, platform.Name, platform.AccountPlatform, platform.Status, platform.LegacyGroupID,
	).Scan(&platformID, &createdAt, &updatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return service.ErrPlatformExists
		}
		return fmt.Errorf("insert platform: %w", err)
	}

	createdRules := make([]service.PlatformModelRule, len(platform.ModelRules))
	for index := range platform.ModelRules {
		rule, err := insertPlatformModelRule(ctx, tx, platformID, platform.Code, platform.LegacyGroupID, platform.ModelRules[index])
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

	updated, err := tx.ExecContext(ctx,
		`UPDATE platforms
		 SET code = $1, name = $2, account_platform = $3, status = $4, legacy_group_id = $5, updated_at = NOW()
		 WHERE id = $6`,
		platform.Code, platform.Name, platform.AccountPlatform, platform.Status, platform.LegacyGroupID, platform.ID,
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
		rule, err := insertPlatformModelRule(ctx, tx, platform.ID, platform.Code, platform.LegacyGroupID, platform.ModelRules[index])
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
	var legacyGroupID sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT id, code, name, account_platform, status, legacy_group_id, created_at, updated_at
		 FROM platforms WHERE id = $1`, id,
	).Scan(
		&platform.ID,
		&platform.Code,
		&platform.Name,
		&platform.AccountPlatform,
		&platform.Status,
		&legacyGroupID,
		&platform.CreatedAt,
		&platform.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, service.ErrPlatformNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get platform: %w", err)
	}
	if legacyGroupID.Valid {
		platform.LegacyGroupID = &legacyGroupID.Int64
	}
	rules, err := r.listPlatformRules(ctx, platform.ID, platform.Code, platform.AccountPlatform, platform.LegacyGroupID)
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
		`SELECT id, code, name, account_platform, status, legacy_group_id, created_at, updated_at
		 FROM platforms ORDER BY code ASC, id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list platforms: %w", err)
	}
	defer func() { _ = rows.Close() }()

	platforms := make([]service.Platform, 0)
	for rows.Next() {
		var platform service.Platform
		var legacyGroupID sql.NullInt64
		if err := rows.Scan(
			&platform.ID,
			&platform.Code,
			&platform.Name,
			&platform.AccountPlatform,
			&platform.Status,
			&legacyGroupID,
			&platform.CreatedAt,
			&platform.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan platform: %w", err)
		}
		if legacyGroupID.Valid {
			platform.LegacyGroupID = &legacyGroupID.Int64
		}
		rules, err := r.listPlatformRules(ctx, platform.ID, platform.Code, platform.AccountPlatform, platform.LegacyGroupID)
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
	legacyGroupID *int64,
	rule service.PlatformModelRule,
) (service.PlatformModelRule, error) {
	capabilities, err := marshalEndpointCapabilities(rule.EndpointCapabilities)
	if err != nil {
		return service.PlatformModelRule{}, err
	}
	status := service.StatusDisabled
	if rule.Enabled {
		status = service.StatusActive
	}
	var createdAt, updatedAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`INSERT INTO platform_model_rules
		 (platform_id, model_pattern, upstream_model, endpoint_capabilities, status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		platformID, rule.ModelPattern, rule.UpstreamModel, capabilities, status,
	).Scan(&rule.ID, &createdAt, &updatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return service.PlatformModelRule{}, service.ErrPlatformModelRule
		}
		return service.PlatformModelRule{}, fmt.Errorf("insert platform model rule: %w", err)
	}
	rule.PlatformID = platformID
	rule.PlatformCode = platformCode
	rule.LegacyGroupID = cloneInt64Pointer(legacyGroupID)
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
		`SELECT r.id, r.platform_id, p.code, p.account_platform, p.legacy_group_id, r.model_pattern, r.upstream_model,
		        r.endpoint_capabilities, r.created_at, r.updated_at
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
		var legacyGroupID sql.NullInt64
		var capabilities []byte
		if err := rows.Scan(
			&rule.ID,
			&rule.PlatformID,
			&rule.PlatformCode,
			&rule.AccountPlatform,
			&legacyGroupID,
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
		if legacyGroupID.Valid {
			rule.LegacyGroupID = &legacyGroupID.Int64
		}
		rule.Enabled = true
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active platform model rules: %w", err)
	}
	return rules, nil
}

func (r *platformRepository) listPlatformRules(ctx context.Context, platformID int64, platformCode, accountPlatform string, legacyGroupID *int64) ([]service.PlatformModelRule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, platform_id, model_pattern, upstream_model, endpoint_capabilities, status, created_at, updated_at
		 FROM platform_model_rules WHERE platform_id = $1 ORDER BY id ASC`, platformID,
	)
	if err != nil {
		return nil, fmt.Errorf("query platform model rules: %w", err)
	}
	defer func() { _ = rows.Close() }()

	rules := make([]service.PlatformModelRule, 0)
	for rows.Next() {
		var rule service.PlatformModelRule
		var capabilities []byte
		var status string
		if err := rows.Scan(
			&rule.ID,
			&rule.PlatformID,
			&rule.ModelPattern,
			&rule.UpstreamModel,
			&capabilities,
			&status,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan platform model rule: %w", err)
		}
		endpointCapabilities, err := decodeEndpointCapabilities(capabilities)
		if err != nil {
			return nil, err
		}
		rule.PlatformCode = platformCode
		rule.AccountPlatform = accountPlatform
		rule.LegacyGroupID = cloneInt64Pointer(legacyGroupID)
		rule.EndpointCapabilities = endpointCapabilities
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
		return nil, fmt.Errorf("decode platform model rule capabilities: %w", err)
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
