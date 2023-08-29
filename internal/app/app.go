package app

import (
	"github.com/VrMolodyakov/segment-api/internal/config"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

type app struct {
	cfg    *config.Config
	logger logging.Logger
	// deps   Deps
}

func New() *app {
	return &app{}
}
