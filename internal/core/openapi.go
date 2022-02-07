package core

import (
	"aureole/internal/plugins"
	_ "embed"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/swaggo/swag"
	"strings"
)

var (
	//go:embed swagger.json
	baseSwaggerJson []byte

	swaggerDocs    spec.Swagger
	defaultErrResp spec.Response
	mfaListResp    spec.Response
)

const (
	defaultErrSwagger = `
{
	"description": "Unauthorized error",
	"schema": {
	  "$ref": "#/definitions/ErrorMessage"
	}
}
`

	mfaListSwagger = `
{
	"description": "A map of available 2FA plugins and their links, and token for authenticate request to 2FA",
	"schema": {
		"$ref": "#/definitions/2FAMap"
	}
}
`
)

type swagger struct{}

func (s *swagger) ReadDoc() string {
	docsBytes, err := swaggerDocs.MarshalJSON()
	if err != nil {
		fmt.Printf("cannot marshal swagger docs: %v", err)
		return ""
	}
	return string(docsBytes)
}

func init() {
	swag.Register("swagger", &swagger{})
}

func assembleSwagger() error {
	err := swaggerDocs.UnmarshalJSON(baseSwaggerJson)
	if err != nil {
		return err
	}

	err = defaultErrResp.UnmarshalJSON([]byte(defaultErrSwagger))
	if err != nil {
		return err
	}

	err = mfaListResp.UnmarshalJSON([]byte(mfaListSwagger))
	if err != nil {
		return err
	}

	swaggerDocs.Info.Version = p.apiVersion
	swaggerDocs.Host = "localhost:3000"
	swaggerDocs.Paths = &spec.Paths{Paths: map[string]spec.PathItem{}}

	for _, a := range p.apps {
		authzResp, err := assembleAuthzResp(a)
		if err != nil {
			return err
		}

		err = assembleAuthNSwagger(a, authzResp)
		if err != nil {
			return err
		}

		err = assemble2FASwagger(a, authzResp)
		if err != nil {
			return err
		}

		err = assemblePluginsSwagger(a)
		if err != nil {
			return err
		}
	}
	return nil
}

