package history

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/VrMolodyakov/segment-api/internal/domain/history"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
)

const (
	historyTable string = "segment_history"
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

func (r *repo) Get(ctx context.Context, date history.Date) ([]history.History, error) {
	sql, args, err := r.builder.
		Select("user_id", "segment_name", "operation", "operation_timestamp").
		From(historyTable).
		Where(sq.And{
			sq.Eq{"DATE_PART('year', operation_timestamp)": date.Year},
			sq.Eq{"DATE_PART('month', operation_timestamp)": date.Month},
		}).
		OrderBy("user_id").
		OrderBy("operation_timestamp").
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
		if err := rows.Scan(&history.UserID, &history.Segment, &history.Operation, &history.Time); err != nil {
			return nil, err
		}
		histories = append(histories, history)
	}

	return histories, nil
}
