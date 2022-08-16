package sms

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	_ "embed"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/mitchellh/mapstructure"
)

//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

// init initializes package by register pluginCreator
func init() {
	meta = core.MFARepo.Register(rawMeta, Create)
}

type (
	sms struct {
		pluginAPI     core.PluginAPI
		rawConf       configs.PluginConfig
		conf          *config
		sender        core.Sender
		tmpl, tmplExt string
	}

	token struct {
		Token string `json:"token"`
	}

	InitMFAReqBody struct {
		token
	}

	VerifyReqBody struct {
		token
		Otp string `json:"otp"`
	}
)

func Create(conf configs.PluginConfig) core.MFA {
	return &sms{rawConf: conf}
}

func (s *sms) Init(api core.PluginAPI) (err error) {
	s.pluginAPI = api
	s.conf, err = initConfig(&s.rawConf.Config)
	if err != nil {
		return err
	}

	var ok bool
	s.sender, ok = s.pluginAPI.GetSender(s.conf.Sender)
	if !ok {
		return fmt.Errorf("sender for app '%s' is not declared", s.pluginAPI.GetAppName())
	}

	if err != nil {
		fmt.Printf("sms 2fa: cannot marshal swagger docs: %v", err)
	}

	tmpl, err := os.ReadFile(s.conf.TmplPath)
	if err != nil {
		s.tmpl = defaultTmpl
		s.tmplExt = "txt"
	} else {
		s.tmpl = string(tmpl)
		s.tmplExt = path.Ext(s.conf.TmplPath)
	}

	return nil
}

func (s *sms) GetMetadata() core.Metadata {
	return meta
}

func (s *sms) IsEnabled(cred *core.Credential) (bool, error) {
	return s.pluginAPI.IsMFAEnabled(cred)
}

func (s *sms) InitMFA() core.MFAInitFunc {
	return func(c fiber.Ctx) (fiber.Map, error) {
		strToken := &InitMFAReqBody{}
		if err := c.BodyParser(strToken); err != nil {
			return nil, err
		}
		if strToken.Token == "" {
			return nil, errors.New("token are required")
		}

		var (
			provider string
			cred     core.Credential
		)
		t, err := s.pluginAPI.ParseJWTService(strToken.Token)
		if err != nil {
			return nil, err
		}
		err = s.pluginAPI.GetFromJWT(t, "provider", &provider)
		if err != nil {
			return nil, errors.New("cannot get provider from token")
		}
		err = s.pluginAPI.GetFromJWT(t, "credential", &cred)
		if err != nil {
			return nil, errors.New("cannot get credential from token")
		}

		otp, err := s.pluginAPI.GetRandStr(s.conf.Otp.Length, s.conf.Otp.Alphabet)
		if err != nil {
			return nil, err
		}

		token, err := s.pluginAPI.CreateJWT(
			map[string]interface{}{
				"phone":    cred.Value,
				"provider": provider,
				"attempts": 0,
			},
			s.conf.Otp.Exp)
		if err != nil {
			return nil, err
		}

		encOtp, err := s.pluginAPI.Encrypt(otp)
		if err != nil {
			return nil, err
		}
		err = s.pluginAPI.SaveToService(cred.Value, encOtp, s.conf.Otp.Exp)
		if err != nil {
			return nil, err
		}
		fmt.Println(otp)
		err = s.sender.Send(cred.Value, "", s.tmpl, s.tmplExt, map[string]interface{}{"otp": otp})
		if err != nil {
			return nil, err
		}

		return fiber.Map{"token": token}, nil
	}
}

func (s *sms) GetOAS3AuthRequestBody() *openapi3.RequestBody {
	schema, _ := openapi3gen.NewSchemaRefForValue(&InitMFAReqBody{}, nil)
	return &openapi3.RequestBody{
		Description: "Token",
		Required:    true,
		Content: map[string]*openapi3.MediaType{
			fiber.MIMEApplicationJSON: {
				Schema: schema,
			},
		},
	}
}

func (s *sms) GetOAS3AuthParameters() openapi3.Parameters {
	return openapi3.Parameters{}
}

