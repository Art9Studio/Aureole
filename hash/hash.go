package hash

func New(algoName string, data RawHashData) (Hasher, error) {
	adapter, err := GetAdapter(algoName)
	if err != nil {
		return nil, err
	}

	conf, err := adapter.NewConfig(data)
	if err != nil {
		return nil, err
	}

	return adapter.GetHasher(conf), nil
}
