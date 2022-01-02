module github.com/buraksezer/olric

go 1.13

require (
	github.com/RoaringBitmap/roaring v0.9.4
	github.com/buraksezer/consistent v0.0.0-20191006190839-693edf70fd72
	github.com/cespare/xxhash/v2 v2.1.2
	github.com/go-redis/redis/v8 v8.11.4
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-sockaddr v1.0.2
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/memberlist v0.3.0
	github.com/miekg/dns v1.1.45 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/redcon v1.4.3
	github.com/vmihailenco/msgpack/v5 v5.3.5
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/tidwall/redcon v1.4.3 => /Users/buraksezer/go/src/github.com/buraksezer/redcon
)
