package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

//HTTPResponse A rest response containing
//either a data payload or list of errors
type HTTPResponse struct {
	Data   interface{}  `json:"data"`
	Errors []*HTTPError `json:"errors"`
}

//HTTPError An error message
type HTTPError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Link    string `json:"link"`
	Status  int    `json:"status"`
}

//NewHTTPError creates a new Error linking back to cxmate.
//Use this to transmit Errors back in a Response to the client.
func NewHTTPError(service string, message string, status int) *HTTPError {
	return &HTTPError{
		Type:    fmt.Sprint("urn:cytoscape:ci:cxmate:", service, ":", status),
		Message: message,
		Link:    "http://github.com/ericsage/cxmate",
		Status:  status,
	}
}

//NewHTTPResponse create a New Response for transmitting back to the client.
func NewHTTPResponse(data interface{}, errors []*HTTPError) *HTTPResponse {
	return &HTTPResponse{
		Data:   data,
		Errors: errors,
	}
}

//Encode writes a response to a provided writer as JSON, usually the writter is a http.ResponseWriter.
func (r *HTTPResponse) toJSON(w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(r)
}

//writeHTTPError is a convience function for writing a single error message back to an http client.
func writeHTTPError(res http.ResponseWriter, service string, message string, httpStatus int) {
	logDebugln("writing http error into an http response writer")
	e := NewHTTPError(service, message, httpStatus)
	r := NewHTTPResponse("", []*HTTPError{e})
	res.WriteHeader(httpStatus)
	r.toJSON(res)
}

//writeHTTPResponse is a convience function for writing a simple message back to an http client with status 200 OK.
func writeHTTPResponse(res http.ResponseWriter, message string) {
	logDebugln("writing http request into an http response writer")
	r := NewHTTPResponse(message, []*HTTPError{})
	res.WriteHeader(http.StatusOK)
	r.toJSON(res)
}
