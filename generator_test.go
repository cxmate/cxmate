package main

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/cxmate/cxmate/proto"
)

func TestPushInvalidBracket(t *testing.T) {
	bs := newBracketStack()
	if err := bs.pushOpen('}'); err == nil {
		t.Error("Push accepted an invalid bracket '}'")
	}
	if len(bs) != 0 {
		t.Fatalf("Expected 0 brackets, found %d", len(bs))
	}
}

func TestPushOpenBracket(t *testing.T) {
	bs := newBracketStack()
	if err := bs.pushOpen('['); err != nil {
		t.Error(err)
	}
	if len(bs) != 1 {
		t.Fatalf("Expected 1 bracket, found %d", len(bs))
	}
	if bs[0] != ']' {
		t.Errorf("Expected ']', found %q", bs[0])
	}
}

func TestPushOpenBrace(t *testing.T) {
	bs := newBracketStack()
	if err := bs.pushOpen('{'); err != nil {
		t.Error(err)
	}
	if len(bs) != 1 {
		t.Fatalf("Expected 1 brace, found %d", len(bs))
	}
	if bs[0] != '}' {
		t.Errorf("Expected '}', found %q", bs[0])
	}
}

func TestPushMultipleOpenBrackets(t *testing.T) {
	bs := newBracketStack()
	testBrackets := []rune{'{', '{', '[', '{'}
	for _, b := range testBrackets {
		if err := bs.pushOpen(b); err != nil {
			t.Error(err)
		}
	}
	if len(bs) != len(testBrackets) {
		t.Fatalf("Expected %d brackets, found %d", len(testBrackets), len(bs))
	}
	matchBrackets := []rune{'}', '}', ']', '}'}
	for i, m := range matchBrackets {
		if bs[i] != m {
			t.Fatalf("Expected %q, found %q", m, bs[i])
		}
	}
}

func TestPopCloseOffEmptyStack(t *testing.T) {
	bs := newBracketStack()
	if r, ok := bs.popClose(); ok {
		t.Fatalf("Should not allow pop from empty stack, received %q", r)
	}
}

func TestPopCloseBracket(t *testing.T) {
	bs := newBracketStack()
	if err := bs.pushOpen('['); err != nil {
		t.Error(err)
	}
	b, ok := bs.popClose()
	if !ok {
		t.Error("Bracket stack should not be empty")
	}
	if b != ']' {
		t.Errorf("Expected ']', found %q", b)
	}
}

func TestPopCloseBrace(t *testing.T) {
	bs := newBracketStack()
	if err := bs.pushOpen('{'); err != nil {
		t.Error(err)
	}
	b, ok := bs.popClose()
	if !ok {
		t.Error("Bracket stack should not be empty")
	}
	if b != '}' {
		t.Errorf("Expected '}', found %q", b)
	}
}

func TestPopMultipleCloseBrackets(t *testing.T) {
	bs := newBracketStack()
	testBrackets := []rune{'{', '{', '[', '{'}
	for _, b := range testBrackets {
		if err := bs.pushOpen(b); err != nil {
			t.Error(err)
		}
	}
	matchBrackets := []rune{'}', ']', '}', '}'}
	for _, m := range matchBrackets {
		if b, ok := bs.popClose(); !ok {
			t.Error("Bracket stack should not be empty")
		} else if m != b {
			t.Fatalf("Expected %q, found %q", m, b)
		}
	}
}

func TestCreateNewElementStream(t *testing.T) {
	s := make(chan *Message, 1)
	e := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{},
		},
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	es, err := newElementStream(s)
	if err != nil {
		t.Error(err)
	}
	if es.err == io.EOF {
		t.Error("Stream should not be closed")
	}
	if err != nil {
		t.Error(err)
	}
	if es.network != "test network" {
		t.Errorf("expected network test network, found %s", es.network)
	}
	if es.aspect != "nodes" {
		t.Errorf("expected aspect nodes, found %s", es.aspect)
	}
	if es.curr != e {
		t.Errorf("current element stream element did not match sent node element")
	}
}

func TestCreateNewElementStreamWithError(t *testing.T) {
	s := make(chan *Message, 1)
	s <- &Message{
		ele: nil,
		err: errors.New("ERROR"),
	}
	if _, err := newElementStream(s); err == nil {
		t.Error("Expected error from creating stream with error message")
	}
}

