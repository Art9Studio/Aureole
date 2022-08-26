package core

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const UserID = "userID"

type (
	PluginAPI struct {
		app       *app
		project   *project
		router    *router
		keyPrefix string
	}

	option func(api *PluginAPI)
)

func initAPI(options ...option) PluginAPI {
	api := PluginAPI{}

	for _, option := range options {
		option(&api)
	}

	return api
}

func withKeyPrefix(prefix string) option {
	return func(api *PluginAPI) {
		api.keyPrefix = prefix
	}
}

func withProject(project *project) option {
	return func(api *PluginAPI) {
		api.project = project
	}
}

func withApp(app *app) option {
	return func(api *PluginAPI) {
		api.app = app
	}
}

func withRouter(r *router) option {
	return func(api *PluginAPI) {
		api.router = r
	}
}

func (api PluginAPI) IsTestRun() bool {
	return api.project.testRun
}

func (api PluginAPI) IsMFAEnabled(cred *Credential) (bool, error) {
	manager, ok := api.app.getIDManager()
	if !ok {
		return false, nil
	}

	ret, err := manager.IsMFAEnabled(cred)
	if err != nil && !errors.Is(err, ErrNoUser) {
		return false, nil
	}
	if ret {
		return true, nil
	}
	return false, nil
}

func (api PluginAPI) GetAppName() string {
	// todo: удлаить этот метод и сделать логирование через плагин апи, при получении плагина логировать ошибку
	return api.app.name
}

func (api PluginAPI) GetAppUrl() url.URL {
	return *api.app.url
}

func (api PluginAPI) GetAppPathPrefix() string {
	return api.app.pathPrefix
}

func (api PluginAPI) GetAuthSessionExp() int {
	return api.app.authSessionExp
}

func (api PluginAPI) GetIssuer() (Issuer, bool) {
	return api.app.getIssuer()
}

func (api PluginAPI) GetSecondFactors() (map[string]MFA, bool) {
	return api.app.getSecondFactors()
}

func (api PluginAPI) GetStorage(name string) (Storage, bool) {
	return api.app.getStorage(name)
}

func (api PluginAPI) GetIDManager() (IDManager, bool) {
	return api.app.getIDManager()
}

func (api PluginAPI) GetCryptoStorage(name string) (CryptoStorage, bool) {
	return api.app.getCryptoStorage(name)
}

func (api PluginAPI) GetSender(name string) (Sender, bool) {
	return api.app.getSender(name)
}

func (api PluginAPI) GetCryptoKey(name string) (CryptoKey, bool) {
	return api.app.getCryptoKey(name)
}

func (api PluginAPI) AddProjectRoutes(routes []*Route) {
	api.router.addProjectRoutes(routes)
}

func (api PluginAPI) GetAppRoutes() map[string][]*ExtendedRoute {
	return api.router.getAppRoutes()
}

func (api PluginAPI) GetProjectRoutes() []*Route {
	return api.router.getProjectRoutes()
}

func (api PluginAPI) SaveToService(k string, v interface{}, exp int) error {
	serviceStorage, ok := api.app.getServiceStorage()
	if !ok {
		return errors.New("can't find internal storage")
	}
	return serviceStorage.Set(api.keyPrefix+k, v, exp)
}

func (api PluginAPI) GetFromService(k string, v interface{}) (ok bool, err error) {
	serviceStorage, ok := api.app.getServiceStorage()
	if !ok {
		return false, errors.New("can't find internal storage")
	}
	return serviceStorage.Get(api.keyPrefix+k, v)
}

func (api PluginAPI) Encrypt(data interface{}) ([]byte, error) {
	return encrypt(api.app, data)
}

func (api PluginAPI) Decrypt(data []byte, value interface{}) error {
	return decrypt(api.app, data, value)
}

func (api PluginAPI) GetRandStr(length int, alphabet string) (string, error) {
	return getRandStr(length, alphabet)
}

func (api PluginAPI) CreateJWT(payload map[string]interface{}, exp int) (string, error) {
	return createJWT(api.app, payload, exp)
}

func (api PluginAPI) ParseJWTService(rawToken string) (jwt.Token, error) {
	return parseJWTService(api.app, rawToken)
}

func (api PluginAPI) ParseJWT(rawToken string) (jwt.Token, error) {
	return parseJWT(api.app, rawToken)
}

func (api PluginAPI) InvalidateJWT(token jwt.Token) error {
	return invalidateJWT(api.app, token)
}

