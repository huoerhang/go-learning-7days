module demo

go 1.16

require (
	geecache v0.0.0
	lru v0.0.0
)

replace (
	geecache => ./geecache
	lru => ./geecache/lru
)