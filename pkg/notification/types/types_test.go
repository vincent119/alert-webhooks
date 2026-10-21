package types

import "testing"

func TestPtrReturnsPointerToValue(t *testing.T) {
	value := Ptr("notification")
	if value == nil || *value != "notification" {
		t.Fatalf("Ptr() = %v", value)
	}
}
