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
			Name:    "Test 'true' to int8",
			Want:    int8(1),
			WantErr: false,
			Type:    MTInt8,
			Value:   true,
		},
		{
			Name:    "Test 'false' to int8",
			Want:    int8(0),
			WantErr: false,
			Type:    MTInt8,
			Value:   false,
		},
		{
			Name:    "Test 'true' to int16",
			Want:    int16(1),
			WantErr: false,
			Type:    MTInt16,
			Value:   true,
		},
		{
			Name:    "Test 'false' to int16",
			Want:    int16(0),
			WantErr: false,
			Type:    MTInt16,
			Value:   false,
		},
		{
			Name:    "Test 'true' to int32",
			Want:    int32(1),
			WantErr: false,
			Type:    MTInt32,
			Value:   true,
		},
		{
			Name:    "Test 'false' to int32",
			Want:    int32(0),
			WantErr: false,
			Type:    MTInt32,
			Value:   false,
		},
		{
			Name:    "Test 'true' to int64",
			Want:    int64(1),
			WantErr: false,
			Type:    MTInt64,
			Value:   true,
		},
		{
			Name:    "Test 'false' to int64",
			Want:    int64(0),
			WantErr: false,
			Type:    MTInt64,
			Value:   false,
		},
		{
			Name:    "Test 'false' to float64",
			Want:    float64(0),
			WantErr: false,
			Type:    MTFloat64,
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
			Name:    "Test '10,1' to int8",
			Want:    int8(10),
			WantErr: false,
			Type:    MTInt8,
			Value:   "10",
		},
		{
			Name:    "Test '10.5' to int16",
			Want:    int16(10),
			WantErr: false,
			Type:    MTInt16,
			Value:   "10",
		},
		{
			Name:    "Test '10.5' to int32",
			Want:    int32(10),
			WantErr: false,
			Type:    MTInt32,
			Value:   "10",
		},
		{
			Name:    "Test '10.5' to int64",
			Want:    int64(10),
			WantErr: false,
			Type:    MTInt64,
			Value:   "10",
		},
		{
			Name:    "Test '15,8' to float64",
			Want:    float64(15.8),
			WantErr: false,
			Type:    MTFloat64,
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
			Name:    "Test '-12.009' to int8",
			Want:    int8(-12),
			WantErr: false,
			Type:    MTInt8,
			Value:   -12.009,
		},
		{
			Name:    "Test '32766.5' to int16",
			Want:    int16(32767),
			WantErr: false,
			Type:    MTInt16,
			Value:   32_767.5,
		},
		{
			Name:    "Test '200_000_000.6' to int32",
			Want:    int32(200_000_000),
			WantErr: false,
			Type:    MTInt32,
			Value:   200_000_000.5,
		},
		{
			Name:    "Test '99933.3232' to int64",
			Want:    int64(99933),
			WantErr: false,
			Type:    MTInt64,
			Value:   99933.3232,
		},
		{
			Name:    "Test '21.332' to float64",
			Want:    21.332,
			WantErr: false,
			Type:    MTFloat64,
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
