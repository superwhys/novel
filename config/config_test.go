package config

import "testing"

func TestServerConfigDefaults(t *testing.T) {
	config := new(ServerConfig)
	config.SetDefault()

	if config.Addr != defaultAddr {
		t.Fatalf("Addr = %q, want %q", config.Addr, defaultAddr)
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestServerConfigRejectsInvalidAddress(t *testing.T) {
	config := &ServerConfig{Addr: "localhost"}
	if err := config.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}
