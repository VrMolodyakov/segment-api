package segment

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment/service"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	"github.com/jackc/pgx/v5"
)

const (
	segmentTable string = "segments"
)

type repo struct {
	builder sq.StatementBuilderType
	client  psql.Client
}

func New(client psql.Client) *repo {
	return &repo{
		client:  client,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *repo) Create(ctx context.Context, segment model.Segment) (int, error) {
	sql, args, err := r.builder.
		Insert(segmentTable).
		Columns(
			"segment_name").
		Values(segment.Name).
		Suffix("RETURNING segment_id").
		ToSql()
	if err != nil {
		return 0, err
	}
	var id int
	err = r.client.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repo) Get(ctx context.Context, name string) (model.Segment, error) {
	sql, args, err := r.builder.
		Select(
			"segment_id",
			"segment_name").
		From(segmentTable).
		Where(sq.Eq{"segment_name": name}).
		ToSql()
	if err != nil {
		return model.Segment{}, err
	}
	var segment model.Segment
	err = r.client.
		QueryRow(ctx, sql, args...).
		Scan(&segment.ID, &segment.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Segment{}, service.ErrSegmentNotFound
		}
		return model.Segment{}, err
	}
	return segment, nil
}

func (r *repo) Delete(ctx context.Context, name string) error {
	sql, args, err := r.builder.
		Delete(segmentTable).
		Where(sq.Eq{"segment_name": name}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.client.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return service.ErrSegmentNotFound
	}

	return nil
}

func (r *repo) GetAll(ctx context.Context) ([]model.Segment, error) {
	sql, args, err := r.builder.
		Select(
			"segment_id",
			"segment_name").
		From(segmentTable).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []model.Segment

	for rows.Next() {
		var segment model.Segment
		if err := rows.Scan(&segment.ID, &segment.Name); err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return segments, nil
}
