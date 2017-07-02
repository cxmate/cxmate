package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ericsage/cxmate/proto"
)

// Encoder represents an encoder for a single CX network
type Encoder struct {
	w          io.Writer
	closeStack []string
	nextEle    proto.NetworkElement
	config     EncoderConfig
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
func NewEncoder(w io.Writer, c EncoderConfig) (*Encoder, error) {
	if c.Label == "" {
		return nil, errors.New("invalid config, label missing")
	}
	if len(c.Aspects) == 0 {
		return nil, errors.New("invalid config, aspects must not be empty")
	}
	return &Encoder{
		w:      w,
		config: c,
	}, nil
}

func (e *Encoder) EncodeNetwork() error {
	defer e.emitCloseRemaining()
	e.emitOpen("[")
	if err := e.emitNumberVerfication(); err != nil {
		return err
	}
	if err := e.emitMetadata(); err != nil {
		return err
	}
	var err error
	for {
		err = emitAspect()
		if err == io.EOF {
			break
		} else {
			return err
		}
	}
	if err := e.emitMetadata(); err != nil {
		return err
	}
	e.emitClose("]")
}

// EncodeNumberVerfication writes out the standard number verification stanza
func (e *Encoder) emitNumberVerfication() error {
	l := NumberVerification{
		LongNumber: supportedLongNumber,
	}
	nv := map[string]NumberVerification{"numberVerification": l}
	s, err := e.encode(nv)
	if err != nil {
		return err
	}
	e.emit(s)
	return nil
}

// EncodeMetadata writes out the standard metadata
func (e *Encoder) emitMetadata() error {
	e.emit(",")
	e.emitOpen("{")
	e.emit("\"metadata\":")
	e.emitOpen("[")
	for i := range e.config.Aspects {
		if i != 0 {
			e.emit(",")
		}
		s, err := e.encode(Metadata{}) //TODO
		if err != nil {
			return err
		}
		e.emit(s)
	}
	e.emitClose("]}")
	return nil
}

func (e *Encoder) emitAspect() error {
	e.emit(",")
	e.emitOpen("{")
	e.emit("\"" + name + "\":")
	e.emitOpen("[")
	n := 0
	for {
		eleType, err := e.stream.peek() //TODO
		if eleType == currType {
			if n != 0 {
				e.emit(",")
			}
			e.emitElement(e.stream.pop()) //TODO
		} else if err != nil {
			return err
		}
	}
	e.emitClose("]}")
	return nil
}

// emit emits a token
func (e *Encoder) emit(token string) {
	io.WriteString(e.w, token)
}

// emitUnclosed emits token and pushs its closing partner chracter onto a stack
func (e *Encoder) emitOpen(token string) {
	var partner string
	switch token {
	case "[":
		partner = "]"
	case "{":
		partner = "}"
	case "\"":
		partner = "\""
	}
	e.closeStack = append(e.closeStack, partner)
}

func (e *Encoder) emitClose(tokens string) error {
	numChars := len([]rune(tokens))
	if numChars > len(e.closeStack) {
		return errors.New(fmt.Sprintln("Not enough items on the close stack to close", numChars, "characters"))
	}
	return emitCloseNum(numChars)
}

func (e *Encoder) emitCloseRemaining(tokens string) error {
	return emitCloseNum(len(e.closeStack))
}

// emitCloseStack emits all of the closing partner characters on the stack, then clears the stack
func (e *Encoder) emitCloseNum(num int) {
	for i := (num - 1); i >= 0; i-- {
		e.emit(e.closeStack[i])
	}
	e.closeStack = e.closeStack[:len(e.closeStack)-num]
}

// encode marshals v to JSON and then converts it to a string
func (e *Encoder) encode(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
