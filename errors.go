package citedhealth

import "fmt"

// CitedHealthError is the base error type for CITED Health API errors.
type CitedHealthError struct {
	Message string
}

func (e *CitedHealthError) Error() string {
	return e.Message
}

// NotFoundError is returned when a resource is not found (404).
type NotFoundError struct {
	Resource   string
	Identifier string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.Identifier)
}

// RateLimitError is returned when the API rate limit is exceeded (429).
type RateLimitError struct {
	RetryAfter int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %d seconds", e.RetryAfter)
}
