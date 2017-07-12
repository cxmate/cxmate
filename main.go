package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
	serviceConf := config.Service
	if err != nil {
		return nil, fmt.Errorf("loading configuration failed: %v", err)
	}
	if err = config.Service.Input.validate(); err != nil {
		return nil, fmt.Errorf("input validation error, err")
	}
	if err = config.Service.Output.validate(); err != nil {
		return nil, fmt.Errorf("output validation error:", err)
	}
	conn, err := NewServiceConn(config.Service.Location)
	if err != nil {
		return nil, fmt.Errorf("connecting to service failed: %v", err)
	}
	logger, err := config.General.Logger.NewLogger(serviceConf.Name, serviceConf.Version)
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
		writeHTTPError(res, m.Config.Service.Name, methodNotAllowedMessage, http.StatusMethodNotAllowed)
		m.Logger.Errorln("Root endpoint requires method: POST received:", req.Method)
		return
	}
	if req.Header.Get("Content-Type") != "application/json" {
		writeHTTPError(res, m.Config.Service.Name, unsupportedMediaTypeMessage, http.StatusUnsupportedMediaType)
		m.Logger.Errorln("Root endpoint requires content-type: application/json received:", req.Header.Get("Content-Type"))
		return
	}
	stream, err := m.Conn.NewServiceStream()
	if err != nil {
		writeHTTPError(res, m.Config.Service.Name, errorEstablishingConnectionMessage, http.StatusFailedDependency)
		m.Logger.Errorln("Could not create service stream error:", err)
		return
	}
	err = m.processCX(stream, req.URL.Query(), req.Body, res)
	if err != nil {
		writeHTTPError(res, m.Config.Service.Name, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *Mate) processCX(s *ServiceStream, p map[string][]string, r io.ReadCloser, w io.Writer) error {
	if err := m.parseCX(s, p, r); err != nil {
		m.Logger.Errorln("Parser error:", err)
		return err
	}
	if err := m.generateCX(s, w); err != nil {
		m.Logger.Errorln("Generator error:", err)
		return err
	}
	return nil
}

func (m *Mate) parseCX(s *ServiceStream, p map[string][]string, r io.ReadCloser) error {
	m.Logger.Debugln("Parsing CX")
	send := make(chan *Message)
	s.OpenSend(send)
	if err := m.Config.Service.Parameters.send(send, p); err != nil {
		return err
	}
	if err := m.Config.Service.Input.parse(r, send); err != nil {
		return err
	}
	close(send)
	return nil
}

func (m *Mate) generateCX(s *ServiceStream, w io.Writer) error {
	m.Logger.Debugln("Generating CX")
	receive := make(chan *Message)
	s.OpenReceive(receive)
	if err := m.Config.Service.Output.generate(w, receive); err != nil {
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
	cxmate.Logger.Infoln("cxMate now listening on", cxmate.Config.General.Location, "connected to service on", cxmate.Config.Service.Location)
	logFatalln(http.ListenAndServe(cxmate.Config.General.Location, mux))
}
