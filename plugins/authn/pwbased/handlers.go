package pwbased

import (
	"aureole/internal/plugins/authn/types"
	storageT "aureole/internal/plugins/storage/types"
	"encoding/base64"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
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
			token, err := uuid.NewV4()
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			tokenHash := p.verif.hasher().Sum([]byte(token.String()))
			verifData := &storageT.EmailVerifData{
				Email:   identity.Email,
				Token:   base64.StdEncoding.EncodeToString(tokenHash),
				Expires: time.Now().Add(time.Duration(p.conf.Verif.Token.Exp) * time.Second).Format(time.RFC3339),
				Invalid: false,
			}

			verifSpecs := &p.verif.coll.Spec
			err = p.storage.InvalidateEmailVerif(verifSpecs, []storageT.Filter{
				{Name: verifSpecs.FieldsMap["email"].Name, Value: identity.Email},
			})
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			_, err = p.storage.InsertEmailVerif(verifSpecs, verifData)
			if err != nil {
				return sendError(c, fiber.StatusInternalServerError, err.Error())
			}

			link := initConfirmLink(p.verif.confirmLink, token.String())
			err = p.verif.sender.Send(verifData.Email.(string),
				"Verify your email",
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

		token, err := uuid.NewV4()
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		tokenHash := p.reset.hasher().Sum(token.Bytes())
		resetData := &storageT.PwResetData{
			Email:   identityData.Email,
			Token:   base64.StdEncoding.EncodeToString(tokenHash),
			Expires: time.Now().Add(time.Duration(p.conf.Reset.Token.Exp) * time.Second).Format(time.RFC3339),
			Invalid: false,
		}

		collSpec := &p.reset.coll.Spec
		err = p.storage.InvalidateReset(collSpec, []storageT.Filter{
			{Name: collSpec.FieldsMap["email"].Name, Value: identityData.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = p.storage.InsertReset(&p.reset.coll.Spec, resetData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := initConfirmLink(p.reset.confirmLink, token.String())
		err = p.reset.sender.Send(resetData.Email.(string),
			"Reset your password",
			p.conf.Reset.Template,
			map[string]interface{}{"link": link})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(&fiber.Map{"status": "success"})
	}
}

func ResetConfirm(p *pwBased) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		t := c.Query("token")
		if t == "" {
			return sendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := uuid.FromString(strings.TrimRight(t, "\n"))
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		tokenHash := p.verif.hasher().Sum(token.Bytes())

		resetSpecs := &p.reset.coll.Spec
		tokenName := p.reset.coll.Spec.FieldsMap["token"].Name
		rawReset, err := p.storage.GetReset(resetSpecs, []storageT.Filter{
			{Name: tokenName, Value: base64.StdEncoding.EncodeToString(tokenHash)},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		reset, ok := rawReset.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get reset data from database")
		}

		if reset[resetSpecs.FieldsMap["invalid"].Name].(bool) {
			return sendError(c, fiber.StatusUnauthorized, "invalid token")
		}

		expires, err := time.Parse(time.RFC3339, reset[resetSpecs.FieldsMap["expires"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if time.Now().After(expires) {
			return sendError(c, fiber.StatusUnauthorized, "link expire")
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

		err = p.storage.InvalidateReset(resetSpecs, []storageT.Filter{
			{Name: tokenName, Value: base64.StdEncoding.EncodeToString(tokenHash)},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		err = p.reset.sender.SendRaw(reset[resetSpecs.FieldsMap["email"].Name].(string),
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

		fieldName := p.coll.Parent.Spec.FieldsMap["email"].Name
		exist, err := p.storage.IsIdentityExist(p.identity, []storageT.Filter{
			{Name: fieldName, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if !exist {
			return sendError(c, fiber.StatusUnauthorized, "user doesn't exist")
		}

		token, err := uuid.NewV4()
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		tokenHash := p.verif.hasher().Sum(token.Bytes())
		verifData := &storageT.EmailVerifData{
			Email:   identity.Email,
			Token:   base64.StdEncoding.EncodeToString(tokenHash),
			Expires: time.Now().Add(time.Duration(p.conf.Verif.Token.Exp) * time.Second).Format(time.RFC3339),
			Invalid: false,
		}

		verifSpecs := &p.verif.coll.Spec
		err = p.storage.InvalidateEmailVerif(verifSpecs, []storageT.Filter{
			{Name: verifSpecs.FieldsMap["email"].Name, Value: identity.Email},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		_, err = p.storage.InsertEmailVerif(verifSpecs, verifData)
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}

		link := initConfirmLink(p.verif.confirmLink, token.String())
		err = p.verif.sender.Send(verifData.Email.(string),
			"Verify your email",
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
		t := c.Query("token")
		if t == "" {
			return sendError(c, fiber.StatusNotFound, "token not found")
		}

		token, err := uuid.FromString(strings.TrimRight(t, "\n"))
		if err != nil {
			return sendError(c, fiber.StatusBadRequest, err.Error())
		}
		tokenHash := p.verif.hasher().Sum(token.Bytes())

		verifSpecs := &p.verif.coll.Spec
		tokenName := p.verif.coll.Spec.FieldsMap["token"].Name
		rawVerif, err := p.storage.GetEmailVerif(verifSpecs, []storageT.Filter{
			{Name: tokenName, Value: base64.StdEncoding.EncodeToString(tokenHash)},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, errors.Wrap(err, "error get email verify").Error())
		}

		verif, ok := rawVerif.(map[string]interface{})
		if !ok {
			return sendError(c, fiber.StatusInternalServerError, "cannot get magic link data from database")
		}

		if verif[verifSpecs.FieldsMap["invalid"].Name].(bool) {
			return sendError(c, fiber.StatusUnauthorized, "invalid token")
		}

		expires, err := time.Parse(time.RFC3339, verif[verifSpecs.FieldsMap["expires"].Name].(string))
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, err.Error())
		}
		if time.Now().After(expires) {
			return sendError(c, fiber.StatusUnauthorized, "link expire")
		}

		err = p.storage.InvalidateEmailVerif(verifSpecs, []storageT.Filter{
			{Name: tokenName, Value: base64.StdEncoding.EncodeToString(tokenHash)},
		})
		if err != nil {
			return sendError(c, fiber.StatusInternalServerError, errors.Wrap(err, "error invalidate email verify").Error())
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
