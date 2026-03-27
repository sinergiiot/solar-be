package tier

import "fmt"

// LimitError is returned when a tier limit is reached.
type LimitError struct {
	Feature string `json:"feature"`
	Current int    `json:"current"`
	Limit   int    `json:"limit"`
	Tier    string    `json:"tier"`
	Message string `json:"message"`
}

func (e *LimitError) Error() string {
	return e.Message
}

func NewLimitError(feature string, current, limit int, tierName string) *LimitError {
	return &LimitError{
		Feature: feature,
		Current: current,
		Limit:   limit,
		Tier:    tierName,
		Message: fmt.Sprintf("Batas limit paket %s tercapai (%d %s). Silakan upgrade ke Pro/Enterprise.", tierName, limit, feature),
	}
}
