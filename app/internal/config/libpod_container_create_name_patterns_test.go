package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLibpodContainerCreateRequiredNamePatternsDefaultOffAndMaps pins
// request_body.libpod_container_create.required_name_patterns: empty by
// default, reaching the native filter option from YAML, and not opening the
// Docker-compat gate.
func TestLibpodContainerCreateRequiredNamePatternsDefaultOffAndMaps(t *testing.T) {
	loaded, err := Load("/nonexistent-so-defaults-only.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.RequestBody.ToFilterOptions().LibpodContainerCreate.RequiredNamePatterns) != 0 {
		t.Fatal("libpod required_name_patterns is non-empty by default, want empty")
	}

	path := filepath.Join(t.TempDir(), "sockguard.yaml")
	yaml := "request_body:\n  libpod_container_create:\n    required_name_patterns: [\"funlytv-*\"]\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	opts := cfg.RequestBody.ToFilterOptions()
	if len(opts.LibpodContainerCreate.RequiredNamePatterns) != 1 ||
		opts.LibpodContainerCreate.RequiredNamePatterns[0] != "funlytv-*" {
		t.Fatalf("yaml: libpod required_name_patterns did not reach the filter option: %v",
			opts.LibpodContainerCreate.RequiredNamePatterns)
	}
}

// TestContainerCreateRequiredNamePatternsMaps pins the compat twin
// request_body.container_create.required_name_patterns reaching its filter
// option from YAML without filling the libpod one.
func TestContainerCreateRequiredNamePatternsMaps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sockguard.yaml")
	yaml := "request_body:\n  container_create:\n    required_name_patterns: [\"funlytv-*\"]\n"
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	opts := cfg.RequestBody.ToFilterOptions()
	if len(opts.ContainerCreate.RequiredNamePatterns) != 1 ||
		opts.ContainerCreate.RequiredNamePatterns[0] != "funlytv-*" {
		t.Fatalf("yaml: compat required_name_patterns did not reach the filter option: %v",
			opts.ContainerCreate.RequiredNamePatterns)
	}
	if len(opts.LibpodContainerCreate.RequiredNamePatterns) != 0 {
		t.Fatal("yaml: the compat key leaked into libpod_container_create")
	}
}

// TestLibpodContainerCreateRequiredNamePatternsValidation pins startup
// refusal of empty and uncompilable pattern entries, the same posture
// response.name_patterns takes.
func TestLibpodContainerCreateRequiredNamePatternsValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{"empty entry", "request_body:\n  libpod_container_create:\n    required_name_patterns: [\"\"]\n"},
		{"compat empty entry", "request_body:\n  container_create:\n    required_name_patterns: [\"\"]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sockguard.yaml")
			if err := os.WriteFile(path, []byte(tc.yaml), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := Validate(cfg); err == nil {
				t.Fatalf("Validate() accepted %s, want a validation error", tc.name)
			}
		})
	}
}
