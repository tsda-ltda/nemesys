package types

import (
	"testing"
)

func TestParseBool(t *testing.T) {
	tests := []struct {
		Name    string
		Want    any
		WantErr bool
		Type    MetricType
		Value   bool
	}{
		{
			Name:    "Test 'true' to string",
			Want:    "true",
			WantErr: false,
			Type:    MTString,
			Value:   true,
		},
		{
			Name:    "Test 'false' to string",
			Want:    "false",
			WantErr: false,
			Type:    MTString,
			Value:   false,
		},
		{
			Name:    "Test 'true' to bool",
			Want:    true,
			WantErr: false,
			Type:    MTBool,
			Value:   true,
		},
		{
			Name:    "Test 'false' to bool",
			Want:    false,
			WantErr: false,
			Type:    MTBool,
			Value:   false,
		},
		{
			Name:    "Test 'true' to integer",
			Want:    int64(1),
			WantErr: false,
			Type:    MTInt,
			Value:   true,
		},
		{
			Name:    "Test 'false' to integer",
			Want:    int64(0),
			WantErr: false,
			Type:    MTInt,
			Value:   false,
		},
		{
			Name:    "Test 'false' to float",
			Want:    float64(0),
			WantErr: false,
			Type:    MTFloat,
			Value:   false,
		},
	}

	// run tests
	for _, test := range tests {
		r, err := ParseBool(test.Value, test.Type)
		if (err != nil) != test.WantErr {
			t.Errorf("Test: %s, WantErr: %v, got err: %s", test.Name, test.WantErr, err)
		}
		if r != test.Want {
			t.Errorf("Test: %s, Want: %v, got: %v", test.Name, test.Want, r)
		}
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		Name    string
		Want    any
		WantErr bool
		Type    MetricType
		Value   string
	}{
		{
			Name:    "Test 'True' to bool",
			Want:    true,
			WantErr: false,
			Type:    MTBool,
			Value:   "True",
		},
		{
			Name:    "Test 'False' to bool",
			Want:    false,
			WantErr: false,
			Type:    MTBool,
			Value:   "false",
		},
		{
			Name:    "Test '1' to bool",
			Want:    true,
			WantErr: false,
			Type:    MTBool,
			Value:   "1",
		},
		{
			Name:    "Test '0' to bool",
			Want:    false,
			WantErr: false,
			Type:    MTBool,
			Value:   "0",
		},
		{
			Name:    "Test '10.5' to integer",
			Want:    int64(10),
			WantErr: false,
			Type:    MTInt,
			Value:   "10",
		},
		{
			Name:    "Test '15,8' to float",
			Want:    float64(15.8),
			WantErr: false,
			Type:    MTFloat,
			Value:   "15,8",
		},
	}

	// run tests
	for _, test := range tests {
		r, err := ParseString(test.Value, test.Type)
		if (err != nil) != test.WantErr {
			t.Errorf("Test: %s, WantErr: %v, got err: %s", test.Name, test.WantErr, err)
		}
		if r != test.Want {
			t.Errorf("Test: %s, Want: %v, got: %v", test.Name, test.Want, r)
		}
	}
}

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		Name    string
		Want    any
		WantErr bool
		Type    MetricType
		Value   float64
	}{
		{
			Name:    "Test '0.1' to bool",
			Want:    false,
			WantErr: false,
			Type:    MTBool,
			Value:   0.1,
		},
		{
			Name:    "Test '1.1' to bool",
			Want:    true,
			WantErr: false,
			Type:    MTBool,
			Value:   1.1,
		},
		{
			Name:    "Test '23.87002' to string",
			Want:    "23.87002",
			WantErr: false,
			Type:    MTString,
			Value:   23.87002,
		},
		{
			Name:    "Test '99933.3232' to integer",
			Want:    int64(99933),
			WantErr: false,
			Type:    MTInt,
			Value:   99933.3232,
		},
		{
			Name:    "Test '21.332' to float",
			Want:    21.332,
			WantErr: false,
			Type:    MTFloat,
			Value:   21.332,
		},
	}

	// run tests
	for _, test := range tests {
		r, err := ParseFloat64(test.Value, test.Type)
		if (err != nil) != test.WantErr {
			t.Errorf("Test: %s, WantErr: %v, got err: %s", test.Name, test.WantErr, err)
		}
		if r != test.Want {
			t.Errorf("Test: %s, Want: %v, got: %v", test.Name, test.Want, r)
		}
	}
}
