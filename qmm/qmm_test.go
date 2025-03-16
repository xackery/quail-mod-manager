package qmm

import "testing"

func TestGenerateMod(t *testing.T) {
	err := generateMod()
	if err != nil {
		t.Fatalf("generate mod: %v", err)
	}
}

func TestAddModzip(t *testing.T) {
	New()
	err := AddModZip("../bin/mods/classic-robes.zip")
	if err != nil {
		t.Fatalf("add mod zip: %v", err)
	}

}
