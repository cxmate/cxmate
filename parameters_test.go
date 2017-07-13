package main

import (
	"testing"

	"github.com/cxmate/cxmate/proto"
)

func TestConvertNumber(t *testing.T) {
	p := &Parameter{
		Name:   "Test Number",
		Type:   "number",
		Format: "test number",
	}
	m, err := p.convert("2.0")
	if err != nil {
		t.Error(err)
	}
	if m.Name != "Test Number" {
		t.Errorf("Expected Test Number found %s", m.Name)
	}
	if m.Format != "test number" {
		t.Errorf("Expected test number found %s", m.Format)
	}
	val, ok := m.GetValue().(*proto.Parameter_NumberValue)
	if !ok {
		t.Error("Could not cast parameter to number")
	}
	if val.NumberValue != 2.0 {
		t.Errorf("Expected 2.0 found %f", val.NumberValue)
	}
}

func TestConvertBoolean(t *testing.T) {
	p := &Parameter{
		Name:   "Test Boolean",
		Type:   "boolean",
		Format: "test boolean",
	}
	m, err := p.convert("true")
	if err != nil {
		t.Error(err)
	}
	if m.Name != "Test Boolean" {
		t.Errorf("Expected Test Boolean found %s", m.Name)
	}
	if m.Format != "test boolean" {
		t.Errorf("Expected test boolean found %s", m.Format)
	}
	val, ok := m.GetValue().(*proto.Parameter_BooleanValue)
	if !ok {
		t.Error("Could not cast parameter to boolean")
	}
	if val.BooleanValue != true {
		t.Errorf("Expected true found %t", val.BooleanValue)
	}
}

func TestConvertInteger(t *testing.T) {
	p := &Parameter{
		Name:   "Test Integer",
		Type:   "integer",
		Format: "test integer",
	}
	m, err := p.convert("7")
	if err != nil {
		t.Error(err)
	}
	if m.Name != "Test Integer" {
		t.Errorf("Expected Test Integer found %s", m.Name)
	}
	if m.Format != "test integer" {
		t.Errorf("Expected test integer found %s", m.Format)
	}
	val, ok := m.GetValue().(*proto.Parameter_IntegerValue)
	if !ok {
		t.Error("Could not cast parameter to integer")
	}
	if val.IntegerValue != 7 {
		t.Errorf("Expected 7 found %d", val.IntegerValue)
	}
}

func TestConvertString(t *testing.T) {
	p := &Parameter{
		Name:   "Test String",
		Format: "test string",
	}
	m, err := p.convert("7")
	if err != nil {
		t.Error(err)
	}
	if m.Name != "Test String" {
		t.Errorf("Expected Test String found %s", m.Name)
	}
	if m.Format != "test string" {
		t.Errorf("Expected test string found %s", m.Format)
	}
	val, ok := m.GetValue().(*proto.Parameter_StringValue)
	if !ok {
		t.Error("Could not cast parameter to string")
	}
	if val.StringValue != "7" {
		t.Errorf("Expected 7 found %s", val.StringValue)
	}
}

func TestParameterSend(t *testing.T) {
	send := make(chan *Message)
	go func(t *testing.T, send chan *Message) {
		m, ok := <-send
		if !ok {
			t.Fatal("Send should not be closed")
		}
		param, ok := m.ele.GetElement().(*proto.NetworkElement_Parameter)
		if !ok {
			t.Error("Could not cast network element to parameter")
		}
		ele, ok := param.Parameter.GetValue().(*proto.Parameter_StringValue)
		if !ok {
			t.Error("Could not cast parameter to string")
		}
		if param.Parameter.Name != "Test String" {
			t.Errorf("Expected Test String found %s", param.Parameter.Name)
		}
		if param.Parameter.Format != "test string" {
			t.Errorf("Expected test string found %s", param.Parameter.Format)
		}
		if ele.StringValue != "test value" {
			t.Errorf("Expected test value found %s", ele.StringValue)
		}
		close(m.errChan)
	}(t, send)
	p := &Parameter{
		Name:   "Test String",
		Format: "test string",
	}
	if err := p.send(send, "test value"); err != nil {
		t.Error("Should not receive error from send")
	}
}

func TestParameterConfigSend(t *testing.T) {
	send := make(chan *Message)
	query := map[string][]string{
		"test1": []string{"val1", "val2"},
		"test2": []string{"val3"},
	}
	params := []Parameter{
		Parameter{
			Name:    "test1",
			Default: "default 1",
		},
		Parameter{
			Name:    "test2",
			Default: "default 2",
		},
		Parameter{
			Name:    "test3",
			Default: "default 3",
		},
	}
	config := ParameterConfig(params)
	go func(t *testing.T, s chan *Message) {
		testStringParam(t, s, "test1", "", "val1")
		testStringParam(t, s, "test1", "", "val2")
		testStringParam(t, s, "test2", "", "val3")
		testStringParam(t, s, "test3", "", "default 3")
	}(t, send)
	config.send(send, query)
}

func testStringParam(t *testing.T, c chan *Message, name string, format string, expected string) {
	m, ok := <-c
	if !ok {
		t.Error("Channel is closed")
	}
	netParam, ok := m.ele.GetElement().(*proto.NetworkElement_Parameter)
	if !ok {
		t.Error("Could not cast message element to parameter")
	}
	ele := netParam.Parameter
	if ele.Name != name {
		t.Errorf("Expected %s found %s", name, ele.Name)
	}
	if ele.Format != format {
		t.Errorf("Expected %s found %s", format, ele.Format)
	}
	val, ok := ele.GetValue().(*proto.Parameter_StringValue)
	if !ok {
		t.Error("Could not cast parameter value to string")
	}
	if val.StringValue != expected {
		t.Errorf("Expected %s found %s", expected, val)
	}
	close(m.errChan)
}
