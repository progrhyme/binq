package client

import "time"

func init() {
	// Maximum for large item to download
	DefaultHTTPTimeout = 300 * time.Second
	httpTimeoutToQueryIndex = 5 * time.Second
}