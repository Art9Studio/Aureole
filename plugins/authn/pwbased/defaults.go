package pwbased

import "aureole/configs"

func (c *сonfig) setDefaults() {
	configs.SetDefault(&c.CompatHashers, []string{})
	c.Login.setDefaults()
	c.Register.setDefaults()
}

func (l *login) setDefaults() {
	configs.SetDefault(&l.Path, "/login")

	if l.FieldsMap == nil {
		l.FieldsMap = map[string]string{
			"identity": "{$.username}",
			"password": "{$.password}",
		}
	} else {
		if _, ok := l.FieldsMap["identity"]; !ok {
			l.FieldsMap["identity"] = "{$.username}"
		}

		if _, ok := l.FieldsMap["password"]; !ok {
			l.FieldsMap["password"] = "{$.password}"
		}
	}
}

func (r *register) setDefaults() {
	configs.SetDefault(&r.Path, "/register")

	if r.FieldsMap == nil {
		r.FieldsMap = map[string]string{
			"identity": "{$.username}",
			"password": "{$.password}",
		}
	} else {
		if _, ok := r.FieldsMap["identity"]; !ok {
			r.FieldsMap["identity"] = "{$.username}"
		}

		if _, ok := r.FieldsMap["password"]; !ok {
			r.FieldsMap["password"] = "{$.password}"
		}
	}
}
