package auth

import "testing"

func TestPassword_HashAndCompare(t *testing.T) {
	raw := "P@ssw0rd-123"
	hash, err := HashPassword(raw)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if hash == raw {
		t.Fatal("hash should not equal raw password")
	}
	if err = ComparePassword(hash, raw); err != nil {
		t.Fatalf("compare should pass: %v", err)
	}
}

func TestPassword_CompareWrongPassword(t *testing.T) {
	hash, err := HashPassword("right-password")
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if err = ComparePassword(hash, "wrong-password"); err == nil {
		t.Fatal("compare should fail for wrong password")
	}
}
