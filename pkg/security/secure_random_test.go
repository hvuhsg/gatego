package security

import "testing"

func TestSecureRandomString(t *testing.T) {
	t.Parallel()

	length := 20

	r1 := GenerateRandomString(length)

	if len(r1) != length {
		t.Errorf("Expected random string with length %d, got strign with length %d", length, len(r1))
	}

	r2 := GenerateRandomString(length)

	if r1 == r2 {
		t.Errorf("Expected random string but got the same string twice (very very improbable)")
	}
}
