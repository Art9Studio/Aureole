package session

import (
	"aureole/configs"
	contextTypes "aureole/context/types"
	"aureole/internal/collections"
	"aureole/internal/plugins/authz"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/mitchellh/mapstructure"
	"time"
)

type session struct {
	rawConf        *configs.Authz
	conf           *config
	projectContext *contextTypes.ProjectCtx
	storage        storageTypes.Storage
	collection     *collections.Collection
}

func (s *session) Initialize() error {
	projectCtx := authz.Repository.ProjectCtx
	adapterConf := &config{}
	if err := mapstructure.Decode(s.rawConf.Config, adapterConf); err != nil {
		return err
	}
	adapterConf.setDefaults()

	s.conf = adapterConf
	s.projectContext = projectCtx

	collection, ok := projectCtx.Collections[s.conf.Collection]
	if !ok {
		return fmt.Errorf("collection named '%s' is not declared", s.conf.Collection)
	}

	storage, ok := projectCtx.Storages[s.conf.Storage]
	if !ok {
		return fmt.Errorf("storage named '%s' is not declared", s.conf.Storage)
	}

	isCollExist, err := storage.IsCollExists(collection.Spec)
	if err != nil {
		return err
	}

	if !isCollExist {
		err := storage.CreateSessionColl(collection.Spec)
		if err != nil {
			return err
		}
	}

	storage.SetCleanInterval(s.conf.CleanInterval)
	storage.StartCleaning(collection.Spec)

	s.storage = storage
	s.collection = collection

	return s.storage.CheckFeaturesAvailable([]string{s.collection.Type})
}

func (s *session) Authorize(ctx *fiber.Ctx, fields map[string]interface{}) error {
	userId := fields["user_id"].(int32)
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
