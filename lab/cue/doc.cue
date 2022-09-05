#Project: close({
	api_version: string
  test_run:    bool
	apps: [...#App]
})

_AppType: "App"

#App: close({
	_type:        _AppType & =~"^\(_AppType)$"
	name:         string
	path_prefix?: string
	internal?: {
		enc_key:  #CryptoKeyPlugin
		sign_key: #CryptoKeyPlugin
		storage:  #StoragePlugin
	}

	auth: [...#AuthPlugin]
	crypto_keys: [...#CryptoKeyPlugin]
})

#PluginConfig: {...}

#Plugin: {
	plugin: string
	name:   string
	config: #PluginConfig
}

_AuthPluginType: "AuthPlugin"

#AuthPlugin: #Plugin & {_type: _AuthPluginType & =~"^\(_AuthPluginType)$"}

_CryptoKeyPluginType: "CryptoKeyPlugin"

#CryptoKeyPlugin: #Plugin & {_type: _CryptoKeyPluginType & =~"^\(_CryptoKeyPluginType)$"}

#StoragePlugin: #Plugin

//
//
//

#EmailAuthPlugin: close(#AuthPlugin & {plugin: "email"})

#JWKCryptoKeyPlugin: close(#CryptoKeyPlugin & {plugin: "jwk"})

//
//
//

_emailAuth: #EmailAuthPlugin & {
	name: "email"
	config: {
		random_data: 1
	}
}

_cryptoKey: #JWKCryptoKeyPlugin & {
	name: "jwk"
	config: {
		random_data: 1
	}
}

#Project & {
	api_version: "0.1"
	apps: [
		#App & {
			name: "App_Lzy"
			auth: [_emailAuth]
			crypto_keys: [_cryptoKey]
		},

		#App & {
			name:        "App_Fl"
			path_prefix: "/full"
		},
	]
}
