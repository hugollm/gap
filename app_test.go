package gap

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApi(t *testing.T) {

	app := New()
	app.Route("GET", "/profiles/read", readProfile)

	t.Run("implements http handler", func(t *testing.T) {
		httptest.NewServer(app)
	})

	t.Run("invalid route responds not found", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/hello", nil)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		if response.Result().StatusCode != 404 {
			t.Errorf("failed to set status code to 404")
		}
		body, _ := ioutil.ReadAll(response.Result().Body)
		out := map[string]string{}
		json.Unmarshal(body, &out)
		if out["error"] != "not found" {
			t.Errorf("failed to set not found error on json body")
		}
	})

	t.Run("valid route with wrong method responds method not allowed", func(t *testing.T) {
		request := httptest.NewRequest("POST", "/profiles/read", nil)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		if response.Result().StatusCode != 405 {
			t.Errorf("failed to set status code to 405")
		}
		body, _ := ioutil.ReadAll(response.Result().Body)
		out := map[string]string{}
		json.Unmarshal(body, &out)
		if out["error"] != "method not allowed" {
			t.Errorf("failed to set method not allowed error on json body")
		}
	})

	t.Run("invalid json is answered with bad request", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/profiles/read", nil)
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		if response.Result().StatusCode != 400 {
			t.Errorf("failed to set status code to 400")
		}
		body, _ := ioutil.ReadAll(response.Result().Body)
		out := map[string]string{}
		json.Unmarshal(body, &out)
		if out["error"] != "invalid json" {
			t.Errorf("failed to set invalid json error on body")
		}
	})

	t.Run("valid route responds with output", func(t *testing.T) {
		requestBody := strings.NewReader(`{"profile_id": 1}`)
		request := httptest.NewRequest("GET", "/profiles/read", requestBody)
		request.Header.Set("auth", "api-token")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		if response.Result().StatusCode != 200 {
			t.Errorf("failed to set status code to 200")
		}
		body, _ := ioutil.ReadAll(response.Result().Body)
		out := readProfileOutput{}
		json.Unmarshal(body, &out)
		if out.Id != 1 || out.Email != "johndoe@example.org" {
			t.Errorf("failed to set json body")
		}
	})
}

type readProfileInput struct {
	Auth      string `header:"auth"`
	ProfileId int    `json:"profile_id"`
}

type readProfileOutput struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

func readProfile(input readProfileInput) (readProfileOutput, error) {
	if input.Auth != "api-token" {
		return readProfileOutput{}, errors.New("auth failed")
	}
	return readProfileOutput{1, "johndoe@example.org"}, nil
}
