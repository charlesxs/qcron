package libs

import (
	"sync"
	"time"
)

var InfoCache = &TimeCache{cm: make(map[string][]time.Time)}

type TimeCache struct {
	sync.Mutex
	cm map[string][]time.Time
}

func (c *TimeCache) WriteCache(m map[string]time.Time) {
	c.Lock()
	defer c.Unlock()
	for k, v := range m {
		if _, ok := c.cm[k]; !ok {
			c.cm[k] = make([]time.Time, 0, 6)
		}

		c.cm[k] = append(c.cm[k], v)
	}
}

func (c *TimeCache) CleanCache()  {
	c.Lock()
	defer c.Unlock()

	now := time.Now()
	for k, v := range c.cm {
		tmp := make([]time.Time, 0, 3)
		for i := range v {
			// 清理超过1周的
			if now.Sub(v[i]).Hours() > 168 {
				continue
			}
			tmp = append(tmp, v[i])
		}
		c.cm[k] = tmp
	}
}

func (c *TimeCache) Set(key string)  {
	c.Lock()
	defer c.Unlock()
	tmp := make([]time.Time, 0, 3)
	c.cm[key] = tmp
}

func (c *TimeCache) Get(key string) ([]time.Time, bool)  {
	v, ok := c.cm[key]
	if !ok {
		return nil, false
	}
	return v, ok
}

func (c *TimeCache) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.cm, key)
}

func (c *TimeCache) Append(key string, value time.Time)  {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.cm[key]; !ok {
		c.cm[key] = make([]time.Time, 0, 3)
	}
	c.cm[key] = append(c.cm[key], value)
}

func (c *TimeCache) ClearAll() {
	c.Lock()
	defer c.Unlock()

	for k := range c.cm {
		delete(c.cm, k)
	}
}

func (c *TimeCache) ForEach(fn func(k string, v []time.Time) bool)  {
	for k, v := range c.cm {
		if ok := fn(k, v); !ok {
			return
		}
	}
}
