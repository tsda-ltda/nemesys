package auth

import "testing"

var v = 0

func TestNewToken(t *testing.T) {
	t.Log(v)
	v++
	expected := 32
	token, err := NewToken(expected)
	if err != nil {
		t.Fatalf("fail to generate token, err: %s", err)
	}
	received := len(token)
	if received != expected {
		t.Errorf("invalid token size, want: %d, got: %d", expected, received)
	}
}
