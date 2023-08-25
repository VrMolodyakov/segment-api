package user

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VrMolodyakov/segment-api/internal/domain/user/model"
	"github.com/VrMolodyakov/segment-api/internal/domain/user/service"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

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

func (r *repo) Create(ctx context.Context, user model.User) (int64, error) {
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
		return 0, fmt.Errorf("couldn't create query : %w", err)
	}
	var id int64
	err = r.client.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return 0, fmt.Errorf("couldn't create an account: %w", service.ErrUserAlreadyExist)
			}
		}

		return 0, fmt.Errorf("couldn't create an account: %w", err)
	}
	return id, nil
}

func (r *repo) Get(ctx context.Context, userID int64) (model.User, error) {
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
		return model.User{}, fmt.Errorf("couldn't create query : %w", err)
	}
	var user model.User
	err = r.client.
		QueryRow(ctx, sql, args...).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, fmt.Errorf("couldn't get an account: %w", service.ErrUserNotFound)
		}
		return model.User{}, fmt.Errorf("couldn't get an account: %w", err)
	}
	return user, nil
}
