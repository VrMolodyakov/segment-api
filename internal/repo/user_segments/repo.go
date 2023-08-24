package usersegments

import (
	"context"
	"errors"
	"fmt"

	psql "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	"github.com/VrMolodyakov/segment-api/pkg/slice"
	"github.com/jackc/pgx/v5"
)

type repo struct {
	client psql.Client
}

func New(client psql.Client) *repo {
	return &repo{
		client: client,
	}
}

func (r *repo) UpdateUserSegments(ctx context.Context, userID int, addSegments []string, deleteSegments []string) error {
	tx, err := r.client.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction: %w", err)
			}

		}
	}()

	// addIDs, err := r.getIDs(ctx, tx, addSegments)
	// if err != nil {
	// 	return err
	// }
	deleteIDs, err := r.getIDs(ctx, tx, deleteSegments)
	if err != nil {
		return err
	}

	if err := r.delete(ctx, tx, deleteIDs); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *repo) getIDs(ctx context.Context, tx pgx.Tx, names []string) ([]any, error) {
	query := "SELECT segment_id FROM segments WHERE segment_name = ANY($1)"
	rows, err := tx.Query(ctx, query, slice.ToAnySlice(names)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]any, 0, len(names))
	for rows.Next() {
		var segmentID int
		if err := rows.Scan(&segmentID); err != nil {
			return nil, err
		}
		ids = append(ids, segmentID)
	}

	if len(ids) != len(names) {
		return nil, fmt.Errorf("not all segments were found in the query")
	}

	return ids, nil
}

func (r *repo) delete(ctx context.Context, tx pgx.Tx, ids []any) error {
	query := "DELETE FROM user_segments WHERE segment_id = ANY($1)"
	rows, err := tx.Exec(ctx, query, ids...)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != int64(len(ids)) {
		return errors.New("couldn't delete all the necessary ids")
	}

	return nil
}

// func (r *repo) UpdateUserSegments(ctx context.Context, addNameSegments []string, deleteNameSegments []string, userID int) error {
//     tx, err := r.client.Begin(ctx)
//     if err != nil {
//         return err
//     }
//     defer func() {
//         if txErr := tx.Rollback(ctx); txErr != nil && err == nil {
//             err = txErr
//         }
//     }()

//     // Выполняем запросы для получения segment_id для addNameSegments
//     addSegmentIDs, err := r.getSegmentIDsByNamesTx(ctx, tx, addNameSegments)
//     if err != nil {
//         return err
//     }

//     // Выполняем запросы для получения segment_id для deleteNameSegments
//     deleteSegmentIDs, err := r.getSegmentIDsByNamesTx(ctx, tx, deleteNameSegments)
//     if err != nil {
//         return err
//     }

//     // Удаляем связи пользователя с сегментами из deleteSegmentIDs
//     err = r.removeUserFromSegmentsTx(ctx, tx, deleteSegmentIDs, userID)
//     if err != nil {
//         return err
//     }

//     // Вставляем новые связи пользователя с сегментами из addSegmentIDs
//     err = r.addUserToSegmentsTx(ctx, tx, addSegmentIDs, userID)
//     if err != nil {
//         return err
//     }

//     if err := tx.Commit(ctx); err != nil {
//         return err
//     }

//     return nil
// }

// func (r *repo) getSegmentIDsByNamesTx(ctx context.Context, tx psql.Tx, segmentNames []string) ([]int, error) {
//     var segmentIDs []int
//     query, args, err := r.builder.
//         Select("segment_id").
//         From(segmentTable).
//         Where(sq.Eq{"segment_name": segmentNames}).
//         ToSql()
//     if err != nil {
//         return nil, err
//     }

//     rows, err := tx.Query(ctx, query, args...)
//     if err != nil {
//         return nil, err
//     }
//     defer rows.Close()

//     for rows.Next() {
//         var segmentID int
//         if err := rows.Scan(&segmentID); err != nil {
//             return nil, err
//         }
//         segmentIDs = append(segmentIDs, segmentID)
//     }

//     return segmentIDs, nil
// }

// func (r *repo) removeUserFromSegmentsTx(ctx context.Context, tx psql.Tx, segmentIDs []int, userID int) error {
//     if len(segmentIDs) == 0 {
//         return nil
//     }

//     _, err := tx.Exec(ctx,
//         "DELETE FROM "+userSegmentsTable+" WHERE user_id = $1 AND segment_id = ANY($2)",
//         userID, pq.Array(segmentIDs))
//     if err != nil {
//         return err
//     }

//     return nil
// }

// func (r *repo) addUserToSegmentsTx(ctx context.Context, tx psql.Tx, segmentIDs []int, userID int) error {
//     if len(segmentIDs) == 0 {
//         return nil
//     }

//     values := make([]interface{}, len(segmentIDs))
//     for i, id := range segmentIDs {
//         values[i] = struct {
//             UserID    int
//             SegmentID int
//         }{
//             UserID:    userID,
//             SegmentID: id,
//         }
//     }

//     _, err := tx.Exec(ctx,
//         "INSERT INTO "+userSegmentsTable+" (user_id, segment_id) VALUES "+psql.InsertValues(len(segmentIDs), 2),
//         values...)
//     if err != nil {
//         return err
//     }

//     return nil
// }
