package github

import (
	"errors"
	"fmt"
	"time"
)

// ErrGraphQLUnavailable indicates that GitHub GraphQL API could not be queried
// (e.g. unauthenticated request without token, or rate limit exceeded).
var ErrGraphQLUnavailable = errors.New("github graphql api unavailable: requires authentication or rate limited")

// RateLimitError represents a GitHub API 403/429 rate limit exhaustion.
type RateLimitError struct {
	Limit     int
	Remaining int
	Reset     time.Time
	Resource  string
	Message   string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("github api rate limit exceeded (limit: %d, remaining: %d, reset: %s, resource: %s): %s",
		e.Limit, e.Remaining, e.Reset.UTC().Format(time.RFC3339), e.Resource, e.Message)
}

// NotFoundError represents a missing GitHub entity (e.g. user or repo).
type NotFoundError struct {
	Resource   string
	Identifier string
	Message    string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("github %s '%s' not found: %s", e.Resource, e.Identifier, e.Message)
}

// IsRateLimit checks if an error is a RateLimitError.
func IsRateLimit(err error) bool {
	var rle *RateLimitError
	return errors.As(err, &rle)
}

// IsNotFound checks if an error is a NotFoundError.
func IsNotFound(err error) bool {
	var nfe *NotFoundError
	return errors.As(err, &nfe)
}
