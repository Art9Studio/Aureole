package email

import (
	"aureole/internal/configs"
	"fmt"
)

func (c *config) setDefaults() {
	c.Login.setDefaults()
	c.Register.setDefaults()
	c.Link.setDefaults()
}

func (l *login) setDefaults() {
	configs.SetDefault(&l.Path, "/login")
	l.FieldsMap = setDefaultMap(l.FieldsMap, []string{"email"})
}

func (r *register) setDefaults() {
	configs.SetDefault(&r.Path, "/register")
	r.FieldsMap = setDefaultMap(r.FieldsMap, []string{"username", "email", "phone"})
}

func (t *token) setDefaults() {
	configs.SetDefault(&t.Exp, 600)
	configs.SetDefault(&t.HashFunc, "sha256")
}

func (m *magicLinkConf) setDefaults() {
	m.Token.setDefaults()
	configs.SetDefault(&m.Path, "/email-confirm")
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
