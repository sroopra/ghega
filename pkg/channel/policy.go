package channel

import (
	"fmt"
	"strings"
)

// ValidatePolicies validates the policies defined on a channel.
func ValidatePolicies(channel *Channel) []ValidationError {
	var errs []ValidationError

	// Network policy
	for i, host := range channel.Policies.Network.AllowedHosts {
		if strings.TrimSpace(host) == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("policies.network.allowedHosts[%d]", i),
				Message: "allowed host must not be empty",
			})
			continue
		}
		if strings.Contains(host, "*") {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("policies.network.allowedHosts[%d]", i),
				Message: fmt.Sprintf("allowed host %q must not contain wildcards", host),
			})
		}
	}

	// Payload policy
	if channel.Policies.Payload.MaxSizeBytes > 0 {
		const maxSize = 100 * 1024 * 1024 // 100MB
		if channel.Policies.Payload.MaxSizeBytes > maxSize {
			errs = append(errs, ValidationError{
				Field:   "policies.payload.maxSizeBytes",
				Message: fmt.Sprintf("maxSizeBytes must be <= 100MB (%d), got %d", maxSize, channel.Policies.Payload.MaxSizeBytes),
			})
		}
	}

	// Time policy
	if channel.Policies.Time.MaxProcessingSeconds > 0 {
		if channel.Policies.Time.MaxProcessingSeconds > 300 {
			errs = append(errs, ValidationError{
				Field:   "policies.time.maxProcessingSeconds",
				Message: fmt.Sprintf("maxProcessingSeconds must be <= 300, got %d", channel.Policies.Time.MaxProcessingSeconds),
			})
		}
	}

	return errs
}
