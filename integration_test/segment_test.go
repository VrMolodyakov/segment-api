package integrationtest

import (
	"bytes"
	"encoding/json"
	"io"

	api "github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/errors"
	segmentDto "github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/segment"
)

func (s *TestSuite) TestSuccessCreateSegment() {
	requestBody := s.loader.LoadString("fixtures/api/create_segment.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/segments", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	var response segmentDto.CreateSegmentResponse
	err = json.Unmarshal(bodyBytes, &response)
	s.Require().NoError(err)
	s.Require().True(response.ID > 0)
	s.Require().Equal("test_name_segment", response.Name)
	s.Require().Equal(201, resp.StatusCode)
}

func (s *TestSuite) TestSegmentAlreadyExists() {
	requestBody := s.loader.LoadString("fixtures/api/create_segment_already_exists.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/segments", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	var got api.ErrorResponse
	json.Unmarshal(bodyBytes, &got)
	s.Require().NoError(err)
	s.Require().Equal("Segment already exists", got.Error())
	s.Require().Equal(400, resp.StatusCode)
}
