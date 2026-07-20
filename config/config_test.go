package config

import "testing"

func TestServerConfigDefaults(t *testing.T) {
	config := new(ServerConfig)
	config.SetDefault()

	if config.Addr != defaultAddr {
		t.Fatalf("Addr = %q, want %q", config.Addr, defaultAddr)
	}
	if config.ContentDir != defaultContentDir {
		t.Fatalf("ContentDir = %q, want %q", config.ContentDir, defaultContentDir)
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestServerConfigRejectsInvalidAddress(t *testing.T) {
	config := &ServerConfig{Addr: "localhost", ContentDir: defaultContentDir}
	if err := config.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestServerConfigRejectsEmptyContentDir(t *testing.T) {
	config := &ServerConfig{Addr: defaultAddr}
	if err := config.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}
