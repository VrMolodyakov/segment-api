package integrationtest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/apierror"
	membrDto "github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/membership"
)

func (s *TestSuite) TestSuccessfulGetUserSegments() {
	resp, err := s.server.Client().Get(s.server.URL + "/api/v1/users/1")
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	var response membrDto.GetUserMembershipResponse
	err = json.Unmarshal(bodyBytes, &response)
	expected := membrDto.GetUserMembershipResponse{
		Memberships: []membrDto.UserResponseInfo{{UserID: int64(1), SegmentName: "test_name"}},
	}
	s.Require().NoError(err)
	s.Require().Equal(expected.Memberships[0].UserID, response.Memberships[0].UserID)
	s.Require().Equal(expected.Memberships[0].SegmentName, response.Memberships[0].SegmentName)
	s.Require().True(response.Memberships[0].ExpiredAt.After(time.Now()))
	s.Require().Equal(200, resp.StatusCode)
}

func (s *TestSuite) TestUserNotFound() {
	resp, err := s.server.Client().Get(s.server.URL + "/api/v1/users/2")
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	var got apierror.ErrorResponse
	json.Unmarshal(bodyBytes, &got)
	s.Require().NoError(err)
	s.Require().Equal("No data was found for the specified user", got.Error())
	s.Require().Equal(404, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsAddNew() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_add.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(200, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsDelete() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_delete.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(200, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsAddAndDelete() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_add_delete.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(200, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsUserNotFount() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_not_found.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(404, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsSegmentNotFount() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_segment_not_found.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(404, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsAssignedSegment() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_assigned_segment.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(409, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsEmptyReq() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_empty.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(400, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsEqual() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_equal.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(400, resp.StatusCode)
}

func (s *TestSuite) TestUpdateUserSegmentsDeleteNotAssigned() {
	requestBody := s.loader.LoadString("fixtures/api/update_user_segments_delete_not_assigned.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/membership/update", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(409, resp.StatusCode)
}

func (s *TestSuite) TestSuccessfullyDeleteSegment() {
	req, err := http.NewRequest(http.MethodDelete, s.server.URL+"/api/v1/segments/test_name_6", nil)
	s.Require().NoError(err)
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(200, resp.StatusCode)
}

func (s *TestSuite) TestSuccessfullyDeleteSegmentNotFound() {
	req, err := http.NewRequest(http.MethodDelete, s.server.URL+"/api/v1/segments/test_name_7", nil)
	s.Require().NoError(err)
	resp, err := s.server.Client().Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	s.Require().NoError(err)
	s.Require().Equal(404, resp.StatusCode)
}
