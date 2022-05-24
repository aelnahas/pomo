package task

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
)

type Status string
type FilterTask func(t Task) bool

const (
	Pending  Status = "pending"
	Complete        = "complete"
)

var (
	CurrentTaskKey = []byte("current")
)

type Task struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Status    Status     `json:"status"`
	Sessions  int        `json:"sessions"`
	CreatedAT time.Time  `json:"created_at"`
	UpdatedAT *time.Time `json:"updated_at"`
}

func NewTask(title string) *Task {
	return &Task{
		ID:        uuid.New(),
		Title:     title,
		Status:    Pending,
		Sessions:  0,
		CreatedAT: time.Now(),
		UpdatedAT: nil,
	}
}

func (t Task) Key() []byte {
	return []byte(t.ID.String())
}

func (t Task) Format() string {
	line := []string{t.ID.String(), t.Title, string(t.Status), fmt.Sprintf("%d", t.Sessions)}
	return strings.Join(line, "\t")
}

func (t Task) Describe() string {
	return fmt.Sprintf("Task: %s, sessions: %d", t.Title, t.Sessions)
}

type Store interface {
	Add(title string) (*Task, error)
	Remove(id uuid.UUID) error
	List(filter FilterTask) ([]Task, error)
	SetState(id uuid.UUID, status Status) (*Task, error)
	AddSessions(id uuid.UUID) (*Task, error)

	ClearCurrentTask(id uuid.UUID) error
	SetCurrentTask(id uuid.UUID) error
	GetCurrentTask() (task *Task, err error)
}

type store struct {
	db *badger.DB
}

var _ Store = &store{}

func NewStore(path string) (*store, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &store{
		db: db,
	}, nil
}

func (s *store) Close() error {
	return s.db.Close()
}

func (s *store) Add(title string) (*Task, error) {
	task := NewTask(title)
	err := s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(task)
		if err != nil {
			return err
		}

		if err := txn.Set(task.Key(), data); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *store) Remove(id uuid.UUID) error {
	return s.db.Update(func(txn *badger.Txn) error {
		byteID := []byte(id.String())
		item, err := txn.Get(byteID)
		if err != nil {
			return err
		}

		if !item.IsDeletedOrExpired() {
			if err := txn.Delete(byteID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *store) List(filter FilterTask) ([]Task, error) {
	tasks := make([]Task, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			if _, err := uuid.ParseBytes(item.Key()); err == nil {
				err := item.Value(func(val []byte) error {
					var t Task
					if err := json.Unmarshal(val, &t); err != nil {
						return err
					}

					if filter(t) {
						tasks = append(tasks, t)
					}
					return nil
				})

				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *store) SetState(id uuid.UUID, status Status) (task *Task, err error) {
	byteID := []byte(id.String())

	err = s.db.Update(func(txn *badger.Txn) (err error) {
		task, err = s.getTaskByID(id, txn)
		if err != nil {
			return err
		}

		task.Status = status
		data, err := json.Marshal(task)
		if err != nil {
			return err
		}

		return txn.Set(byteID, data)
	})

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *store) AddSessions(id uuid.UUID) (*Task, error) {
	var task *Task
	byteID := []byte(id.String())

	err := s.db.Update(func(txn *badger.Txn) error {
		var err error
		task, err = s.getTaskByID(id, txn)
		if err != nil {
			return err
		}

		task.Sessions++
		data, err := json.Marshal(task)
		if err != nil {
			return err
		}

		return txn.Set(byteID, data)
	})

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *store) GetTask(id uuid.UUID) (*Task, error) {
	var task *Task
	err := s.db.View(func(txn *badger.Txn) (err error) {
		task, err = s.getTaskByID(id, txn)
		return err
	})

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *store) getTaskByID(id uuid.UUID, txn *badger.Txn) (*Task, error) {
	var task Task
	byteID := []byte(id.String())
	item, err := txn.Get(byteID)
	if err != nil {
		return nil, err
	}

	err = item.Value(func(val []byte) error {
		return json.Unmarshal(val, &task)
	})

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *store) GetCurrentTask() (*Task, error) {
	var task *Task
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(CurrentTaskKey)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			id, err := uuid.Parse(string(val))
			if err != nil {
				return err
			}

			task, err = s.getTaskByID(id, txn)

			return err
		})

		if err != nil {
			return err
		}

		return err
	})

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *store) SetCurrentTask(id uuid.UUID) error {
	return s.db.Update(func(txn *badger.Txn) error {
		task, err := s.getTaskByID(id, txn)
		if err != nil {
			return err
		}

		return txn.Set(CurrentTaskKey, task.Key())
	})
}

func (s *store) ClearCurrentTask(id uuid.UUID) error {
	return s.db.Update(func(txn *badger.Txn) error {
		if _, err := s.getTaskByID(id, txn); err != nil {
			return err
		}

		return txn.Delete(CurrentTaskKey)
	})
}
