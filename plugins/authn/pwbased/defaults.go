package pwbased

import (
	"aureole/internal/configs"
	"fmt"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.CompatHashers, []string{})
	c.Login.setDefaults()
	c.Register.setDefaults()
	c.Reset.setDefaults()
	c.Verif.setDefaults()
}

func (l *login) setDefaults() {
	configs.SetDefault(&l.Path, "/login")
	l.FieldsMap = setDefaultMap(l.FieldsMap, []string{"username", "email", "phone", "password"})
}

func (r *register) setDefaults() {
	configs.SetDefault(&r.Path, "/register")
	r.FieldsMap = setDefaultMap(r.FieldsMap, []string{"username", "email", "phone", "password"})
}

func (c *resetConf) setDefaults() {
	configs.SetDefault(&c.Path, "/password/reset")
	configs.SetDefault(&c.ConfirmUrl, "/password/reset/confirm")
	c.FieldsMap = setDefaultMap(c.FieldsMap, []string{"email"})
	configs.SetDefault(&c.Token.Exp, 3600)
	configs.SetDefault(&c.Token.HashFunc, "sha256")
}

func (c *verifConf) setDefaults() {
	configs.SetDefault(&c.Path, "/email-verify")
	configs.SetDefault(&c.ConfirmUrl, "/email-verify/confirm")
	c.FieldsMap = setDefaultMap(c.FieldsMap, []string{"email"})
	configs.SetDefault(&c.Token.Exp, 3600)
	configs.SetDefault(&c.Token.HashFunc, "sha256")
}

func setDefaultMap(fieldsMap map[string]string, keys []string) map[string]string {
	if fieldsMap == nil {
		fieldsMap = map[string]string{}
		for _, key := range keys {
			fieldsMap[key] = fmt.Sprintf("{$.%s}", key)
		}
	} else {
		for _, key := range keys {
			if _, ok := fieldsMap[key]; !ok {
				fieldsMap[key] = fmt.Sprintf("{$.%s}", key)
			}
		}
	}
	return fieldsMap
}
