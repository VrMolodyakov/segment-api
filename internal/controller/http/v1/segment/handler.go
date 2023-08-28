package segment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/validator"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
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

func (h *handler) CreateSegment(w http.ResponseWriter, r *http.Request) {
	var segmentReq CreateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&segmentReq); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	errs := validator.Validate(segmentReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		http.Error(w, string(jsonErr), http.StatusBadRequest)
		return
	}

	id, err := h.segment.CreateSegment(r.Context(), segmentReq.Name, segmentReq.Percentage)
	if err != nil {
		switch {
		case errors.Is(err, segment.ErrSegmentAlreadyExists):
			http.Error(w, "Segment already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Create segment error", http.StatusInternalServerError)
		return
	}

	response := NewSegmentResponse(id, segmentReq.Name, segmentReq.Percentage)
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}
