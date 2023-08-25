package usersegments

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	history "github.com/VrMolodyakov/segment-api/internal/domain/history/model"
	segment "github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	"github.com/jackc/pgx/v5"
)

const (
	segmentTable      string = "segments"
	userSegmentsTable string = "segment_history"
	historyTable      string = "segment_history"
)

type repo struct {
	client  psql.Client
	builder sq.StatementBuilderType
}

func New(client psql.Client) *repo {
	return &repo{
		client:  client,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *repo) UpdateUserSegments(ctx context.Context, userID int, addSegments []segment.Segment, deleteSegments []string) error {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w", err)
			}

		}
	}()

	if err := r.getInsertIDs(ctx, tx, addSegments); err != nil {
		return err
	}

	deleteIDs, err := r.getDeleteIDs(ctx, tx, deleteSegments)
	if err != nil {
		return err
	}

	if err := r.delete(ctx, tx, userID, deleteIDs); err != nil {
		return err
	}

	if err := r.recordDeleteHistory(ctx, tx, userID, deleteIDs, history.Deleted, time.Now()); err != nil {
		return err
	}

	if err := r.insert(ctx, tx, userID, addSegments); err != nil {
		return err
	}

	if err := r.recordInsertHistory(ctx, tx, userID, addSegments, history.Added, time.Now()); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *repo) GetUserSegments(ctx context.Context, userID int) ([]history.Histpry, error) {
	sql, args, err := r.builder.
		Select("history_id", "user_id", "segment_id", "operation", "operation_timestamp").
		From(historyTable).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	histories := make([]history.Histpry, 0)
	for rows.Next() {
		var history history.Histpry
		if err := rows.Scan(&history); err != nil {
			return nil, err
		}
		histories = append(histories, history)
	}

	return histories, nil
}

func (r *repo) getDeleteIDs(ctx context.Context, tx pgx.Tx, names []string) ([]int, error) {
	sql, args, err := r.builder.
		Select("segment_id").
		From(segmentTable).
		Where(sq.Eq{"segment_name": names}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int, 0, len(names))
	for rows.Next() {
		var segmentID int
		if err := rows.Scan(&segmentID); err != nil {
			return nil, err
		}
		ids = append(ids, segmentID)
	}

	if len(ids) != len(names) {
		return nil, fmt.Errorf(
			"not all segments were found in the query, want %d , got %d",
			len(names),
			len(ids),
		)
	}

	return ids, nil
}

func (r *repo) getInsertIDs(ctx context.Context, tx pgx.Tx, segments []segment.Segment) error {
	names := make([]string, len(segments))
	for i := range segments {
		names[i] = segments[i].Name
	}
	sql, args, err := r.builder.
		Select("segment_id,segment_name").
		From(segmentTable).
		Where(sq.Eq{"segment_name": names}).
		ToSql()
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	ids := make(map[string]int64)
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		ids[name] = id
	}

	if len(ids) != len(names) {
		return fmt.Errorf(
			"not all segments were found in the query, want %d , got %d",
			len(names),
			len(ids),
		)
	}

	for i := range segments {
		segments[i].ID = ids[segments[i].Name]
	}

	return nil
}

func (r *repo) delete(ctx context.Context, tx pgx.Tx, userID int, segmentIDs []int) error {
	sql, args, err := r.builder.
		Delete(userSegmentsTable).
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"segment_id": segmentIDs}).
		ToSql()
	if err != nil {
		return err
	}

	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != int64(len(segmentIDs)) {
		return fmt.Errorf(
			"couldn't delete all the necessary rows, want %d , got %d",
			len(segmentIDs),
			rows.RowsAffected(),
		)
	}

	return nil
}

func (r *repo) insert(ctx context.Context, tx pgx.Tx, userID int, segments []segment.Segment) error {
	insertState := r.builder.Insert(historyTable).Columns("user_id", "segment_id", "expired_at")
	for i := range segments {
		insertState = insertState.Values(userID, segments[i].ID, segments[i].ExpiredAt)
	}

	sql, args, err := insertState.ToSql()
	if err != nil {
		return err
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != int64(len(segments)) {
		return fmt.Errorf(
			"couldn't insert all the necessary rows, want %d , got %d",
			len(segments),
			rows.RowsAffected(),
		)
	}

	return nil
}

func (r *repo) recordDeleteHistory(
	ctx context.Context,
	tx pgx.Tx,
	userID int,
	segmentIDs []int,
	operation history.Operation,
	timestamp time.Time,
) error {
	insertState := r.builder.Insert(historyTable).Columns("user_id", "segment_id", "operation", "operation_timestamp")
	for _, id := range segmentIDs {
		insertState = insertState.Values(userID, id, operation, timestamp)
	}
	sql, args, err := insertState.ToSql()
	if err != nil {
		return err
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != int64(len(segmentIDs)) {
		return fmt.Errorf(
			"couldn't insert all the necessary rows, want %d , got %d",
			len(segmentIDs),
			rows.RowsAffected(),
		)
	}

	return nil

}

func (r *repo) recordInsertHistory(
	ctx context.Context,
	tx pgx.Tx,
	userID int,
	segments []segment.Segment,
	operation history.Operation,
	timestamp time.Time,
) error {
	insertState := r.builder.Insert(historyTable).Columns("user_id", "segment_id", "operation", "operation_timestamp")
	for i := range segments {
		insertState = insertState.Values(userID, segments[i].ID, operation, timestamp)
	}
	sql, args, err := insertState.ToSql()
	if err != nil {
		return err
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != int64(len(segments)) {
		return fmt.Errorf(
			"couldn't insert all the necessary rows, want %d , got %d",
			len(segments),
			rows.RowsAffected(),
		)
	}

	return nil

}
