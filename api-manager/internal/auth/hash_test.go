package auth

import "testing"

func TestHash(t *testing.T) {
	secret := "secret12345"
	got, err := Hash(secret, 1)
	if err != nil {
		t.Errorf("fail to hash, err: %s", err)
	}
	if !CheckHash(secret, got) {
		t.Error("hash could not be checked, want: true, got: false")
	}
}

func TestCheckHash(t *testing.T) {
	secret := "secret12345"
	hash := "$2a$04$OQiL8iMzCC0TU2a1BznCkeTo3MO8JrpXeZjfMgvfyhjFztZLiDZX."

	if !CheckHash(secret, hash) {
		t.Error("check hash failed, want: true, got: false")
	}
}
