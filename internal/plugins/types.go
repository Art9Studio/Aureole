package plugins

type (
	MetaDataGetter interface {
		GetMetaData() Meta
	}

	Meta struct {
		Type string
		Name string
		ID   string
	}
)