func (api PluginAPI) GetFromJWT(token jwt.Token, name string, value interface{}) error {
	return getFromJWT(token, name, value)
}

func (api PluginAPI) Filter(fields, filters map[string]string) (bool, error) {
	for fieldName, pattern := range filters {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		if !re.MatchString(fields[fieldName]) {
			return false, nil
		}
	}
	return true, nil
}

func (api PluginAPI) GetAuthRoute(shortName string) string {
	return path.Clean(getPluginPathPrefix(api.app.pathPrefix, shortName) + AuthPipelinePath)
}

func (api PluginAPI) GetUserID(ctx *fiber.Ctx) string {
	idRaw := ctx.Locals(UserID)
	id, ok := idRaw.(string)

	if !ok {
		return ""
	}
	return id
}

func generateScratchCodes(num int, alphabet string) ([]string, error) {
	scratchCodes := make([]string, num)
	var err error
	for i := 0; i < num; i++ {
		scratchCodes[i], err = getRandStr(8, alphabet)
		if err != nil {
			return nil, err
		}
	}
	return scratchCodes, err
}

type GetScratchCodesBody struct {
	Id string `json:"id"`
}

func authMiddleware(pluginAPI PluginAPI, next fiber.Handler) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		bearer := ctx.Get(fiber.HeaderAuthorization)
		tokenSplit := strings.Split(bearer, "Bearer ")

		var rawToken string
		if len(tokenSplit) == 2 && tokenSplit[1] != "" {
			rawToken = tokenSplit[1]
		} else {
			return ctx.SendStatus(http.StatusForbidden)
		}

		token, err := pluginAPI.ParseJWT(rawToken)
		if err != nil {
			return SendError(ctx, http.StatusForbidden, err.Error())
		}

		var id string
		if err = pluginAPI.GetFromJWT(token, Sub, &id); err != nil {
			return SendError(ctx, http.StatusForbidden, err.Error())
		}
		ctx.Locals(UserID, id)

		return next(ctx)
	}
}

func GetScratchCodes(app *app) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		id := c.Locals(UserID).(string)
		cred := &Credential{Name: ID, Value: id}

		manager, ok := app.getIDManager()
		if !ok {
			return errors.New("cannot get IDManager")
		}

		ok, err := manager.IsMFAEnabled(cred)
		if err != nil {
			return SendError(c, http.StatusInternalServerError, err.Error())
		}
		if !ok {
			return SendError(c, http.StatusBadRequest, "mfa not enabled")
		}

		scratchCodes, err := generateScratchCodes(app.scratchCode.num, app.scratchCode.alphabet)
		if err != nil {
			return SendError(c, http.StatusInternalServerError, err.Error())
		}

		toString := func() *string {
			sb := strings.Builder{}
			for i, s := range scratchCodes {
				sb.WriteString(s)
				if err != nil {
					return nil
				}
				if i < len(scratchCodes)-1 {
					sb.WriteByte(',')

				}
			}
			res := sb.String()
			return &res
		}
		res := toString()
		if err != nil {
			return SendError(c, http.StatusInternalServerError, err.Error())
		}
		if _, err = manager.RegisterOrUpdate(
			&AuthResult{
				Cred:       cred,
				ProviderId: "0",
				Secrets:    &Secrets{scrCodes: res},
			},
		); err != nil {
			return SendError(c, http.StatusInternalServerError, err.Error())
		}
		return c.JSON(&getScratchCodeResp{scrCodes: *res})
	}
}

type AuthScratchCodesBody struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func GetAuthRecoveryCodes(app *app) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		in := &AuthScratchCodesBody{}
		if err := ctx.BodyParser(in); err != nil {
			return SendError(ctx, http.StatusBadRequest, err.Error())
		}
		if in.Email == "" || in.Code == "" {
			return SendError(ctx, http.StatusBadRequest, "email and code are required")
		}

		cred := &Credential{Name: Email, Value: in.Email}
		manager, ok := app.getIDManager()
		if !ok {
			return SendError(ctx, http.StatusInternalServerError, "cannot get id manager")
		}

		if err := manager.UseScratchCode(cred, in.Code); err != nil {
			return SendError(ctx, http.StatusBadRequest, fmt.Sprintf("incorrect or invalid code: %s", err.Error()))
		}
		return authorize(ctx, app, &User{Email: &in.Email})
	}
}
