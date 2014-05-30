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

type Monitor struct {
}

func (m *Monitor) Display() string {
	return "Display()"
}

func TestRegistration(t *testing.T) {
	f := New()
	f.Register(&service{}, &Monitor{})
	if len(f.Dump()) != 6 {
		t.Error("Registered methods should be 6")
	}
}

func TestMethodCalls(t *testing.T) {
	f := New()
	f.Register(&service{}, &Monitor{})
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

func createMonitor() Monitor {
	return Monitor{}
}

func delegateRegister(f *FuncUtil, vars ...interface{}) {
	f.Register(vars...)
}

func TestNamespace(t *testing.T) {
	expect := "com.example.device.Monitor.Display() string"
	f := New("com.example.device")
	monitor := createMonitor()
	//f.Register(&monitor)
	delegateRegister(f, &monitor)
	m := f.Dump()[0]
	if m != expect {
		t.Errorf("Should be %s got %s", expect, m)
	}
	if rets, err := f.Call("com.example.device.Monitor.Display"); err != nil {
		t.Error(err)
	} else {
		if rets[0] != "Display()" {
			t.Error("Should be Display()")
		}
	}
}
