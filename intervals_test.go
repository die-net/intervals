package intervals

import (
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"
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
		if empty := test.interval.Empty(); empty != test.empty {
			t.Errorf("Empty(%v) = %v, want %v", test.interval, empty, test.empty)
		}
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
		if v != test.expected || ok != test.ok {
			t.Errorf("%v.Search(%d) = (%v, %v), want (%v, %v)", test.intervals, test.offset, v, ok, test.expected, test.ok)
		}
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
		if !slices.Equal(vs, test.expected) {
			t.Errorf("inserting %v = %v, want %v", test.inserts, vs, test.expected)
		}
	}
}

func TestInsertRandom(t *testing.T) {
	for range 1000 {
		vs := Intervals[int64]{}
		count := rand.IntN(10) + 1 //nolint:gosec // Not security sensitive.
		start := int64(10000)
		end := int64(-1)
		for range count {
			s := rand.Int64N(1024)    //nolint:gosec // Not security sensitive.
			e := s + rand.Int64N(128) //nolint:gosec // Not security sensitive.
			v := Interval[int64]{s, e}
			vs = vs.Insert(v)

			if s == e {
				continue
			}

			// Track our own idea of the first start and verify it.
			if s < start {
				start = s
			}
			if vs[0].Start != start {
				t.Fatalf("vs[0].Start = %d, want %d", vs[0].Start, start)
			}

			// Track our own idea of the last end and verify it.
			if e > end {
				end = e
			}
			if vs[len(vs)-1].End != end {
				t.Fatalf("vs[%d].End = %d, want %d", len(vs)-1, vs[len(vs)-1].End, end)
			}
		}

		if len(vs) > count {
			t.Fatalf("len(vs) = %d, want <= %d", len(vs), count)
		}

		for i := range vs {
			// Make sure each of the records looks legit.
			if vs[i].Start >= vs[i].End {
				t.Fatalf("vs[%d] = %v, want Start < End", i, vs[i])
			}

			// And there's a gap to the next one.
			if i < len(vs)-1 && vs[i].End >= vs[i+1].Start {
				t.Fatalf("vs[%d].End = %d >= vs[%d].Start = %d, want a gap", i, vs[i].End, i+1, vs[i+1].Start)
			}
		}
	}
}

func TestInsertPanic(t *testing.T) {
	// End before start should panic.
	assertPanics(t, func() { Intervals[int64]{}.Insert(Interval[int64]{2, 1}) })
}

// assertPanics fails the test if fn does not panic when called.
func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected panic, but function returned normally")
		}
	}()
	fn()
}

func BenchmarkInsertDisjoint(b *testing.B) {
	benchInsert(b, 1024, 1, 0)
}

func BenchmarkInsertMergeable(b *testing.B) {
	benchInsert(b, 1024, 0, 0)
}

func BenchmarkInsertOverlapping(b *testing.B) {
	benchInsert(b, 1024, 0, 10240)
}

func benchInsert(b *testing.B, step, gap, overlap int64) {
	b.ReportAllocs()
	for _, num := range []int{1, 10, 100, 1000, 10000} {
		b.Run(strconv.Itoa(num), func(sb *testing.B) {
			sb.SetBytes(1)
			sb.RunParallel(func(pb *testing.PB) {
				ovs := make(Intervals[int64], 0, num)

				ivs := make(Intervals[int64], num)
				n := len(ivs)

				for pb.Next() {
					n++
					if n >= len(ivs) {
						ovs = ovs[:0]
						randIntervals(ivs, step, gap, overlap)
						n = 0
					}

					ovs = ovs.Insert(ivs[n])
				}
			})
		})
	}
}

func randIntervals(vs Intervals[int64], step, gap, overlap int64) {
	s := int64(0)
	for i := range vs {
		l := rand.Int64N(step) //nolint:gosec // Not security sensitive.
		e := s + l

		v := Interval[int64]{s, e}
		if overlap > 0 {
			v.End += rand.Int64N(overlap) //nolint:gosec // Not security sensitive.
		}
		vs[i] = v

		s = e + gap
	}

	rand.Shuffle(len(vs), func(i, j int) {
		vs[i], vs[j] = vs[j], vs[i]
	})
}

func BenchmarkSearchDisjoint(b *testing.B) {
	benchSearch(b, 1024, 1, 0)
}

func BenchmarkSearchMerged(b *testing.B) {
	benchSearch(b, 1024, 0, 0)
}

func benchSearch(b *testing.B, step, gap, overlap int64) {
	b.ReportAllocs()
	for _, num := range []int{1, 10, 100, 1000, 10000} {
		b.Run(strconv.Itoa(num), func(sb *testing.B) {
			sb.SetBytes(1)
			sb.RunParallel(func(pb *testing.PB) {
				ivs := make(Intervals[int64], num)
				randIntervals(ivs, step, gap, overlap)

				ovs := make(Intervals[int64], 0, len(ivs))
				for _, iv := range ivs {
					ovs = ovs.Insert(iv)
				}

				start := ovs[0].Start
				end := ovs[len(ovs)-1].End

				n := end

				for pb.Next() {
					n++
					if n >= end {
						n = start
					}

					_, _ = ovs.Search(n)
				}
			})
		})
	}
}
