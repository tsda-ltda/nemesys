package auth

import "testing"

func TestNewToken(t *testing.T) {
	l := 32
	token, err := NewToken(l)
	if err != nil {
		t.Fatalf("fail to generate token, err: %s", err)
	}

	if len(token) != l {
		t.Errorf("invalid token size, want: %d, got: %d", l, len(token))
	}
}
