package history

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	history "github.com/VrMolodyakov/segment-api/internal/domain/history/model"
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

func (r *repo) Get(ctx context.Context, year int, month int) ([]history.History, error) {
	sql, args, err := r.builder.
		Select("user_id", "segment_name", "operation", "operation_timestamp").
		From(historyTable).
		Join("segments USING (segment_id)").
		Where(sq.And{
			sq.Eq{"DATE_PART('year', operation_timestamp)": year},
			sq.Eq{"DATE_PART('month', operation_timestamp)": month},
		}).
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
