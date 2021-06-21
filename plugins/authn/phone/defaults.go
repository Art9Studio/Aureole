package phone

import (
	"aureole/internal/configs"
	"fmt"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.ResendUrl, "/login/resend")
	c.Login.setDefaults()
	c.Register.setDefaults()
	c.Verification.setDefaults()
}

func (l *login) setDefaults() {
	configs.SetDefault(&l.Path, "/login")
	l.FieldsMap = setDefaultMap(l.FieldsMap, []string{"username", "email", "phone"})
}

func (r *register) setDefaults() {
	configs.SetDefault(&r.Path, "/register")
	r.FieldsMap = setDefaultMap(r.FieldsMap, []string{"username", "email", "phone"})
}

func (c *verification) setDefaults() {
	configs.SetDefault(&c.MaxAttempts, 3)
	configs.SetDefault(&c.Path, "/login/verify")
	c.Code.setDefaults()
	c.FieldsMap = setDefaultMap(c.FieldsMap, []string{"id", "code"})
}
func (c *verificationCode) setDefaults() {
	configs.SetDefault(&c.Length, 6)
	configs.SetDefault(&c.Alphabet, "1234567890")
	configs.SetDefault(&c.Prefix, "A-")
	configs.SetDefault(&c.Exp, 300)
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
