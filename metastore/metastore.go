package metastore

import (
	"container/list"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/disksing/luson/config"
	"github.com/disksing/luson/util"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Store struct {
	dataDir       string
	cacheCapacity int

	sync.Mutex
	access *list.List
	cache  map[string]*list.Element
}

func NewStore(dataDir config.DataDir, conf *config.Config) *Store {
	return &Store{
		dataDir:       string(dataDir),
		cacheCapacity: conf.MetaCacheSize,
		access:        list.New(),
		cache:         make(map[string]*list.Element),
	}
}

func (s *Store) Create() (string, error) {
	s.Lock()
	defer s.Unlock()
	for i := 0; i < 10; i++ {
		id := uuid.NewV4().String()
		err := os.Mkdir(filepath.Join(s.dataDir, id), 0755)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		return id, nil
	}
	return "", errors.Errorf("failed to allocate valid uuid")
}

func (s *Store) Get(id string) (*MetaData, error) {
	s.Lock()
	defer s.Unlock()
	if e, ok := s.cache[id]; ok {
		s.access.MoveToFront(e)
		return e.Value.(*MetaData), nil
	}
	m, err := s.load(id)
	if err != nil {
		return nil, err
	}
	s.in(m)
	s.evict()
	return m, nil
}

func (s *Store) Put(m *MetaData) error {
	s.Lock()
	defer s.Unlock()
	if !util.IsUUID(m.ID) {
		return errors.Errorf("id is invalid")
	}
	if !validAccess(m.Access) {
		return errors.Errorf("access is invalid")
	}
	if e, ok := s.cache[m.ID]; ok {
		s.out(e)
	}
	err := s.save(m)
	if err != nil {
		return err
	}
	s.in(m)
	return nil
}

func (s *Store) evict() {
	for len(s.cache) > s.cacheCapacity {
		s.out(s.access.Back())
	}
}

func (s *Store) in(m *MetaData) {
	e := s.access.PushFront(m)
	s.cache[m.ID] = e
}

func (s *Store) out(e *list.Element) {
	m := e.Value.(MetaData)
	s.access.Remove(e)
	delete(s.cache, m.ID)
}

func (s *Store) load(id string) (*MetaData, error) {
	f, err := os.OpenFile(s.fname(id), os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var v MetaData
	if err = json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	if !validAccess(v.Access) {
		v.Access = Private
	}
	return &v, nil
}

func (s *Store) save(m *MetaData) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(s.fname(m.ID), b, 0644)
}

func (s *Store) fname(id string) string {
	return filepath.Join(s.dataDir, id, "meta.json")
}

const (
	Public    string = "public"    // everyone can read/write
	Protected        = "protected" // everyone can read, write with api key
	Private          = "private"   // read/write with api key
)

func validAccess(s string) bool {
	return s == Public || s == Protected || s == Private
}

type MetaData struct {
	ID     string `json:"id"`
	Access string `json:"access"`
}
