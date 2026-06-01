Intervals [![Build Status](https://github.com/die-net/intervals/actions/workflows/go-test.yml/badge.svg)](https://github.com/die-net/intervals/actions/workflows/go-test.yml) [![Coverage Status](https://coveralls.io/repos/github/die-net/intervals/badge.svg?branch=main)](https://coveralls.io/github/die-net/intervals?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/die-net/intervals)](https://goreportcard.com/report/github.com/die-net/intervals)
=========

Package `intervals` provides a fast, insertion-sort based, zero-dependency
solution to the "merge overlapping intervals" problem.  It keeps a slice of
ranges sorted and non-overlapping as you insert new ranges, and lets you
quickly look up which range (if any) contains a given value.

The package is generic: it works with any [ordered](https://pkg.go.dev/cmp#Ordered)
type, including the various sizes of `int`, `uint`, and `float`, as well as
`uintptr` and `string`.

Concepts
--------

### Interval

An `Interval[T]` represents a half-open range of the form `[Start, End)`,
which contains all `x` where `Start <= x < End`. Half-open ranges compose
cleanly: the `End` of one interval can equal the `Start` of the next without
the two values overlapping.

```go
type Interval[T cmp.Ordered] struct {
    Start T
    End   T
}
```

An interval is considered *empty* when `Start == End`. An interval where
`Start > End` is invalid and will panic when inserted. If `T` is a float,
neither bound may be `NaN` (because `NaN` is unordered); a `NaN` bound is
treated as empty.

### Intervals

`Intervals[T]` is a slice of `Interval[T]` kept in ascending order with no two
intervals overlapping. You build and maintain it through the `Insert` method,
which:

- inserts a new interval in sorted position, or
- extends an adjacent or overlapping interval, or
- merges and compacts any intervals the new range now spans.

Two intervals are merged when the `End` of one is `>=` the `Start` of the next,
so adjacent ranges like `[0, 2)` and `[2, 3)` collapse into `[0, 3)`.

API
---

| Method | Description |
| --- | --- |
| `(Interval[T]) Empty() bool` | Reports whether the interval is empty (`Start == End`). Panics if `Start > End`. |
| `(Intervals[T]) Insert(v Interval[T]) Intervals[T]` | Returns a new `Intervals` with `v` added, keeping order and merging overlaps. An empty `v` is a no-op; an invalid `v` panics. |
| `(Intervals[T]) Search(off T) (Interval[T], bool)` | Returns the interval containing `off` and `true`, or the zero `Interval` and `false`. |

Because `Insert` may reallocate the underlying slice (just like the built-in
`append`), always assign its result back:

```go
done = done.Insert(iv)
```

Concurrency
-----------

Neither `Insert` nor `Search` is safe for concurrent access. `Insert` mutates
the receiver, and `Search` requires the slice to remain correctly ordered while
it runs. If you share an `Intervals` across goroutines, guard it with a
`sync.RWMutex` (or equivalent): a write lock around `Insert`, and a read lock
around `Search`.

Installation
------------

```sh
go get github.com/die-net/intervals
```

This package requires Go 1.22 or newer and pulls in no third-party
dependencies.

Example
-------

A motivating use case is a concurrent download, such as fetching chunks of a
large file from S3 with `n` goroutines. Rather than waiting for the last byte
before sending the first, you can record each completed byte range as it lands
and start streaming the contiguous prefix as soon as it is ready.

```go
package main

import (
    "fmt"

    "github.com/die-net/intervals"
)

func main() {
    var done intervals.Intervals[int64]

    // Chunks complete out of order; record each finished byte range.
    done = done.Insert(intervals.Interval[int64]{Start: 0, End: 1024})
    done = done.Insert(intervals.Interval[int64]{Start: 2048, End: 4096})
    done = done.Insert(intervals.Interval[int64]{Start: 1024, End: 2048}) // fills the gap

    fmt.Println(done) // [{0 4096}]

    // How much contiguous data is ready starting at offset 0?
    if iv, ok := done.Search(0); ok {
        fmt.Printf("bytes [%d, %d) are ready to stream\n", iv.Start, iv.End)
    }
}
```

Performance
-----------

Both `Insert` and `Search` use a binary search (`sort.Search`) to locate the
region of interest. `Insert` reuses the existing slice whenever possible,
copying in-place in the common cases, aggressively merging intervals where
possible and only allocating to double the size of the underlying slice when
it is full. `Search` doesn't allocate.

The `Insert` benchmarks cover three scenarios: inserting disjoint intervals
(the worst case, since each must be kept separate), inserting mergeable
adjacent intervals, and inserting heavily overlapping intervals.

The `Search` benchmarks cover the disjoint case (many separate intervals to
binary-search across) and the merged case (a single interval).  `Search` is
read-only and runs in roughly logarithmic time in the number of unmerged
intervals remaining.

```
$ go test -cpu=1 -bench=. -benchmem
goos: darwin
goarch: arm64
pkg: github.com/die-net/intervals
cpu: Apple M2 Pro
BenchmarkInsertDisjoint/1        107374227        10.96 ns/op    91.23 MB/s   0 B/op   0 allocs/op
BenchmarkInsertDisjoint/10        45950748        26.01 ns/op    38.44 MB/s   0 B/op   0 allocs/op
BenchmarkInsertDisjoint/100       26269006        45.47 ns/op    21.99 MB/s   0 B/op   0 allocs/op
BenchmarkInsertDisjoint/1000      12517956        95.94 ns/op    10.42 MB/s   0 B/op   0 allocs/op
BenchmarkInsertDisjoint/10000      2251401       532.3 ns/op      1.88 MB/s   0 B/op   0 allocs/op
BenchmarkInsertMergeable/1       100000000        10.96 ns/op    91.22 MB/s   0 B/op   0 allocs/op
BenchmarkInsertMergeable/10       48691087        24.37 ns/op    41.04 MB/s   0 B/op   0 allocs/op
BenchmarkInsertMergeable/100      30892900        38.76 ns/op    25.80 MB/s   0 B/op   0 allocs/op
BenchmarkInsertMergeable/1000     19887729        60.10 ns/op    16.64 MB/s   0 B/op   0 allocs/op
BenchmarkInsertMergeable/10000     7792966       157.4 ns/op      6.35 MB/s   0 B/op   0 allocs/op
BenchmarkInsertOverlapping/1      70237390        16.43 ns/op    60.88 MB/s   0 B/op   0 allocs/op
BenchmarkInsertOverlapping/10     47493560        25.14 ns/op    39.78 MB/s   0 B/op   0 allocs/op
BenchmarkInsertOverlapping/100    48296054        24.92 ns/op    40.12 MB/s   0 B/op   0 allocs/op
BenchmarkInsertOverlapping/1000   40313540        29.73 ns/op    33.63 MB/s   0 B/op   0 allocs/op
BenchmarkInsertOverlapping/10000  30862311        38.81 ns/op    25.77 MB/s   0 B/op   0 allocs/op
BenchmarkSearchDisjoint/1        479366676         2.506 ns/op  399.11 MB/s   0 B/op   0 allocs/op
BenchmarkSearchDisjoint/10       279172508         4.494 ns/op  222.52 MB/s   0 B/op   0 allocs/op
BenchmarkSearchDisjoint/100      194244874         6.083 ns/op  164.39 MB/s   0 B/op   0 allocs/op
BenchmarkSearchDisjoint/1000     145232559         8.269 ns/op  120.94 MB/s   0 B/op   0 allocs/op
BenchmarkSearchDisjoint/10000    110869339        10.91 ns/op    91.70 MB/s   0 B/op   0 allocs/op
BenchmarkSearchMerged/1          477927547         2.552 ns/op  391.81 MB/s   0 B/op   0 allocs/op
BenchmarkSearchMerged/10         478269381         2.502 ns/op  399.74 MB/s   0 B/op   0 allocs/op
BenchmarkSearchMerged/100        479289055         2.503 ns/op  399.54 MB/s   0 B/op   0 allocs/op
BenchmarkSearchMerged/1000       477311384         2.504 ns/op  399.28 MB/s   0 B/op   0 allocs/op
BenchmarkSearchMerged/10000      475930947         2.507 ns/op  398.91 MB/s   0 B/op   0 allocs/op
PASS
```

(Numbers above were measured on an Apple M2 Pro and will vary by machine.)

Testing
-------

The package has 100% statement coverage, including a randomized fuzz-style test
and a number of awkward corner cases (empty ranges, `NaN` bounds, gap filling,
and full replacement).

```sh
go test -cover ./...
```

License
-------

Copyright 2021-2026 Aaron Hopkins and contributors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at: http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
