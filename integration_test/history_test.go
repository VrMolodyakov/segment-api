package integrationtest

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/apierror"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/history"
)

func (s *TestSuite) TestCreateLink() {
	requestBody := s.loader.LoadString("fixtures/api/create_link.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/history/link", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	var response history.CreateLinkResponse
	err = json.Unmarshal(bodyBytes, &response)
	s.Require().NoError(err)
	s.Require().Equal(200, resp.StatusCode)
	s.Require().Equal("http://localhost:8081/api/v1/history/download/2023/8", response.Link)
}

func (s *TestSuite) TestCreateLinkWrongYear() {
	requestBody := s.loader.LoadString("fixtures/api/create_link_wrong_year.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/history/link", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Require().Equal(400, resp.StatusCode)
	var got apierror.ErrorResponse
	json.Unmarshal(bodyBytes, &got)
	s.Require().Equal("Incorrect date, history for dates before 2007 year is not available", got.Error())
}

func (s *TestSuite) TestCreateLinkWrongMonth() {
	requestBody := s.loader.LoadString("fixtures/api/create_link_wrong_month.json")
	resp, err := s.server.Client().Post(s.server.URL+"/api/v1/history/link", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Require().Equal(400, resp.StatusCode)
	var got apierror.ErrorResponse
	json.Unmarshal(bodyBytes, &got)
	s.Require().Equal("Incorrect date, impossible to get information for a month that has not yet come", got.Error())
}

func (s *TestSuite) TestDownload() {
	requestBody := s.loader.LoadString("fixtures/api/create_link.json")
	_, err := s.server.Client().Post(s.server.URL+"/api/v1/history/link", "", bytes.NewBufferString(requestBody))
	s.Require().NoError(err)
	resp, err := s.server.Client().Get(s.server.URL + "/api/v1/history/download/2023/8")
	s.Require().NoError(err)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Require().Equal("attachment; filename=history-for-2023-8.csv", resp.Header.Get("Content-Disposition"))
	expectedCSV := []byte("ID,UserID,Segment,Operation,Time\n3,test_name_3,added,2023-08-31 03:00:00\n3,test_name_4,added,2023-08-31 03:00:00\n")
	s.Require().Equal(string(expectedCSV), string(bodyBytes))
}
