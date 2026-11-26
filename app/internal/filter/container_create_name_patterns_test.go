package filter

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// required_name_patterns on the compat route: the container name rides the
// ?name= query parameter, its legacy ?fn= twin, or the body Name field, so
// every supplied candidate must match and a create that supplies none is
// denied like an out-of-scope name.

func makeNamedInspectRequest(t *testing.T, target string, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func inspectNamedCreate(t *testing.T, policy containerCreatePolicy, target, body string) string {
	t.Helper()
	reason, err := policy.inspect(nil, makeNamedInspectRequest(t, target, body), "/containers/create")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return reason
}

func TestContainerCreateRequiredNamePatterns(t *testing.T) {
	// Unset patterns keep today's behavior: any name, and none, pass.
	unset := newContainerCreatePolicy(ContainerCreateOptions{})
	for _, target := range []string{
		"/containers/create",
		"/containers/create?name=anything",
	} {
		if reason := inspectNamedCreate(t, unset, target, `{"Image":"docker.io/library/alpine:3.21"}`); reason != "" {
			t.Fatalf("want allow: required_name_patterns unset at %s, got deny %q", target, reason)
		}
	}

	scoped := newContainerCreatePolicy(ContainerCreateOptions{RequiredNamePatterns: []string{"funlytv-*"}})
	inspect := func(target, body string) string {
		return inspectNamedCreate(t, scoped, target, body)
	}
	if reason := inspect("/containers/create?name=funlytv-build-7", `{"Image":"docker.io/library/alpine:3.21"}`); reason != "" {
		t.Fatalf("want allow: funlytv-build-7 matches funlytv-*, got deny %q", reason)
	}
	if reason := inspect("/containers/create?name=electric-rose-tool", `{"Image":"docker.io/library/alpine:3.21"}`); reason == "" {
		t.Fatal("want deny: query name outside every pattern")
	}
	if reason := inspect("/containers/create", `{"Image":"docker.io/library/alpine:3.21","Name":"electric-rose-tool"}`); reason == "" {
		t.Fatal("want deny: body Name outside every pattern")
	}
	// podman-remote sends no name at all when --name is absent.
	if reason := inspect("/containers/create", `{"Image":"docker.io/library/alpine:3.21"}`); reason == "" {
		t.Fatal("want deny: engine-assigned (no supplied name) under required_name_patterns")
	}
	// Every supplied candidate is weighed, so an in-scope query cannot
	// launder an out-of-scope body name: whichever the engine picks is scoped.
	if reason := inspect("/containers/create?name=funlytv-7", `{"Image":"docker.io/library/alpine:3.21","Name":"other-team"}`); reason == "" {
		t.Fatal("want deny: body Name out of scope alongside an in-scope query name")
	}
	// The legacy ?fn= twin is weighed like ?name=.
	if reason := inspect("/containers/create?fn=funlytv-7", `{"Image":"docker.io/library/alpine:3.21"}`); reason != "" {
		t.Fatalf("want allow: fn=funlytv-7 matches funlytv-*, got deny %q", reason)
	}
}
