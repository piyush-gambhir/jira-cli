package cmd

import "testing"

func TestTruthy(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{{"true", true}, {"TRUE", true}, {"1", true}, {"false", false}, {"0", false}, {"anything", false}} {
		if got := truthy(tc.value); got != tc.want {
			t.Errorf("truthy(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
