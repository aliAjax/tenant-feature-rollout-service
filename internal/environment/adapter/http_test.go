package adapter

import (
	"strings"
	"testing"
)

func TestEnvironmentHTTPInitializesPolicy(t *testing.T) {
	e, err := DecodeEnvironmentRequest(strings.NewReader(`{"id":"prod","project_id":"p-1","name":"production"}`))
	if err != nil {
		t.Fatal(err)
	}
	e.Policies["approval"] = "required"
	if e.Policies["approval"] != "required" {
		t.Fatal("decoded policy map is not writable")
	}
}
