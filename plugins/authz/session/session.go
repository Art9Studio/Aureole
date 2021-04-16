package session

import (
	"aureole/configs"
	"aureole/internal/collections"
	"aureole/internal/plugins/authz"
	storageTypes "aureole/internal/plugins/storage/types"
	"aureole/internal/router"
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

func (s *session) GetRoutes() []*router.Route {
	return []*router.Route{}
}

func (s *session) Initialize() (err error) {
	s.conf, err = initConfig(&s.rawConf.Config)
	if err != nil {
		return err
	}

	pluginsApi := authz.Repository.PluginsApi
	s.collection, err = pluginsApi.GetCollection(s.conf.Collection)
	if err != nil {
		return fmt.Errorf("collection named '%s' is not declared", s.conf.Collection)
	}

	s.storage, err = pluginsApi.GetStorage(s.conf.Storage)
	if err != nil {
		return fmt.Errorf("storage named '%s' is not declared", s.conf.Storage)
	}

	isCollExist, err := s.storage.IsCollExists(s.collection.Spec)
	if err != nil {
		return err
	}
	if !isCollExist {
		err = s.storage.CreateSessionColl(s.collection.Spec)
		if err != nil {
			return err
		}
	}

	s.storage.SetCleanInterval(s.conf.CleanInterval)
	s.storage.StartCleaning(s.collection.Spec)
	return s.storage.CheckFeaturesAvailable([]string{s.collection.Type})
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	adapterConf.setDefaults()

	return adapterConf, nil
}

func (s *session) Authorize(ctx *fiber.Ctx, fields map[string]interface{}) error {
	userId := fields["user_id"]
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
