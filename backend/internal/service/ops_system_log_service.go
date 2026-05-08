package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

const (
	opsAccountSwitchComponent = "ops.account_switch"
	opsAccountSwitchMessage   = "account_switch"
	opsAccountSelectedMessage = "account_selected"
	opsAccountSwitchMaxLimit  = 30
)

func (s *OpsService) ListSystemLogs(ctx context.Context, filter *OpsSystemLogFilter) (*OpsSystemLogList, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	if s.opsRepo == nil {
		return &OpsSystemLogList{
			Logs:     []*OpsSystemLog{},
			Total:    0,
			Page:     1,
			PageSize: 50,
		}, nil
	}
	if filter == nil {
		filter = &OpsSystemLogFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 50
	}
	if filter.PageSize > 200 {
		filter.PageSize = 200
	}

	result, err := s.opsRepo.ListSystemLogs(ctx, filter)
	if err != nil {
		return nil, infraerrors.InternalServer("OPS_SYSTEM_LOG_LIST_FAILED", "Failed to list system logs").WithCause(err)
	}
	return result, nil
}

func (s *OpsService) CleanupSystemLogs(ctx context.Context, filter *OpsSystemLogCleanupFilter, operatorID int64) (int64, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return 0, err
	}
	if s.opsRepo == nil {
		return 0, infraerrors.ServiceUnavailable("OPS_REPO_UNAVAILABLE", "Ops repository not available")
	}
	if operatorID <= 0 {
		return 0, infraerrors.BadRequest("OPS_SYSTEM_LOG_CLEANUP_INVALID_OPERATOR", "invalid operator")
	}
	if filter == nil {
		filter = &OpsSystemLogCleanupFilter{}
	}
	if filter.EndTime != nil && filter.StartTime != nil && filter.StartTime.After(*filter.EndTime) {
		return 0, infraerrors.BadRequest("OPS_SYSTEM_LOG_CLEANUP_INVALID_RANGE", "invalid time range")
	}

	deletedRows, err := s.opsRepo.DeleteSystemLogs(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "requires at least one filter") {
			return 0, infraerrors.BadRequest("OPS_SYSTEM_LOG_CLEANUP_FILTER_REQUIRED", "cleanup requires at least one filter condition")
		}
		return 0, infraerrors.InternalServer("OPS_SYSTEM_LOG_CLEANUP_FAILED", "Failed to cleanup system logs").WithCause(err)
	}

	if auditErr := s.opsRepo.InsertSystemLogCleanupAudit(ctx, &OpsSystemLogCleanupAudit{
		CreatedAt:   time.Now().UTC(),
		OperatorID:  operatorID,
		Conditions:  marshalSystemLogCleanupConditions(filter),
		DeletedRows: deletedRows,
	}); auditErr != nil {
		// 审计失败不影响主流程，避免运维清理被阻塞。
		log.Printf("[OpsSystemLog] cleanup audit failed: %v", auditErr)
	}
	return deletedRows, nil
}

func marshalSystemLogCleanupConditions(filter *OpsSystemLogCleanupFilter) string {
	if filter == nil {
		return "{}"
	}
	payload := map[string]any{
		"level":             strings.TrimSpace(filter.Level),
		"component":         strings.TrimSpace(filter.Component),
		"request_id":        strings.TrimSpace(filter.RequestID),
		"client_request_id": strings.TrimSpace(filter.ClientRequestID),
		"platform":          strings.TrimSpace(filter.Platform),
		"model":             strings.TrimSpace(filter.Model),
		"query":             strings.TrimSpace(filter.Query),
	}
	if filter.UserID != nil {
		payload["user_id"] = *filter.UserID
	}
	if filter.AccountID != nil {
		payload["account_id"] = *filter.AccountID
	}
	if filter.StartTime != nil && !filter.StartTime.IsZero() {
		payload["start_time"] = filter.StartTime.UTC().Format(time.RFC3339Nano)
	}
	if filter.EndTime != nil && !filter.EndTime.IsZero() {
		payload["end_time"] = filter.EndTime.UTC().Format(time.RFC3339Nano)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func (s *OpsService) GetSystemLogSinkHealth() OpsSystemLogSinkHealth {
	if s == nil || s.systemLogSink == nil {
		return OpsSystemLogSinkHealth{}
	}
	return s.systemLogSink.Health()
}

func (s *OpsService) RecordAccountSelection(
	ctx context.Context,
	requestID string,
	clientRequestID string,
	userID *int64,
	group *Group,
	accountID int64,
	accountName string,
	platform string,
) {
	if s == nil || s.systemLogSink == nil || accountID <= 0 || !s.IsMonitoringEnabled(ctx) {
		return
	}

	platform = strings.TrimSpace(platform)
	if platform == "" && group != nil {
		platform = strings.TrimSpace(group.Platform)
	}

	fields := map[string]any{
		"request_id":       strings.TrimSpace(requestID),
		"client_request_id": strings.TrimSpace(clientRequestID),
		"platform":         platform,
		"account_id":       accountID,
		"account_name":     strings.TrimSpace(accountName),
		"event_type":       "account_selected",
	}
	if userID != nil && *userID > 0 {
		fields["user_id"] = *userID
	}
	if group != nil && group.ID > 0 {
		fields["group_id"] = group.ID
		fields["group_name"] = strings.TrimSpace(group.Name)
	}

	logger.WriteSinkEvent("warn", opsAccountSwitchComponent, opsAccountSelectedMessage, fields)
}

func (s *OpsService) GetAccountSwitchSummary(
	ctx context.Context,
	filter *OpsDashboardFilter,
	limit int,
) (*OpsAccountSwitchSummary, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	if s.opsRepo == nil {
		return &OpsAccountSwitchSummary{RecentSwitches: []*OpsAccountSwitchRecord{}}, nil
	}
	if filter == nil {
		return nil, infraerrors.BadRequest("OPS_FILTER_REQUIRED", "filter is required")
	}

	if limit <= 0 || limit > opsAccountSwitchMaxLimit {
		limit = opsAccountSwitchMaxLimit
	}

	result, err := s.opsRepo.GetAccountSwitchSummary(ctx, filter, limit)
	if err != nil {
		return nil, infraerrors.InternalServer("OPS_ACCOUNT_SWITCH_SUMMARY_FAILED", "Failed to load account switch summary").WithCause(err)
	}
	if result == nil {
		return &OpsAccountSwitchSummary{RecentSwitches: []*OpsAccountSwitchRecord{}}, nil
	}
	if result.RecentSwitches == nil {
		result.RecentSwitches = []*OpsAccountSwitchRecord{}
	}
	return result, nil
}
