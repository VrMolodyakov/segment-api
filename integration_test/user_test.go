package integrationtest

import (
	"bytes"
	"encoding/json"
	"io"

	membrDto "github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/membership"
)

func (s *TestSuite) TestSuccessCreateUser() {
	requestBody := s.loader.LoadString("fixtures/api/create_user.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/users", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	var response membrDto.CreateUserResponse
	err = json.Unmarshal(bodyBytes, &response)
	s.Require().NoError(err)
	s.Require().True(response.ID > 0)
	s.Require().Equal("example@example2.com", response.Email)
	s.Require().Equal(201, resp.StatusCode)
}

func (s *TestSuite) TestUserAlreadyExists() {
	requestBody := s.loader.LoadString("fixtures/api/create_user_already_exists.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/users", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Require().Equal("User already exists\n", string(bodyBytes))
	s.Require().Equal(400, resp.StatusCode)
}
