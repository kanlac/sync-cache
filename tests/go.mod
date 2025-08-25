module sync-cache-tests

go 1.23.12

require github.com/redis/go-redis/v9 v9.12.1 // indirect

replace sync-cache => ../

require sync-cache v0.0.0-00010101000000-000000000000

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgraph-io/ristretto/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/kanlac/sync-cache v0.0.0-20250821083416-be19175bc452 // indirect
	golang.org/x/sys v0.31.0 // indirect
)
