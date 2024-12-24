package util

import "testing"

func TestCheckImagePath(t *testing.T) {
	tests := []struct {
		image    string
		expected bool
	}{
		{"ubuntu:latest", true},
		{"myorg/myrepo/myimage:v1", true},
		{"myrepo/myimage", true},
		{"invalid/image@latest", false},
		{"1234/valid_image:v1", true},
		{"myorg/myrepo/invalid@tag", false},
		{"myorg/myrepo/myimage", true},
		{"invalid@tag", false},
	}
	for _, test := range tests {
		result := FmtCheck.CheckImagePath(test.image)
		if result != test.expected {
			t.Errorf("Expected %t for image %s, got %t", test.expected, test.image, result)
		}
	}
}
