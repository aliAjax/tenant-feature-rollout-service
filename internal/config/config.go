package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address                        string
	ReadTimeout, WriteTimeout      time.Duration
	MaxFlags, MaxRules, MaxEvalQPS int
}

func Default() Config {
	return Config{Address: ":8087", ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, MaxFlags: 1000, MaxRules: 10000, MaxEvalQPS: 1000}
}
func Load() Config {
	c := Default()
	path := os.Getenv("ROLLOUT_CONFIG")
	if path == "" {
		path = "configs/config.yaml"
	}
	_ = loadYAML(path, &c)
	if v := os.Getenv("ROLLOUT_ADDR"); v != "" {
		c.Address = v
	}
	return c
}
func loadYAML(path string, c *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		p := strings.SplitN(strings.TrimSpace(s.Text()), ":", 2)
		if len(p) != 2 {
			continue
		}
		k := strings.TrimSpace(p[0])
		v := strings.Trim(strings.TrimSpace(p[1]), "\"'")
		switch k {
		case "address":
			c.Address = v
		case "read_timeout":
			if d, e := time.ParseDuration(v); e == nil {
				c.ReadTimeout = d
			}
		case "write_timeout":
			if d, e := time.ParseDuration(v); e == nil {
				c.WriteTimeout = d
			}
		case "max_flags":
			if n, e := strconv.Atoi(v); e == nil {
				c.MaxFlags = n
			}
		case "max_rules":
			if n, e := strconv.Atoi(v); e == nil {
				c.MaxRules = n
			}
		case "max_eval_qps":
			if n, e := strconv.Atoi(v); e == nil {
				c.MaxEvalQPS = n
			}
		}
	}
	return s.Err()
}
func (c Config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("address required")
	}
	if c.MaxFlags <= 0 || c.MaxRules <= 0 {
		return fmt.Errorf("quotas must be positive")
	}
	return nil
}
