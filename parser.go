package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ericsage/cxmate/proto"
)

// supportedLongNumber is the value a CX parser must be able to represent.
const supportedLongNumber = 281474976710655

// LongNumber holds an integer value that must be representable by a CX parser.
type LongNumber struct {
	LongNumber int64 `json:"longNumber"`
}

// Metadata contains the metadata of a single CX aspect.
type Metadata struct {
	Name             string      `json:"name,omitempty"`
	Version          string      `json:"version,omitempty"`
	IDCounter        float64     `json:"idCounter,omitempty"`
	Properties       interface{} `json:"properties,omitempty"`
	ElementCount     float64     `json:"elementCount,omitempty"`
	ConsistencyGroup float64     `json:"consistencyGroup,omitempty"`
	Checksum         string      `json:"checksum,omitempty"`
}

// ParserConfig contains a description of each network the service will input.
type ParserConfig []NetworkDescription

// NetworkDescription describes a single CX network.
type NetworkDescription struct {
	//Label is a descriptor for the network that is used to identify it in a stream.
	Label string `json:"label"`
	//Description is a short description of the network, what it represents.
	Description string `json:"description"`
	//Aspects should contain a list of strings, the required aspects for this network.
	Aspects []string `json:"aspects"`
}

// validate performs validation on the ParserConfig.
func (c ParserConfig) validate() error {
	used := map[string]bool{}
	for i, n := range c {
		if n.Label == "" {
			return fmt.Errorf("invalid config for input network  at position %d: label missing", i)
		}
		if _, exists := used[n.Label]; exists {
			return fmt.Errorf("invalid config: output position: %d error: duplicate label found: %s", i, n.Label)
		}
		used[n.Label] = true
		if len(n.Aspects) == 0 {
			return fmt.Errorf("invalid config for input network %s at position %d: aspect list must not be empty", n.Label, i)
		}
		logDebug("Valid configuration found for input network", n.Label, "with required apsects", n.Aspects)
	}
	return nil
}

// Parser parses CX networks into protocol buffer messages.
type Parser struct {
	dec      *json.Decoder
	config   ParserConfig
	sendChan chan<- *Message
}

// parse uses a parserConfig as a guide for reading a CX stream from a reader, and sends any required aspect elements to the message stream.
func (c ParserConfig) parse(r io.Reader, s chan<- *Message) error {
	p := &Parser{
		dec:      json.NewDecoder(r),
		config:   c,
		sendChan: s,
	}
	if err := p.parseNetworks(); err != nil {
		return err
	}
	return nil
}

// parseNetworks iterates through the parser config and attempts to parse each network in the config.
func (p *Parser) parseNetworks() error {
	if err := p.parseDelim('[', "an opening brace of a CX stream"); err != nil {
		return err
	}
	for i, n := range p.config {
		if err := p.parseNetwork(n.Label, n.Aspects); err != nil {
			return fmt.Errorf("parse error for network at position %d: %v", i, err)
		}
	}
	if err := p.parseDelim(']', "a closing brace of a CX stream"); err != nil {
		return err
	}
	return nil
}

// parseNetwork parses a single network, streaming any required aspects to the service.
func (p *Parser) parseNetwork(label string, aspects []string) error {
	logDebug("Parsing network", label, "with required aspects", aspects)
	if err := p.parseDelim('[', "an opening brace of a CX encoded network"); err != nil {
		return err
	}
	if err := p.parseNumberVerification(); err != nil {
		return err
	}
	if err := p.parseMetadata(aspects); err != nil {
		return err
	}
	for p.more() {
		if err := p.parseAspect(label, aspects); err != nil {
			return err
		}
	}
	if err := p.parseDelim(']', "a closing brace of a CX encoded network"); err != nil {
		return err
	}
	return nil
}

