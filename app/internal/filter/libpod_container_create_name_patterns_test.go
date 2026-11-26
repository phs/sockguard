package filter

import (
	"testing"
)

// required_name_patterns is the write-side counterpart of
// response.name_patterns: it scopes the create body's name field so every
// container reachable by name in a per-ID route was named in scope at create
// time — soft ownership without labels, filter injection, or daemon lookups.

func libpodCreateBody(name string) []byte {
	if name == "" {
		return []byte(`{"image":"localhost/tool:latest","systemd":"false"}`)
	}
	return []byte(`{"image":"localhost/tool:latest","systemd":"false","name":"` + name + `"}`)
}

func TestLibpodContainerCreateRequiredNamePatterns(t *testing.T) {
	// Unset patterns keep today's behavior: any name, and none, pass.
	unset := newLibpodContainerCreatePolicy(LibpodContainerCreateOptions{AllowSystemdMode: true})
	if reason := inspectLibpod(t, unset, libpodCreateBody("anything")); reason != "" {
		t.Fatalf("want allow: required_name_patterns unset, got deny %q", reason)
	}
	if reason := inspectLibpod(t, unset, libpodCreateBody("")); reason != "" {
		t.Fatalf("want allow: engine-assigned name with patterns unset, got deny %q", reason)
	}

	scoped := newLibpodContainerCreatePolicy(LibpodContainerCreateOptions{
		AllowSystemdMode:     true,
		RequiredNamePatterns: []string{"funlytv-*", "funlytv"},
	})
	if reason := inspectLibpod(t, scoped, libpodCreateBody("funlytv-build-7")); reason != "" {
		t.Fatalf("want allow: funlytv-build-7 matches funlytv-*, got deny %q", reason)
	}
	if reason := inspectLibpod(t, scoped, libpodCreateBody("electric-rose-tool")); reason == "" {
		t.Fatal("want deny: name outside every pattern")
	}
	// podman-remote sends an empty name whenever --name is absent; the gate
	// forces the client to name its own containers in scope.
	if reason := inspectLibpod(t, scoped, libpodCreateBody("")); reason == "" {
		t.Fatal("want deny: engine-assigned (empty) name under required_name_patterns")
	}

	// The glob dialect matches rule paths and visibility patterns: * stops at
	// a slash, so a slash-shaped pattern cannot accidentally match the
	// hyphenated names the same convention also admits.
	separated := newLibpodContainerCreatePolicy(LibpodContainerCreateOptions{
		AllowSystemdMode:     true,
		RequiredNamePatterns: []string{"funlytv/*"},
	})
	if reason := inspectLibpod(t, separated, libpodCreateBody("funlytv-build")); reason == "" {
		t.Fatal("want deny: funlytv-build does not match funlytv/*")
	}
}
