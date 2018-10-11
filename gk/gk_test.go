package gk

import "testing"

func TestGetIP(t *testing.T) {
	ip, err := GetIP()
	if ip == "" || err != nil {
		t.Fail()
	}
}
