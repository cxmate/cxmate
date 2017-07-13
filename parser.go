package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cxmate/cxmate/proto"
)

// NetworkDescription describes a single CX network.
type NetworkDescription struct {
	//Label is a descriptor for the network that is used to identify it in a stream.
	Label string `json:"label"`
	//Description is a short description of the network, what it represents.
	Description string `json:"description"`
	//Aspects should contain a list of strings, the required aspects for this network.
	Aspects []string `json:"aspects"`
}

// ParserConfig contains a description of each network the service will input.
type ParserConfig []NetworkDescription

// validate performs validation on the ParserConfig.
func (c ParserConfig) validate() error {
	used := map[string]bool{}
	if len(c) == 0 {
		return errors.New("must have at least one input network")
	}
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
	sendChan chan<- *Message
}

// parse uses a parserConfig as a guide for reading a CX stream from a reader, and sends any required aspect elements to the message stream.
func (c ParserConfig) parse(r io.Reader, s chan<- *Message, singleton bool) error {
	p := &Parser{
		dec:      json.NewDecoder(r),
		sendChan: s,
	}
	var err error
	if singleton {
		err = p.network(c[0].Label, c[0].Aspects)
	} else {
		err = p.stream(c)
	}
	if err != nil {
		return err
	}
	return nil
}

// networks iterates through the parser config and attempts to parse each network in the config.
func (p *Parser) stream(networks []NetworkDescription) error {
	if err := p.bracket('[', "an opening brace of a CX stream"); err != nil {
		return err
	}
	for i, n := range networks {
		if err := p.network(n.Label, n.Aspects); err != nil {
			return fmt.Errorf("error parsing %s at position %d: %v", n.Label, i, err)
		}
	}
	if err := p.bracket(']', "a closing brace of a CX stream"); err != nil {
		return err
	}
	return nil
}

// network parses a single network, streaming any required aspects to the service.
func (p *Parser) network(network string, aspects []string) error {
	logDebug("Parsing", network, "with required aspects", aspects)
	if err := p.bracket('[', "an opening brace of a CX encoded network"); err != nil {
		return err
	}
	if err := p.numberVerification(network); err != nil {
		return err
	}
	if err := p.preMetadata(network, aspects); err != nil {
		return err
	}
	for p.more() {
		if err := p.aspect(network, aspects); err != nil {
			return err
		}
	}
	if err := p.bracket(']', "a closing brace of a CX encoded network"); err != nil {
		return err
	}
	return nil
}

// maxInt is the value a CX parser must be able to represent.
const maxInt = 281474976710655

// numberVerification parses the numberVerification aspect and checks that the correct longNumber has been set.
func (p *Parser) numberVerification(network string) error {
	logDebug("Parsing number verification in", network)
	var aspect map[string][]map[string]int
	if err := p.value(&aspect, "the CX NumberVerification aspect"); err != nil {
		return err
	}
	nv, ok := aspect["numberVerification"]
	if !ok {
		return errors.New(`missing key "numberVerification", in numberVerification aspect`)
	}
	if len(nv) != 1 {
		return fmt.Errorf("expected one numberVerification aspect element, found %d", len(nv))
	}
	ln, ok := nv[0]["longNumber"]
	if !ok {
		return errors.New(`missing key "longNumber" in numberVerification aspect element`)
	}
	if ln != maxInt {
		return fmt.Errorf("expected longNumber %d found %d", maxInt, ln)
	}
	return nil
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

// preMetadata parses the metadata aspect that goes before any user defined aspects. preMetadata garuntees that the parsed metadata contains metadata element for each aspect in the list of aspects provided.
func (p *Parser) preMetadata(network string, aspects []string) error {
	logDebugln("Parsing preMetadata in", network)
	var aspect map[string][]Metadata
	if err := p.value(&aspect, "the CX pre-metadata aspect"); err != nil {
		return err
	}
	metadata, ok := aspect["metaData"]
	if !ok {
		return errors.New(`could not find pre-metadata aspect identifier "metaData"`)
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

// aspect parses a user defined aspect.
func (p *Parser) aspect(network string, aspects []string) error {
	if err := p.bracket('{', "an opening bracket of an aspect fragment"); err != nil {
		return err
	}
	aspect, err := p.key("an aspect identifier")
	logDebug("Parsing", aspect, "aspect in", network)
	if err != nil {
		return err
	}
	if aspect == "metaData" {
		logDebugln("Parsing post-metadata in", network, "returning EOF")
		return io.EOF
	}
	var shouldRead bool
	for _, requiredAspect := range aspects {
		if aspect == requiredAspect {
			shouldRead = true
		}
	}
	if err := p.bracket('[', "an opening brace for a list of "+aspect+" aspect elements"); err != nil {
		return err
	}
	logDebugln("Parsing aspect", aspect, "required:", shouldRead)
	for p.more() {
		if shouldRead {
			if err := p.element(network, aspect); err != nil {
				return err
			}
		} else {
			if err := p.opaque(aspect); err != nil {
				return err
			}
		}
	}
	if err := p.bracket(']', "a closing brace of an aspect element list containing "+aspect+" aspect elements"); err != nil {
		return err
	}
	if err := p.bracket('}', "a closing bracket of an aspect of type "+aspect); err != nil {
		return err
	}
	return nil
}

// more checks if the next parsable token is a closing bracket.
func (p *Parser) more() bool {
	return p.dec.More()
}

// bracket parses a token tok match against a provided bracket rune.
func (p *Parser) bracket(expects rune, description string) error {
	token, err := p.dec.Token()
	if err != nil {
		return fmt.Errorf("Could not fetch token %s error: %v", description, err)
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return fmt.Errorf("expected %s but found non-delimeter: %v", description, token)
	}
	if rune(delim) != expects {
		return fmt.Errorf("expected bracket %q, %s, found bracket: %q", expects, description, rune(delim))
	}
	return nil
}

// key parses an object field token.
func (p *Parser) key(description string) (string, error) {
	token, err := p.dec.Token()
	if err != nil {
		return "", fmt.Errorf("Could not fetch token %s error: %v", description, err)
	}
	name, ok := token.(string)
	if !ok {
		return "", fmt.Errorf("expected %s but found non-field key: %v", description, token)
	}
	return name, nil
}

// value parses any value into an empty interface.
func (p *Parser) value(value interface{}, description string) error {
	err := p.dec.Decode(value)
	if err != nil {
		return fmt.Errorf("expected %s but could not decode value, error: %v", description, err)
	}
	return nil
}

//element parses an aspect element to protobuf and sends it to the service.
func (p *Parser) element(network string, aspect string) error {
	ele, err := proto.NetworkElementFromJSON(network, aspect, p.dec)
	if err != nil {
		return fmt.Errorf("error parsing required aspect element in %s error: %v", aspect, err)
	}
	return SendMessage(ele, p.sendChan)
}

// opaque parses an opaque aspect element, discarding all of its tokens.
func (p *Parser) opaque(aspect string) error {
	counter := 0
	for {
		t, err := p.dec.Token()
		if err != nil {
			return fmt.Errorf("error parsing opaque aspect element in %s error: %v", aspect, err)
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
