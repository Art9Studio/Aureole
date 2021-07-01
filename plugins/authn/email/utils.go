package email

import (
	coll "aureole/internal/collections"
	"aureole/internal/identity"
	storageT "aureole/internal/plugins/storage/types"
	"aureole/pkg/jsonpath"
	crand "crypto/rand"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/big"
	"reflect"
)

func sendError(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(&fiber.Map{
		"success": false,
		"message": message,
	})
}

func getJsonData(json interface{}, fieldPath string, confirmData *interface{}) (int, error) {
	jsonVal, err := jsonpath.GetJsonPath(fieldPath, json)
	if err != nil {
		return fiber.StatusBadRequest, err
	}

	*confirmData = jsonVal
	return 0, nil
}

func isCredential(trait *identity.Trait) bool {
	return trait.IsCredential && trait.IsRequired && trait.IsUnique
}

func getRegisterData(context *email, json interface{}, jsonMap map[string]string, iData *storageT.IdentityData) (int, error) {
	collMap := context.coll.Spec.FieldsMap
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

func getRandomString(length int, alphabet string) (string, error) {
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		ret[i] = alphabet[num.Int64()]
	}

	return string(ret), nil
}
