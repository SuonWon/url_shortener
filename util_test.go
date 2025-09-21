package util_test

import (
	"testing"

	tiny "github.com/url_shortener/internal/util"
)

func TestAdd(t *testing.T) {
	result := tiny.Add(5, 3)
	expected := 8

	if result != expected {
		t.Errorf("Add(5, 3) returned %d, expected %d", result, expected)
	}
}

func TestTiny(t *testing.T) {
	result, _ := tiny.RandCode(7)
	expected := 7
	resLen := len(result)
	if resLen != expected {
		t.Errorf("Add(5, 3) returned %d, expected %d", resLen, expected)
	}
}
