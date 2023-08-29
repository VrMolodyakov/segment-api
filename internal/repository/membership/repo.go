package membership

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/VrMolodyakov/segment-api/internal/domain/history"
	"github.com/VrMolodyakov/segment-api/internal/domain/membership"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/internal/domain/user"
	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	"github.com/VrMolodyakov/segment-api/pkg/clock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	segmentTable      string = "segments"
	userTable         string = "users"
	userSegmentsTable string = "user_segments"
	historyTable      string = "segment_history"
)

var (
	maxFutureTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
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
		return fmt.Errorf("couldn't begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w, initial error: %s", rollbackErr, err.Error())
			}

		}
	}()

	if err = r.insertIfExists(ctx, tx, userID, addSegments); err != nil {
		return err
	}

	if err = r.deleteIfExists(ctx, tx, userID, deleteSegments); err != nil {
		return err
	}

	if err = r.registerUpdateUserEvent(ctx, tx, userID, addSegments, deleteSegments, r.clock.Now()); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("couldn't commit transaction: %w", err)
	}

	return nil
}

func (r *repo) DeleteSegment(ctx context.Context, name string) error {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return fmt.Errorf("couldn't begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w, initial error: %s", rollbackErr, err.Error())
			}

		}
	}()

	segmentID, err := r.getDeleteID(ctx, tx, name)
	if err != nil {
		return err
	}

	users, err := r.getUsersBySegmentId(ctx, tx, segmentID)
	if err != nil {
		return err
	}

	if len(users) > 0 {
		if err = r.deleteBySegmentID(ctx, tx, segmentID); err != nil {
			return err
		}

		if err = r.registerDeleteUsersEvent(ctx, tx, users, name, r.clock.Now()); err != nil {
			return err
		}
	}

	if err = r.deleteSegment(ctx, tx, segmentID); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("couldn't commit transaction: %w", err)
	}

	return nil
}

func (r *repo) GetUserSegments(ctx context.Context, id int64) ([]membership.MembershipInfo, error) {
	sql, args, err := r.builder.
		Select("user_id", "segment_name", "expired_at").
		From(userSegmentsTable).
		Join("segments USING (segment_id)").
		Where(sq.Eq{"user_id": id}).
		Where(sq.Gt{"expired_at": r.clock.Now()}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("couldn't create query : %w", err)
	}

	rows, err := r.client.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("couldn't run query : %w", err)
	}
	defer rows.Close()

	memberships := make([]membership.MembershipInfo, 0)
	for rows.Next() {
		var m membership.MembershipInfo
		if err := rows.Scan(
			&m.UserID,
			&m.SegmentName,
			&m.ExpiredAt); err != nil {
			return nil, fmt.Errorf("couldn't scan membership info : %w", err)
		}
		memberships = append(memberships, m)
	}

	return memberships, nil
}

func (r *repo) CreateUser(ctx context.Context, user user.User, hitPercentage int) (int64, error) {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("couldn't begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w, initial error: %s", rollbackErr, err.Error())
			}

		}
	}()

	userID, err := r.createUser(ctx, tx, user)
	if err != nil {
		return 0, err
	}

	segments, err := r.hitPercentage(ctx, tx, hitPercentage)
	if err != nil {
		return 0, err
	}
	if len(segments) > 0 {
		if err = r.insertDefault(ctx, tx, userID, segments); err != nil {
			return 0, err
		}

		if err = r.registerInsertUserEvents(ctx, tx, userID, segments, r.clock.Now()); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("couldn't commit transaction: %w", err)
	}

	return userID, nil
}

func (r *repo) DeleteExpired(ctx context.Context) error {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return fmt.Errorf("couldn't begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w, initial error: %s", rollbackErr, err.Error())
			}

		}
	}()

	expired, err := r.getExpiredRows(ctx, tx)
	if err != nil {
		return err
	}

	if len(expired) > 0 {
		if err = r.registerCleanupUserEvents(ctx, tx, expired); err != nil {
			return err
		}

		if err = r.deleteExpired(ctx, tx); err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("couldn't commit transaction: %w", err)
	}

	return nil
}

func (r *repo) getExpiredRows(ctx context.Context, tx pgx.Tx) ([]membership.MembershipInfo, error) {
	sql, args, err := r.builder.
		Select("user_id", "segment_name", "expired_at").
		From(userSegmentsTable).
		Join("segments USING (segment_id)").
		Where(sq.Lt{"expired_at": r.clock.Now()}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("couldn't create query : %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("couldn't run query : %w", err)
	}
	defer rows.Close()

	memberships := make([]membership.MembershipInfo, 0)
	for rows.Next() {
		var m membership.MembershipInfo
		if err := rows.Scan(
			&m.UserID,
			&m.SegmentName,
			&m.ExpiredAt); err != nil {
			return nil, fmt.Errorf("couldn't scan membership data : %w", err)
		}
		memberships = append(memberships, m)
	}

	return memberships, nil
}

func (r *repo) deleteExpired(ctx context.Context, tx pgx.Tx) error {
	sql, args, err := r.builder.
		Delete(userSegmentsTable).
		Where(sq.Lt{"expired_at": r.clock.Now()}).
		ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run query : %w", err)
	}
	return nil
}

