package membership

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/apierror"
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

// @Summary Create new user
// @Description Create user
// @Tags Users
// @Accept json
// @Produce json
// @Param userReq body CreateUserRequest true "Create user request"
// @Success 201 {object} CreateUserResponse "Create user response"
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 409 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /users [post]
func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userReq CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, fmt.Sprintf("invalid request: %s", err.Error()))
		return
	}

	errs := validator.Validate(userReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, string(jsonErr))
		return
	}

	newUser := userReq.ToModel()
	id, err := h.membership.CreateUser(r.Context(), newUser)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			w.WriteHeader(http.StatusBadRequest)
			apierror.WriteErrorMessage(w, fmt.Sprintf("Invalid email: %s", err.Error()))
			return
		case errors.Is(err, user.ErrUserAlreadyExist):
			w.WriteHeader(http.StatusConflict)
			apierror.WriteErrorMessage(w, "User already exists")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, "Create user error")
		return
	}

	jsonResponse, err := json.Marshal(NewCreateUserResponse(id, newUser.FirstName, newUser.LastName, newUser.Email))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, http.StatusText(http.StatusInternalServerError))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

// @Summary Update user segments
// @Description Update user segments
// @Tags Membership
// @Accept json
// @Produce json
// @Param updateReq body UpdateUserRequest true "Update request"
// @Success 200
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 404 {object} apierror.ErrorResponse
// @Failure 409 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /membership/update [post]
func (h *handler) UpdateUserMembership(w http.ResponseWriter, r *http.Request) {
	var updateReq UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, fmt.Sprintf("Invalid request: %s", err.Error()))
		return
	}

	errs := validator.Validate(updateReq)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, string(jsonErr))
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
			apierror.WriteErrorMessage(w, "Not all segments with the specified names were found or adding/removing one segment multiple times")
			return
		case errors.Is(err, membership.ErrSegmentAlreadyAssigned):
			w.WriteHeader(http.StatusConflict)
			apierror.WriteErrorMessage(w, "Attempt to add segments that the user already belongs to")
			return
		case errors.Is(err, user.ErrUserNotFound):
			w.WriteHeader(http.StatusNotFound)
			apierror.WriteErrorMessage(w, "Attempt to update the data of a non-existent user")
			return
		case errors.Is(err, membership.ErrEmptyData):
			w.WriteHeader(http.StatusBadRequest)
			apierror.WriteErrorMessage(w, "Data for update and delete cannot be empty at the same time")
			return
		case errors.Is(err, membership.ErrIncorrectData):
			w.WriteHeader(http.StatusBadRequest)
			apierror.WriteErrorMessage(w, "Attempt to add and remove the same segment")
			return
		case errors.Is(err, membership.ErrSegmentNotAssigned):
			w.WriteHeader(http.StatusConflict)
			apierror.WriteErrorMessage(w, "Attempt to delete a segment unassigned to the user")
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, "Update user segments")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// @Summary Delete segment
// @Description Delete segment
// @Tags Segments
// @Accept json
// @Produce json
// @Param  segmentName   path string  true "Segment name"
// @Success 200
// @Failure 404 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /segments/{segmentName} [delete]
func (h *handler) DeleteMembership(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "segmentName")
	err := h.membership.DeleteMembership(r.Context(), name)
	if err != nil {
		switch {
		case errors.Is(err, segment.ErrSegmentNotFound):
			w.WriteHeader(http.StatusNotFound)
			apierror.WriteErrorMessage(w, "Segment with the specified name wasn't found")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, "Delete segment")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// @Summary Get user segments
// @Description Get user segments
// @Tags Users
// @Accept json
// @Produce json
// @Param  userID   path int  true "User id"
// @Success 200 {object} GetUserMembershipResponse "User segment info"
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 404 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /users/{userID} [get]
func (h *handler) GetUserMembership(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, "Invalid user id parameter")
		return
	}
	data, err := h.membership.GetUserMembership(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, "Get user membership segment")
		return
	}

	if len(data) == 0 {
		w.WriteHeader(http.StatusNotFound)
		apierror.WriteErrorMessage(w, "No data was found for the specified user")
		return
	}

	response := make([]UserResponseInfo, len(data))
	for i, d := range data {
		response[i] = NewUserResponseInfo(d.UserID, d.SegmentName, d.ExpiredAt)
	}

	jsonResponse, err := json.Marshal(NewUserMembershipResponse(response))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, http.StatusText(http.StatusInternalServerError))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
