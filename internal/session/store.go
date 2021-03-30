package session

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/niubaoshu/gotiny"
)

type Store struct {
	*Config
}

var mux sync.Mutex

func New(cfg *Config) *Store {
	cfg.setDefaults()

	return &Store{
		cfg,
	}
}

// Get will get/create a session
func (s *Store) Get(c *fiber.Ctx) (*Session, error) {
	var fresh bool

	// Get key from cookie
	id := c.Cookies(s.CookieName)

	// If no key exist, create new one
	if len(id) == 0 {
		id = s.KeyGenerator()
		fresh = true
	}

	// Create session object
	sess := acquireSession()
	sess.ctx = c
	sess.config = s
	sess.id = id
	sess.fresh = fresh

	// Fetch existing data
	if !fresh {
		raw, err := s.Storage.Get(id)
		// Unmashal if we found data
		if raw != nil && err == nil {
			mux.Lock()
			gotiny.Unmarshal(raw, &sess.data)
			mux.Unlock()
			sess.fresh = false
		} else if err != nil {
			return nil, err
		} else {
			sess.fresh = true
		}
	}

	return sess, nil
}

// Reset will delete all session from the storage
func (s *Store) Reset() error {
	return s.Storage.Reset()
}