func assembleAuthzResp(a *app) (*spec.Responses, error) {
	authz, ok := a.getAuthorizer()
	if !ok {
		return nil, fmt.Errorf("cannot get athorizer for app %s", a.name)
	}

	resp, def := authz.GetResponseData()
	respBytes, err := resp.MarshalJSON()
	if err != nil {
		return nil, err
	}

	respJson := appendDefinitions(def, string(respBytes), "authz", authz.GetMetaData().Type)
	err = resp.UnmarshalJSON([]byte(respJson))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func assembleAuthNSwagger(a *app, authzResp *spec.Responses) error {
	for _, authn := range a.authenticators {
		if authn != nil {
			authzRespCopy, err := copyAuthzResp(authzResp)
			if err != nil {
				return err
			}

			paths, defs := authn.GetHandlersSpec()
			pathsJsonBytes, err := paths.MarshalJSON()
			if err != nil {
				return err
			}
			pathsJson := appendDefinitions(defs, string(pathsJsonBytes), "authN", authn.GetMetaData().Type)
			err = paths.UnmarshalJSON([]byte(pathsJson))
			if err != nil {
				return err
			}

			loginPathItem := paths.Paths["/login"]
			delete(paths.Paths, "/login")

			var handler *spec.Operation
			if loginPathItem.Get != nil {
				handler = loginPathItem.Get
			} else if loginPathItem.Post != nil {
				handler = loginPathItem.Post
			}

			handler.Produces = []string{"application/json"}
			if handler.Responses != nil {
				errResp, ok := handler.Responses.StatusCodeResponses[401]
				if ok {
					authzRespCopy.StatusCodeResponses[401] = errResp
				}
			}
			handler.Responses = authzRespCopy
			handler.Responses.StatusCodeResponses[202] = mfaListResp
			handler.Responses.Default = &defaultErrResp

			loginPath := fmt.Sprintf("/%s/%s/login", a.name, strings.ReplaceAll(authn.GetMetaData().Type, "_", "-"))
			swaggerDocs.Paths.Paths[loginPath] = loginPathItem

			for path, pathItem := range paths.Paths {
				path = "/" + a.name + path
				swaggerDocs.Paths.Paths[path] = pathItem
			}
		}
	}
	return nil
}

func assemble2FASwagger(a *app, authzResp *spec.Responses) error {
	for _, mfa := range a.secondFactors {
		if mfa != nil {
			authzRespCopy, err := copyAuthzResp(authzResp)
			if err != nil {
				return err
			}

			paths, defs := mfa.GetHandlersSpec()
			pathsJsonBytes, err := paths.MarshalJSON()
			if err != nil {
				return err
			}
			pathsJson := appendDefinitions(defs, string(pathsJsonBytes), "2fa", mfa.GetMetaData().Type)
			err = paths.UnmarshalJSON([]byte(pathsJson))
			if err != nil {
				return err
			}

			start2FAPathItem, err := assemble2FAStartDocs(paths)
			if err != nil {
				return err
			}
			verify2FAPathItem, err := assemble2FAVerifyDocs(paths, authzRespCopy)
			if err != nil {
				return err
			}

			pluginType := strings.ReplaceAll(mfa.GetMetaData().Type, "_", "-")
			start2FAPath := fmt.Sprintf("/%s/2fa/%s/start", a.name, pluginType)
			verify2FAPath := fmt.Sprintf("/%s/2fa/%s/verify", a.name, pluginType)

			swaggerDocs.Paths.Paths[start2FAPath] = *start2FAPathItem
			swaggerDocs.Paths.Paths[verify2FAPath] = *verify2FAPathItem

			for path, pathItem := range paths.Paths {
				path = "/" + a.name + "/2fa" + path
				swaggerDocs.Paths.Paths[path] = pathItem
			}
		}
	}
	return nil
}

func assemble2FAVerifyDocs(paths *spec.Paths, authzResp *spec.Responses) (*spec.PathItem, error) {
	pathItem := paths.Paths["/verify"]
	delete(paths.Paths, "/verify")

	var handler *spec.Operation
	if pathItem.Get != nil {
		handler = pathItem.Get
	} else if pathItem.Post != nil {
		handler = pathItem.Post
	}

	handler.Produces = []string{"application/json"}
	errResp, ok := handler.Responses.StatusCodeResponses[401]
	if ok {
		authzResp.StatusCodeResponses[401] = errResp
	}
	handler.Responses = authzResp
	handler.Responses.Default = &defaultErrResp

	return &pathItem, nil
}

func assemble2FAStartDocs(paths *spec.Paths) (*spec.PathItem, error) {
	pathItem := paths.Paths["/start"]
	delete(paths.Paths, "/start")

	var handler *spec.Operation
	if pathItem.Get != nil {
		handler = pathItem.Get
	} else if pathItem.Post != nil {
		handler = pathItem.Post
	}

	handler.Produces = []string{"application/json"}
	handler.Responses.StatusCodeResponses = map[int]spec.Response{200: handler.Responses.StatusCodeResponses[200]}
	handler.Responses.StatusCodeResponses[401] = defaultErrResp

	return &pathItem, nil
}

func assemblePluginsSwagger(a *app) error {
	if a.authorizer != nil {
		err := appendPluginSpec(a.authorizer, a, "authZ", a.authorizer.GetMetaData().Type)
		if err != nil {
			return err
		}
	}

	if a.idManager != nil {
		err := appendPluginSpec(a.idManager, a, "id_manager", a.idManager.GetMetaData().Type)
		if err != nil {
			return err
		}
	}

	if len(a.cryptoKeys) != 0 {
		for _, key := range a.cryptoKeys {
			if key != nil {
				err := appendPluginSpec(key, a, "crypto_key", key.GetMetaData().Type)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(a.cryptoStorages) != 0 {
		for _, storage := range a.cryptoStorages {
			if storage != nil {
				err := appendPluginSpec(storage, a, "crypto_storage", storage.GetMetaData().Type)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(a.storages) != 0 {
		for _, storage := range a.storages {
			if storage != nil {
				err := appendPluginSpec(storage, a, "storage", storage.GetMetaData().Type)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(a.senders) != 0 {
		for _, sender := range a.senders {
			if sender != nil {
				err := appendPluginSpec(sender, a, "sender", sender.GetMetaData().Type)
				if err != nil {
					return err
				}
			}
		}
	}

	if len(a.admins) != 0 {
		for _, adminPlugin := range a.admins {
			if adminPlugin != nil {
				err := appendPluginSpec(adminPlugin, a, "admin", adminPlugin.GetMetaData().Type)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func appendPluginSpec(plugin interface{}, a *app, pluginKind, pluginType string) error {
	pluginSwagger, ok := plugin.(plugins.OpenAPISpecGetter)
	if ok {
		paths, defs := pluginSwagger.GetHandlersSpec()
		pathsJsonBytes, err := paths.MarshalJSON()
		if err != nil {
			return err
		}

		pathsJson := appendDefinitions(defs, string(pathsJsonBytes), pluginKind, pluginType)
		err = paths.UnmarshalJSON([]byte(pathsJson))
		if err != nil {
			return err
		}

		for path, pathItem := range paths.Paths {
			path = "/" + a.name + path
			swaggerDocs.Paths.Paths[path] = pathItem
		}
	}

	return nil
}

func appendDefinitions(defs spec.Definitions, pluginSpecsJson, pluginKind, pluginType string) string {
	for name, d := range defs {
		newName := name

		_, ok := swaggerDocs.Definitions[name]
		if ok {
			newName = pluginType + "." + name
			_, ok = swaggerDocs.Definitions[newName]
			if ok {
				newName += pluginKind + "." + newName
			}
		}
		swaggerDocs.Definitions[newName] = d

		if newName != name {
			pluginSpecsJson = renameRefsToDefs(pluginSpecsJson, name, newName)
		}
	}
	return pluginSpecsJson
}

func renameRefsToDefs(pluginSpecsJson, oldName, newName string) string {
	oldDefRef := "#/definitions/" + oldName
	newDefRef := "#/definitions/" + newName
	return strings.ReplaceAll(pluginSpecsJson, oldDefRef, newDefRef)
}

func copyAuthzResp(authzResp *spec.Responses) (*spec.Responses, error) {
	var resp spec.Responses
	bytes, err := authzResp.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = resp.UnmarshalJSON(bytes)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
