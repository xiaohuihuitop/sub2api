//go:build unit

package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func newPlatformDeleteRepo(t *testing.T) (*platformRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return newPlatformRepository(db), mock
}

func expectPlatformDeletePrelude(mock sqlmock.Sqlmock, id int64) {
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM platforms WHERE id = $1 FOR UPDATE")).
		WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectExec("LOCK TABLE scheduler_outbox.*IN SHARE MODE").WillReturnResult(sqlmock.NewResult(0, 0))
}

func platformReferenceRows(values ...int64) *sqlmock.Rows {
	columns := []string{"accounts", "api_keys", "usage_logs", "prompt_audit_jobs", "prompt_audit_events", "content_moderation_logs", "scheduler_outbox", "ops_error_logs", "ops_system_metrics", "ops_metrics_hourly", "ops_metrics_daily", "ops_alert_silences", "ops_alert_rules", "ops_alert_events", "content_moderation_config", "prompt_audit_config"}
	row := make([]driver.Value, len(values))
	for i := range values {
		row[i] = values[i]
	}
	return sqlmock.NewRows(columns).AddRow(row...)
}

func TestPlatformRepositoryDeleteUnusedDeletesOnlyUnreferencedPlatform(t *testing.T) {
	repo, mock := newPlatformDeleteRepo(t)
	expectPlatformDeletePrelude(mock, 7)
	mock.ExpectQuery("SELECT.*accounts.*content_moderation_config.*prompt_audit_config").WithArgs(int64(7)).
		WillReturnRows(platformReferenceRows(make([]int64, 16)...))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM platforms WHERE id = $1")).WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.DeleteUnused(context.Background(), 7))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPlatformRepositoryDeleteUnusedReportsEveryReferenceClass(t *testing.T) {
	repo, mock := newPlatformDeleteRepo(t)
	expectPlatformDeletePrelude(mock, 7)
	mock.ExpectQuery("SELECT.*accounts.*content_moderation_config.*prompt_audit_config").WithArgs(int64(7)).
		WillReturnRows(platformReferenceRows(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16))
	mock.ExpectRollback()

	err := repo.DeleteUnused(context.Background(), 7)
	require.ErrorIs(t, err, service.ErrPlatformInUse)
	metadata := errors.FromError(err).Metadata
	require.Equal(t, "1", metadata["accounts"])
	require.Equal(t, "2", metadata["api_keys"])
	require.Equal(t, "15", metadata["audits"])
	require.Equal(t, "84", metadata["ops"])
	require.Equal(t, "31", metadata["configs"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPlatformRepositoryDeleteUnusedReturnsNotFound(t *testing.T) {
	repo, mock := newPlatformDeleteRepo(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM platforms WHERE id = $1 FOR UPDATE")).WithArgs(int64(7)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	require.ErrorIs(t, repo.DeleteUnused(context.Background(), 7), service.ErrPlatformNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
