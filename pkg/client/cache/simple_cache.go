/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"kubesphere.io/devops/pkg/server/errors"
)

var ErrNoSuchKey = errors.New("no such key")

type simpleObject struct {
	value       string
	neverExpire bool
	expiredAt   time.Time
}

// SimpleCache implements cache.Interface use memory objects, it should be used only for testing
type simpleCache struct {
	store  map[string]simpleObject
	rwLock sync.RWMutex
}

func NewSimpleCache() Interface {
	return &simpleCache{store: make(map[string]simpleObject), rwLock: sync.RWMutex{}}
}

func (s *simpleCache) Keys(pattern string) ([]string, error) {
	// There is a little difference between go regexp and redis key pattern
	// In redis, * means any character, while in go . means match everything.
	pattern = strings.Replace(pattern, "*", ".", -1)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var keys []string
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	for k := range s.store {
		if re.MatchString(k) { // TODO need add expire check.
			keys = append(keys, k)
		}
	}

	return keys, nil
}

func (s *simpleCache) Set(key string, value string, duration time.Duration) error {
	sobject := simpleObject{
		value:       value,
		neverExpire: false,
		expiredAt:   time.Now().Add(duration),
	}

	if duration == NeverExpire {
		sobject.neverExpire = true
	}
	s.rwLock.Lock()
	s.store[key] = sobject
	s.rwLock.Unlock()
	return nil
}

func (s *simpleCache) Del(keys ...string) error {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	for _, key := range keys {
		delete(s.store, key)
	}
	return nil
}

func (s *simpleCache) Get(key string) (string, error) {
	var bExpire bool
	s.rwLock.RLock()
	if sobject, ok := s.store[key]; ok {
		if sobject.neverExpire || time.Now().Before(sobject.expiredAt) {
			s.rwLock.RUnlock()
			return sobject.value, nil
		}
		if !time.Now().Before(sobject.expiredAt) {
			bExpire = true
		}
	}
	s.rwLock.RUnlock()
	if bExpire {
		s.rwLock.Lock()
		delete(s.store, key)
		s.rwLock.Unlock()
	}
	return "", ErrNoSuchKey
}

func (s *simpleCache) Exists(keys ...string) (bool, error) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	for _, key := range keys {
		if _, ok := s.store[key]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func (s *simpleCache) Expire(key string, duration time.Duration) error {
	value, err := s.Get(key)
	if err != nil {
		return err
	}

	sobject := simpleObject{
		value:       value,
		neverExpire: false,
		expiredAt:   time.Now().Add(duration),
	}

	if duration == NeverExpire {
		sobject.neverExpire = true
	}
	s.rwLock.Lock()
	s.store[key] = sobject
	s.rwLock.Unlock()
	return nil
}
