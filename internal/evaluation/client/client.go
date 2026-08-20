package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Value struct {
	Key     string `json:"key"`
	Value   any    `json:"value"`
	Type    string `json:"type"`
	Version int64  `json:"version"`
	Reason  string `json:"reason"`
}

type Request struct {
	ProjectID, Key, UserID string
	Attributes             map[string]any
}

type Transport interface {
	Do(context.Context, Request) (Value, error)
}

type HTTPTransport struct {
	BaseURL, Token string
	Client         *http.Client
}

func (t HTTPTransport) Do(ctx context.Context, r Request) (Value, error) {
	if t.Client == nil {
		t.Client = &http.Client{Timeout: 3 * time.Second}
	}
	client := t.Client
	if client.Timeout == 0 {
		copy := *client
		copy.Timeout = 200 * time.Millisecond
		client = &copy
	}
	body, err := json.Marshal(map[string]any{"project_id": r.ProjectID, "key": r.Key, "user_id": r.UserID, "attributes": r.Attributes})
	if err != nil {
		return Value{}, fmt.Errorf("encode evaluation: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, t.BaseURL+"/v1/evaluate", bytes.NewReader(body))
	if err != nil {
		return Value{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if t.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.Token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Value{}, errors.New(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return Value{}, fmt.Errorf("evaluate status %d", resp.StatusCode)
	}
	var v Value
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return Value{}, fmt.Errorf("decode evaluation: %w", err)
	}
	return v, nil
}

type Client struct {
	transport Transport
	mu        sync.RWMutex
	cache     map[string]Value
	offline   bool
	hook      func(Value)
}

func New(t Transport) *Client                 { return &Client{transport: t, cache: make(map[string]Value)} }
func (c *Client) SetOffline(v bool)           { c.mu.Lock(); c.offline = v; c.mu.Unlock() }
func (c *Client) OnEvaluation(fn func(Value)) { c.mu.Lock(); c.hook = fn; c.mu.Unlock() }

func (c *Client) Evaluate(ctx context.Context, r Request) (Value, error) {
	c.mu.RLock()
	off := c.offline
	cached, ok := c.cache[r.Key]
	hook := c.hook
	c.mu.RUnlock()
	if off && ok {
		return cached, nil
	}
	v, err := c.transport.Do(ctx, r)
	if err != nil && ok {
		return cached, nil
	}
	if err != nil {
		return Value{}, err
	}
	c.mu.Lock()
	c.cache[r.Key] = v
	c.mu.Unlock()
	if hook != nil {
		hook(v)
	}
	return v, nil
}

func (c *Client) Warm(ctx context.Context, requests []Request) error {
	for _, request := range requests {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := c.Evaluate(ctx, request); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Clear() { c.mu.Lock(); c.cache = make(map[string]Value); c.mu.Unlock() }

func (c *Client) Snapshot() map[string]Value {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]Value, len(c.cache))
	for key, value := range c.cache {
		out[key] = value
	}
	return out
}
