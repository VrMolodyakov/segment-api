package membership

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	api "github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/errors"
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
		w.WriteHeader(http.StatusBadRequest)
		api.WriteErrorMessage(w, fmt.Sprintf("invalid request: %s", err.Error()))
		return
	}

	errs := validator.Validate(userReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		w.WriteHeader(http.StatusBadRequest)
		api.WriteErrorMessage(w, string(jsonErr))
		return
	}

	newUser := userReq.ToModel()
	id, err := h.membership.CreateUser(r.Context(), newUser)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			w.WriteHeader(http.StatusBadRequest)
			api.WriteErrorMessage(w, fmt.Sprintf("Invalid email: %s", err.Error()))
			return
		case errors.Is(err, user.ErrUserAlreadyExist):
			w.WriteHeader(http.StatusBadRequest)
			api.WriteErrorMessage(w, "User already exists")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		api.WriteErrorMessage(w, "Create user error")
		return
	}

	jsonResponse, err := json.Marshal(NewCreateUserResponse(id, newUser.FirstName, newUser.LastName, newUser.Email))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.WriteErrorMessage(w, http.StatusText(http.StatusInternalServerError))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func (h *handler) UpdateUserMembership(w http.ResponseWriter, r *http.Request) {
	var updateReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		api.WriteErrorMessage(w, fmt.Sprintf("Invalid request: %s", err.Error()))
		return
	}

	errs := validator.Validate(updateReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		w.WriteHeader(http.StatusBadRequest)
		api.WriteErrorMessage(w, string(jsonErr))
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
			w.WriteHeader(http.StatusNotFound)
			api.WriteErrorMessage(w, "Not all segments with the specified names were found")
			return
		case errors.Is(err, membership.ErrSegmentAlreadyAssigned):
			w.WriteHeader(http.StatusBadRequest)
			api.WriteErrorMessage(w, "Attempt to add segments that the user already belongs to")
			return
		case errors.Is(err, user.ErrUserNotFound):
			w.WriteHeader(http.StatusNotFound)
			api.WriteErrorMessage(w, "Attempt to update the data of a non-existent user")
			return
		case errors.Is(err, membership.ErrEmptyData):
			w.WriteHeader(http.StatusBadRequest)
			api.WriteErrorMessage(w, "Data for update and delete cannot be empty at the same time")
			return
		case errors.Is(err, membership.ErrIncorrectData):
			w.WriteHeader(http.StatusBadRequest)
			api.WriteErrorMessage(w, "Attempt to add and remove the same segment")
			return
		case errors.Is(err, membership.ErrSegmentNotAssigned):
			w.WriteHeader(http.StatusBadRequest)
			api.WriteErrorMessage(w, "Attempt to delete a segment unassigned to the user")
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		api.WriteErrorMessage(w, "Update user segments")
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
			w.WriteHeader(http.StatusNotFound)
			api.WriteErrorMessage(w, "Segment with the specified name wasn't found")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		api.WriteErrorMessage(w, "Delete segment")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *handler) GetUserMembership(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		api.WriteErrorMessage(w, "Invalid user id parameter")
		return
	}

	data, err := h.membership.GetUserMembership(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.WriteErrorMessage(w, "Get user membership segment")
		return
	}

	if len(data) == 0 {
		w.WriteHeader(http.StatusNotFound)
		api.WriteErrorMessage(w, "No data was found for the specified user")
		return
	}

	response := make([]UserResponseInfo, len(data))
	for i, d := range data {
		response[i] = NewUserResponseInfo(d.UserID, d.SegmentName, d.ExpiredAt)
	}

	jsonResponse, err := json.Marshal(NewUserMembershipResponse(response))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.WriteErrorMessage(w, http.StatusText(http.StatusInternalServerError))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
