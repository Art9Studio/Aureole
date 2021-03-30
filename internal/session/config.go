package session

import (
	"aureole/configs"
	storageTypes "aureole/internal/plugins/storage/types"
	"github.com/gofiber/fiber/v2/utils"
	"time"
)

// Config defines the config for middleware.
type Config struct {
	// Allowed session duration
	// Optional. Default value 24 * time.Hour
	Expiration time.Duration

	// Storage interface to store the session data
	Storage storageTypes.Storage

	// Name of the session cookie. This cookie will store session key.
	// Optional. Default value "session_id".
	CookieName string

	// Domain of the CSRF cookie.
	// Optional. Default value "".
	CookieDomain string

	// Path of the CSRF cookie.
	// Optional. Default value "".
	CookiePath string

	// Indicates if CSRF cookie is secure.
	// Optional. Default value false.
	CookieSecure bool

	// Indicates if CSRF cookie is HTTP only.
	// Optional. Default value false.
	CookieHTTPOnly bool

	// Indicates if CSRF cookie is HTTP only.
	// Optional. Default value false.
	CookieSameSite string

	// KeyGenerator generates the session key.
	// Optional. Default value utils.UUIDv4
	KeyGenerator func() string
}

func (c *Config) setDefaults() {
	configs.SetDefault(&c.Expiration, 24*time.Hour)
	configs.SetDefault(&c.CookieName, "session_id")
	configs.SetDefault(&c.KeyGenerator, utils.UUIDv4)
}
