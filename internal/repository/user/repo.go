package user

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VrMolodyakov/segment-api/internal/domain/user"
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

func (r *repo) Get(ctx context.Context, userID int64) (user.User, error) {
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
		return user.User{}, fmt.Errorf("couldn't create query : %w", err)
	}
	var u user.User
	err = r.client.
		QueryRow(ctx, sql, args...).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, fmt.Errorf("couldn't get an account: %w", user.ErrUserNotFound)
		}
		return user.User{}, fmt.Errorf("couldn't get an account: %w", err)
	}
	return u, nil
}
