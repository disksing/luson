package jsonstore

import (
	"time"

	"github.com/pkg/errors"
)

type Txn struct {
	s                    *Store
	writes               map[string]interface{}
	hashConditions       map[string]string
	modifyTimeConditions map[string]time.Time
}

func (s *Store) NewTxn() *Txn {
	return &Txn{
		s:                    s,
		writes:               make(map[string]interface{}),
		hashConditions:       make(map[string]string),
		modifyTimeConditions: make(map[string]time.Time),
	}
}

func (t *Txn) Get(id string) (interface{}, error) {
	if v, ok := t.writes[id]; ok {
		return v, nil
	}

	v, hash, err := t.s.Get(id)
	if err != nil {
		return nil, err
	}
	t.hashConditions[id] = hash
	return v, nil
}

func (t *Txn) Put(id string, v interface{}) {
	t.writes[id] = v
}

func (t *Txn) IfMatchHash(id, hash string) {
	t.hashConditions[id] = hash
}

func (t *Txn) IfUnmodifiedSince(id string, v time.Time) {
	t.modifyTimeConditions[id] = v
}

func (t *Txn) Commit() error {
	t.s.Lock()
	defer t.s.Unlock()
	for id, hash := range t.hashConditions {
		j, err := t.s.get(id)
		if err != nil {
			return err
		}
		if j.hash != hash {
			return errors.Errorf("hash condition not match, id=" + id)
		}
	}
	for id, v := range t.modifyTimeConditions {
		j, err := t.s.get(id)
		if err != nil {
			return err
		}
		if j.lastModify.After(v) {
			return errors.Errorf("modify time condition not match, id=" + id)
		}
	}
	// FIXME: writes not atomic, need some sort of WAL.
	for id, v := range t.writes {
		err := t.s.put(id, v)
		if err != nil {
			return err
		}
	}
	return nil
}
