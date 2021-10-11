// Package intervals provides an insertion-sort based solution to the merge
// overlapping intervals problem.
package intervals

import (
	"sort"
)

type Interval struct {
	Start int64
	End   int64
}

func (v Interval) Empty() bool {
	if v.Start < v.End {
		return false
	}

	if v.Start > v.End {
		panic("start can't be after end")
	}

	return true
}

type Intervals []Interval

func (ivs Intervals) Insert(v Interval) Intervals {
	if v.Empty() {
		return ivs
	}

	// Find the first interval that matters, whose end is at least our
	// start.  Everything before that won't be changing.
	skip := sort.Search(len(ivs), func(i int) bool { return ivs[i].End >= v.Start })

	// If there are any intervals that v overlaps with, merge their
	// range into v and consume them from after.
	after := ivs[skip:]
	for len(after) > 0 && v.End >= after[0].Start {
		if after[0].Start < v.Start {
			v.Start = after[0].Start
		}
		if after[0].End > v.End {
			v.End = after[0].End
		}
		after = after[1:]
	}

	// If we are going to outgrow cap(ivs), use make+copy instead of
	// append.  Append will have to double-allocate to avoid overwriting
	// "after".
	outlen := skip + 1 + len(after)
	if outlen > cap(ivs) {
		ovs := make(Intervals, outlen, outlen*2)

		copy(ovs, ivs[:skip])
		ovs[skip] = v
		copy(ovs[skip+1:], after)

		return ovs
	}

	// Otherwise we can re-use this slice.  If we are shrinking or
	// growing, after needs to move.
	if outlen != len(ivs) {
		ivs = ivs[:outlen]
		copy(ivs[skip+1:], after)
	}

	ivs[skip] = v
	return ivs
}
