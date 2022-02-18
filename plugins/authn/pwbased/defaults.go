package pwbased

import (
	"aureole/internal/configs"
	"aureole/plugins/authn/pwbased/pwhasher"
)

func (c *config) setDefaults() {
	configs.SetDefault(&c.CompatHashers, []pwhasher.Config{})
	configs.SetDefault(&c.Reset.Exp, 3600)
	configs.SetDefault(&c.Verify.Exp, 3600)
}