func (s *sms) Verify() core.MFAVerifyFunc {
	return func(c fiber.Ctx) (*core.Credential, fiber.Map, error) {
		otp := &VerifyReqBody{}
		if err := c.BodyParser(otp); err != nil {
			return nil, nil, err
		}
		if otp.Token == "" || otp.Otp == "" {
			return nil, nil, errors.New("token and otp are required")
		}

		var (
			phone    string
			attempts int
		)

		t, err := s.pluginAPI.ParseJWTService(otp.Token)
		if err != nil {
			return nil, nil, err
		}
		err = s.pluginAPI.GetFromJWT(t, "phone", &phone)
		if err != nil {
			return nil, nil, errors.New("cannot get otp from token")
		}
		err = s.pluginAPI.GetFromJWT(t, "attempts", &attempts)
		if err != nil {
			return nil, nil, errors.New("cannot get attempts from token")
		}
		if err := s.pluginAPI.InvalidateJWT(t); err != nil {
			return nil, nil, err
		}

		if attempts >= s.conf.MaxAttempts {
			return nil, nil, errors.New("too much attempts")
		}

		var (
			encOtp  []byte
			decrOtp string
		)
		ok, err := s.pluginAPI.GetFromService(phone, &encOtp)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, errors.New("otp has expired")
		}
		err = s.pluginAPI.Decrypt(encOtp, &decrOtp)
		if err != nil {
			return nil, nil, err
		}

		if decrOtp == otp.Otp {
			return &core.Credential{
				Name:  "phone",
				Value: phone,
			}, nil, nil
		} else {
			token, err := s.pluginAPI.CreateJWT(
				map[string]interface{}{
					"phone":    phone,
					"attempts": int(attempts) + 1,
				},
				s.conf.Otp.Exp)
			if err != nil {
				return nil, nil, err
			}
			return nil, fiber.Map{"token": token}, errors.New("wrong otp")
		}
	}
}

func (s *sms) GetOAS3VerifyRequestBody() *openapi3.RequestBody {
	schema, _ := openapi3gen.NewSchemaRefForValue(&VerifyReqBody{}, nil)
	return &openapi3.RequestBody{
		Description: "Token & OTP",
		Required:    true,
		Content: map[string]*openapi3.MediaType{
			fiber.MIMEApplicationJSON: {
				Schema: schema,
			},
		},
	}
}

func (s *sms) GetOAS3VerifyParameters() openapi3.Parameters {
	return openapi3.Parameters{}
}

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}
	PluginConf.setDefaults()
	return PluginConf, nil
}

func (s *sms) GetCustomAppRoutes() []*core.Route {
	Init2FASchema, _ := openapi3gen.NewSchemaRefForValue(&InitMFAReqBody{}, nil)
	resendResSchema, _ := openapi3gen.NewSchemaRefForValue(&token{}, nil)
	verifyReqSchema, _ := openapi3gen.NewSchemaRefForValue(&VerifyReqBody{}, nil)
	sendReqSchema, _ := openapi3gen.NewSchemaRefForValue(&sendOTPReqBody{}, nil)
	return []*core.Route{
		{
			Method:        http.MethodPost,
			Path:          resendUrl,
			Handler:       resend(s),
			OAS3Operation: assembleOAS3Operation(Init2FASchema, resendResSchema),
		},
		{
			Method:        http.MethodPost,
			Path:          initMFA,
			Handler:       authMiddleware(s, initMFASMS(s)),
			OAS3Operation: assembleOAS3Operation(verifyReqSchema, nil),
		},
		{
			Method:        http.MethodPost,
			Path:          send,
			Handler:       authMiddleware(s, sendOTP(s)),
			OAS3Operation: assembleOAS3Operation(sendReqSchema, resendResSchema),
		},
	}
}

func assembleOAS3Operation(reqSchema, resSchema *openapi3.SchemaRef) *openapi3.Operation {
	okResponse := "OK"
	badReqResponse := "Bad Request"
	internalErrorResponse := "Internal Server Error"
	operation := &openapi3.Operation{
		OperationID: meta.ShortName,
		Description: meta.DisplayName,
		Tags:        []string{fmt.Sprintf("mfa by %s", meta.DisplayName)},
		RequestBody: &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required: true,
				Content: map[string]*openapi3.MediaType{
					fiber.MIMEApplicationJSON: {
						Schema: reqSchema,
					},
				},
			},
		},
		Responses: map[string]*openapi3.ResponseRef{
			strconv.Itoa(http.StatusOK): {
				Value: core.AssembleOAS3OKResponse(&okResponse, resSchema),
			},
			strconv.Itoa(http.StatusBadRequest): {
				Value: core.AssembleOAS3ErrResponse(&badReqResponse),
			},
			strconv.Itoa(http.StatusInternalServerError): {
				Value: core.AssembleOAS3ErrResponse(&internalErrorResponse),
			},
		},
	}

	return operation
}
