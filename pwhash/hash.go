package pwhash

// New returns desired PwHasher depends on the given config
func New(algoName string, rawConf *RawHashConfig) (PwHasher, error) {
	adapter, err := GetAdapter(algoName)
	if err != nil {
		return nil, err
	}

	return adapter.GetPwHasher(rawConf)
}
