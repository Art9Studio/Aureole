package session

import (
	"aureole/internal/collections"
	"aureole/internal/configs"
	"aureole/internal/plugins/authz"
	"aureole/internal/plugins/authz/types"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/mitchellh/mapstructure"
	"time"
)

type session struct {
	rawConf    *configs.Authz
	conf       *config
	storage    storageTypes.Storage
	collection *collections.Collection
}

func (s *session) Init(appName string) (err error) {
	s.conf, err = initConfig(&s.rawConf.Config)
	if err != nil {
		return err
	}

	pluginApi := authz.Repository.PluginApi
	s.collection, err = pluginApi.Project.GetCollection(s.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", s.conf.Collection)
	}

	s.storage, err = pluginApi.Project.GetStorage(s.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", s.conf.Storage)
	}

	/*isCollExist, err := s.storage.IsCollExists(s.collection.Spec)
	if err != nil {
		return err
	}
	if !isCollExist {
		err = s.storage.CreateSessionColl(s.collection.Spec)
		if err != nil {
			return err
		}
	}*/

	s.storage.SetCleanInterval(s.conf.CleanInterval)
	s.storage.StartCleaning(s.collection.Spec)
	if err := s.storage.CheckFeaturesAvailable([]string{s.collection.Type}); err != nil {
		return err
	}

	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func (s *session) Authorize(ctx *fiber.Ctx, authzCtx *types.Context) error {
	userId := authzCtx.Id
	expires := time.Now().Add(time.Duration(s.conf.MaxAge) * time.Second)

	sessionToken, err := uuid.NewV4()
	if err != nil {
		return err
	}

	sessionData := storageTypes.InsertSessionData{
		UserId:       userId,
		SessionToken: sessionToken,
		Expiration:   expires,
	}
	_, err = s.storage.InsertSession(s.collection.Spec, sessionData)
	if err != nil {
		return err
	}

	cookie := &fiber.Cookie{
		Name:     "session_token",
		Value:    sessionToken.String(),
		Path:     s.conf.Path,
		Domain:   s.conf.Domain,
		MaxAge:   s.conf.MaxAge,
		Expires:  expires,
		Secure:   s.conf.Secure,
		HTTPOnly: s.conf.HttpOnly,
		SameSite: s.conf.SameSite,
	}
	ctx.Cookie(cookie)

	return nil
}

func (s *session) GetNativeQueries() map[string]string {
	return nil
}
