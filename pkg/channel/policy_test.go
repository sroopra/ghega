package channel

import "testing"

func TestValidatePolicies_NetworkPolicy(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Network.AllowedHosts = []string{"example.com", "api.example.com"}
	errs := ValidatePolicies(&ch)
	if len(errs) > 0 {
		t.Errorf("expected no policy errors, got: %v", errs)
	}
}

func TestValidatePolicies_NetworkPolicyEmptyHost(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Network.AllowedHosts = []string{"example.com", "", "api.example.com"}
	errs := ValidatePolicies(&ch)
	if !hasErrorField(errs, "policies.network.allowedHosts[1]") {
		t.Errorf("expected error on empty host, got: %v", errorFields(errs))
	}
}

func TestValidatePolicies_NetworkPolicyWildcard(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Network.AllowedHosts = []string{"*.example.com"}
	errs := ValidatePolicies(&ch)
	if !hasErrorField(errs, "policies.network.allowedHosts[0]") {
		t.Errorf("expected error on wildcard host, got: %v", errorFields(errs))
	}
}

func TestValidatePolicies_NetworkPolicyWhitespaceHost(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Network.AllowedHosts = []string{"   "}
	errs := ValidatePolicies(&ch)
	if !hasErrorField(errs, "policies.network.allowedHosts[0]") {
		t.Errorf("expected error on whitespace-only host, got: %v", errorFields(errs))
	}
}

func TestValidatePolicies_PayloadPolicy(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Payload.MaxSizeBytes = 1024 * 1024
	errs := ValidatePolicies(&ch)
	if len(errs) > 0 {
		t.Errorf("expected no policy errors, got: %v", errs)
	}
}

func TestValidatePolicies_PayloadPolicyTooLarge(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Payload.MaxSizeBytes = 100*1024*1024 + 1
	errs := ValidatePolicies(&ch)
	if !hasErrorField(errs, "policies.payload.maxSizeBytes") {
		t.Errorf("expected error on oversized payload, got: %v", errorFields(errs))
	}
}

func TestValidatePolicies_PayloadPolicyZeroIsValid(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Payload.MaxSizeBytes = 0
	errs := ValidatePolicies(&ch)
	if len(errs) > 0 {
		t.Errorf("expected no policy errors for zero maxSizeBytes, got: %v", errs)
	}
}

func TestValidatePolicies_TimePolicy(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Time.MaxProcessingSeconds = 300
	errs := ValidatePolicies(&ch)
	if len(errs) > 0 {
		t.Errorf("expected no policy errors, got: %v", errs)
	}
}

func TestValidatePolicies_TimePolicyTooLarge(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Time.MaxProcessingSeconds = 301
	errs := ValidatePolicies(&ch)
	if !hasErrorField(errs, "policies.time.maxProcessingSeconds") {
		t.Errorf("expected error on oversized time limit, got: %v", errorFields(errs))
	}
}

func TestValidatePolicies_TimePolicyZeroIsValid(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies.Time.MaxProcessingSeconds = 0
	errs := ValidatePolicies(&ch)
	if len(errs) > 0 {
		t.Errorf("expected no policy errors for zero maxProcessingSeconds, got: %v", errs)
	}
}

func TestValidatePolicies_NoPoliciesIsValid(t *testing.T) {
	ch := makeValidChannel()
	ch.Policies = Policies{}
	errs := ValidatePolicies(&ch)
	if len(errs) > 0 {
		t.Errorf("expected no policy errors when policies are empty, got: %v", errs)
	}
}
