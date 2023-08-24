package repo

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/VrMolodyakov/segment-api/internal/domain/user/model"
	"github.com/VrMolodyakov/segment-api/internal/domain/user/service"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	"github.com/jackc/pgx/v5"
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

func (r *repo) Get(ctx context.Context, userID int) (model.User, error) {
	sql, args, err := r.builder.
		Select(
			"user_id",
			"first_name",
			"last_name",
			"email").
		From(userTable).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return model.User{}, err
	}
	var user model.User
	err = r.client.
		QueryRow(ctx, sql, args...).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, service.ErrUserNotFound
		}
		return model.User{}, err
	}
	return user, nil
}
