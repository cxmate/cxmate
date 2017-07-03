package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
func NewMate() *Mate {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Could not load configuration file, error:", err)
		os.Exit(0)
	}
	configureLogger(config.General.Logger)
	conn, err := NewServiceConn(config.Service.Location)
	if err != nil {
		logFatalln("Could not connect to the service,", err)
	}
	return &Mate{
		Config: config,
		Conn:   conn,
	}
}

func (m *Mate) handleRoot(res http.ResponseWriter, req *http.Request) {
	logDebugln("request recieved")
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
	done := make(chan bool)
	go func() {
		if err := m.decodeRequestBody(s, p, r); err != nil {
			logDebug(err)
			errChan <- err
			logDebug("error sent to channel")
			close(done)
		} else {
			done <- true
		}
	}()
	go func() {
		if err := m.encodeResponseBody(s, w); err != nil {
			errChan <- err
			close(done)
		} else {
			done <- true
		}
	}()
	if _, ok := <-done; !ok {
		err := <-errChan
		close(errChan)
		return err
	}
	if _, ok := <-done; !ok {
		err := <-errChan
		close(errChan)
		return err
	}
	return nil
}

func (m *Mate) decodeRequestBody(s *ServiceStream, p map[string][]string, r io.ReadCloser) error {
	send := make(chan *Message)
	s.OpenSend(send)
	if err := m.processParameters(send, p); err != nil {
		return err
	}
	dec, err := NewDecoder(r, send, m.Config.Service.Input[0])
	if err != nil {
		return err
	}
	if err := dec.DecodeNetwork(); err != nil {
		return err
	}
	close(send)
	return nil
}

func (m *Mate) encodeResponseBody(s *ServiceStream, w io.Writer) error {
	receive := make(chan *Message)
	s.OpenReceive(receive)
	enc, err := NewEncoder(w, receive, m.Config.Service.Output[0])
	if err != nil {
		return err
	}
	if err := enc.EncodeNetwork(); err != nil {
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
	cxmate := NewMate()
	mux := newMateMux(cxmate)
	logFatal(http.ListenAndServe(cxmate.Config.General.Location, mux))
}
