package intervals

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestSearchExpected(t *testing.T) {
	tests := []struct {
		intervals Intervals
		offset    int64
		expected  Interval
		ok        bool
	}{
		{intervals: Intervals{{0, 2}, {4, 5}}, offset: -1, expected: Interval{0, 0}, ok: false},
		{intervals: Intervals{{0, 2}, {4, 5}}, offset: 0, expected: Interval{0, 2}, ok: true},
		{intervals: Intervals{{0, 2}, {4, 5}}, offset: 1, expected: Interval{0, 2}, ok: true},
		{intervals: Intervals{{0, 2}, {4, 5}}, offset: 2, expected: Interval{0, 2}, ok: true},
		{intervals: Intervals{{0, 2}, {4, 5}}, offset: 3, expected: Interval{0, 0}, ok: false},
		{intervals: Intervals{{0, 2}, {4, 5}}, offset: 4, expected: Interval{4, 5}, ok: true},
		{intervals: Intervals{{0, 2}, {4, 5}}, offset: 5, expected: Interval{4, 5}, ok: true},
		{intervals: Intervals{{0, 2}, {4, 5}}, offset: 6, expected: Interval{0, 0}, ok: false},
	}

	for _, test := range tests {
		v, ok := test.intervals.Search(test.offset)
		assert.Equal(t, test.expected, v)
		assert.Equal(t, test.ok, ok)
	}
}

func TestInsertExpected(t *testing.T) {
	tests := []struct {
		inserts  Intervals
		expected Intervals
	}{
		// Merge two adjacent things
		{inserts: Intervals{{0, 2}, {2, 3}}, expected: Intervals{{0, 3}}},
		// Same thing, reversed
		{inserts: Intervals{{2, 3}, {0, 2}}, expected: Intervals{{0, 3}}},
		// A gap shouldn't be merged
		{inserts: Intervals{{0, 4}, {5, 8}, {10, 12}}, expected: Intervals{{0, 4}, {5, 8}, {10, 12}}},
		// Filling in the gap should work fine
		{inserts: Intervals{{2, 6}, {9, 11}, {6, 9}}, expected: Intervals{{2, 11}}},
		// Replacing some or all of them is fine
		{inserts: Intervals{{1, 4}, {10, 12}, {4, 16}}, expected: Intervals{{1, 16}}},
		{inserts: Intervals{{0, 3}, {10, 12}, {2, 12}}, expected: Intervals{{0, 12}}},
		{inserts: Intervals{{1, 4}, {10, 12}, {0, 7}}, expected: Intervals{{0, 7}, {10, 12}}},
		// Inserting nothing gets nothing
		{inserts: Intervals{{12, 12}}, expected: Intervals{}},
	}

	for _, test := range tests {
		vs := Intervals{}
		for _, v := range test.inserts {
			vs = vs.Insert(Interval{v.Start, v.End})
		}
		assert.Equal(t, test.expected, vs)
	}
}

func TestInsertRandom(t *testing.T) {
	for n := 0; n < 1000; n++ {
		vs := Intervals{}
		count := rand.Intn(10) + 1
		start := int64(10000)
		end := int64(-1)
		for i := 0; i < count; i++ {
			s := rand.Int63n(1024)
			e := s + rand.Int63n(128)
			v := Interval{s, e}
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
	assert.Panics(t, func() { Intervals{}.Insert(Interval{2, 1}) })
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
				r := rand.New(rand.NewSource(time.Now().UnixNano()))

				ovs := make(Intervals, 0, num)

				ivs := make(Intervals, num)
				n := len(ivs)

				for pb.Next() {
					n++
					if n >= len(ivs) {
						ovs = ovs[:0]
						randIntervals(r, ivs, step, overlap)
						n = 0
					}

					ovs = ovs.Insert(ivs[n])
				}
			})
		})
	}
}

func randIntervals(r *rand.Rand, vs Intervals, step, overlap int64) {
	s := int64(0)
	for i := 0; i < len(vs); i++ {
		l := r.Int63n(step)
		e := s + l

		v := Interval{s, e}
		if overlap > 0 {
			v.End += r.Int63n(overlap)
		}
		vs[i] = v

		s = e
	}

	r.Shuffle(len(vs), func(i, j int) {
		vs[i], vs[j] = vs[j], vs[i]
	})
}
