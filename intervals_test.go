package intervals

import (
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	tests := []struct {
		interval Interval[float64]
		empty    bool
	}{
		{interval: Interval[float64]{0.0, 0.0}, empty: true},
		{interval: Interval[float64]{math.Copysign(0, -1), +0.0}, empty: true},
		{interval: Interval[float64]{0.0, math.SmallestNonzeroFloat64}, empty: false},
		{interval: Interval[float64]{0.0, 1.0}, empty: false},
		{interval: Interval[float64]{-1.0, -1.0}, empty: true},
		{interval: Interval[float64]{-1.0, 0.0}, empty: false},
		{interval: Interval[float64]{math.Inf(-1), -1}, empty: false},
		{interval: Interval[float64]{1, math.Inf(1)}, empty: false},

		// NaN can't be compared to anything
		{interval: Interval[float64]{math.NaN(), 1}, empty: true},
		{interval: Interval[float64]{1, math.NaN()}, empty: true},
		{interval: Interval[float64]{math.NaN(), math.NaN()}, empty: true},
	}

	for _, test := range tests {
		empty := test.interval.Empty()
		assert.Equal(t, test.empty, empty)
	}
}

func TestSearchExpected(t *testing.T) {
	tests := []struct {
		intervals Intervals[int64]
		offset    int64
		expected  Interval[int64]
		ok        bool
	}{
		{intervals: Intervals[int64]{{0, 2}, {4, 5}}, offset: -1, expected: Interval[int64]{0, 0}, ok: false},
		{intervals: Intervals[int64]{{0, 2}, {4, 5}}, offset: 0, expected: Interval[int64]{0, 2}, ok: true},
		{intervals: Intervals[int64]{{0, 2}, {4, 5}}, offset: 1, expected: Interval[int64]{0, 2}, ok: true},
		{intervals: Intervals[int64]{{0, 2}, {4, 5}}, offset: 2, expected: Interval[int64]{0, 0}, ok: false},
		{intervals: Intervals[int64]{{0, 2}, {4, 5}}, offset: 3, expected: Interval[int64]{0, 0}, ok: false},
		{intervals: Intervals[int64]{{0, 2}, {4, 5}}, offset: 4, expected: Interval[int64]{4, 5}, ok: true},
		{intervals: Intervals[int64]{{0, 2}, {4, 5}}, offset: 5, expected: Interval[int64]{0, 0}, ok: false},
		{intervals: Intervals[int64]{{0, 2}, {4, 5}}, offset: 6, expected: Interval[int64]{0, 0}, ok: false},
	}

	for _, test := range tests {
		v, ok := test.intervals.Search(test.offset)
		assert.Equal(t, test.expected, v)
		assert.Equal(t, test.ok, ok)
	}
}

func TestInsertExpected(t *testing.T) {
	tests := []struct {
		inserts  Intervals[int64]
		expected Intervals[int64]
	}{
		// Merge two adjacent things
		{inserts: Intervals[int64]{{0, 2}, {2, 3}}, expected: Intervals[int64]{{0, 3}}},
		// Same thing, reversed
		{inserts: Intervals[int64]{{2, 3}, {0, 2}}, expected: Intervals[int64]{{0, 3}}},
		// A gap shouldn't be merged
		{inserts: Intervals[int64]{{0, 4}, {5, 8}, {10, 12}}, expected: Intervals[int64]{{0, 4}, {5, 8}, {10, 12}}},
		// Filling in the gap should work fine
		{inserts: Intervals[int64]{{2, 6}, {9, 11}, {6, 9}}, expected: Intervals[int64]{{2, 11}}},
		// Replacing some or all of them is fine
		{inserts: Intervals[int64]{{1, 4}, {10, 12}, {4, 16}}, expected: Intervals[int64]{{1, 16}}},
		{inserts: Intervals[int64]{{0, 3}, {10, 12}, {2, 12}}, expected: Intervals[int64]{{0, 12}}},
		{inserts: Intervals[int64]{{1, 4}, {10, 12}, {0, 7}}, expected: Intervals[int64]{{0, 7}, {10, 12}}},
		// Inserting nothing gets nothing
		{inserts: Intervals[int64]{{12, 12}}, expected: Intervals[int64]{}},
	}

	for _, test := range tests {
		vs := Intervals[int64]{}
		for _, v := range test.inserts {
			vs = vs.Insert(Interval[int64]{v.Start, v.End})
		}
		assert.Equal(t, test.expected, vs)
	}
}

func TestInsertRandom(t *testing.T) {
	for n := 0; n < 1000; n++ {
		vs := Intervals[int64]{}
		count := rand.Intn(10) + 1
		start := int64(10000)
		end := int64(-1)
		for i := 0; i < count; i++ {
			s := rand.Int63n(1024)
			e := s + rand.Int63n(128)
			v := Interval[int64]{s, e}
			vs = vs.Insert(v)

			if s == e {
				continue
			}

			// Track our own idea of the first start and verify it.
			if s < start {
				start = s
			}
			assert.Equal(t, start, vs[0].Start)

			// Track our own idea of the last end and verify it.
			if e > end {
				end = e
			}
			assert.Equal(t, end, vs[len(vs)-1].End)
		}

		assert.LessOrEqual(t, len(vs), count)

		for i := 0; i < len(vs); i++ {
			// Make sure each of the records looks legit.
			assert.Less(t, vs[i].Start, vs[i].End)

			// And there's a gap to the next one.
			if i < len(vs)-1 {
				assert.Less(t, vs[i].End, vs[i+1].Start)
			}
		}
	}
}

func TestInsertPanic(t *testing.T) {
	// End before start should panic.
	assert.Panics(t, func() { Intervals[int64]{}.Insert(Interval[int64]{2, 1}) })
}

func BenchmarkInsertNonOverlapping(b *testing.B) {
	benchInsert(b, 1024, 0)
}

func BenchmarkInsertOverlapping(b *testing.B) {
	benchInsert(b, 1024, 10240)
}

func benchInsert(b *testing.B, step, overlap int64) {
	b.ReportAllocs()
	for _, num := range []int{1, 10, 100, 1000, 10000} {
		b.Run(strconv.Itoa(num), func(sb *testing.B) {
			sb.RunParallel(func(pb *testing.PB) {
				ovs := make(Intervals[int64], 0, num)

				ivs := make(Intervals[int64], num)
				n := len(ivs)

				for pb.Next() {
					n++
					if n >= len(ivs) {
						ovs = ovs[:0]
						randIntervals(ivs, step, overlap)
						n = 0
					}

					ovs = ovs.Insert(ivs[n])
				}
			})
		})
	}
}

func randIntervals(vs Intervals[int64], step, overlap int64) {
	s := int64(0)
	for i := 0; i < len(vs); i++ {
		l := rand.Int63n(step)
		e := s + l

		v := Interval[int64]{s, e}
		if overlap > 0 {
			v.End += rand.Int63n(overlap)
		}
		vs[i] = v

		s = e
	}

	rand.Shuffle(len(vs), func(i, j int) {
		vs[i], vs[j] = vs[j], vs[i]
	})
}