func (r *repo) hitPercentage(ctx context.Context, tx pgx.Tx, percentage int) ([]segment.SegmentInfo, error) {
	sql, args, err := r.builder.
		Select(
			"segment_id",
			"segment_name").
		From(segmentTable).
		Where(sq.Lt{"automatic_percentage": percentage}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("couldn't create query : %w", err)
	}
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segmentIDs []segment.SegmentInfo

	for rows.Next() {
		var s segment.SegmentInfo
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, fmt.Errorf("couldn't scan id : %w", err)
		}
		segmentIDs = append(segmentIDs, s)
	}

	return segmentIDs, nil
}

func (r *repo) createUser(ctx context.Context, tx pgx.Tx, newUser user.User) (int64, error) {
	sql, args, err := r.builder.
		Insert(userTable).
		Columns(
			"first_name",
			"last_name",
			"email").
		Values(newUser.FirstName, newUser.LastName, newUser.Email).
		Suffix("RETURNING user_id").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("couldn't create query : %w", err)
	}
	var id int64
	err = tx.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return 0, fmt.Errorf("couldn't create an account: %w", user.ErrUserAlreadyExist)
			}
		}

		return 0, fmt.Errorf("couldn't create an account: %w", err)
	}
	return id, nil
}

func (r *repo) deleteSegment(ctx context.Context, tx pgx.Tx, segmentID int64) error {
	sql, args, err := r.builder.
		Delete(segmentTable).
		Where(sq.Eq{"segment_id": segmentID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run query : %w", err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return segment.ErrSegmentNotFound
	}

	return nil
}

func (r *repo) insertIfExists(ctx context.Context, tx pgx.Tx, userID int64, addSegments []segment.Segment) error {
	if len(addSegments) > 0 {
		if err := r.fillInsertIDs(ctx, tx, addSegments); err != nil {
			return err
		}

		if err := r.insertWithExpirity(ctx, tx, userID, addSegments); err != nil {
			return err
		}
	}
	return nil
}

func (r *repo) deleteIfExists(ctx context.Context, tx pgx.Tx, userID int64, deleteSegments []string) error {
	var deleteIDs []int64
	var err error
	if len(deleteSegments) > 0 {
		if deleteIDs, err = r.getDeleteIDs(ctx, tx, deleteSegments...); err != nil {
			return err
		}
		if err = r.deleteUserSegments(ctx, tx, userID, deleteIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *repo) getUsersBySegmentId(ctx context.Context, tx pgx.Tx, segmentID int64) ([]int64, error) {
	sql, args, err := r.builder.
		Select("user_id").
		From(userSegmentsTable).
		Where(sq.Eq{"segment_id": segmentID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("couldn't create query : %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("couldn't run query : %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("couldn't scan user id : %w", err)
		}
		ids = append(ids, userID)
	}
	return ids, nil
}

func (r *repo) getDeleteIDs(ctx context.Context, tx pgx.Tx, names ...string) ([]int64, error) {
	sql, args, err := r.builder.
		Select("segment_id").
		From(segmentTable).
		Where(sq.Eq{"segment_name": names}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("couldn't create query : %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("couldn't run select segment id query : %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0, len(names))
	for rows.Next() {
		var segmentID int64
		if err := rows.Scan(&segmentID); err != nil {
			return nil, fmt.Errorf("couldn't scan segment id : %w", err)
		}
		ids = append(ids, segmentID)
	}

	if len(ids) != len(names) {
		return nil, segment.ErrSegmentNotFound
	}

	return ids, nil
}

func (r *repo) getDeleteID(ctx context.Context, tx pgx.Tx, name string) (int64, error) {
	sql, args, err := r.builder.
		Select("segment_id").
		From(segmentTable).
		Where(sq.Eq{"segment_name": name}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("couldn't create query : %w", err)
	}
	var segmentID int64
	err = tx.QueryRow(ctx, sql, args...).Scan(&segmentID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, segment.ErrSegmentNotFound
		}
		return 0, fmt.Errorf("couldn't find segment id : %w", err)
	}

	return segmentID, nil
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
		return fmt.Errorf("couldn't create query : %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run fill insert id query: %w", err)
	}
	defer rows.Close()

	ids := make(map[string]int64)
	for rows.Next() {
		var id int64
		var s string
		if err := rows.Scan(&id, &s); err != nil {
			return fmt.Errorf("scan segment id,name: %w", err)
		}
		ids[s] = id
	}

	if len(ids) != len(names) {
		return segment.ErrSegmentNotFound
	}

	for i := range segments {
		segments[i].ID = ids[segments[i].Name]
	}

	return nil
}

func (r *repo) deleteUserSegments(ctx context.Context, tx pgx.Tx, userID int64, segmentIDs []int64) error {
	sql, args, err := r.builder.
		Delete(userSegmentsTable).
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"segment_id": segmentIDs}).
		ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}

	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run query : %w", err)
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

func (r *repo) deleteBySegmentID(ctx context.Context, tx pgx.Tx, segmentID int64) error {
	sql, args, err := r.builder.
		Delete(userSegmentsTable).
		Where(sq.Eq{"segment_id": segmentID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run query : %w", err)
	}
	if rows.RowsAffected() == int64(0) {
		return fmt.Errorf(
			"couldn't delete rows, rows affected %d", rows.RowsAffected(),
		)
	}
	return nil
}

func (r *repo) insertWithExpirity(ctx context.Context, tx pgx.Tx, userID int64, segments []segment.Segment) error {
	insertState := r.builder.Insert(userSegmentsTable).Columns("user_id", "segment_id", "expired_at")
	for i := range segments {
		if segments[i].ExpiredAt.IsZero() {
			insertState = insertState.Values(userID, segments[i].ID, maxFutureTime)
		} else {
			insertState = insertState.Values(userID, segments[i].ID, segments[i].ExpiredAt)
		}
	}

	sql, args, err := insertState.ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return fmt.Errorf("insert user: %w", membership.ErrSegmentAlreadyAssigned)
			}
			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				return fmt.Errorf("insert user: %w", user.ErrUserNotFound)
			}
		}
		return fmt.Errorf("couldn't run insert query : %w", err)
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

func (r *repo) insertDefault(ctx context.Context, tx pgx.Tx, userID int64, segments []segment.SegmentInfo) error {
	insertState := r.builder.Insert(userSegmentsTable).Columns("user_id", "segment_id", "expired_at")
	for i := range segments {
		insertState = insertState.Values(userID, segments[i].ID, maxFutureTime)
	}

	sql, args, err := insertState.ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run insert query : %w", err)
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

func (r *repo) registerUpdateUserEvent(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	inserted []segment.Segment,
	deleted []string,
	timestamp time.Time,
) error {

	insertState := r.builder.Insert(historyTable).Columns("user_id", "segment_name", "operation", "operation_timestamp")

	for i := range inserted {
		insertState = insertState.Values(userID, inserted[i].Name, history.Added, timestamp)
	}

	for _, id := range deleted {
		insertState = insertState.Values(userID, id, history.Deleted, timestamp)
	}

	sql, args, err := insertState.ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run query : %w", err)
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

func (r *repo) registerDeleteUsersEvent(
	ctx context.Context,
	tx pgx.Tx,
	users []int64,
	segment string,
	timestamp time.Time,
) error {

	insertState := r.builder.Insert(historyTable).Columns("user_id", "segment_name", "operation", "operation_timestamp")

	for i := range users {
		insertState = insertState.Values(users[i], segment, history.Deleted, timestamp)
	}

	sql, args, err := insertState.ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run query : %w", err)
	}
	neededLen := int64(len(users))
	if rows.RowsAffected() != neededLen {
		return fmt.Errorf(
			"couldn't insert all the necessary rows, want %d , got %d",
			neededLen,
			rows.RowsAffected(),
		)
	}

	return nil
}

func (r *repo) registerInsertUserEvents(
	ctx context.Context,
	tx pgx.Tx,
	user int64,
	segments []segment.SegmentInfo,
	timestamp time.Time,
) error {

	insertState := r.builder.Insert(historyTable).Columns("user_id", "segment_name", "operation", "operation_timestamp")

	for i := range segments {
		insertState = insertState.Values(user, segments[i].Name, history.Added, timestamp)
	}

	sql, args, err := insertState.ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run query : %w", err)
	}
	neededLen := int64(len(segments))
	if rows.RowsAffected() != neededLen {
		return fmt.Errorf(
			"couldn't insert all the necessary rows, want %d , got %d",
			neededLen,
			rows.RowsAffected(),
		)
	}

	return nil
}

func (r *repo) registerCleanupUserEvents(
	ctx context.Context,
	tx pgx.Tx,
	memberships []membership.MembershipInfo,
) error {

	insertState := r.builder.Insert(historyTable).Columns("user_id", "segment_name", "operation", "operation_timestamp")

	for i := range memberships {
		insertState = insertState.Values(
			memberships[i].UserID,
			memberships[i].SegmentName,
			history.Deleted,
			memberships[i].ExpiredAt,
		)
	}

	sql, args, err := insertState.ToSql()
	if err != nil {
		return fmt.Errorf("couldn't create query : %w", err)
	}
	rows, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("couldn't run query : %w", err)
	}
	neededLen := int64(len(memberships))
	if rows.RowsAffected() != neededLen {
		return fmt.Errorf(
			"couldn't insert all the necessary rows, want %d , got %d",
			neededLen,
			rows.RowsAffected(),
		)
	}

	return nil
}
