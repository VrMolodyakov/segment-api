package segment

import (
	"context"
	"errors"
	"fmt"

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

func (r *repo) Create(ctx context.Context, name string) (int64, error) {
	sql, args, err := r.builder.
		Insert(segmentTable).
		Columns(
			"segment_name").
		Values(name).
		Suffix("RETURNING segment_id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("couldn't create query : %w", err)
	}
	var id int64
	err = r.client.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("couldn't run query : %w", err)
	}
	return id, nil
}

func (r *repo) Get(ctx context.Context, name string) (model.SegmentInfo, error) {
	sql, args, err := r.builder.
		Select(
			"segment_id",
			"segment_name").
		From(segmentTable).
		Where(sq.Eq{"segment_name": name}).
		ToSql()
	if err != nil {
		return model.SegmentInfo{}, fmt.Errorf("couldn't create query : %w", err)
	}
	var segment model.SegmentInfo
	err = r.client.
		QueryRow(ctx, sql, args...).
		Scan(&segment.ID, &segment.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.SegmentInfo{}, service.ErrSegmentNotFound
		}
		return model.SegmentInfo{}, fmt.Errorf("couldn't run query : %w", err)
	}
	return segment, nil
}

func (r *repo) GetAll(ctx context.Context) ([]model.SegmentInfo, error) {
	sql, args, err := r.builder.
		Select(
			"segment_id",
			"segment_name").
		From(segmentTable).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("couldn't create query : %w", err)
	}

	rows, err := r.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []model.SegmentInfo

	for rows.Next() {
		var segment model.SegmentInfo
		if err := rows.Scan(&segment.ID, &segment.Name); err != nil {
			return nil, fmt.Errorf("couldn't scan query : %w", err)
		}
		segments = append(segments, segment)
	}

	return segments, nil
}
