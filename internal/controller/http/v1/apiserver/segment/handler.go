package segment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	api "github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/errors"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/validator"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
)

const (
	max int = 100
)

type SegmentService interface {
	CreateSegment(ctx context.Context, name string, percentage int) (int64, error)
}

type handler struct {
	segment SegmentService
}

func New(segment SegmentService) *handler {
	return &handler{
		segment: segment,
	}
}

// @Summary Create a new segment
// @Description Creates a new segment with the provided details.
// @Tags Segments
// @Accept json
// @Produce json
// @Param segmentReq body CreateSegmentRequest true "Segment creation request"
// @Success 201 {object} CreateSegmentResponse "Segment created successfully"
// @Failure 400 {string} string "Bad request or validation error"
// @Failure 500 {string} string "Internal server error"
// @Router /create-segment [post]
func (h *handler) CreateSegment(w http.ResponseWriter, r *http.Request) {
	var segmentReq CreateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&segmentReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		api.WriteErrorMessage(w, fmt.Sprintf("invalid request: %s", err.Error()))
		return
	}

	errs := validator.Validate(segmentReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		w.WriteHeader(http.StatusBadRequest)
		api.WriteErrorMessage(w, string(jsonErr))
		return
	}

	id, err := h.segment.CreateSegment(r.Context(), segmentReq.Name, max-segmentReq.HitPercentage)
	if err != nil {
		switch {
		case errors.Is(err, segment.ErrSegmentAlreadyExists):
			w.WriteHeader(http.StatusBadRequest)
			api.WriteErrorMessage(w, "Segment already exists")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		api.WriteErrorMessage(w, "Create segment error")
		return
	}

	jsonResponse, err := json.Marshal(NewSegmentResponse(id, segmentReq.Name, segmentReq.HitPercentage))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.WriteErrorMessage(w, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}
