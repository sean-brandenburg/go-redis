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

func (s *Server) Set(key string, value any, timeout int64) {
	s.storeDataMu.Lock()
	defer s.storeDataMu.Unlock()

	if timeout == 0 {
		s.storeData[key] = storeValue{
			data: value,
		}
		return
	}

	expiryTime := time.Now().Add(time.Duration(timeout) * time.Second)
	s.storeData[key] = storeValue{
		data:      value,
		expiresAt: &expiryTime,
	}
}

func (s *Server) Get(key string) (any, bool) {
	s.storeDataMu.Lock()
	defer s.storeDataMu.Unlock()

	value, ok := s.storeData[key]
	if !ok {
		return nil, false
	}

	// If we find that the key is expired, delete it
	if value.isExpired() {
		s.Logger.Debug(fmt.Sprintf("found expired key for value %q", key))
		delete(s.storeData, key)
		return nil, false
	}

	return value.data, true
}
