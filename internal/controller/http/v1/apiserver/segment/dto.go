package segment

type CreateSegmentRequest struct {
	Name          string `json:"name" validate:"required,min=6"`
	HitPercentage int    `json:"hitPercentage" validate:"gt=-1"`
}

type CreateSegmentResponse struct {
	ID         int64  `json:"segmentID"`
	Name       string `json:"name"`
	Percentage int    `json:"hitPercentage"`
}

func NewSegmentResponse(id int64, name string, percentage int) CreateSegmentResponse {
	return CreateSegmentResponse{
		ID:         id,
		Name:       name,
		Percentage: percentage,
	}
}
