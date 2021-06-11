package pwbased

import (
	coll "aureole/internal/collections"
	"aureole/internal/identity"
	storageT "aureole/internal/plugins/storage/types"
	"aureole/pkg/jsonpath"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"reflect"
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

// todo: refactor getLoginData and getRegisterData methods, maybe create parser struct for this stuff?
func getLoginData(context *pwBased, json interface{}, jsonMap map[string]string, iData *storageT.IdentityData) {
	collMap := context.coll.Parent.Spec.FieldsMap
	i := context.identity

	getLoginTraitData(&i.Username, json, jsonMap["username"], collMap["username"].Default, &iData.Username)
	getLoginTraitData(&i.Email, json, jsonMap["email"], collMap["email"].Default, &iData.Email)
	getLoginTraitData(&i.Phone, json, jsonMap["phone"], collMap["phone"].Default, &iData.Phone)
}

func getLoginTraitData(trait *identity.Trait, json interface{}, fieldPath string, defaultVal interface{}, iDataField *interface{}) {
	if trait.IsEnabled {
		jsonVal, _ := jsonpath.GetJsonPath(fieldPath, json)
		*iDataField = getValueOrDefault(jsonVal, defaultVal)
	}
}

func getValueOrDefault(value, defaultValue interface{}) interface{} {
	if !isZeroVal(value) {
		return value
	} else if !isZeroVal(defaultValue) {
		return defaultValue
	} else {
		return nil
	}
}

func isZeroVal(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func getRegisterData(context *pwBased, json interface{}, jsonMap map[string]string, iData *storageT.IdentityData) (int, error) {
	collMap := context.coll.Parent.Spec.FieldsMap
	i := context.identity

	statusCode, err := getRegisterTraitData(&i.Username, json, jsonMap["username"], collMap["username"].Default, &iData.Username)
	if err != nil {
		return statusCode, err
	}

	statusCode, err = getRegisterTraitData(&i.Email, json, jsonMap["email"], collMap["email"].Default, &iData.Email)
	if err != nil {
		return statusCode, err
	}

	statusCode, err = getRegisterTraitData(&i.Phone, json, jsonMap["phone"], collMap["phone"].Default, &iData.Phone)
	if err != nil {
		return statusCode, err
	}

	statusCode, err = getExtraTraitsData(i.Additional, json, jsonMap, collMap, iData)
	if err != nil {
		return statusCode, err
	}
	return 0, nil
}

func getRegisterTraitData(trait *identity.Trait, json interface{}, fieldPath string, defaultVal interface{}, iDataField *interface{}) (int, error) {
	if trait.IsEnabled {
		jsonVal, err := jsonpath.GetJsonPath(fieldPath, json)
		val := getValueOrDefault(jsonVal, defaultVal)

		if val == nil && trait.IsRequired {
			return fiber.StatusBadRequest, err
		}
		*iDataField = val
	}
	return 0, nil
}

func getExtraTraitsData(traits map[string]identity.ExtraTrait, json interface{}, jsonFieldsMap map[string]string, collFieldsMap map[string]coll.FieldSpec, iData *storageT.IdentityData) (int, error) {
	for traitName, trait := range traits {
		if trait.IsInternal {
			if collFieldsMap[traitName].Default == nil && traits[traitName].IsRequired {
				return fiber.StatusInternalServerError, fmt.Errorf("%s: required value isn't passed", traitName)
			}

			iData.Additional[traitName] = collFieldsMap[traitName].Default
		} else {
			fieldPath, ok := jsonFieldsMap[traitName]
			if !ok {
				fieldPath = fmt.Sprintf("{$.%s}", traitName)
			}

			jsonVal, err := jsonpath.GetJsonPath(fieldPath, json)
			value := getValueOrDefault(jsonVal, collFieldsMap[traitName].Default)
			if value == nil && traits[traitName].IsRequired {
				return fiber.StatusBadRequest, err
			}

			iData.Additional[traitName] = value
		}
	}
	return 0, nil
}

func getPwData(json interface{}, fieldsMap map[string]string, data *storageT.PwBasedData) (int, error) {
	password, err := jsonpath.GetJsonPath(fieldsMap["password"], json)
	if err != nil {
		return fiber.StatusBadRequest, err
	}

	data.Password = password
	return 0, nil
}

func getCredField(context *pwBased, iData *storageT.IdentityData) (string, interface{}, int, error) {
	var credName string
	credVals := map[string]interface{}{}
	i := context.identity

	if iData.Username != nil && isCredential(i.Username) {
		credVals["username"] = iData.Username
		credName = "username"
	}

	if iData.Email != nil && isCredential(i.Email) {
		credVals["email"] = iData.Email
		credName = "email"
	}

	if iData.Phone != nil && isCredential(i.Phone) {
		credVals["phone"] = iData.Phone
		credName = "phone"
	}

	if l := len(credVals); l != 1 {
		return "", nil, fiber.StatusInternalServerError, fmt.Errorf("expects 1 credential, %d got", l)
	}

	return credName, credVals[credName], 0, nil
}

func isCredential(trait identity.Trait) bool {
	return trait.IsCredential && trait.IsRequired && trait.IsUnique
}
