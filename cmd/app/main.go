package main

import (
	"context"
	"log"
	"os"

	"github.com/VrMolodyakov/segment-api/internal/app"
	_ "github.com/VrMolodyakov/segment-api/internal/domain/segment"
	_ "github.com/VrMolodyakov/segment-api/internal/domain/user"
	_ "github.com/VrMolodyakov/segment-api/internal/repository/history"
	_ "github.com/VrMolodyakov/segment-api/internal/repository/user"
	_ "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	_ "github.com/VrMolodyakov/segment-api/pkg/clock"
)

func main() {
	ctx := context.Background()
	a := app.New()

	defer func() {
		a.Close(ctx)
	}()

	if err := a.ReadConfig(); err != nil {
		log.Fatal(err, "read config")
		return
	}

	logFile, err := os.Create("temp.txt")
	if err != nil {
		log.Fatal(err)
	}

	a.InitLogger(os.Stderr, logFile)

	if err := a.Setup(ctx); err != nil {
		log.Fatal(err, "setup dependencies")
		return
	}

	a.Start(ctx)

}
