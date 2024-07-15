Intervals [![Build Status](https://github.com/die-net/intervals/actions/workflows/go-test.yml/badge.svg)](https://github.com/die-net/intervals/actions/workflows/go-test.yml) [![Coverage Status](https://coveralls.io/repos/github/die-net/intervals/badge.svg?branch=main)](https://coveralls.io/github/die-net/intervals?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/die-net/intervals)](https://goreportcard.com/report/github.com/die-net/intervals)
=========

Package intervals provides a fast insertion-sort based solution to the merge
overlapping intervals problem.

An Interval represents a range of the form `[Start, End)`, which contains
all `x` in `Start <= x < End`.  `x` can be any type that is ordered,
including the various sizes of `int`, `uint`, `float` types, `uintptr` and
`string`.

The `Intervals` type is a slice of `Interval` as maintained by the `Insert()`
method, which keeps them in ascending order and with any two overlapping
`Interval` merged.  Two `Interval` overlap if the `End` of one is `>=` the
`Start` from the next one.

An example use case for this might be trying to do a concurrent download,
such as chunks of a large file from S3.  You could kick off `n` goroutines,
each waiting to download `1/n` of the file, and use something like
`sync.WaitGroup` to wait until they are all done until you start to use
them.  However, that would require waiting until the last byte is received
before start sending the first byte.  If you want to start streaming the
answer from the beginning as soon as the relevant chunks are written to
disk, you could maintain a `done Intervals[int64]` and `done.Insert()` the
byte range written after each successful `Write()`, start streaming as soon
as `done.Search(0)` returns true, up to the whatever the returned Interval's
`End - 1` is.

The overhead of `Insert()` is very low, with the worst case of `Insert()`ing
10,000 non-overlapping intervals only taking 147ns on my Apple Mac M1:

```
$ go test -cpu=1 -bench=.
goos: darwin
goarch: arm64
pkg: github.com/die-net/intervals
BenchmarkInsertNonOverlapping/1         	115100521	        10.30 ns/op
BenchmarkInsertNonOverlapping/10        	45857023	        26.01 ns/op
BenchmarkInsertNonOverlapping/100       	29844249	        39.27 ns/op
BenchmarkInsertNonOverlapping/1000      	21177404	        57.57 ns/op
BenchmarkInsertNonOverlapping/10000     	 8115252	       147.0 ns/op
BenchmarkInsertOverlapping/1            	91738153	        12.94 ns/op
BenchmarkInsertOverlapping/10           	52171077	        23.19 ns/op
BenchmarkInsertOverlapping/100          	54989328	        22.00 ns/op
BenchmarkInsertOverlapping/1000         	47051826	        25.75 ns/op
BenchmarkInsertOverlapping/10000        	36945717	        32.76 ns/op
PASS
```

And there's 100% test coverage, with a bunch of weird corner cases tested. 

License
-------

Copyright 2021-2024 Aaron Hopkins and contributors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at: http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
