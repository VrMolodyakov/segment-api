package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/config"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

type app struct {
	cfg    *config.Config
	logger logging.Logger
	deps   Deps
}

func New() *app {
	return &app{}
}

func (a *app) Setup(ctx context.Context) error {
	return a.deps.Setup(ctx, a.cfg, a.logger)
}

func (a *app) InitLogger(console io.Writer, files ...io.Writer) {
	loggerCfg := logging.NewLogerConfig(
		a.cfg.Logger.Development,
		a.cfg.Logger.Level,
	)
	a.logger = logging.NewLogger(loggerCfg)
	a.logger.InitLogger(console, files...)
}

func (a *app) ReadConfig() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	a.cfg = cfg
	return nil
}

func (a *app) Start(ctx context.Context) {
	ctx, stop := signal.NotifyContext(ctx, os.Kill, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		a.logger.Infof("segemnt server started on host %s port %d", a.cfg.HTTP.Host, a.cfg.HTTP.Port)
		if err := a.deps.server.ListenAndServe(); err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				a.logger.Warn("server shutdown")
			default:
				a.logger.Fatal(err.Error())
			}
		}
		err := a.deps.server.Shutdown(ctx)
		if err != nil {
			a.logger.Fatal(err.Error())
		}
	}()

	go func() {
		a.deps.cleaner.Start(ctx, time.Duration(a.cfg.Cleaner.Interval)*time.Second)
	}()

	<-ctx.Done()
}

func (a *app) Close(ctx context.Context) {
	a.deps.Close(ctx, a.logger)
}
