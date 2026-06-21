package dto

import "fmt"

// HTTPError representa um erro HTTP
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implementa a interface error
func (e *HTTPError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("HTTP %d: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}
