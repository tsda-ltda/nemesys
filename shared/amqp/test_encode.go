package amqp

import "testing"

func TestEncode(t *testing.T) {
	d := struct{ data int }{data: 100}
	b, err := Encode(d)
	if err != nil {
		t.Errorf("fail to encode, err: %s", err)
		return
	}
	if len(b) < 1 {
		t.Errorf("returned bytes are empty")
	}
}

func TestDecode(t *testing.T) {
	d := struct{ data int }{data: 100}
	b, err := Encode(d)
	if err != nil {
		t.Errorf("fail to encode, err: %s", err)
		return
	}
	if len(b) < 1 {
		t.Errorf("returned bytes are empty")
	}

	var r struct{ data int }
	err = Decode(b, &r)
	if err != nil {
		t.Errorf("fail to decode, err: %s", err)
		return
	}
	if r.data != 100 {
		t.Errorf("data returned is wrong, expected: %d, got: %d", 100, r.data)
	}
}
