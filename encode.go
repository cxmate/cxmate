package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ericsage/cxmate/proto"
	"github.com/golang/protobuf/jsonpb"
)

// Encoder represents an encoder for a single CX network
type Encoder struct {
	w        io.Writer
	config   EncoderConfig
	stream   *elementQueue
	brackets *bracketStack
}

// EncoderConfig configures the details of networks being emitted by an algorithm
type EncoderConfig struct {
	// Label is a descriptor for the network that is used to identify it in a stream
	Label string `json:"label"`
	// Description is a short description of the network, what it represents
	Description string `json:"description"`
	// Aspects should contain a list of strings, the required aspects for this network
	Aspects []string `json:"aspects"`
}

// NewEncoder validates an Encoder config and returns a new encoder or an error
func NewEncoder(w io.Writer, s <-chan *Message, c EncoderConfig) (*Encoder, error) {
	logDebugln("Creating a new decoder")
	if c.Label == "" {
		return nil, errors.New("invalid config, label missing")
	}
	if len(c.Aspects) == 0 {
		return nil, errors.New("invalid config, aspects must not be empty")
	}
	q, err := newElementQueue(s)
	if err != nil {
		return nil, err
	}
	return &Encoder{
		w:        w,
		config:   c,
		stream:   q,
		brackets: &bracketStack{},
	}, nil
}

// EncodeNetwork writes out a network
func (e *Encoder) EncodeNetwork() error {
	defer e.emitRemainingBrackets()
	logDebugln("Generating network", e.config.Label, "with aspects", e.config.Aspects)
	e.emitOpenBrackets("[")
	if err := e.emitNumberVerfication(); err != nil {
		return err
	}
	if err := e.emitMetadata(); err != nil {
		return err
	}
	for i := 0; e.stream.hasNext(); i++ {
		name := e.stream.peekType()
		e.emitAspect(name)
	}
	e.emitCloseBrackets("]")
	return nil
}

// EncodeNumberVerfication writes out the standard number verification stanza
func (e *Encoder) emitNumberVerfication() error {
	logDebugln("Generating number verification")
	l := LongNumber{
		LongNumber: supportedLongNumber,
	}
	nv := map[string]LongNumber{"numberVerification": l}
	s, err := e.encode(nv)
	if err != nil {
		return err
	}
	e.emit(s)
	return nil
}

// EncodeMetadata writes out the standard metadata
func (e *Encoder) emitMetadata() error {
	logDebugln("Generating pre-metadata")
	e.emit(",")
	e.emitOpenBrackets("{")
	e.emit("\"metaData\":")
	e.emitOpenBrackets("[")
	for i, name := range e.config.Aspects {
		if i != 0 {
			e.emit(",")
		}
		s, err := e.encode(Metadata{
			Name: name,
		})
		if err != nil {
			return err
		}
		e.emit(s)
	}
	e.emitCloseBrackets("]}")
	return nil
}

func (e *Encoder) emitAspect(name string) error {
	logDebugln("Generating", name, "aspect")
	e.emit(",")
	if err := e.emitOpenBrackets("{"); err != nil {
		return err
	}
	e.emit("\"" + name + "\":")
	if err := e.emitOpenBrackets("["); err != nil {
		return err
	}
	for i := 0; e.stream.hasNext(); i++ {
		eleType := e.stream.peekType()
		if eleType != name {
			break
		}
		if i != 0 {
			e.emit(",")
		}
		ele, err := e.stream.popElement()
		if err != nil {
			return err
		}
		proto.NetworkElementToJSON(e.w, ele)
	}
	if err := e.emitCloseBrackets("]}"); err != nil {
		return err
	}
	return nil
}

// encode marshals v to JSON and then converts it to a string
func (e *Encoder) encode(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// emit emits a token
func (e *Encoder) emit(token string) {
	io.WriteString(e.w, token)
}

func (e *Encoder) emitElement(ele *proto.NetworkElement) {
	m := &jsonpb.Marshaler{}
	m.Marshal(e.w, ele)
}

func (e *Encoder) emitOpenBrackets(brackets string) error {
	for _, b := range brackets {
		if err := e.brackets.pushOpenBracket(string(b)); err != nil {
			return err
		}
		e.emit(string(b))
	}
	return nil
}

type bracketStack struct {
	stack []string
}

func (e *Encoder) emitCloseBrackets(brackets string) error {
	for _, b := range brackets {
		match, err := e.brackets.popCloseBracket()
		if err != nil {
			return err
		}
		if string(b) != match {
			return errors.New(fmt.Sprint("expected bracket", b, "found bracket", match))
		}
		e.emit(match)
	}
	return nil
}

func (e *Encoder) emitRemainingBrackets() error {
	for {
		b, err := e.brackets.popCloseBracket()
		if err != nil {
			return err
		}
		if b == "" {
			return nil
		}
		e.emit(b)
	}
}

func (s *bracketStack) pushOpenBracket(b string) error {
	if b != "[" && b != "{" {
		return errors.New(fmt.Sprint(b, "is not a recognized opening bracket"))
	}
	s.stack = append(s.stack, b)
	return nil
}

func (s *bracketStack) popCloseBracket() (string, error) {
	if len(s.stack) == 0 {
		return "", nil
	}
	b := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return s.matchBracket(b)
}

func (s *bracketStack) matchBracket(b string) (string, error) {
	switch b {
	case "[":
		return "]", nil
	case "{":
		return "}", nil
	default:
		return "", errors.New(fmt.Sprint("could not match bracket", b))
	}
}

type elementQueue struct {
	stream  <-chan *Message
	eleType string
	ele     *proto.NetworkElement
	eof     bool
}

func newElementQueue(s <-chan *Message) (*elementQueue, error) {
	q := &elementQueue{stream: s, eof: true}
	if err := q.loadNext(); err != nil {
		return nil, err
	}
	return q, nil
}

func (q *elementQueue) hasNext() bool {
	return q.eof
}

func (q *elementQueue) peekType() string {
	return q.eleType
}

func (q *elementQueue) popElement() (*proto.NetworkElement, error) {
	ele := q.ele
	err := q.loadNext()
	if err != nil {
		return nil, err
	}
	return ele, nil
}

func (q *elementQueue) loadNext() error {
	ele, err := ReceiveMessage(q.stream)
	if err == io.EOF {
		q.eof = false
		return nil
	}
	if err != nil {
		return err
	}
	q.eleType = proto.GetAspectName(ele)
	q.ele = ele
	return nil
}
