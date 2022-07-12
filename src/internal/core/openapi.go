package core

import (
	_ "embed"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
)

var (
	openapiDocStruct = &openapiDoc{}
	defaultErrResp   = openapi3.NewResponse().
				WithDescription("Unauthorized Error").
				WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/ErrorMessage"})

	mfaListResp = openapi3.NewResponse().
			WithDescription("List of MFA methods").
			WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/MFAList"})
)

// todo: delete when end up with swagger hub
type openapiDoc struct {
	doc *openapi3.T
}

// todo: rename ReadJsonString when end up with swagger hub
func (s *openapiDoc) ReadDoc() string {
	docsBytes, err := s.doc.MarshalJSON()
	if err != nil {
		fmt.Printf("cannot marshal openapiDoc docs: %v", err)
		return ""
	}
	return string(docsBytes)
}

func (s *openapiDoc) Assemble(p *project, r *router) error {
	s.doc = &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Aureole Public API",
			Version: p.apiVersion,
		},
	}

	for _, app := range p.apps {
		var err error
		err, s.doc.Paths = assemblePaths(r, app)
		if err != nil {
			return err
		}
	}

	//for _, a := range p.apps {
	//_, err := assembleIssuerResp(a)
	//if err != nil {
	//	return err
	//}

	//err = assembleAuthNDoc(a, authzResp)
	//if err != nil {
	//	return err
	//}
	//
	//err = assemble2FADoc(a, authzResp)
	//if err != nil {
	//	return err
	//}
	//
	//err = assemblePluginsDoc(a)
	//if err != nil {
	//	return err
	//}
	//}
	return nil
}

func assemblePaths(r *router, app *app) (error, openapi3.Paths) {
	paths := openapi3.Paths{}
	for _, route := range r.getAppRoutes()[app.name] {
		pathItem := openapi3.PathItem{}
		fieldName := toCamelCase(route.Method)
		reflect.ValueOf(&pathItem).Elem().FieldByName(fieldName).Set(reflect.ValueOf(route.OAS3Operation))
		paths[route.Path] = &pathItem
	}
	return nil, paths
}

func NewOpenAPIOperation(schema *openapi3.SchemaRef) *openapi3.Operation {
	return &openapi3.Operation{
		RequestBody: &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Content: map[string]*openapi3.MediaType{
					"application/json": {
						Schema: schema,
					},
				},
			},
		},
		Responses: map[string]*openapi3.ResponseRef{
			"200": {
				Value: &openapi3.Response{
					Content: map[string]*openapi3.MediaType{
						"application/json": {},
					},
				},
			},
		},
	}
}

