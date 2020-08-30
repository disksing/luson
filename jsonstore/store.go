package jsonstore

import (
	"container/list"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/disksing/luson/config"
)

type Store struct {
	dataDir       string
	cacheCapacity int64

	sync.Mutex
	access    *list.List
	cache     map[string]*list.Element
	totalSize int64
}

func NewStore(dataDir config.DataDir, conf *config.Config) *Store {
	return &Store{
		dataDir:       string(dataDir),
		cacheCapacity: int64(conf.JSONCacheSize) * 1024 * 1024,
		access:        list.New(),
		cache:         make(map[string]*list.Element),
	}
}

type jData struct {
	id         string
	value      interface{}
	hash       string
	lastModify time.Time
	size       int64
}

func (s *Store) Get(id string) (interface{}, string, error) {
	s.Lock()
	defer s.Unlock()
	j, err := s.get(id)
	if err != nil {
		return nil, "", err
	}
	return j.value, j.hash, nil
}

func (s *Store) Put(id string, v interface{}) error {
	s.Lock()
	defer s.Unlock()
	return s.put(id, v)
}

func (s *Store) get(id string) (*jData, error) {
	if e, ok := s.cache[id]; ok {
		s.access.MoveToFront(e)
		return e.Value.(*jData), nil
	}
	j, err := s.load(id)
	if err != nil {
		return nil, err
	}
	s.in(j)
	s.evict()
	return j, nil
}

func (s *Store) put(id string, v interface{}) error {
	if e, ok := s.cache[id]; ok {
		s.out(e)
	}
	j, err := s.save(id, v)
	if err != nil {
		return err
	}
	s.in(j)
	return nil
}

func (s *Store) evict() {
	for s.totalSize > s.cacheCapacity {
		s.out(s.access.Back())
	}
}

func (s *Store) in(j *jData) {
	e := s.access.PushFront(j)
	s.cache[j.id] = e
	s.totalSize += j.size
}

func (s *Store) out(e *list.Element) {
	j := e.Value.(*jData)
	s.access.Remove(e)
	delete(s.cache, j.id)
	s.totalSize -= j.size
}

func (s *Store) load(id string) (*jData, error) {
	f, err := os.OpenFile(s.fname(id), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err = json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return &jData{
		id:         id,
		value:      v,
		hash:       s.sha1(b),
		lastModify: stat.ModTime(),
		size:       int64(len(b)),
	}, nil
}

func (s *Store) save(id string, v interface{}) (*jData, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(s.fname(id), data, 0644)
	if err != nil {
		return nil, err
	}
	return &jData{
		id:         id,
		value:      v,
		hash:       s.sha1(data),
		lastModify: time.Now(),
		size:       int64(len(data)),
	}, nil
}

func (s *Store) fname(id string) string {
	return filepath.Join(s.dataDir, id, "data.json")
}

func (s *Store) sha1(b []byte) string {
	sh := sha1.New()
	_, _ = sh.Write(b)
	return hex.EncodeToString(sh.Sum(nil)[:8])
}
