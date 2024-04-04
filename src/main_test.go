package main

import "testing"

func TestHello(t *testing.T) {
	str := "xxx"
	if str != "xxx" {
		t.Fail()
	}
}
