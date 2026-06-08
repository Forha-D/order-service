package utils

import "testing"

func TestBuildIdempotencyKey_IsStableAndBounded(t *testing.T) {
	key1 := BuildIdempotencyKey("user-123", "create_order", "abc-001")
	key2 := BuildIdempotencyKey("user-123", "create_order", "abc-001")

	if key1 != key2 {
		t.Fatalf("expected deterministic key, got %q and %q", key1, key2)
	}

	if len(key1) != 64 {
		t.Fatalf("expected 64-char hashed key, got %d chars", len(key1))
	}
}

func TestBuildIdempotencyKey_DiffersForDifferentInputs(t *testing.T) {
	key1 := BuildIdempotencyKey("user-123", "create_order", "abc-001")
	key2 := BuildIdempotencyKey("user-123", "create_order", "abc-002")

	if key1 == key2 {
		t.Fatal("expected different keys for different client values")
	}
}
