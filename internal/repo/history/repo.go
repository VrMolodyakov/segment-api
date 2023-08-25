package history

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
)

const (
	userTable string = "segment_history"
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

func (r *repo) Get(ctx context.Context)
