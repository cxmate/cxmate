package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cxmate/cxmate/proto"
)

// GeneratorConfig contains a description of each network that cxMate will generate for the service.
type GeneratorConfig []NetworkDescription

// validate performs validation on the GeneratorConfig.
func (c GeneratorConfig) validate() error {
	used := map[string]bool{}
	if len(c) == 0 {
		return errors.New("must have at least one output network")
	}
	for i, n := range c {
		if n.Label == "" {
			return fmt.Errorf("invalid input output position: %d error: label missing", i)
		}
		if _, exists := used[n.Label]; exists {
			return fmt.Errorf("invalid config: output position: %d error: duplicate label found: %s", i, n.Label)
		}
		used[n.Label] = true
		if len(n.Aspects) == 0 {
			return fmt.Errorf("invalid config: output position: %d networks: %s error: aspect list must not be empty", i, n.Label)
		}
		logDebugln("Config loaded: output position:", i, "network:", n.Label, "required apsects:", n.Aspects)
	}
	return nil
}

// Generator generates CX networks from protocol buffer messages.
type Generator struct {
	w        io.Writer
	elements *elements
	brackets brackets
}

// run initializes and runs a CX generator.
func (c GeneratorConfig) generate(w io.Writer, s <-chan *Message, singleton bool) error {
	logDebugln("Generator initializing")
	stream, err := newElementStream(s)
	if err != nil {
		return err
	}
	g := &Generator{
		w:        w,
		elements: stream,
		brackets: newBracketStack(),
	}
	defer g.closeRemainingBrackets()
	if singleton {
		err = g.network(c[0].Label, c[0].Aspects)
	} else {
		err = g.stream(c)
	}
	if err != nil {
		return err
	}
	return nil
}

// stream generates a list of networks from a list of NetworkDescriptions.
func (g *Generator) stream(networks []NetworkDescription) error {
	logDebugln("Generating a stream of networks")
	if err := g.openBrackets("["); err != nil {
		return err
	}
	for i, n := range networks {
		if i == 1 {
			if err := g.rune(','); err != nil {
				return err
			}
		}
		if err := g.network(n.Label, n.Aspects); err != nil {
			return fmt.Errorf("error while generating %s at position %d: %v", n.Label, i, err)
		}
	}
	if err := g.closeBrackets("]"); err != nil {
		return err
	}
	return nil
}

// network generates a single network.
func (g *Generator) network(network string, aspects []string) error {
	logDebugln("Generating", network, "with aspects", aspects)
	if err := g.openBrackets("["); err != nil {
		return err
	}
	if err := g.numberVerification(network); err != nil {
		return err
	}
	if err := g.rune(','); err != nil {
		return err
	}

	if err := g.preMetadata(network, aspects); err != nil {
		return err
	}
	for g.elements.hasNext() {
		if elementNetwork, _ := g.elements.peekNetwork(); elementNetwork != network {
			break
		}
		elementAspect, _ := g.elements.peekAspect()
		if err := g.rune(','); err != nil {
			return err
		}
		if err := g.aspect(network, elementAspect); err != nil {
			return err
		}
	}
	if err := g.closeBrackets("]"); err != nil {
		return err
	}
	return nil
}

// numberVerfication generates a numberVerification aspect.
func (g *Generator) numberVerification(network string) error {
	logDebugln("Generating number verification in", network)
	nv := map[string][]map[string]int64{
		"numberVerification": []map[string]int64{
			map[string]int64{"longNumber": 281474976710655},
		},
	}
	if err := g.value(nv); err != nil {
		return err
	}
	return nil
}

// preMetadata generates the metadata aspect that goes before any user defined aspects. preMetadata takes a list of aspect names, which is uses to populate the name field for each aspect metadata element (the only required field in preMetadata).
func (g *Generator) preMetadata(network string, aspects []string) error {
	logDebugln("Generating preMetadata in", network)
	md := []Metadata{}
	for _, a := range aspects {
		md = append(md, Metadata{
			Name: a,
		})
	}
	pm := map[string][]Metadata{
		"metaData": md,
	}
	if err := g.value(pm); err != nil {
		return err
	}
	return nil
}

// aspect generates a single aspect by reading elements until the elements aspect or network no longer match the provided parameters.
func (g *Generator) aspect(network string, aspect string) error {
	logDebugln("Generating", aspect, "aspect in", network)
	if err := g.openBrackets("{"); err != nil {
		return err
	}
	if err := g.string("\"" + aspect + "\":"); err != nil {
		return err
	}
	if err := g.openBrackets("["); err != nil {
		return err
	}
	for i := 0; g.elements.hasNext(); i++ {
		elementAspect, _ := g.elements.peekAspect()
		if elementAspect != aspect {
			break
		}
		elementNetwork, _ := g.elements.peekNetwork()
		if elementNetwork != network {
			break
		}
		element, err := g.elements.pop()
		if err != nil {
			return err
		}
		if i != 0 {
			if err := g.rune(','); err != nil {
				return err
			}
		}
		proto.NetworkElementToJSON(g.w, element)
	}
	if err := g.closeBrackets("]}"); err != nil {
		return err
	}
	return nil
}

// value generates a JSON encoded value.
func (g *Generator) value(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err = g.w.Write(b); err != nil {
		return err
	}
	return nil
}

