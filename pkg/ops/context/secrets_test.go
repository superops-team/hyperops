package context

import (
	"testing"
)

func TestSeretsManager(t *testing.T) {
	sm := NewSecretsManager()
	sm.AddSecret("test")
	ok := sm.HasSecret("test")
	if !ok {
		t.Errorf("not match secrets")
	}
}

func TestSafeReplace(t *testing.T) {
	sm := NewSecretsManager()
	sm.AddSecret("wlc")
	sm.AddSecret("123")
	newmsg := sm.SafeReplace("my password is wlc, 123")
	expect := "my password is ***, ***"
	if newmsg != expect {
		t.Errorf("expected: %s, got: %s", expect, newmsg)
	}
}
