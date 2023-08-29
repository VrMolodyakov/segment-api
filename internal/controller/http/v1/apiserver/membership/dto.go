package membership

import (
	"time"

	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/internal/domain/user"
)

var (
	location, _ = time.LoadLocation("Europe/Moscow")
)

type CreateUserRequest struct {
	FirstName string `json:"firsName" validate:"required,min=3"`
	LastName  string `json:"lastName" validate:"required,min=3"`
	Email     string `json:"email" validate:"required,min=5"`
}

type CreateUserResponse struct {
	ID        int64  `json:"userID"`
	FirstName string `json:"firsName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type UpdateUserRequest struct {
	UserID int64           `json:"userID" validate:"required,gt=0"`
	Update []UpdateSegment `json:"update" validate:"dive"`
	Delete []string        `json:"delete"`
}

type UpdateSegment struct {
	Name string `json:"name" validate:"required,min=6"`
	TTL  int    `json:"ttl"`
}

type DeleteSegmentRequest struct {
	Name string `json:"name" validate:"required,min=6"`
}

type GetUserMembershipResponse struct {
	Memberships []UserResponseInfo `json:"memberships"`
}

type UserResponseInfo struct {
	UserID      int64     `json:"userID"`
	SegmentName string    `json:"segmentName"`
	ExpiredAt   time.Time `json:"expiredAt"`
}

func (u UpdateSegment) ToModel() segment.Segment {
	var expired time.Time
	if u.TTL > 0 {
		expired = time.Now().Add(time.Second * time.Duration(u.TTL))
	}
	return segment.Segment{
		Name:      u.Name,
		ExpiredAt: expired,
	}
}

func NewUserMembershipResponse(info []UserResponseInfo) GetUserMembershipResponse {
	return GetUserMembershipResponse{
		Memberships: info,
	}
}

func NewUserResponseInfo(id int64, segment string, expiredAt time.Time) UserResponseInfo {
	return UserResponseInfo{
		UserID:      id,
		SegmentName: segment,
		ExpiredAt:   expiredAt.In(location),
	}
}

func (u *UpdateUserRequest) GetUpdatedSegments() []segment.Segment {
	segments := make([]segment.Segment, len(u.Update))
	for i := range segments {
		segments[i] = u.Update[i].ToModel()
	}
	return segments
}

func (u *UpdateUserRequest) GetDeletedSegments() []string {
	return u.Delete
}

func (c CreateUserRequest) ToModel() user.User {
	return user.User{
		FirstName: c.FirstName,
		LastName:  c.LastName,
		Email:     c.Email,
	}
}

func NewCreateUserResponse(id int64, firstName string, lastName string, email string) CreateUserResponse {
	return CreateUserResponse{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
}
