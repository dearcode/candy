package main

import (
	"fmt"

	"github.com/biogo/store/llrb"
)

const (
	defaultEnd = 9973
)

//Region 用户ID区间begin~end
type Region struct {
	Begin  int
	End    int
	Server string
}

func (r Region) Compare(b llrb.Comparable) int {
	if b.(Region).Begin == r.Begin {
		return 0
	}

	if b.(Region).Begin < r.Begin && r.Begin < b.(Region).End {
		return 0
	}

	return b.(Region).Begin - r.Begin
}

func main() {
	values := []Region{
		{0, 5, ""},
		{5, 10, ""},
		{10, 15, ""},
		{15, 20, ""},
		{25, 100, ""},
	}

	t := &llrb.Tree{}
	for _, v := range values {
		t.Insert(v)
	}
	for i := 0; i < 30; i++ {
		c := t.Get(Region{i, 0, ""})
		if c == nil {
			fmt.Printf("get %d not find\n", i)
			continue
		}
		r := c.(Region)
		fmt.Printf("get %d find:%v\n", i, r)
	}
}
