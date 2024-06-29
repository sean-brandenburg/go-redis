package server

import (
	"fmt"
	"time"
)

type storeValue struct {
	data      any
	expiresAt *time.Time
}

func (v storeValue) isExpired() bool {
	return v.expiresAt != nil && v.expiresAt.Before(time.Now())
}

func (s *BaseServer) Set(key string, value any, expiryTimeMs int64) {
	s.storeDataMu.Lock()
	defer s.storeDataMu.Unlock()

	if expiryTimeMs == 0 {
		s.storeData[key] = storeValue{
			data: value,
		}
		return
	}

	expiryTime := time.Now().Add(time.Duration(expiryTimeMs) * time.Millisecond)
	s.storeData[key] = storeValue{
		data:      value,
		expiresAt: &expiryTime,
	}
}

func (s *BaseServer) Get(key string) (any, bool) {
	s.storeDataMu.Lock()
	defer s.storeDataMu.Unlock()

	value, ok := s.storeData[key]
	if !ok {
		return nil, false
	}

	// If we find that the key is expired, delete it
	if value.isExpired() {
		s.logger.Debug(fmt.Sprintf("found expired key for value %q", key))
		delete(s.storeData, key)
		return nil, false
	}

	return value.data, true
}
