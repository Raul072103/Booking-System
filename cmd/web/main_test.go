package main

import (
	"github.com/raul/BookingSystem/internal/handlers"
	"testing"
)

func TestRun(t *testing.T) {
	err := run(handlers.NewTestRepo(&app))
	if err != nil {
		t.Errorf("failed run()")
	}
}
