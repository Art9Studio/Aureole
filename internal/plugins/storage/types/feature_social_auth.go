package types

import (
	"aureole/internal/collections"
	"encoding/json"
)

type SocialAuthData struct {
	Id         interface{}
	SocialId   interface{}
	Email      interface{}
	Provider   interface{}
	UserData   interface{}
	UserId     interface{}
	Additional map[string]interface{}
}

func NewSocialAuthData(rawData JSONCollResult, specs map[string]collections.FieldSpec) *SocialAuthData {
	data := rawData.(map[string]interface{})
	userDataJson, _ := json.MarshalIndent(data[specs["user_data"].Name], "", "	")

	socAuth := &SocialAuthData{
		Id:         data[specs["id"].Name],
		SocialId:   data[specs["social_id"].Name],
		Email:      data[specs["email"].Name],
		Provider:   data[specs["provider"].Name],
		UserData:   string(userDataJson),
		UserId:     data[specs["user_id"].Name],
		Additional: map[string]interface{}{},
	}

	for fieldName, fieldVal := range data {
		if fieldName != specs["id"].Name &&
			fieldName != specs["social_id"].Name &&
			fieldName != specs["email"].Name &&
			fieldName != specs["provider"].Name &&
			fieldName != specs["user_data"].Name {
			socAuth.Additional[fieldName] = fieldVal
		}
	}

	return socAuth
}

type SocialAuth interface {
	InsertSocialAuth(*collections.Spec, *SocialAuthData) (JSONCollResult, error)

	GetSocialAuth(*collections.Spec, []Filter) (JSONCollResult, error)

	IsSocialAuthExist(*collections.Spec, []Filter) (bool, error)

	LinkAccount(*collections.Spec, []Filter, interface{}) error
}
