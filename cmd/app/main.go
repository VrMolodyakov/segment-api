package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/config"
	_ "github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	_ "github.com/VrMolodyakov/segment-api/internal/domain/user/model"
	"github.com/VrMolodyakov/segment-api/internal/repository/history"
	_ "github.com/VrMolodyakov/segment-api/internal/repository/user"
	"github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	_ "github.com/VrMolodyakov/segment-api/pkg/clock"
)

// usersegments "github.com/VrMolodyakov/segment-api/internal/repository/user_segments"
func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	pgCfg := postgresql.NewPgConfig(
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.PoolSize,
		cfg.Postgres.SSLMode,
	)

	ctx := context.Background()
	client, err := postgresql.NewClient(ctx, 5, 10*time.Second, pgCfg)
	if err != nil {
		log.Fatal(err)
	}

	// userRepo := user.New(client)
	// userRepo.Create(ctx, userModel.User{
	// 	FirstName: "Denis",
	// 	LastName:  "Best",
	// 	Email:     "d@mail.ru",
	// })

	// segmentRepo.Create(ctx, "1-seg")
	// segmentRepo.Create(ctx, "2-seg")
	// segmentRepo.Create(ctx, "3-seg")
	// segmentRepo.Create(ctx, "4-seg")

	// mclock := clock.New()
	// usrepo := usersegments.New(client, mclock)
	// s1 := model.Segment{Name: "1-seg", ExpiredAt: time.Now().Add(1 * time.Hour)}
	// s2 := model.Segment{Name: "2-seg", ExpiredAt: time.Now().Add(1 * time.Hour)}
	// addSegments := []model.Segment{s1, s2}
	// err = usrepo.UpdateUserSegments(ctx, 1, addSegments, []string{"3-seg", "4-seg"})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	history := history.New(client)
	data, err := history.Get(ctx, 2016, 6)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range data {
		fmt.Println(d)
	}
}
