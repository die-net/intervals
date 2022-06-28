// Package intervals provides a fast insertion-sort based solution to the
// merge overlapping intervals problem.
package intervals

import (
	"sort"
)

// Interval represents range of the form [Start, End), which contains all
// integer x in Start <= x < End.
type Interval struct {
	Start int64
	End   int64
}

// Empty returns true if Start is less than end, false if they are equal,
// and panics if Start is greater than End.
func (v Interval) Empty() bool {
	if v.Start < v.End {
		return false
	}

	if v.Start > v.End {
		panic("start can't be after end")
	}

	return true
}

// Intervals is an ordered representation of a slice of Interval.
type Intervals []Interval

// Search will return the Interval in the given Intervals containing the
// value of off and true if found.  Otherwise, it will return an empty
// Interval and false.
func (ivs Intervals) Search(off int64) (Interval, bool) {
	// Skip search if we definitely don't contain offset.
	if len(ivs) == 0 || ivs[0].Start > off || ivs[len(ivs)-1].End < off {
		return Interval{}, false
	}

	// Find the first interval that matches, whose end is at least off.
	i := sort.Search(len(ivs), func(i int) bool { return ivs[i].End >= off })
	if i < len(ivs) && ivs[i].Start <= off {
		return ivs[i], true
	}

	return Interval{}, false
}

// Insert adds a given Interval to Intervals, possibly inserting in the
// correct order between two intervals, extending an existing interval, or
// compacting existing intervals as necessary.  ivs must be correctly
// ordered.  An empty v will result in a noop, and an invalid v may panic.
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
