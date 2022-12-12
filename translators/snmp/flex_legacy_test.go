package snmp

import "testing"

func TestGetFlexLegacyAlarmOID(t *testing.T) {
	tests := []struct {
		oid      string
		err      error
		expected string
	}{
		{
			oid:      "invalid oid",
			err:      ErrNoAlarmOID,
			expected: "",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.1.2.0",
			err:      ErrNoAlarmOID,
			expected: "",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.1.1.4.2",
			err:      ErrNoAlarmOID,
			expected: "",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.2.1.17",
			err:      ErrNoAlarmOID,
			expected: "",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.2.1.14.23",
			err:      nil,
			expected: ".1.3.6.1.4.1.31957.1.3.2.1.13.23",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.3.1.9.1",
			err:      nil,
			expected: ".1.3.6.1.4.1.31957.1.3.3.1.10.1",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.4.1.4.2",
			err:      nil,
			expected: ".1.3.6.1.4.1.31957.1.3.4.1.11.2",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.6.1.14.100",
			err:      nil,
			expected: ".1.3.6.1.4.1.31957.1.3.6.1.13.100",
		},
	}
	for _, test := range tests {
		r, err := getFlexLegacyAlarmOID(test.oid)
		if err != test.err {
			t.Errorf("\noid: %s\n err expected: %s\n got err: %s", test.oid, test.err, err)
			return
		}
		if r != test.expected {
			t.Errorf("\noid: %s\n oid expected: %s\n got: %s", test.oid, test.expected, r)
		}
	}
}

func TestGetFlexLegacyCategoryOID(t *testing.T) {
	tests := []struct {
		oid      string
		err      error
		expected string
	}{
		{
			oid:      "invalid oid",
			err:      ErrNoAlarmOID,
			expected: "",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.1.2.0",
			err:      ErrNoAlarmOID,
			expected: "",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.1.1.4.2",
			err:      ErrNoAlarmOID,
			expected: "",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.2.1.17",
			err:      ErrNoAlarmOID,
			expected: "",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.2.1.14.23",
			err:      nil,
			expected: ".1.3.6.1.4.1.31957.1.3.2.1.12.23",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.3.1.9.1",
			err:      nil,
			expected: ".1.3.6.1.4.1.31957.1.3.3.1.8.1",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.4.1.4.2",
			err:      nil,
			expected: ".1.3.6.1.4.1.31957.1.3.4.1.8.2",
		},
		{
			oid:      ".1.3.6.1.4.1.31957.1.3.6.1.14.100",
			err:      nil,
			expected: ".1.3.6.1.4.1.31957.1.3.6.1.12.100",
		},
	}
	for _, test := range tests {
		r, err := getFlexLegacyCategoryOID(test.oid)
		if err != test.err {
			t.Errorf("\noid: %s\n err expected: %s\n got err: %s", test.oid, test.err, err)
			return
		}
		if r != test.expected {
			t.Errorf("\noid: %s\n oid expected: %s\n got: %s", test.oid, test.expected, r)
		}
	}
}
