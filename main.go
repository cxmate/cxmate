package main

import (
	"fmt"
	"io"
	"net/http"
)

//Error messages for http replies.
const (
	specRequiresGetMethodMessage       = "you must use the GET method with this endpoint"
	couldNotGenerateSpecMessage        = "could not generate a specification for this service"
	methodNotAllowedMessage            = "you must use the POST method with this endpoint"
	unsupportedMediaTypeMessage        = "you must set the content type header to application/json"
	errorEstablishingConnectionMessage = "failed to establish a client connection to the backing service. Try again later or contact the service author"
)

//Mate holds the configuration and connection the backing service
type Mate struct {
	Config *Config
	Conn   *ServiceConn
}

//NewMate loads configuration and connects to the backing service before returning a new instance of Mate
func NewMate() (*Mate, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("could not load configuration file, error: %v", err)
	}
	configureLogger(config.General.Logger)
	if err = config.Service.Input.validate(); err != nil {
		return nil, err
	}
	if err = config.Service.Output.validate(); err != nil {
		return nil, err
	}
	conn, err := NewServiceConn(config.Service.Location)
	if err != nil {
		return nil, fmt.Errorf("could not connect to the service error: %v", err)
	}
	return &Mate{
		Config: config,
		Conn:   conn,
	}, nil
}

func (m *Mate) handleRoot(res http.ResponseWriter, req *http.Request) {
	logDebugln("Request recieved")
	if req.Method != "POST" {
		writeHTTPError(res, m.Config.Service.Name, methodNotAllowedMessage, http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Content-Type") != "application/json" {
		writeHTTPError(res, m.Config.Service.Name, unsupportedMediaTypeMessage, http.StatusUnsupportedMediaType)
		return
	}
	stream, err := m.Conn.NewServiceStream()
	if err != nil {
		writeHTTPError(res, m.Config.Service.Name, errorEstablishingConnectionMessage, http.StatusFailedDependency)
		return
	}
	err = m.processCX(stream, req.URL.Query(), req.Body, res)
	if err != nil {
		writeHTTPError(res, m.Config.Service.Name, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *Mate) processCX(s *ServiceStream, p map[string][]string, r io.ReadCloser, w io.Writer) error {
	errChan := make(chan error, 1)
	go func() {
		if err := m.decodeRequestBody(s, p, r); err != nil {
			logDebug("Processing the request returned an error:", err)
			errChan <- err
		} else {
			errChan <- nil
		}
	}()
	go func() {
		if err := m.encodeResponseBody(s, w); err != nil {
			logDebug("Generating the response returned an error:", err)
			errChan <- err
		} else {
			errChan <- nil
		}
	}()
	if err := <-errChan; err != nil {
		logDebug("First routine returned an error")
		return err
	}
	logDebug("First routine finished")
	if err := <-errChan; err != nil {
		logDebug("Second routine returned an error")
		return err
	}
	logDebug("Second routine finished")
	return nil
}

func (m *Mate) decodeRequestBody(s *ServiceStream, p map[string][]string, r io.ReadCloser) error {
	send := make(chan *Message)
	s.OpenSend(send)
	if err := m.processParameters(send, p); err != nil {
		return err
	}
	if err := m.Config.Service.Input.parse(r, send); err != nil {
		return err
	}
	close(send)
	return nil
}

func (m *Mate) encodeResponseBody(s *ServiceStream, w io.Writer) error {
	receive := make(chan *Message)
	s.OpenReceive(receive)
	if err := m.Config.Service.Output.generate(w, receive); err != nil {
		return err
	}
	return nil
}

func newMateMux(cxmate *Mate) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", cxmate.handleRoot)
	return mux
}

func main() {
	logDebug("cxmate starting")
	cxmate, err := NewMate()
	if err != nil {
		logFatalln(err)
	}
	mux := newMateMux(cxmate)
	logFatalln(http.ListenAndServe(cxmate.Config.General.Location, mux))
}