//
//func assembleAuthNDoc(a *app, authzResp *openapi3.Responses) Error {
//	for _, authn := range a.authenticators {
//		if authn != nil {
//			authzRespCopy, err := copyAuthzResp(authzResp)
//			if err != nil {
//				return err
//			}
//
//			paths := authn.GetCustomAppRoutes()
//
//			//pathsJson := appendDefinitions(paths, "authN", authn.GetMetadata().ShortName)
//
//			loginPathItem := (*paths)["/login"]
//			delete(*paths, "/login")
//
//			var handler *openapi3.OAS3Operation
//			if loginPathItem.Get != nil {
//				handler = loginPathItem.Get
//			} else if loginPathItem.Post != nil {
//				handler = loginPathItem.Post
//			}
//
//			handler.Produces = []string{"application/json"}
//			if handler.Responses != nil {
//				errResp, ok := handler.Responses.StatusCodeResponses[401]
//				if ok {
//					authzRespCopy.StatusCodeResponses[401] = errResp
//				}
//			}
//			handler.Responses = authzRespCopy
//			handler.Responses.StatusCodeResponses[202] = mfaListResp
//			handler.Responses.Default = &defaultErrResp
//
//			loginPath := fmt.Sprintf("/%s/%s/login", a.pathPrefix, strings.ReplaceAll(authn.GetMetadata().ShortName, "_", "-"))
//			s.doc.Paths[loginPath] = loginPathItem
//
//			for path, pathItem := range paths.Paths {
//				path = a.pathPrefix + path
//				s.doc.Paths[path] = pathItem
//			}
//		}
//	}
//	return nil
//}
//
//func assemble2FADoc(a *app, authzResp *openapi3.Responses) Error {
//	for _, mfa := range a.mfa {
//		if mfa != nil {
//			authzRespCopy, err := copyAuthzResp(authzResp)
//			if err != nil {
//				return err
//			}
//
//			paths, defs := mfa.GetCustomAppRoutes()
//			pathsJsonBytes, err := paths.MarshalJSON()
//			if err != nil {
//				return err
//			}
//			pathsJson := appendDefinitions(defs, string(pathsJsonBytes), "2fa", mfa.GetMetadata().ShortName)
//			err = paths.UnmarshalJSON([]byte(pathsJson))
//			if err != nil {
//				return err
//			}
//
//			start2FAPathItem, err := assemble2FAStartDocs(paths)
//			if err != nil {
//				return err
//			}
//			verify2FAPathItem, err := assemble2FAVerifyDocs(paths, authzRespCopy)
//			if err != nil {
//				return err
//			}
//
//			pluginType := strings.ReplaceAll(mfa.GetMetadata().ShortName, "_", "-")
//			start2FAPath := fmt.Sprintf("/%s/2fa/%s/start", a.pathPrefix, pluginType)
//			verify2FAPath := fmt.Sprintf("/%s/2fa/%s/verify", a.pathPrefix, pluginType)
//
//			s.doc.Paths.Paths[start2FAPath] = *start2FAPathItem
//			s.doc.Paths.Paths[verify2FAPath] = *verify2FAPathItem
//
//			for path, pathItem := range paths.Paths {
//				path = a.pathPrefix + "/2fa" + path
//				s.doc.Paths.Paths[path] = pathItem
//			}
//		}
//	}
//	return nil
//}
//
//func assemble2FAVerifyDocs(paths *openapi3.Paths, authzResp *openapi3.Responses) (*openapi3.PathItem, Error) {
//	pathItem := paths.Paths["/verify"]
//	delete(paths.Paths, "/verify")
//
//	var handler *openapi3.OAS3Operation
//	if pathItem.Get != nil {
//		handler = pathItem.Get
//	} else if pathItem.Post != nil {
//		handler = pathItem.Post
//	}
//
//	handler.Produces = []string{"application/json"}
//	errResp, ok := handler.Responses.StatusCodeResponses[401]
//	if ok {
//		authzResp.StatusCodeResponses[401] = errResp
//	}
//	handler.Responses = authzResp
//	handler.Responses.Default = &defaultErrResp
//
//	return &pathItem, nil
//}
//
//func assemble2FAStartDocs(paths *openapi3.Paths) (*openapi3.PathItem, Error) {
//	pathItem := paths.Paths["/start"]
//	delete(paths.Paths, "/start")
//
//	var handler *openapi3.OAS3Operation
//	if pathItem.Get != nil {
//		handler = pathItem.Get
//	} else if pathItem.Post != nil {
//		handler = pathItem.Post
//	}
//
//	handler.Produces = []string{"application/json"}
//	handler.Responses.StatusCodeResponses = map[int]openapi3.Response{200: handler.Responses.StatusCodeResponses[200]}
//	handler.Responses.StatusCodeResponses[401] = defaultErrResp
//
//	return &pathItem, nil
//}
//
//func assemblePluginsDoc(a *app) Error {
//	if a.issuer != nil {
//		err := appendPluginSpec(a.issuer, a, "authZ", a.issuer.GetMetadata().ShortName)
//		if err != nil {
//			return err
//		}
//	}
//
//	if a.idManager != nil {
//		err := appendPluginSpec(a.idManager, a, "id_manager", a.idManager.GetMetadata().ShortName)
//		if err != nil {
//			return err
//		}
//	}
//
//	if len(a.cryptoKeys) != 0 {
//		for _, key := range a.cryptoKeys {
//			if key != nil {
//				err := appendPluginSpec(key, a, "crypto_key", key.GetMetadata().ShortName)
//				if err != nil {
//					return err
//				}
//			}
//		}
//	}
//
//	if len(a.cryptoStorages) != 0 {
//		for _, storage := range a.cryptoStorages {
//			if storage != nil {
//				err := appendPluginSpec(storage, a, "crypto_storage", storage.GetMetadata().ShortName)
//				if err != nil {
//					return err
//				}
//			}
//		}
//	}
//
//	if len(a.storages) != 0 {
//		for _, storage := range a.storages {
//			if storage != nil {
//				err := appendPluginSpec(storage, a, "storage", storage.GetMetadata().ShortName)
//				if err != nil {
//					return err
//				}
//			}
//		}
//	}
//
//	if len(a.senders) != 0 {
//		for _, sender := range a.senders {
//			if sender != nil {
//				err := appendPluginSpec(sender, a, "sender", sender.GetMetadata().ShortName)
//				if err != nil {
//					return err
//				}
//			}
//		}
//	}
//
//	if len(a.rootPlugins) != 0 {
//		for _, adminPlugin := range a.rootPlugins {
//			if adminPlugin != nil {
//				err := appendPluginSpec(adminPlugin, a, adminPlugin.GetMetadata().Type, adminPlugin.GetMetadata().ShortName)
//				if err != nil {
//					return err
//				}
//			}
//		}
//	}
//
//	return nil
//}
//
//func appendPluginSpec(Plugin interface{}, a *app, pluginKind Plugin.PluginType, pluginName string) Error {
//	pluginSwagger, ok := Plugin.(Plugin.OpenAPISpecGetter)
//	if ok {
//		paths, defs := pluginSwagger.GetCustomAppRoutes()
//		pathsJsonBytes, err := paths.MarshalJSON()
//		if err != nil {
//			return err
//		}
//
//		pathsJson := appendDefinitions(defs, string(pathsJsonBytes), pluginKind, pluginName)
//		err = paths.UnmarshalJSON([]byte(pathsJson))
//		if err != nil {
//			return err
//		}
//
//		for path, pathItem := range paths.Paths {
//			path = a.pathPrefix + path
//			s.doc.Paths[path] = pathItem
//		}
//	}
//
//	return nil
//}
//
//func appendDefinitions(responses openapi3.Responses, pluginType Plugin.PluginType, pluginName string) *openapi3.Responses {
//	for name, d := range defs {
//		newName := name
//
//		_, ok := s.doc.Definitions[name]
//		if ok {
//			newName = pluginName + "." + name
//			_, ok = s.doc.Definitions[newName]
//			if ok {
//				newName += string(pluginType) + "." + newName
//			}
//		}
//		s.doc.Definitions[newName] = d
//
//		if newName != name {
//			responses = renameRefsToDefs(responses, name, newName)
//		}
//	}
//	return responses
//}
//
//func renameRefsToDefs(pluginSpecsJson, oldName, newName string) string {
//	oldDefRef := "#/definitions/" + oldName
//	newDefRef := "#/definitions/" + newName
//	return strings.ReplaceAll(pluginSpecsJson, oldDefRef, newDefRef)
//}
//
//func copyAuthzResp(authzResp *openapi3.Responses) (*openapi3.Responses, Error) {
//	var resp openapi3.Responses
//	bytes, err := authzResp.MarshalJSON()
//	if err != nil {
//		return nil, err
//	}
//	err = resp.UnmarshalJSON(bytes)
//	if err != nil {
//		return nil, err
//	}
//	return &resp, nil
//}
