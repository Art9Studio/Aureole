package session

import (
	contextTypes "aureole/context/types"
	"aureole/internal/collections"
	"aureole/internal/plugins/authn"
	storageTypes "aureole/internal/plugins/storage/types"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"time"
)

type session struct {
	Conf           *config
	ProjectContext *contextTypes.ProjectCtx
	Storage        storageTypes.Storage
	Collection     *collections.Collection
}

func (s *session) Initialize() error {
	projectCtx := authn.Repository.ProjectCtx

	collection, ok := projectCtx.Collections[s.Conf.Collection]
	if !ok {
		return fmt.Errorf("collection named '%s' is not declared", s.Conf.Collection)
	}

	storage, ok := projectCtx.Storages[s.Conf.Storage]
	if !ok {
		return fmt.Errorf("storage named '%s' is not declared", s.Conf.Storage)
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

	storage.SetCleanInterval(s.Conf.CleanInterval)
	storage.StartCleaning(collection.Spec)

	s.ProjectContext = projectCtx
	s.Storage = storage
	s.Collection = collection

	return s.Storage.CheckFeaturesAvailable([]string{s.Collection.Type})
}

func (s *session) Authorize(ctx *fiber.Ctx, fields map[string]interface{}) error {
	userId := fields["user_id"].(int)
	expires := time.Now().Add(time.Duration(s.Conf.MaxAge) * time.Second)

	sessionToken, err := uuid.NewV4()
	if err != nil {
		return err
	}

	sessionData := storageTypes.InsertSessionData{
		UserId:       userId,
		SessionToken: sessionToken,
		Expiration:   expires,
	}
	_, err = s.Storage.InsertSession(s.Collection.Spec, sessionData)
	if err != nil {
		return err
	}

	cookie := &fiber.Cookie{
		Name:     "session_token",
		Value:    sessionToken.String(),
		Path:     s.Conf.Path,
		Domain:   s.Conf.Domain,
		MaxAge:   s.Conf.MaxAge,
		Expires:  expires,
		Secure:   s.Conf.Secure,
		HTTPOnly: s.Conf.HttpOnly,
		SameSite: s.Conf.SameSite,
	}
	ctx.Cookie(cookie)

	return nil
}
