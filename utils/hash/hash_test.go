package hash

import "testing"

func TestPassword(t *testing.T) {
	h, err := Password("s3cret")
	if err != nil || h == "s3cret" {
		t.Fatal(err)
	}
	if !ComparePassword(h, "s3cret") || ComparePassword(h, "wrong") {
		t.Fatal("compare")
	}
}
