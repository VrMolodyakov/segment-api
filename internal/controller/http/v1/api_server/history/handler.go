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

func (h *handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	var linkRequest CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&linkRequest); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %s", err.Error()), http.StatusBadRequest)
		return
	}

	errs := validator.Validate(linkRequest)
	if errs != nil {
		jsonErr, _ := json.Marshal(errs)
		http.Error(w, string(jsonErr), http.StatusBadRequest)
		return
	}

	err := h.history.PrepareHistoryData(r.Context(), linkRequest.ToModel())
	if err != nil {
		switch {
		case errors.Is(err, history.ErrIncorrectYear):
			http.Error(w, fmt.Sprintf("Incorrect date, %s", err.Error()), http.StatusBadRequest)
			return
		case errors.Is(err, history.ErrIncorrectMonth):
			http.Error(w, fmt.Sprintf("Incorrect date, %s", err.Error()), http.StatusBadRequest)
			return
		}
		http.Error(w, "Couldn't prepare history data", http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

func (h *handler) DownloadCSVData(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	monthStr := chi.URLParam(r, "month")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year parameter", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		http.Error(w, "Invalid month parameter", http.StatusBadRequest)
		return
	}

	date := history.NewDate(year, month)
	data, err := h.history.GetUsersHistory(r.Context(), date)
	if err != nil {
		switch {
		case errors.Is(err, history.ErrIncorrectYear):
			http.Error(w, fmt.Sprintf("Incorrect date, %s", err.Error()), http.StatusBadRequest)
			return
		case errors.Is(err, history.ErrIncorrectMonth):
			http.Error(w, fmt.Sprintf("Incorrect date, %s", err.Error()), http.StatusBadRequest)
			return
		case errors.Is(err, history.ErrExpiredData):
			http.Error(w, "Data lifetime for the link has expired, create a new one", http.StatusNotFound)
			return
		}
	}

	buffer := h.pool.Get()
	defer h.pool.Release(buffer)

	if err := h.writer.Write(buffer, data); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't create a csv file, %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=history-for-%s.csv", date.ToString()))

	_, err = io.Copy(w, buffer)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

}
