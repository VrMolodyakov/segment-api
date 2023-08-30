package main

import (
	"context"
	"log"
	"os"

	"github.com/VrMolodyakov/segment-api/internal/app"
)

//  @title          Segment api
//  @version        1.0
//  @description    Api for working with segments

//  @contact.name   Vyacheslav Molodyakov
//  @contact.email  vrmolodyakov@mail.ru

//  @host       localhost:8080
//  @basePath   /api/v1

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
