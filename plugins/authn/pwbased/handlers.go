package pwbased

import (
	"aureole/internal/plugins/authn/types"
	storageT "aureole/internal/plugins/storage/types"
	"github.com/lestrrat-go/jwx/jwt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Login(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		input, err := types.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		identityData := &storageT.IdentityData{
			Id:       input.Id,
			Username: input.Username,
			Phone:    input.Phone,
			Email:    input.Email,
		}

		credName, credVal, err := getCredField(p.identity, identityData)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		if input.Password == nil {
			return sendError(c, fiber.StatusBadRequest, "password required")
		}
		pwData := &storageT.PwBasedData{Password: input.Password}

		_ = credName
		_ = credVal
		_ = pwData

		/*f := []storageT.Filter{{Name: credName, Value: credVal}}
		exist, err := p.storage.IsIdentityExist(p.identity, f)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}

		rawIdentity, err := p.storage.GetIdentity(p.identity, f)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		identity, ok := rawIdentity.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get identity from database")
		}

		pw, err := p.storage.GetPassword(p.coll, f)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		isMatch, err := p.pwHasher.ComparePw(pwData.Password.(string), pw.(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if isMatch {
			collSpec := p.identity.Collection.Spec
			payload := authzT.NewPayload(identity, collSpec.FieldsMap)
			// todo: refactor this
			payload.NativeQ = func(queryName string, args ...interface{}) string {
				queries := p.authorizer.GetNativeQueries()

				q, ok := queries[queryName]
				if !ok {
					return "--an error occurred during render--"
				}

				rawRes, err := p.storage.NativeQuery(q, args...)
				if err != nil {
					return "--an error occurred during render--"
				}

				res, err := json.Marshal(rawRes)
				if err != nil {
					return "--an error occurred during render--"
				}

				return string(res)
			}
			return p.authorizer.Authorize(c, payload)
		} else {
			return sendError(c, fiber.StatusUnauthorized, fmt.Sprintf("wrong password or %s", credName))
		}*/
		return sendError(c, fiber.StatusInternalServerError, "pwbased not available now")
	}
}

func Register(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		input, err := types.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if err := input.Init(p.identity, p.identity.Collection.Spec.FieldsMap, true); err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		identity := &storageT.IdentityData{
			Id:         input.Id,
			Username:   input.Username,
			Phone:      input.Phone,
			Email:      input.Email,
			Additional: input.Additional,
		}

		if input.Password == nil {
			return sendError(c, fiber.StatusBadRequest, "password required")
		}
		pwData := &storageT.PwBasedData{Password: input.Password}

		pwHash, err := p.pwHasher.HashPw(pwData.Password.(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		pwData.PasswordHash = pwHash

		credName, credVal, err := getCredField(p.identity, identity)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}

		_ = credName
		_ = credVal

		/*i := p.identity
		exist, err := p.storage.IsIdentityExist(i, []storageT.Filter{{Name: credName, Value: credVal}})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if exist {
			return sendError(c, fiber.StatusBadRequest, "user already exist")
		}

		userId, err := p.storage.InsertPwBased(i, p.coll, identity, pwData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}*/

		if p.conf.Register.IsVerifyAfter {
			token, err := createToken(p, map[string]interface{}{
				"email":           identity.Email,
				jwt.ExpirationKey: time.Now().Add(time.Duration(p.conf.Verif.Exp) * time.Second).Unix(),
			})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			link := attachToken(p.verif.confirmLink, token)
			err = p.verif.sender.Send(identity.Email.(string),
				"",
				p.conf.Verif.Template,
				map[string]interface{}{"link": link})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}
		}

		/*if p.conf.Register.IsLoginAfter {
			payload := authzT.Payload{
				Id:         userId,
				Username:   identity.Username,
				Phone:      identity.Phone,
				Email:      identity.Email,
				Additional: identity.Additional,
			}
			// todo: refactor this
			payload.NativeQ = func(queryName string, args ...interface{}) string {
				queries := p.authorizer.GetNativeQueries()

				q, ok := queries[queryName]
				if !ok {
					return "--an error occurred during render--"
				}

				rawRes, err := p.storage.NativeQuery(q, args)
				if err != nil {
					return "--an error occurred during render--"
				}

				res, err := json.Marshal(rawRes)
				if err != nil {
					return "--an error occurred during render--"
				}

				return string(res)
			}
			return p.authorizer.Authorize(c, &payload)
		} else {
			return c.JSON(&fiber.Map{"user_id": userId})
		}*/
		return sendError(c, fiber.StatusInternalServerError, "pwbased not available now")
	}
}

