package email

import (
	authnTypes "aureole/internal/plugins/authn/types"
	authzT "aureole/internal/plugins/authz/types"
	storageT "aureole/internal/plugins/storage/types"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

func SendMagicLink(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if !e.identity.Email.IsEnabled || !isCredential(&e.identity.Email) {
			return sendError(c, fiber.StatusInternalServerError, "expects 1 credential, 0 got")
		}

		input, err := authnTypes.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		identity := &storageT.IdentityData{Email: input.Email}

		/*emailCol := e.coll.Spec.FieldsMap["email"].Name
		exist, err := e.storage.IsIdentityExist(e.identity, []storageT.Filter{
			{Name: emailCol, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			_, err := e.storage.InsertIdentity(e.identity, identity)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}*/

		token, err := createToken(e, map[string]interface{}{"email": identity.Email})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(e.magicLink, token)
		err = e.sender.Send(identity.Email.(string),
			"",
			e.conf.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

/*func Register(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var (
			rawInput interface{}
			input    authnTypes.Input
		)
		if err := c.BodyParser(&rawInput); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if err := mapstructure.Decode(rawInput, &input); err != nil {
			return err
		}
		if err := input.Init(e.identity, e.identity.Collection.Spec.FieldsMap, true); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		identity := &storageT.IdentityData{
			Id:         input.Id,
			Username:   input.Username,
			Phone:      input.Phone,
			Email:      input.Email,
			Additional: input.Additional,
		}

		emailField := e.coll.Spec.FieldsMap["email"].Name
		exist, err := e.storage.IsIdentityExist(e.identity, []storageT.Filter{
			{Name: emailField, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		userId, err := e.storage.InsertIdentity(e.identity, identity)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		token, err := uuid.NewV4()
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		tokenHash := e.link.hasher().Sum(token.Bytes())
		linkData := &storageT.EmailLinkData{
			Email:   identity.Email,
			Token:   base64.StdEncoding.EncodeToString(tokenHash),
			Expires: time.Now().Add(time.Duration(e.conf.Link.Token.Exp) * time.Second).Format(time.RFC3339),
			Invalid: false,
		}

		linkSpecs := &e.link.coll.Spec
		err = e.storage.InvalidateEmailLink(linkSpecs, []storageT.Filter{
			{Name: linkSpecs.FieldsMap["email"].Name, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = e.storage.InsertEmailLink(linkSpecs, linkData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(e.link.magicLink, token.String())
		err = e.link.sender.Send(linkData.Email.(string),
			"",
			e.conf.Link.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"user_id": userId})
	}
}*/

func Login(e *email) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		rawToken := c.Query("token")
		if rawToken == "" {
			return sendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := jwt.ParseString(
			rawToken,
			jwt.WithIssuer("Aureole Internal"),
			jwt.WithAudience("Aureole Internal"),
			jwt.WithValidate(true),
			jwt.WithKeySet(e.serviceKey.GetPublicSet()),
		)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}

		payload := authzT.NewPayload(e.authorizer, e.storage)
		payload.Email = email
		return e.authorizer.Authorize(c, payload)
	}
}
