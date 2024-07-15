// Package intervals provides a fast insertion-sort based solution to the
// merge overlapping intervals problem.
package intervals

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// Interval represents a range of the form `[Start, End)`, which contains
// all x in Start <= x < End.  x can be any type that is ordered, including
// the various sizes of int, uint, float types, uintptr and string.
//
// If T is a float, neither Start nor End can be NaN, since NaN is not
// ordered.
type Interval[T constraints.Ordered] struct {
	Start T
	End   T
}

// Empty returns true if Start is less than end, false if they are equal,
// and panics if Start is greater than End.
//
// If T is a float and Start and/or End are NaN, this will return true.
func (v Interval[T]) Empty() bool {
	if v.Start < v.End {
		return false
	}

	if v.Start > v.End {
		panic("start can't be after end")
	}

	return true
}

// Intervals is an ascending ordered representation of a slice of
// non-overlapping Interval.
type Intervals[T constraints.Ordered] []Interval[T]

// Search will return the Interval in the given Intervals containing the
// value of off and true if found.  Otherwise, it will return an empty
// Interval and false.  ivs must be correctly ordered.
//
// Search is not safe for concurrent access with Interval; add a read lock
// if necessary.
func (ivs Intervals[T]) Search(off T) (Interval[T], bool) {
	// Skip search if we definitely don't contain offset.
	if len(ivs) == 0 || ivs[0].Start > off || ivs[len(ivs)-1].End < off {
		return Interval[T]{}, false
	}

	// Find the first interval that matches, whose end is greater than offset.
	i := sort.Search(len(ivs), func(i int) bool { return ivs[i].End > off })
	if i < len(ivs) && ivs[i].Start <= off {
		return ivs[i], true
	}

	return Interval[T]{}, false
}

// Insert adds a given Interval to Intervals, possibly inserting in the
// correct order between two intervals, extending an existing interval, or
// compacting existing intervals as necessary.  ivs must be correctly
// ordered.  An empty v will result in a noop, and an invalid v may panic.
//
// Insert is not safe for concurrent access; add a lock if necessary.
func (ivs Intervals[T]) Insert(v Interval[T]) Intervals[T] {
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
		ovs := make(Intervals[T], outlen, outlen*2)

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