// parseNumberVerification parses the numberVerification aspect and checks that the correct longNumber has been set.
func (p *Parser) parseNumberVerification() error {
	logDebug("Parsing number verification")
	var aspect map[string][]LongNumber
	if err := p.parseValue(&aspect, "the CX NumberVerification aspect"); err != nil {
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

// parseMetadata parses a pre-metadata aspect and fails if the aspects list is not included in the parsed metadata elements.
func (p *Parser) parseMetadata(aspects []string) error {
	logDebugln("Parsing pre-metadata")
	var aspect map[string][]Metadata
	if err := p.parseValue(&aspect, "the CX pre-metadata aspect"); err != nil {
		return err
	}
	metadata, ok := aspect["metaData"]
	if !ok {
		return errors.New("could not find pre-metadata aspect identifier \"metaData\"")
	}
	notFound := make([]string, len(aspects))
	copy(notFound, aspects)
	for _, m := range metadata {
		for i, n := range notFound {
			if m.Name == n {
				notFound = append(notFound[:i], notFound[i+1:]...)
			}
		}
	}
	if len(notFound) != 0 {
		return fmt.Errorf("could not find required aspects %v in the pre-metadata", notFound)
	}
	return nil
}

// parseAspect parses an aspect and determines if its aspect elements are required, it returns io.EOF if a post-metadata aspect is encountered.
func (p *Parser) parseAspect(label string, aspects []string) error {
	if err := p.parseDelim('{', "an opening bracket of an aspect fragment"); err != nil {
		return err
	}
	id, err := p.parseKey("an aspect identifier")
	if err != nil {
		return err
	}
	if id == "metaData" {
		logDebugln("Parsing post-metadata, returning EOF")
		return io.EOF
	}
	var shouldRead bool
	for _, requiredID := range aspects {
		if id == requiredID {
			shouldRead = true
		}
	}
	if err := p.parseDelim('[', "an opening brace for a list of "+id+" aspect elements"); err != nil {
		return err
	}
	logDebugln("Parsing aspect", id, "required:", shouldRead)
	for p.more() {
		if shouldRead {
			if err := p.parseElement(label, id); err != nil {
				return err
			}
		} else {
			if err := p.consumeElement(id); err != nil {
				return err
			}
		}
	}
	if err := p.parseDelim(']', "a closing brace of an aspect element list containing "+id+" aspect elements"); err != nil {
		return err
	}
	if err := p.parseDelim('}', "a closing bracket of an aspect of type"+id); err != nil {
		return err
	}
	return nil
}

// more returns true if a closing bracket is not the next token in the stream.
func (p *Parser) more() bool {
	return p.dec.More()
}

// parseDelim reads a token and tries to match it against a provided rune, which should be a JSON delimeter.
func (p *Parser) parseDelim(expects rune, description string) error {
	errMessage := "Expected " + description + ", error:"
	token, err := p.dec.Token()
	if err != nil {
		return fmt.Errorf("%s %v", errMessage, err)
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return fmt.Errorf("%s did not delimeter, found: %q", errMessage, token)
	}
	if rune(delim) != expects {
		return fmt.Errorf("%s found delimeter: %q", errMessage, delim)
	}
	return nil
}

// parseKey reads a token and tries to interpret it as a JSON object key.
func (p *Parser) parseKey(description string) (string, error) {
	errMessage := "Expected " + description + ", error: "
	token, err := p.dec.Token()
	if err != nil {
		return "", fmt.Errorf("%s %v", errMessage, err)
	}
	name, ok := token.(string)
	if !ok {
		return "", fmt.Errorf("%s did not find key", errMessage)
	}
	return name, nil
}

// parseValue tries to decode a JSON object into a value interface.
func (p *Parser) parseValue(value interface{}, description string) error {
	err := p.dec.Decode(value)
	if err != nil {
		return fmt.Errorf("Expected %s but could not decode value, error: %v", description, err)
	}
	return nil
}

// parseElement tries to decode a CX aspect element into a protobuf and send it to the service.
func (p *Parser) parseElement(networkID string, aspectID string) error {
	ele, err := proto.NetworkElementFromJSON(networkID, aspectID, p.dec)
	if err != nil {
		return errors.New("Error parsing aspect element in required aspect " + aspectID + ", error: " + err.Error())
	}
	return SendMessage(ele, p.sendChan)
}

// consumeElement tries to parse all of the tokens in an opaque aspect element.
func (p *Parser) consumeElement(id string) error {
	counter := 0
	for {
		t, err := p.dec.Token()
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