func TestHasNextWithNext(t *testing.T) {
	s := make(chan *Message, 2)
	e := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{},
		},
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	es, err := newElementStream(s)
	if err != nil {
		t.Error(err)
	}
	if !es.hasNext() {
		t.Error("Expected stream to be open")
	}
	if _, err := es.pop(); err != nil {
		t.Error(err)
	}
	if !es.hasNext() {
		t.Error("Expected stream to be open")
	}
}

func TestHasNextWithEOF(t *testing.T) {
	s := make(chan *Message, 2)
	e := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{},
		},
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	es, err := newElementStream(s)
	if err != nil {
		t.Error(err)
	}
	if !es.hasNext() {
		t.Error("Expected stream to be open")
	}
	close(s)
	if _, err := es.pop(); err != nil {
		t.Error(err)
	}
	if es.hasNext() {
		t.Error("Expected stream to be closed")
	}
}

func TestPeekAspect(t *testing.T) {
	s := make(chan *Message, 2)
	e := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{},
		},
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	es, err := newElementStream(s)
	if err != nil {
		t.Error(err)
	}
	if aspect, ok := es.peekAspect(); !ok || aspect != "nodes" {
		t.Errorf("Expected nodes found %s", aspect)
	}
}

func TestPeekNetwork(t *testing.T) {
	s := make(chan *Message, 2)
	e := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{},
		},
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	es, err := newElementStream(s)
	if err != nil {
		t.Error(err)
	}
	if network, ok := es.peekNetwork(); !ok || network != "test network" {
		t.Errorf("Expected test network found %s", network)
	}
}

func TestElementStreamWithError(t *testing.T) {
	s := make(chan *Message, 2)
	e := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{},
		},
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	es, err := newElementStream(s)
	s <- &Message{
		ele: nil,
		err: errors.New("ERROR"),
	}
	ele, err := es.pop()
	if err != nil {
		t.Error(err)
	}
	if ele != e {
		t.Error("Popped element did not match sent element")
	}
	ele, err = es.pop()
	if err == nil || ele != nil {
		t.Error("Expected error from popping off element after error message")
	}
	if aspect, ok := es.peekAspect(); aspect != "" || ok {
		t.Errorf("Expected peekAspect to return \"\" and false, got %s %t", aspect, ok)
	}
	if network, ok := es.peekNetwork(); network != "" || ok {
		t.Errorf("Expected peekNetwork to return \"\" and false, got %s %t", network, ok)
	}
}

func TestValidateValidGeneratorConfig(t *testing.T) {
	net1 := NetworkDescription{
		Label:   "test network",
		Aspects: []string{"nodes"},
	}
	genConf := GeneratorConfig([]NetworkDescription{net1})
	if err := genConf.validate(); err != nil {
		t.Error(err)
	}
}

func TestValidateNoLabelGeneratorConfig(t *testing.T) {
	net1 := NetworkDescription{
		Label:   "",
		Aspects: []string{"nodes"},
	}
	genConf := GeneratorConfig([]NetworkDescription{net1})
	if err := genConf.validate(); err == nil {
		t.Error("Should not accept config with network with no label")
	}
}

func TestValidateNoAspectGeneratorConfig(t *testing.T) {
	net1 := NetworkDescription{
		Label:   "test network",
		Aspects: []string{},
	}
	genconf := GeneratorConfig([]NetworkDescription{net1})
	if err := genconf.validate(); err == nil {
		t.Error("Should not accept config with network with no aspects")
	}
}

func TestValidateNonUniqueLabelsGeneratorConfig(t *testing.T) {
	net1 := NetworkDescription{
		Label:   "test network",
		Aspects: []string{"edges"},
	}
	net2 := NetworkDescription{
		Label:   "test network",
		Aspects: []string{"nodes"},
	}
	genConf := GeneratorConfig([]NetworkDescription{net1, net2})
	if err := genConf.validate(); err == nil {
		t.Error("Should not accept config with duplicate labels")
	}
}

func TestString(t *testing.T) {
	buf := bytes.NewBufferString("")
	g := &Generator{
		w: buf,
	}
	if err := g.string("test"); err != nil {
		t.Error(err)
	}
	if buf.String() != "test" {
		t.Errorf("Expected buffer to contain e found %s", buf.String())
	}
}

