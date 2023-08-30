package segment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/apierror"
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
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 409 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /segments [post]
func (h *handler) CreateSegment(w http.ResponseWriter, r *http.Request) {
	var segmentReq CreateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&segmentReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, fmt.Sprintf("invalid request: %s", err.Error()))
		return
	}

	errs := validator.Validate(segmentReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, string(jsonErr))
		return
	}

	id, err := h.segment.CreateSegment(r.Context(), segmentReq.Name, max-segmentReq.HitPercentage)
	if err != nil {
		switch {
		case errors.Is(err, segment.ErrSegmentAlreadyExists):
			w.WriteHeader(http.StatusConflict)
			apierror.WriteErrorMessage(w, "Segment already exists")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, "Create segment error")
		return
	}

	jsonResponse, err := json.Marshal(NewSegmentResponse(id, segmentReq.Name, segmentReq.HitPercentage))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}