// rune generates a string.
func (g *Generator) string(s string) error {
	if _, err := g.w.Write([]byte(s)); err != nil {
		return err
	}
	return nil
}

// rune generates a single rune.
func (g *Generator) rune(r rune) error {
	if _, err := g.w.Write([]byte(string(r))); err != nil {
		return err
	}
	return nil
}

// openBrackets pushes each bracket in the bracket string onto the bracket stack and then generates the bracket.
func (g *Generator) openBrackets(brackets string) error {
	for _, b := range brackets {
		if err := g.brackets.pushOpen(b); err != nil {
			return err
		}
		if err := g.rune(b); err != nil {
			return err
		}
	}
	return nil
}

// closeBrackets pops a bracket from the bracket stack for each bracket in the bracket string. If the popped bracket and the bracket from the bracket string match, the bracket is emitted, otherwise an error is returned.
func (g *Generator) closeBrackets(brackets string) error {
	for _, b := range brackets {
		if m, ok := g.brackets.popClose(); !ok {
			return fmt.Errorf("generator: expected closing bracket %q found empty bracket stack", b)
		} else if m != b {
			return fmt.Errorf("generator: expected closing bracket %q found %q", m, b)
		}
		if err := g.rune(b); err != nil {
			return err
		}
	}
	return nil
}

// closeRemainingBrackets pops brackets off the bracket stack, and emits them until the bracket stack becomes empty.
func (g *Generator) closeRemainingBrackets() error {
	for {
		m, ok := g.brackets.popClose()
		if !ok {
			break
		}
		if err := g.rune(m); err != nil {
			return err
		}
	}
	return nil
}

// brackets is a stack that pushes the matching bracket of any brackets that are pushed onto it via its pushOpen method. Popping off the matched brackets will yielding the closing brackets in the correct order to close the opened brackets.
type brackets []rune

// newBracketStack creates an empty bracket stack.
func newBracketStack() brackets {
	return make(brackets, 0)
}

// pushOpen validates that b is a bracket ('[' or '{') before pushing the matching bracket onto the stack.
func (s *brackets) pushOpen(b rune) error {
	if b != '[' && b != '{' {
		return fmt.Errorf("brackets: %q is not a recognized opening bracket", b)
	}
	m, err := s.getMatch(b)
	if err != nil {
		return fmt.Errorf("brackets: error pushing bracket: %v", err)
	}
	*s = append(*s, m)
	return nil
}

// popClose pops off a bracket frop the top of the stack and returns it. An extra bool is returned indicating if the pop succeeded (true) or the stack was empty (false).
func (s *brackets) popClose() (rune, bool) {
	if len(*s) == 0 {
		return ' ', false
	}
	b := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return b, true
}

// getMatch is a utility to get the closing bracket for a corrosponding opening bracket.
func (s *brackets) getMatch(b rune) (rune, error) {
	switch b {
	case '[':
		return ']', nil
	case '{':
		return '}', nil
	default:
		return ' ', fmt.Errorf("brackets: could not match bracket %q", b)
	}
}

// elements provides queue based access to a stream of protobuf messages. The queue prefetchesan element before it is popped so that it becomes possible to peek at its aspect and network identifers before consuming the element. If an error occurs when elements tries to fetch the next element, the error will be returned by the constructor or by pop, and it will no longer be possible to fetch elements.
type elements struct {
	stream  <-chan *Message
	curr    *proto.NetworkElement
	err     error
	aspect  string
	network string
}

// newElementStream creates a new stream that will read from the provided channel, and blocks until the first element can be received from the channel. An error fetching the first element will yield a nil stream and an error return value. Note that the element stream will never close the channel.
func newElementStream(s <-chan *Message) (*elements, error) {
	e := &elements{stream: s}
	e.next()
	if e.err != nil {
		return nil, e.err
	}
	return e, nil
}

// pop returns an element from the stream or an error. If an error occurs, the stream should be closed as no values can be read from it in the future. If an element is available, pop blocks until the next element can be fetched before returning the current element.
func (e *elements) pop() (*proto.NetworkElement, error) {
	if e.err != nil {
		return nil, e.err
	}
	c := e.curr
	e.next()
	return c, nil
}

// hasNext returns true if an element can be read from the stream. If hasNext returns false then then elements can no longer be read from.
func (e *elements) hasNext() bool {
	if e.err != nil {
		return false
	}
	return true
}

// peekAspect returns the aspect identifier of the next element or an empty string if the next element could not be read. A boolean is also returned to determine if this error occured.
func (e *elements) peekAspect() (string, bool) {
	if e.err != nil {
		return "", false
	}
	return e.aspect, true
}

// peekNetwork returns the network label of the next element or an empty string if the next element could not be read. A boolean is also returned to determine if this error occured.
func (e *elements) peekNetwork() (string, bool) {
	if e.err != nil {
		return "", false
	}
	return e.network, true
}

// Next moves to the next element in the stream unless an error is recieved.
func (e *elements) next() {
	element, err := ReceiveMessage(e.stream)
	e.err = err
	if err == nil {
		e.curr = element
		e.aspect = proto.GetAspectName(element)
		e.network = element.Label
	}
}
