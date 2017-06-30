package main

import (
	"testing"

	"github.com/ericsage/cxmate/proto"
)

var testParams = map[string][]string{
	"testkey":  []string{"value1", "value2"},
	"testkey2": []string{"value1"},
}

func TestSendQueryStringParams(t *testing.T) {
	send := make(chan *Message)
	alive := make(chan bool)
	go func() {
		sm := <-send
		matchKeyValue(t, sm, "testkey", "value1")
		close(sm.errChan)
		sm = <-send
		matchKeyValue(t, sm, "testkey", "value2")
		close(sm.errChan)
		sm = <-send
		matchKeyValue(t, sm, "testkey2", "value1")
		close(sm.errChan)
		close(send)
		close(alive)
	}()
	config := &Config{
		Algorithm: AlgorithmConfig{
			Parameters: []ParameterConfig{
				ParameterConfig{
					Key: "testkey",
				},
				ParameterConfig{
					Key: "testkey2",
				},
			},
		},
	}
	m := &Mate{config, nil}
	err := m.processParameters(send, testParams)
	if err != nil {
		t.Fatalf("ProcessParameters returned an error: %#v", err)
	}
	<-alive
}

func matchKeyValue(t *testing.T, sm *Message, key string, value string) {
	paramwrapper := sm.ele.Element.(*proto.NetworkElement_Parameter)
	param := paramwrapper.Parameter
	if param.Name != key || param.Value != value {
		t.Fatalf("Expected key %s and value %s, received %#v", key, value, param)
	}
}
