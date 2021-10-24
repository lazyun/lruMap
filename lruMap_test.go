package lruMap

import (
	"testing"
	"time"
)

func Test2(t *testing.T) {
	m := New(2)

	m.Set("qwe", "123")
	m.Set("asd", "124")
	m.Set("qwe", "125")
	m.Set("zxc", "126")
	<-time.After(5 * time.Second)

	RangePrint(m)
}