func TestRune(t *testing.T) {
	buf := bytes.NewBufferString("")
	g := &Generator{
		w: buf,
	}
	if err := g.rune('e'); err != nil {
		t.Error(err)
	}
	if buf.String() != "e" {
		t.Errorf("Expected buffer to contain e found %s", buf.String())
	}
}

func TestGenOpenBrackets(t *testing.T) {
	buf := bytes.NewBufferString("")
	g := &Generator{
		w:        buf,
		brackets: newBracketStack(),
	}
	if err := g.openBrackets("{["); err != nil {
		t.Error(err)
	}
	if buf.String() != "{[" {
		t.Errorf("Expected buffer to contain {[ found %s", buf.String())
	}
	if len(g.brackets) != 2 {
		t.Errorf("Expected buffer to contain two brackets found %d", len(g.brackets))
	}
}

func TestGenCloseBrackets(t *testing.T) {
	buf := bytes.NewBufferString("")
	g := &Generator{
		w:        buf,
		brackets: newBracketStack(),
	}
	if err := g.openBrackets("{["); err != nil {
		t.Error(err)
	}
	if err := g.closeBrackets("]}"); err != nil {
		t.Error(err)
	}
	if buf.String() != "{[]}" {
		t.Errorf("Expected buffer to contain {[]} found %s", buf.String())
	}
	if len(g.brackets) != 0 {
		t.Errorf("Expected buffer to contain zero brackets found %d", len(g.brackets))
	}
}

func TestGenRemainingBrackets(t *testing.T) {
	buf := bytes.NewBufferString("")
	g := &Generator{
		w:        buf,
		brackets: newBracketStack(),
	}
	if err := g.openBrackets("{[[{"); err != nil {
		t.Error(err)
	}
	if err := g.closeRemainingBrackets(); err != nil {
		t.Error(err)
	}
	if buf.String() != "{[[{}]]}" {
		t.Errorf("Expected buffer to contain {[[{}]]} found %s", buf.String())
	}
	if len(g.brackets) != 0 {
		t.Errorf("Expected buffer to contain zero brackets found %d", len(g.brackets))
	}
}

func TestGenerateNumberVerification(t *testing.T) {
	buf := bytes.NewBufferString("")
	g := &Generator{
		w:        buf,
		brackets: newBracketStack(),
	}
	if err := g.numberVerification("test network"); err != nil {
		t.Error(err)
	}
	expected := `{"numberVerification":[{"longNumber":281474976710655}]}`
	if buf.String() != expected {
		t.Errorf("Expected %s found %s", expected, buf.String())
	}
}

func TestGeneratePreMetadata(t *testing.T) {
	buf := bytes.NewBufferString("")
	g := &Generator{
		w:        buf,
		brackets: newBracketStack(),
	}
	if err := g.preMetadata("test network", []string{"nodes", "edges"}); err != nil {
		t.Error(err)
	}
	expected := `{"metaData":[{"name":"nodes"},{"name":"edges"}]}`
	if buf.String() != expected {
		t.Errorf("Expected %s found %s", expected, buf.String())
	}
}

func TestGenerateAspect(t *testing.T) {
	s := make(chan *Message, 10)
	e := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{
				Id: 1,
			},
		},
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	s <- &Message{
		ele: e,
		err: nil,
	}
	s <- &Message{
		ele: nil,
		err: io.EOF,
	}
	stream, err := newElementStream(s)
	if err != nil {
		t.Error(err)
	}
	buf := bytes.NewBufferString("")
	g := &Generator{
		w:        buf,
		brackets: newBracketStack(),
		elements: stream,
	}
	if err := g.aspect("test network", "nodes"); err != nil {
		t.Error(err)
	}
	expected := `{"nodes":[{"@id":"1"},{"@id":"1"},{"@id":"1"},{"@id":"1"}]}`
	if buf.String() != expected {
		t.Errorf("Expected %s found %s", expected, buf.String())
	}
}

