package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//Mate holds the configuration and connection the backing service
type Mate struct {
	Config *Config
	Conn   *ServiceConn
	Logger *Logger
}

//NewMate loads configuration and connects to the backing service before returning a new instance of Mate
func NewMate() (*Mate, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("loading configuration failed: %v", err)
	}
	serviceConf := config.Service
	if err = config.General.validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %v", err)
	}
	if err = config.Service.Parameters.validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %v", err)
	}
	if err = config.Service.Input.validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %v", err)
	}
	if err = config.Service.Output.validate(); err != nil {
		return nil, fmt.Errorf("config validation error: %v", err)
	}
	conn, err := NewServiceConn(config.Service.Location)
	if err != nil {
		return nil, fmt.Errorf("connecting to service failed: %v", err)
	}
	logger, err := config.General.Logger.NewLogger(serviceConf.Title, serviceConf.Version)
	if err != nil {
		return nil, fmt.Errorf("logger creation failed: %v", err)
	}
	return &Mate{
		Config: config,
		Conn:   conn,
		Logger: logger,
	}, nil
}

//Error messages for http replies.
const (
	specRequiresGetMethodMessage       = "you must use the GET method with this endpoint"
	couldNotGenerateSpecMessage        = "could not generate a specification for this service"
	methodNotAllowedMessage            = "you must use the POST method with this endpoint"
	unsupportedMediaTypeMessage        = "you must set the content type header to application/json"
	errorEstablishingConnectionMessage = "failed to establish a client connection to the backing service. Try again later or contact the service author"
)

func (m *Mate) handleRoot(res http.ResponseWriter, req *http.Request) {
	m.Logger.Infoln("Request received")
	if req.Method != "POST" {
		writeHTTPError(res, m.Config.Service.Title, methodNotAllowedMessage, http.StatusMethodNotAllowed)
		m.Logger.Errorln("Root endpoint requires method: POST received:", req.Method)
		return
	}
	if req.Header.Get("Content-Type") != "application/json" {
		writeHTTPError(res, m.Config.Service.Title, unsupportedMediaTypeMessage, http.StatusUnsupportedMediaType)
		m.Logger.Errorln("Root endpoint requires content-type: application/json received:", req.Header.Get("Content-Type"))
		return
	}
	stream, err := m.Conn.NewServiceStream()
	if err != nil {
		writeHTTPError(res, m.Config.Service.Title, errorEstablishingConnectionMessage, http.StatusFailedDependency)
		m.Logger.Errorln("Could not create service stream error:", err)
		return
	}
	if err := m.parseCX(stream, req.URL.Query(), req.Body); err != nil {
		m.Logger.Errorln("Parser error:", err)
		writeHTTPError(res, m.Config.Service.Title, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(res, `{"data":`)
	detector := &WriteDetector{res, false}
	if err := m.generateCX(stream, detector); err != nil {
		m.Logger.Errorln("Generator error:", err)
		if !detector.Wrote {
			io.WriteString(res, `""`)
		}
		io.WriteString(res, `, "errors":[`)
		httpErr := NewHTTPError(m.Config.Service.Title, err.Error(), http.StatusInternalServerError)
		httpErr.toJSON(res)
		io.WriteString(res, `]}`)
	} else {
		io.WriteString(res, `, "errors":[]}`)
	}
}

func (m *Mate) parseCX(s *ServiceStream, p map[string][]string, r io.ReadCloser) error {
	m.Logger.Debugln("Parsing CX")
	send := make(chan *Message)
	s.OpenSend(send)
	if err := m.Config.Service.Parameters.send(send, p); err != nil {
		return err
	}
	if err := m.Config.Service.Input.parse(r, send, m.Config.Service.SingletonInput); err != nil {
		return err
	}
	close(send)
	return nil
}

func (m *Mate) generateCX(s *ServiceStream, w io.Writer) error {
	m.Logger.Debugln("Generating CX")
	receive := make(chan *Message)
	s.OpenReceive(receive)
	if err := m.Config.Service.Output.generate(w, receive, m.Config.Service.SingletonOutput); err != nil {
		return err
	}
	return nil
}

func main() {
	cxmate, err := NewMate()
	if err != nil {
		log.Fatalln("Initialization of cxMate failed with error:", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", cxmate.handleRoot)
	mux.Handle("/metrics", promhttp.Handler())
	cxmate.Logger.Infoln("cxMate now listening on", cxmate.Config.General.Location, "connected to service on", cxmate.Config.Service.Location)
	logFatalln(http.ListenAndServe(cxmate.Config.General.Location, mux))
}

//WriteDetector determines if a write has been made to a writer
type WriteDetector struct {
	http.ResponseWriter
	Wrote bool
}

//Write writes to the response writer
func (w *WriteDetector) Write(b []byte) (int, error) {
	w.Wrote = true
	return w.ResponseWriter.Write(b)
}

//WriteHeader writes to the response header
func (w *WriteDetector) WriteHeader(code int) {
	w.Wrote = true
	w.ResponseWriter.WriteHeader(code)
}
