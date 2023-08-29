package history

import "github.com/VrMolodyakov/segment-api/internal/domain/history"

type CreateLinkRequest struct {
	Year  int `json:"year" validate:"required,gt=-1"`
	Month int `json:"month" validate:"required,gt=-1,lt=13"`
}

type CreateLinkResponse struct {
	Link string `json:"string"`
}

func NewCreateLinkResponse(link string) CreateLinkResponse {
	return CreateLinkResponse{
		Link: link,
	}
}

func (c CreateLinkRequest) ToModel() history.Date {
	return history.Date{
		Year:  c.Year,
		Month: c.Month,
	}
}
