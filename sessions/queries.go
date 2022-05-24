package sessions

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
)

type Session struct {
	Current Type `json:"current"`
	Count   int  `json:"count"`
}

func (s Session) Next(intervals int) Type {
	switch {
	case s.Current == Short || s.Current == Long:
		return Focus
	case s.Count > 0 && s.Count%intervals == 0 && s.Current == Focus:
		return Long
	default:
		return Short
	}
}

type store struct {
	db        *badger.DB
	intervals int
}

var Key = []byte("session")
var defaultSession = Session{
	Current: Focus,
	Count:   0,
}

type Store interface {
	Current() (Type, error)
	Next() (Type, error)
	Reset() error
	Increment() error
	Session() (*Session, error)
}

var _ Store = &store{}

func NewStore(path string, intervals int) (*store, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &store{db: db, intervals: intervals}, nil
}

func (s *store) Reset() error {

	err := s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(&defaultSession)
		if err != nil {
			return err
		}

		return txn.Set(Key, data)
	})

	return err
}

func (s *store) Current() (Type, error) {
	var sessionType Type
	err := s.db.View(func(txn *badger.Txn) error {
		session, err := s.getCurrent(txn)
		if err != nil {
			return err
		}

		sessionType = session.Current
		return nil
	})

	return sessionType, err
}

func (s *store) Next() (Type, error) {
	var sessionType Type
	err := s.db.View(func(txn *badger.Txn) error {
		session, err := s.getCurrent(txn)
		if err != nil {
			return err
		}

		sessionType = session.Next(s.intervals)
		return nil
	})

	return sessionType, err
}

func (s *store) Increment() error {
	err := s.db.Update(func(txn *badger.Txn) error {
		session, err := s.getCurrent(txn)
		if err != nil {
			return err
		}

		if session.Current == Focus {
			session.Count++
		}

		session.Current = session.Next(s.intervals)

		data, err := json.Marshal(session)
		if err != nil {
			return err
		}

		return txn.Set(Key, data)
	})
	return err
}

func (s *store) Session() (*Session, error) {
	var session *Session
	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		session, err = s.getCurrent(txn)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *store) getCurrent(txn *badger.Txn) (*Session, error) {
	item, err := txn.Get(Key)
	if err != nil {
		if err := s.Reset(); err != nil {
			return nil, err
		}
		item, err = txn.Get(Key)
		if err != nil {
			return nil, err
		}
	}

	var session Session
	err = item.Value(func(val []byte) error {
		return json.Unmarshal(val, &session)
	})

	if err != nil {
		return nil, err
	}

	return &session, nil
}
