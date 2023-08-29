package membership

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/validator"
	"github.com/VrMolodyakov/segment-api/internal/domain/membership"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/internal/domain/user"
	"github.com/go-chi/chi/v5"
)

type MembershipService interface {
	CreateUser(ctx context.Context, user user.User) (int64, error)
	DeleteMembership(ctx context.Context, segmentName string) error
	GetUserMembership(ctx context.Context, userID int64) ([]membership.MembershipInfo, error)
	UpdateUserMembership(ctx context.Context, userID int64, addSegments []segment.Segment, deleteSegments []string) error
}

type handler struct {
	membership MembershipService
}

func New(membership MembershipService) *handler {
	return &handler{
		membership: membership,
	}
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userReq CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	errs := validator.Validate(userReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		http.Error(w, string(jsonErr), http.StatusBadRequest)
		return
	}

	newUser := userReq.ToModel()
	id, err := h.membership.CreateUser(r.Context(), newUser)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			http.Error(w, fmt.Sprintf("Invalid email: %s", err.Error()), http.StatusBadRequest)
			return
		case errors.Is(err, user.ErrUserAlreadyExist):
			http.Error(w, "User already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Create user error", http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(NewCreateUserResponse(id, newUser.FirstName, newUser.LastName, newUser.Email))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func (h *handler) UpdateUserMembership(w http.ResponseWriter, r *http.Request) {
	var updateReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	errs := validator.Validate(updateReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		http.Error(w, string(jsonErr), http.StatusBadRequest)
		return
	}

	err := h.membership.UpdateUserMembership(
		r.Context(),
		updateReq.UserID,
		updateReq.GetUpdatedSegments(),
		updateReq.GetDeletedSegments(),
	)

	if err != nil {
		switch {
		case errors.Is(err, segment.ErrSegmentNotFound):
			http.Error(w, "Not all segments with the specified names were found", http.StatusBadRequest)
			return
		case errors.Is(err, membership.ErrSegmentAlreadyAssigned):
			http.Error(w, "Attempt to add segments that the user already belongs to", http.StatusBadRequest)
			return
		case errors.Is(err, user.ErrUserNotFound):
			http.Error(w, "Attempt to update the data of a non-existent user", http.StatusBadRequest)
			return
		case errors.Is(err, membership.ErrEmptyData):
			http.Error(w, "Data for update and delete cannot be empty at the same time", http.StatusBadRequest)
			return
		case errors.Is(err, membership.ErrIncorrectData):
			http.Error(w, "Attempt to add and remove the same segment", http.StatusBadRequest)
			return
		case errors.Is(err, membership.ErrSegmentNotAssigned):
			http.Error(w, "Attempt to delete a segment unassigned to the user", http.StatusBadRequest)
			return

		}
		http.Error(w, "Update user segments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *handler) DeleteMembership(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "segmentName")
	err := h.membership.DeleteMembership(r.Context(), name)
	if err != nil {
		switch {
		case errors.Is(err, segment.ErrSegmentNotFound):
			http.Error(w, "Segment with the specified name wasn't found", http.StatusBadRequest)
			return
		}
		http.Error(w, "Delete segment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *handler) GetUserMembership(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user id parameter", http.StatusBadRequest)
		return
	}

	data, err := h.membership.GetUserMembership(r.Context(), userID)
	if err != nil {
		http.Error(w, "Get user membership segment", http.StatusInternalServerError)
		return
	}

	if len(data) == 0 {
		http.Error(w, "No data was found for the specified user", http.StatusNotFound)
		return
	}

	response := make([]UserResponseInfo, len(data))
	for i, d := range data {
		response[i] = NewUserResponseInfo(d.UserID, d.SegmentName, d.ExpiredAt)
	}

	jsonResponse, err := json.Marshal(NewUserMembershipResponse(response))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
