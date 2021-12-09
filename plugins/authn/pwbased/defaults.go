package pwbased

import (
	"aureole/internal/configs"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.CompatHashers, []string{})
	c.Reset.setDefaults()
	c.Verif.setDefaults()
}

func (c *resetConf) setDefaults() {
	configs.SetDefault(&c.Exp, 3600)
}

func (c *verifConf) setDefaults() {
	configs.SetDefault(&c.Exp, 3600)
}
