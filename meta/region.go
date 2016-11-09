package meta

import (
	"github.com/biogo/store/llrb"
)

const (
    //MaxRegionEnd 最大region
	MaxRegionEnd = 9973
)

//NewRegion init region.
func NewRegion(host string, begin, end int32) *Region {
	return &Region{Begin: begin, End: end, Host: host}
}

//Span 区间的跨度
func (r *Region) Span() int32 {
	return r.End - r.Begin
}

//Max 两个Region最大范围
func (r *Region) Max(c Region) (b int32, e int32) {
	b, e = r.Begin, r.End

	if b > c.Begin {
		b = c.Begin
	}

	if e < c.End {
		e = c.End
	}

	return
}

//Match 用户与region匹配
func (r *Region) Match(id int64) bool {
	i := int32(id % MaxRegionEnd)
	if i >= r.Begin && i < r.End {
		return true
	}
	return false
}

//Compare for llrb.
func (r Region) Compare(c llrb.Comparable) int {
	b := c.(*Region)
	if b.Begin == r.Begin {
		return 0
	}

	if b.Begin < r.Begin && r.Begin < b.End {
		return 0
	}

	return int(b.Begin - r.Begin)
}
