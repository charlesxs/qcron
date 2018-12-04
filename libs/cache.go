package libs

import (
	"sync"
	"time"
)

var InfoCache = struct {
	sync.Mutex
	Cache map[string][]time.Time
}{Cache: make(map[string][]time.Time)}

