package usersegments

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	history "github.com/VrMolodyakov/segment-api/internal/domain/history/model"
	segment "github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	"github.com/VrMolodyakov/segment-api/pkg/clock"
	"github.com/jackc/pgx/v5"
)

const (
	segmentTable      string = "segments"
	userSegmentsTable string = "user_segments"
	historyTable      string = "segment_history"
)

type repo struct {
	client  psql.Client
	builder sq.StatementBuilderType
	clock   clock.Clock
}

func New(client psql.Client, clock clock.Clock) *repo {
	return &repo{
		client:  client,
		clock:   clock,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *repo) UpdateUserSegments(ctx context.Context, userID int64, addSegments []segment.Segment, deleteSegments []string) error {
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

	if err = r.insertIfExists(ctx, tx, userID, addSegments); err != nil {
		return err
	}

	var deleteIDs []int64
	if deleteIDs, err = r.deleteIfExists(ctx, tx, userID, deleteSegments); err != nil {
		return err
	}

	if err = r.registerEvents(ctx, tx, userID, addSegments, deleteIDs, r.clock.Now()); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *repo) GetUserSegments(ctx context.Context, userID int64) ([]history.History, error) {
	sql, args, err := r.builder.
		Select("history_id", "user_id", "segment_name", "operation", "operation_timestamp").
		From(historyTable).
		Join("segments USING (segment_id)").
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

	histories := make([]history.History, 0)
	for rows.Next() {
		var history history.History
		if err := rows.Scan(&history.ID, &history.UserID, &history.Segment, &history.Operation, &history.Time); err != nil {
			return nil, err
		}
		histories = append(histories, history)
	}

	return histories, nil
}

func (r *repo) insertIfExists(ctx context.Context, tx pgx.Tx, userID int64, addSegments []segment.Segment) error {
	if len(addSegments) > 0 {
		if err := r.fillInsertIDs(ctx, tx, addSegments); err != nil {
			return err
		}

		if err := r.insert(ctx, tx, userID, addSegments); err != nil {
			return err
		}
	}
	return nil
}

func (r *repo) deleteIfExists(ctx context.Context, tx pgx.Tx, userID int64, deleteSegments []string) ([]int64, error) {
	var deleteIDs []int64
	var err error
	if len(deleteSegments) > 0 {
		if deleteIDs, err = r.getDeleteIDs(ctx, tx, deleteSegments); err != nil {
			return nil, err
		}
		if err = r.delete(ctx, tx, userID, deleteIDs); err != nil {
			return nil, err
		}
	}
	return deleteIDs, nil
}

func (r *repo) getDeleteIDs(ctx context.Context, tx pgx.Tx, names []string) ([]int64, error) {
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

	ids := make([]int64, 0, len(names))
	for rows.Next() {
		var segmentID int64
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

func (r *repo) fillInsertIDs(ctx context.Context, tx pgx.Tx, segments []segment.Segment) error {
	names := make([]string, len(segments))
	for i := range segments {
		names[i] = segments[i].Name
	}
	sql, args, err := r.builder.
		Select("segment_id", "segment_name").
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

func (r *repo) delete(ctx context.Context, tx pgx.Tx, userID int64, segmentIDs []int64) error {
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

func (r *repo) insert(ctx context.Context, tx pgx.Tx, userID int64, segments []segment.Segment) error {
	insertState := r.builder.Insert(userSegmentsTable).Columns("user_id", "segment_id", "expired_at")
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

func (r *repo) registerEvents(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	inserted []segment.Segment,
	deleted []int64,
	timestamp time.Time,
) error {

	insertState := r.builder.Insert(historyTable).Columns("user_id", "segment_id", "operation", "operation_timestamp")

	for i := range inserted {
		insertState = insertState.Values(userID, inserted[i].ID, history.Added, timestamp)
	}

	for _, id := range deleted {
		insertState = insertState.Values(userID, id, history.Deleted, timestamp)
	}

	sql, args, err := insertState.ToSql()
	if err != nil {
		return err
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	neededLen := int64(len(inserted) + len(deleted))
	if rows.RowsAffected() != neededLen {
		return fmt.Errorf(
			"couldn't insert all the necessary rows, want %d , got %d",
			neededLen,
			rows.RowsAffected(),
		)
	}

	return nil
}
