package errors

import (
	"encoding/json"
	"io"
)

type ErrorResponse struct {
	Ok      bool   `json:"ok" default:"false"`
	Message string `json:"message"`
}

func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Message: message,
	}
}

func (e ErrorResponse) Error() string {
	return e.Message
}

func WriteErrorMessage(writer io.Writer, message string) (int, error) {
	buf, err := json.Marshal(ErrorResponse{
		Message: message,
	})

	if err != nil {
		return writer.Write([]byte(message))
	}

	return writer.Write(buf)
}
