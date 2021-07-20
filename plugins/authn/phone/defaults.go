package phone

import (
	"aureole/internal/configs"
	"fmt"
)

func (c *config) setDefaults() {
	c.Login.setDefaults()
	c.Register.setDefaults()
	c.Verification.setDefaults()
}

func (l *login) setDefaults() {
	configs.SetDefault(&l.Path, "/login")
	l.FieldsMap = setDefaultMap(l.FieldsMap, []string{"phone"})
}

func (r *register) setDefaults() {
	configs.SetDefault(&r.Path, "/register")
	r.FieldsMap = setDefaultMap(r.FieldsMap, []string{"username", "email", "phone"})
}

func (v *verifConf) setDefaults() {
	configs.SetDefault(&v.MaxAttempts, 3)
	configs.SetDefault(&v.Path, "/login/verify")
	configs.SetDefault(&v.ResendUrl, "/login/resend")
	v.Otp.setDefaults()
	v.FieldsMap = setDefaultMap(v.FieldsMap, []string{"id", "otp"})
}
func (c *otp) setDefaults() {
	configs.SetDefault(&c.Length, 1)
	configs.SetDefault(&c.Alphabet, "1234567890")
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
