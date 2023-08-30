package history

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/apierror"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/validator"
	"github.com/VrMolodyakov/segment-api/internal/domain/history"
	"github.com/go-chi/chi/v5"
)

type HistoryService interface {
	GetUsersHistory(ctx context.Context, date history.Date) ([]history.History, error)
	PrepareHistoryData(ctx context.Context, date history.Date) error
}

type BufferPool interface {
	Get() *bytes.Buffer
	Release(buf *bytes.Buffer)
}

type CSVWriter interface {
	Write(w io.Writer, args []history.History) error
}

type LinkParam struct {
	Host string
	Port int
}

func NewLinkParam(host string, port int) LinkParam {
	return LinkParam{
		Host: host,
		Port: port,
	}
}

type handler struct {
	parameters LinkParam
	writer     CSVWriter
	pool       BufferPool
	history    HistoryService
}

func New(history HistoryService, parameters LinkParam, pool BufferPool, writer CSVWriter) *handler {
	return &handler{
		parameters: parameters,
		pool:       pool,
		writer:     writer,
		history:    history,
	}
}

// @Summary Create new download link
// @Description Create new download link
// @Tags History
// @Accept json
// @Produce json
// @Param linkRequest body CreateLinkRequest true "Creaet link request"
// @Success 200 {object} CreateLinkResponse "Create link successfully"
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /history/link [post]
func (h *handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var linkRequest CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&linkRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, fmt.Sprintf("invalid request: %s", err.Error()))
		return
	}

	errs := validator.Validate(linkRequest)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, string(jsonErr))
		return
	}

	err := h.history.PrepareHistoryData(r.Context(), linkRequest.ToModel())
	if err != nil {
		switch {
		case errors.Is(err, history.ErrIncorrectYear):
			w.WriteHeader(http.StatusBadRequest)
			apierror.WriteErrorMessage(w, fmt.Sprintf("Incorrect date, %s", err.Error()))
			return
		case errors.Is(err, history.ErrIncorrectMonth):
			w.WriteHeader(http.StatusBadRequest)
			apierror.WriteErrorMessage(w, fmt.Sprintf("Incorrect date, %s", err.Error()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, "Couldn't prepare history data")
		return
	}

	link := fmt.Sprintf(
		"http://%s:%d/api/v1/history/download/%d/%d",
		h.parameters.Host,
		h.parameters.Port,
		linkRequest.Year,
		linkRequest.Month,
	)

	jsonResponse, err := json.Marshal(NewCreateLinkResponse(link))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, http.StatusText(http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

// @Summary Download history
// @Description Download history
// @Tags History
// @Accept json
// @Param  year   path int  true "Year"
// @Param  month  path int  true "Month"
// @Produce application/csv
// @Success 200
// @Failure 400 {object} apierror.ErrorResponse
// @Failure 404 {object} apierror.ErrorResponse
// @Failure 409 {object} apierror.ErrorResponse
// @Failure 500 {object} apierror.ErrorResponse
// @Router /history/download/{year}/{month} [get]
func (h *handler) DownloadCSVData(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	monthStr := chi.URLParam(r, "month")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, "Invalid year parameter")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apierror.WriteErrorMessage(w, "Invalid month parameter")
		return
	}

	date := history.NewDate(year, month)
	data, err := h.history.GetUsersHistory(r.Context(), date)
	if err != nil {
		switch {
		case errors.Is(err, history.ErrIncorrectYear):
			w.WriteHeader(http.StatusBadRequest)
			apierror.WriteErrorMessage(w, fmt.Sprintf("Incorrect date, %s", err.Error()))
			return
		case errors.Is(err, history.ErrIncorrectMonth):
			w.WriteHeader(http.StatusBadRequest)
			apierror.WriteErrorMessage(w, fmt.Sprintf("Incorrect date, %s", err.Error()))
			return
		case errors.Is(err, history.ErrExpiredData):
			w.WriteHeader(http.StatusNotFound)
			apierror.WriteErrorMessage(w, "Data lifetime for the link has expired, create a new one")
			return
		}
	}

	if len(data) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	buffer := h.pool.Get()
	defer h.pool.Release(buffer)

	if err := h.writer.Write(buffer, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, fmt.Sprintf("Couldn't create a csv file, %s", err.Error()))
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=history-for-%s.csv", date.ToString()))

	_, err = io.Copy(w, buffer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		apierror.WriteErrorMessage(w, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

}
