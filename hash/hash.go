package hash

// New returns desired Hasher depends on the given config
func New(algoName string, rawConf *RawHashConfig) (Hasher, error) {
	adapter, err := GetAdapter(algoName)
	if err != nil {
		return nil, err
	}

	return adapter.GetHasher(rawConf)
}
