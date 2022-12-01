package influxdb

import (
	"testing"
	"time"
)

func BenchmarkParseDuration(b *testing.B) {
	d := "72ms"
	for i := 0; i < b.N; i++ {
		ParseDuration(d)
	}
}

func TestParseDuration(t *testing.T) {
	tests := map[string]time.Duration{
		"1h":   time.Hour,
		"2m":   time.Minute * 2,
		"76ms": time.Millisecond * 76,
		"30d":  time.Hour * 24 * 30,
	}
	for k, d := range tests {
		result, err := ParseDuration(k)
		if err != nil {
			t.Errorf("Fail to parse %s", k)
		}
		if result == d {
			continue
		}
		t.Errorf("ParseDuration failed, expected: %v, got: %v", d, result)
	}
}
