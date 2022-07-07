package core

import (
	_ "embed"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	//go:embed openapi.yaml
	baseDoc []byte

	doc            openapi3.T
	defaultErrResp = openapi3.NewResponse().
			WithDescription("Unauthorized error").
			WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/ErrorMessage"})

	mfaListResp = openapi3.NewResponse().
			WithDescription("List of MFA methods").
			WithJSONSchemaRef(&openapi3.SchemaRef{Ref: "#/components/schemas/MFAList"})
)

type swagger struct{}

func (s *swagger) ReadDoc() string {
	docsBytes, err := doc.MarshalJSON()
	if err != nil {
		fmt.Printf("cannot marshal swagger docs: %v", err)
		return ""
	}
	return string(docsBytes)
}

func assembleDoc(p *project) error {
	err := doc.UnmarshalJSON(baseDoc)
	if err != nil {
		return err
	}

	doc.Info.Version = p.apiVersion
	doc.Paths = openapi3.Paths{}

	for _, a := range p.apps {
		_, err := assembleIssuerResp(a)
		if err != nil {
			return err
		}

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
	}
	return nil
}

func assembleIssuerResp(a *app) (*openapi3.Responses, error) {
	issuer, ok := a.getIssuer()
	if !ok {
		return nil, fmt.Errorf("cannot get issuer for app %s", a.name)
	}

	resp, err := issuer.GetResponseData()

	//resp = appendDefinitions(resp, issuer.GetMetaData().Type, issuer.GetMetaData().Name)

	return resp, err
}

//
//func assembleAuthNDoc(a *app, authzResp *openapi3.Responses) error {
//	for _, authn := range a.authenticators {
//		if authn != nil {
//			authzRespCopy, err := copyAuthzResp(authzResp)
//			if err != nil {
//				return err
//			}
//
//			paths := authn.GetAppRoutes()
//
//			//pathsJson := appendDefinitions(paths, "authN", authn.GetMetaData().Name)
//
//			loginPathItem := (*paths)["/login"]
//			delete(*paths, "/login")
//
//			var handler *openapi3.Operation
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
//			loginPath := fmt.Sprintf("/%s/%s/login", a.pathPrefix, strings.ReplaceAll(authn.GetMetaData().Name, "_", "-"))
//			doc.Paths[loginPath] = loginPathItem
//
//			for path, pathItem := range paths.Paths {
//				path = a.pathPrefix + path
//				doc.Paths[path] = pathItem
//			}
//		}
//	}
//	return nil
//}
//
//func assemble2FADoc(a *app, authzResp *openapi3.Responses) error {
//	for _, mfa := range a.mfa {
//		if mfa != nil {
//			authzRespCopy, err := copyAuthzResp(authzResp)
//			if err != nil {
//				return err
//			}
//
//			paths, defs := mfa.GetAppRoutes()
//			pathsJsonBytes, err := paths.MarshalJSON()
//			if err != nil {
//				return err
//			}
//			pathsJson := appendDefinitions(defs, string(pathsJsonBytes), "2fa", mfa.GetMetaData().Name)
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
//			pluginType := strings.ReplaceAll(mfa.GetMetaData().Name, "_", "-")
//			start2FAPath := fmt.Sprintf("/%s/2fa/%s/start", a.pathPrefix, pluginType)
//			verify2FAPath := fmt.Sprintf("/%s/2fa/%s/verify", a.pathPrefix, pluginType)
//
//			doc.Paths.Paths[start2FAPath] = *start2FAPathItem
//			doc.Paths.Paths[verify2FAPath] = *verify2FAPathItem
//
//			for path, pathItem := range paths.Paths {
//				path = a.pathPrefix + "/2fa" + path
//				doc.Paths.Paths[path] = pathItem
//			}
//		}
//	}
//	return nil
//}
//
//func assemble2FAVerifyDocs(paths *openapi3.Paths, authzResp *openapi3.Responses) (*openapi3.PathItem, error) {
//	pathItem := paths.Paths["/verify"]
//	delete(paths.Paths, "/verify")
//
//	var handler *openapi3.Operation
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
//func assemble2FAStartDocs(paths *openapi3.Paths) (*openapi3.PathItem, error) {
//	pathItem := paths.Paths["/start"]
//	delete(paths.Paths, "/start")
//
//	var handler *openapi3.Operation
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
//func assemblePluginsDoc(a *app) error {
//	if a.issuer != nil {
//		err := appendPluginSpec(a.issuer, a, "authZ", a.issuer.GetMetaData().Name)
//		if err != nil {
//			return err
//		}
//	}
//
//	if a.idManager != nil {
//		err := appendPluginSpec(a.idManager, a, "id_manager", a.idManager.GetMetaData().Name)
//		if err != nil {
//			return err
//		}
//	}
//
//	if len(a.cryptoKeys) != 0 {
//		for _, key := range a.cryptoKeys {
//			if key != nil {
//				err := appendPluginSpec(key, a, "crypto_key", key.GetMetaData().Name)
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
//				err := appendPluginSpec(storage, a, "crypto_storage", storage.GetMetaData().Name)
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
//				err := appendPluginSpec(storage, a, "storage", storage.GetMetaData().Name)
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
//				err := appendPluginSpec(sender, a, "sender", sender.GetMetaData().Name)
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
//				err := appendPluginSpec(adminPlugin, a, adminPlugin.GetMetaData().Type, adminPlugin.GetMetaData().Name)
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
//func appendPluginSpec(Plugin interface{}, a *app, pluginKind Plugin.PluginType, pluginName string) error {
//	pluginSwagger, ok := Plugin.(Plugin.OpenAPISpecGetter)
//	if ok {
//		paths, defs := pluginSwagger.GetAppRoutes()
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
//			doc.Paths[path] = pathItem
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
//		_, ok := doc.Definitions[name]
//		if ok {
//			newName = pluginName + "." + name
//			_, ok = doc.Definitions[newName]
//			if ok {
//				newName += string(pluginType) + "." + newName
//			}
//		}
//		doc.Definitions[newName] = d
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
//func copyAuthzResp(authzResp *openapi3.Responses) (*openapi3.Responses, error) {
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
