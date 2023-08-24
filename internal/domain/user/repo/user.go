package repo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/VrMolodyakov/segment-api/internal/domain/user/model"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
)

const (
	userTable string = "users"
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

func (r *repo) Create(ctx context.Context, user model.User) (int, error) {
	sql, args, err := r.builder.
		Insert(userTable).
		Columns(
			"first_name",
			"last_name",
			"email").
		Values(user.FirstName, user.LastName, user.Email).
		Suffix("RETURNING user_id").
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
