package session

import (
	contextTypes "aureole/context/types"
	"aureole/internal/collections"
	storageTypes "aureole/internal/plugins/storage/types"
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
