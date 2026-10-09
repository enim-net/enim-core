package typ

import "testing"

func TestColor(t *testing.T) {
	var c Color
	c.SetDefault()
	if c != "#000000" {
		t.Fatal(c)
	}
	if v, _ := Color("").Value(); v != "#000000" {
		t.Fatal(v)
	}
	for _, in := range []any{"#fff", []byte("#fff")} {
		var s Color
		if err := s.Scan(in); err != nil || s != "#fff" {
			t.Fatal(s, err)
		}
	}
	var n Color
	_ = n.Scan(nil)
	if n != "#000000" {
		t.Fatal(n)
	}
}
