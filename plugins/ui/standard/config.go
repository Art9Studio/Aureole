package standard

import "aureole/internal/configs"

const (
	getRedirectURL       = "/ui/redirect-url"
	getJWTStorageKeysURL = "/ui/jwt-storage-keys"
)

type config struct {
	SuccessRedirect string `mapstructure:"success_redirect"`
	StorageJWTKeys  struct {
		Access  string `mapstructure:"access"`
		Refresh string `mapstructure:"refresh"`
	} `mapstructure:"storage_jwt_keys"`
}

func (c *config) setDefaults() {
	configs.SetDefault(&c.StorageJWTKeys.Access, "access_jwt")
	configs.SetDefault(&c.StorageJWTKeys.Refresh, "refresh_jwt")
}
