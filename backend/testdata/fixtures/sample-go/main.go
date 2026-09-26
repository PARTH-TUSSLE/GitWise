package sample

import (
	"context"
	"fmt"
	"sync"
)

// Config represents sample service configuration.
type Config struct {
	Host string
	Port int
}

// Greeter defines a sample greeting interface.
type Greeter interface {
	Greet(ctx context.Context, name string) (string, error)
}

// UserID is a custom type alias.
type UserID string

// Service implements Greeter.
type Service struct {
	mu  sync.Mutex
	cfg Config
}

// NewService instantiates a Service.
func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// Greet prints a friendly greeting.
func (s *Service) Greet(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}
	return fmt.Sprintf("Hello, %s!", name), nil
}

// unexportedHelper is an internal function.
func unexportedHelper(val int) int {
	return val * 2
}
