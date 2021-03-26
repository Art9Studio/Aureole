package session

import (
	contextTypes "aureole/context/types"
	"aureole/internal/collections"
	storageTypes "aureole/internal/plugins/storage/types"
	"github.com/gofiber/fiber/v2"
)

type session struct {
	Conf           *config
	ProjectContext *contextTypes.ProjectCtx
	Storage        storageTypes.Storage
	Collection     *collections.Collection
}

func (s *session) Authorize(ctx *fiber.Ctx, fields map[string]interface{}) error {
	panic("implement me")
}
