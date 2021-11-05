package main

import (
	"os"
	"testing"
)

func TestRemove(t *testing.T) {
	err := os.Remove("static\\covers\\c5mes04lv0153h6vgcsg.jpg")
	if err != nil {
		t.Fatal(err)
	}
}
