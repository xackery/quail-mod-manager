package qmm

import "testing"

func TestGenerateMod(t *testing.T) {
	err := generateMod()
	if err != nil {
		t.Fatalf("generate mod: %v", err)
	}
}