func TestGenerateNetwork(t *testing.T) {
	s := make(chan *Message, 20)
	e1 := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{
				Id: 1,
			},
		},
	}
	e2 := &proto.NetworkElement{
		Label: "test network",
		Element: &proto.NetworkElement_Edge{
			Edge: &proto.Edge{
				Id: 1,
			},
		},
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e2,
		err: nil,
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e2,
		err: nil,
	}
	s <- &Message{
		ele: nil,
		err: io.EOF,
	}
	stream, err := newElementStream(s)
	if err != nil {
		t.Error(err)
	}
	buf := bytes.NewBufferString("")
	g := &Generator{
		w:        buf,
		brackets: newBracketStack(),
		elements: stream,
	}
	if err := g.network("test network", []string{"nodes", "edges"}); err != nil {
		t.Error(err)
	}
	expected := `[{"numberVerification":[{"longNumber":281474976710655}]},{"metaData":[{"name":"nodes"},{"name":"edges"}]},{"nodes":[{"@id":"1"}]},{"edges":[{"@id":"1"}]},{"nodes":[{"@id":"1"},{"@id":"1"}]},{"edges":[{"@id":"1"}]}]`
	if buf.String() != expected {
		t.Errorf("Expected %s found %s", expected, buf.String())
	}
}

func TestGenerateStream(t *testing.T) {
	s := make(chan *Message, 20)
	conf := []NetworkDescription{
		NetworkDescription{
			Label:   "test network 1",
			Aspects: []string{"nodes"},
		},
		NetworkDescription{
			Label:   "test network 2",
			Aspects: []string{"edges"},
		},
	}
	e1 := &proto.NetworkElement{
		Label: "test network 1",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{
				Id: 1,
			},
		},
	}
	e2 := &proto.NetworkElement{
		Label: "test network 2",
		Element: &proto.NetworkElement_Edge{
			Edge: &proto.Edge{
				Id: 1,
			},
		},
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e2,
		err: nil,
	}
	s <- &Message{
		ele: e2,
		err: nil,
	}
	s <- &Message{
		ele: nil,
		err: io.EOF,
	}
	stream, err := newElementStream(s)
	if err != nil {
		t.Error(err)
	}
	buf := bytes.NewBufferString("")
	g := &Generator{
		w:        buf,
		brackets: newBracketStack(),
		elements: stream,
	}
	if err := g.stream(conf); err != nil {
		t.Error(err)
	}
	expected := `[[{"numberVerification":[{"longNumber":281474976710655}]},{"metaData":[{"name":"nodes"}]},{"nodes":[{"@id":"1"},{"@id":"1"},{"@id":"1"}]}],[{"numberVerification":[{"longNumber":281474976710655}]},{"metaData":[{"name":"edges"}]},{"edges":[{"@id":"1"},{"@id":"1"}]}]]`
	if buf.String() != expected {
		t.Errorf("Expected %s found %s", expected, buf.String())
	}
}

func TestRun(t *testing.T) {
	s := make(chan *Message, 20)
	conf := []NetworkDescription{
		NetworkDescription{
			Label:   "test network 1",
			Aspects: []string{"nodes"},
		},
		NetworkDescription{
			Label:   "test network 2",
			Aspects: []string{"edges"},
		},
	}
	e1 := &proto.NetworkElement{
		Label: "test network 1",
		Element: &proto.NetworkElement_Node{
			Node: &proto.Node{
				Id: 1,
			},
		},
	}
	e2 := &proto.NetworkElement{
		Label: "test network 2",
		Element: &proto.NetworkElement_Edge{
			Edge: &proto.Edge{
				Id: 1,
			},
		},
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e1,
		err: nil,
	}
	s <- &Message{
		ele: e2,
		err: nil,
	}
	s <- &Message{
		ele: e2,
		err: nil,
	}
	s <- &Message{
		ele: nil,
		err: io.EOF,
	}
	genConf := GeneratorConfig(conf)
	buf := bytes.NewBufferString("")
	if err := genConf.generate(buf, s, false); err != nil {
		t.Error(err)
	}
	expected := `[[{"numberVerification":[{"longNumber":281474976710655}]},{"metaData":[{"name":"nodes"}]},{"nodes":[{"@id":"1"},{"@id":"1"},{"@id":"1"}]}],[{"numberVerification":[{"longNumber":281474976710655}]},{"metaData":[{"name":"edges"}]},{"edges":[{"@id":"1"},{"@id":"1"}]}]]`
	if buf.String() != expected {
		t.Errorf("Expected %s found %s", expected, buf.String())
	}
}
