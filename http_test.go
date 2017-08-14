package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func executeRequest(req *http.Request) *http.Response {

	rr := httptest.NewRecorder()
	logger := &logrus.Logger{Out: ioutil.Discard}
	entry := logrus.NewEntry(logger)
	m := &Mate{
		Config: &Config{
			Service: ServiceConfig{
				Title: "test",
			},
		},
		Logger: &Logger{log: entry},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", m.handleRoot)
	mux.ServeHTTP(rr, req)
	return rr.Result()
}

func decodeResponse(req *http.Request) (HTTPResponse, *http.Response) {
	response := executeRequest(req)
	defer response.Body.Close()
	dec := json.NewDecoder(response.Body)
	var rb HTTPResponse
	dec.Decode(&rb)
	return rb, response
}

func TestIncorrectHTTPMethod(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	r, res := decodeResponse(req)
	if res.StatusCode != 405 {
		t.Errorf("Should return a 405 status code, returned %d", res.StatusCode)
	}
	if len(r.Errors) != 1 {
		t.Fatal("Should return a single error to the client.")
	}
	if r.Errors[0].Status != http.StatusMethodNotAllowed {
		t.Errorf("Should return an error with StatusMethodNotAllowed.")
	}
}

func TestIncorrectMIMEType(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "plain/text")
	r, res := decodeResponse(req)
	if res.StatusCode != 415 {
		t.Errorf("Should return a 415 status code, returned %d", res.StatusCode)
	}
	if len(r.Errors) != 1 {
		t.Fatal("Should return a single error to the client.")
	}
	if r.Errors[0].Status != http.StatusUnsupportedMediaType {
		t.Error("Should return an error with StatusUnsupportedMediaType.")
	}
}

/* WIll ONLY RUN IF SERVICE IS CONNECTED
func TestPassMiddlewareChecks(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	r, res := decodeResponse(req)
	if res.StatusCode != 200 {
		t.Errorf("Should return a 200 status code, returned %d", res.StatusCode)
	}
	if len(r.Errors) != 0 {
		t.Errorf("Should not return any errors.")
	}
}
*/
