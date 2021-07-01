package email

import (
	"aureole/internal/collections"
)

var emailLinkCollType = &collections.CollectionType{
	Name:       "email_link",
	IsAppendix: false,
}
