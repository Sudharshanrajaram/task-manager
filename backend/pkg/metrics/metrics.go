package metrics

import "sync/atomic"

var (
	// TotalHTTPRequests counts all processed HTTP requests
	TotalHTTPRequests atomic.Uint64
)
