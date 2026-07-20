package config

import (
	"errors"
	"fmt"
	"net"
)

const (
	defaultAddr = "0.0.0.0:8080"
)

type ServerConfig struct {
	Addr string `json:"addr"`
}

func (c *ServerConfig) SetDefault() {
	if c.Addr == "" {
		c.Addr = defaultAddr
	}
}

func (c *ServerConfig) Validate() error {
	if c.Addr == "" {
		return errors.New("服务监听地址不能为空")
	}
	if _, _, err := net.SplitHostPort(c.Addr); err != nil {
		return fmt.Errorf("无效的服务监听地址 %q: %w", c.Addr, err)
	}
	return nil
}
