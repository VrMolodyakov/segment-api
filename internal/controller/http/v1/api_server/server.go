package segmentsserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/config"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/api_server/history"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/api_server/membership"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/api_server/segment"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(
	cfg config.HTTP,
	segmentService segment.SegmentService,
	historyService history.HistoryService,
	membershipService membership.MembershipService,
	pool history.BufferPool,
	writer history.CSVWriter,
) *http.Server {

	segmentHandler := segment.New(segmentService)
	historyHandler := history.New(historyService, history.NewLinkParam(cfg.Host, cfg.Port), pool, writer)
	membershipHandler := membership.New(membershipService)

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/segment", func(r chi.Router) {
		r.Post("/create", segmentHandler.CreateSegment)
		r.Route("/{segmentName}", func(r chi.Router) {
			r.Delete("/", membershipHandler.DeleteMembership)
		})
	})

	router.Route("/membership", func(r chi.Router) {
		r.Post("/update", membershipHandler.UpdateUserMembership)
	})

	router.Route("/users", func(r chi.Router) {
		r.Post("/", membershipHandler.CreateUser)
		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/", membershipHandler.GetUserMembership)
		})
	})

	router.Route("/history", func(r chi.Router) {
		r.Post("/link", historyHandler.CreateLink)
		r.Route("/{year}", func(r chi.Router) {
			r.Route("/{month}", func(r chi.Router) {
				r.Get("/", historyHandler.DownloadCSVData)
			})
		})
	})

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &http.Server{
		Addr:         addr,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
	}
}