func Reset(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		input, err := types.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Email == nil {
			return sendError(c, fiber.StatusBadRequest, "email required")
		}
		identityData := &storageT.IdentityData{Email: input.Email}

		/*collMap := p.identity.Collection.Spec.FieldsMap
		exist, err := p.storage.IsIdentityExist(p.identity, []storageT.Filter{{
			Name: collMap["email"].Name, Value: identityData.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}*/

		token, err := createToken(p, map[string]interface{}{
			"email":           identityData.Email,
			jwt.ExpirationKey: time.Now().Add(time.Duration(p.conf.Reset.Exp) * time.Second).Unix(),
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.reset.confirmLink, token)
		err = p.verif.sender.Send(identityData.Email.(string),
			"",
			p.conf.Verif.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func ResetConfirm(p *pwBased) func(*fiber.Ctx) error {
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
			jwt.WithKeySet(p.serviceKey.GetPublicSet()),
		)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		email, ok := token.Get("email")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}

		input, err := types.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Password == nil {
			return sendError(c, fiber.StatusBadRequest, "password required")
		}
		pw := &storageT.PwBasedData{Password: input.Password}

		pwHash, err := p.pwHasher.HashPw(pw.Password.(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		pw.PasswordHash = pwHash

		/*identitySpecs := p.coll.Parent.Spec
		email := reset[resetSpecs.FieldsMap["email"].Name].(string)
		_, err = p.storage.UpdatePassword(p.coll,
			[]storageT.Filter{{Name: identitySpecs.FieldsMap["email"].Name, Value: email}},
			pw.PasswordHash)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}*/

		err = p.reset.sender.SendRaw(email.(string),
			"Reset your password",
			"Your password has been successfully changed")
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		// todo: add expiring any current user session

		redirectUrl := c.Query("redirect_url")
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func Verify(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		input, err := types.NewInput(c)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		if input.Email == nil {
			return sendError(c, fiber.StatusBadRequest, "email required")
		}
		identity := &storageT.IdentityData{Email: input.Email}

		// should we check it?
		if !p.identity.Email.IsEnabled || !isCredential(p.identity.Email) {
			return sendError(c, fiber.StatusInternalServerError, "expects 1 credential, 0 got")
		}

		/*fieldName := p.coll.Parent.Spec.FieldsMap["email"].Name
		exist, err := p.storage.IsIdentityExist(p.identity, []storageT.Filter{
			{Name: fieldName, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}*/

		token, err := createToken(p, map[string]interface{}{
			"email":           identity.Email,
			jwt.ExpirationKey: time.Now().Add(time.Duration(p.conf.Verif.Exp) * time.Second).Unix(),
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := attachToken(p.verif.confirmLink, token)
		err = p.verif.sender.Send(identity.Email.(string),
			"",
			p.conf.Verif.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func VerifyConfirm(p *pwBased) func(*fiber.Ctx) error {
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
			jwt.WithKeySet(p.serviceKey.GetPublicSet()),
		)
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		_, ok := token.Get("email")
		if !ok {
			return sendError(c, fiber.StatusBadRequest, "cannot get email from token")
		}

		/*iCollSpec := &p.identity.Collection.Spec
		err = p.storage.SetEmailVerified(iCollSpec, []storageT.Filter{
			{Name: iCollSpec.FieldsMap["email"].Name, Value: verif[verifSpecs.FieldsMap["email"].Name]},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}*/

		redirectUrl := c.Query("redirect_url")
		if redirectUrl != "" {
			return c.Redirect(redirectUrl)
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}
