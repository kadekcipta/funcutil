package funcutil

import (
	"fmt"
	"testing"
)

type service struct {
	running bool
}

func (s *service) Run() {
	s.running = true
}

func (s *service) Stop(wait bool) {
	s.running = false
}

func (s *service) Pause() {
	s.running = false
}

func (s service) Running() bool {
	return s.running
}

func (s service) Info() string {
	return fmt.Sprintf("Running: %v", s.running)
}

type monitor struct {
}

func (m *monitor) Display() string {
	return "Display()"
}

func TestRegistration(t *testing.T) {
	f := New()
	f.Register(&service{}, &monitor{})
	if len(f.dump()) != 6 {
		t.Error("Registered methods should be 6")
	}
}

func TestMethodCalls(t *testing.T) {
	f := New()
	f.Register(&service{}, &monitor{})
	// test value set
	if _, err := f.Call("service.Run"); err != nil {
		t.Error(err)
	}
	// verify value
	if rets, err := f.Call("service.Running"); err != nil {
		t.Error(err)
	} else {
		if !rets[0].(bool) {
			t.Error("value should be set to true")
		}
	}
	// test wrong arguments
	if _, err := f.Call("service.Stop", 12); err == nil {
		t.Error("should failed due to wrong argument type")
	}
	// test non existing method
	if _, err := f.Call("service.NotExists"); err == nil {
		t.Error("method should not exists")
	}
}

func TestNamespace(t *testing.T) {
	expect := "com.example.device.monitor.Display() string"
	f := New("com.example.device")
	f.Register(&monitor{})
	m := f.dump()[0]
	if m != expect {
		t.Errorf("Should be %s got %s", expect, m)
	}
	if rets, err := f.Call("com.example.device.monitor.Display"); err != nil {
		t.Error(err)
	} else {
		if rets[0] != "Display()" {
			t.Error("Should be Display()")
		}
	}
}
