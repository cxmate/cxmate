package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ericsage/cxmate/proto"
)

//supportedLongNumber is the maximum value CX can support
const supportedLongNumber = 281474976710655

//NumberVerification represents the maximum value supported by CX
type NumberVerification struct {
	LongNumber int64 `json:"longNumber"`
}

//Metadata represents the metadata found in a CX network
type Metadata struct {
	Name             string      `json:"name,omitempty"`
	Version          string      `json:"version,omitempty"`
	IDCounter        float64     `json:"idCounter,omitempty"`
	Properties       interface{} `json:"properties,omitempty"`
	ElementCount     float64     `json:"elementCount,omitempty"`
	ConsistencyGroup float64     `json:"consistencyGroup,omitempty"`
	Checksum         string      `json:"checksum,omitempty"`
}

//Decoder represents a decoder for a single CX network
type Decoder struct {
	dec      *json.Decoder
	config   DecoderConfig
	sendChan chan<- *Message
}

//DecoderConfig configures the details of networks being fed into the algorithm.
type DecoderConfig struct {
	//Label is a descriptor for the network that is used to identify it in a stream.
	Label string `json:"label"`
	//Description is a short description of the network, what it represents.
	Description string `json:"description"`
	//Aspects should contain a list of strings, the required aspects for this network.
	Aspects []string `json:"aspects"`
}

//NewDecoder validates a Decoder config and returns a new decoder or an error
func NewDecoder(r io.Reader, s chan<- *Message, c DecoderConfig) (*Decoder, error) {
	logDebug("creating new decoder")
	if c.Label == "" {
		return nil, errors.New("invalid config, label missing")
	}
	if len(c.Aspects) == 0 {
		return nil, errors.New("invalid config, aspects must not be empty")
	}
	return &Decoder{
		dec:      json.NewDecoder(r),
		config:   c,
		sendChan: s,
	}, nil
}

//DecodeNumberVerification decodes the magic long number from a CX stream
func (d *Decoder) DecodeNumberVerification() error {
	logDebug("decoding number verification")
	var aspect map[string][]NumberVerification
	if err := d.parseValue(&aspect, "the CX NumberVerification aspect"); err != nil {
		return err
	}
	numberVerification, ok := aspect["numberVerification"]
	if !ok {
		return errors.New("could not find key \"numberVerification\", the numberVerification is either missing or damaged")
	}
	if len(numberVerification) != 1 {
		return errors.New(fmt.Sprint("expected one long number in numberVerification, found", len(numberVerification)))
	}
	if numberVerification[0].LongNumber != supportedLongNumber {
		return errors.New(fmt.Sprint("expected long number", supportedLongNumber, "found", numberVerification[0].LongNumber))
	}
	return nil
}

//DecodeMetadata decodes a list of metdata structs from a CX stream
func (d *Decoder) DecodeMetadata() error {
	logDebugln("decoding metadata")
	var aspect map[string][]Metadata
	if err := d.parseValue(&aspect, "the CX metadata aspect"); err != nil {
		return err
	}
	metadata, ok := aspect["metaData"]
	if !ok {
		return errors.New("could not find key \"metadata\", the metadata is either missing or damaged")
	}
	names := make([]string, len(d.config.Aspects))
	copy(names, d.config.Aspects)
	for _, m := range metadata {
		for i, n := range names {
			if m.Name == n {
				names = append(names[:i], names[i+1:]...)
			}
		}
	}
	if len(names) != 0 {
		return errors.New(fmt.Sprintln("could not find required aspects in metadata expected", d.config.Aspects, "missing", names))
	}
	return nil
}

//DecodeAspect decodes a single aspect from a CX stream, it returns io.EOF if metdata is encountered
func (d *Decoder) DecodeAspect() error {
	if err := d.parseDelim('{', "an opening bracket of an aspect fragment"); err != nil {
		return err
	}
	id, err := d.parseKey("an aspect name as a key to a list of fragments")
	if err != nil {
		return err
	}
	if id == "metaData" {
		logDebugln("post-metadata encountered, returning EOF")
		return io.EOF
	}
	var shouldRead bool
	for _, requiredID := range d.config.Aspects {
		if id == requiredID {
			shouldRead = true
		}
	}
	if err := d.parseDelim('[', "an opening brace for a list of "+id+" elements"); err != nil {
		return err
	}
	for d.more() {
		if shouldRead {
			logDebugln("decoding required aspect", id)
			if err := d.parseElement(id); err != nil {
				return err
			}
		} else {
			if err := d.consumeElement(id); err != nil {
				logDebugln("decoding unrequired aspect", id)
				return err
			}
		}
	}
	if err := d.parseDelim(']', "a closing brace for a list of "+id+" elements"); err != nil {
		return err
	}
	if err := d.parseDelim('}', "a closing bracket for an aspect fragment of type"+id); err != nil {
		return err
	}
	return nil
}

//DecodeNetwork decodes a single network
func (d *Decoder) DecodeNetwork() error {
	logDebug("decoding network", d.config.Label, "with required aspects", d.config.Aspects)
	if err := d.parseDelim('[', "an opening brace of a CX encoded network"); err != nil {
		return err
	}
	if err := d.DecodeNumberVerification(); err != nil {
		return err
	}
	if err := d.DecodeMetadata(); err != nil {
		return err
	}
	for d.more() {
		if err := d.DecodeAspect(); err != nil {
			return err
		}
	}
	if err := d.parseDelim(']', "a closing brace for a CX encoded network"); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) more() bool {
	return d.dec.More()
}

func (d *Decoder) parseDelim(expects rune, description string) error {
	errMessage := "Expected " + description + ", error:"
	token, err := d.dec.Token()
	if err != nil {
		return errors.New(fmt.Sprint(errMessage, err.Error()))
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return errors.New(fmt.Sprint(errMessage, "did not find a delimeter, found: ", token))
	}
	if rune(delim) != expects {
		return errors.New(fmt.Sprint(errMessage, "found:", delim))
	}
	return nil
}

func (d *Decoder) parseKey(description string) (string, error) {
	errMessage := "Expected " + description + ", error: "
	token, err := d.dec.Token()
	if err != nil {
		return "", errors.New(errMessage + err.Error())
	}
	name, ok := token.(string)
	if !ok {
		return "", errors.New(errMessage + "did not find key")
	}
	return name, nil
}

func (d *Decoder) parseValue(value interface{}, description string) error {
	err := d.dec.Decode(value)
	if err != nil {
		return errors.New("Expected " + description + " but could not decode, error: " + err.Error())
	}
	return nil
}

func (d *Decoder) parseElement(id string) error {
	ele, err := proto.NetworkElementFromJSON(id, d.dec)
	if err != nil {
		return errors.New("Error parsing aspect element in required aspect " + id + ", error: " + err.Error())
	}
	return SendMessage(ele, d.sendChan)
}

func (d *Decoder) consumeElement(id string) error {
	counter := 0
	for {
		t, err := d.dec.Token()
		if err != nil {
			return errors.New("encountered error reading anonymous aspect element in the " + id + " aspect, error: " + err.Error())
		}
		if delim, ok := t.(json.Delim); ok {
			if delim == '{' {
				counter++
			}
			if delim == '}' {
				counter--
			}
			if counter == 0 {
				break
			}
		}
	}
	return nil
}
