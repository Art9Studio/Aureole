module aureole

go 1.15

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	go.uber.org/atomic => github.com/uber-go/atomic v1.9.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/andybalholm/brotli v1.0.3 // indirect
	github.com/coocood/freecache v1.2.0
	github.com/coreos/bbolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/etcd v3.3.27+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v3 v3.0.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/ecies/go v1.0.1
	github.com/go-playground/locales v0.14.0
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/go-redis/redis/v8 v8.11.4
	github.com/go-test/deep v1.0.8
	github.com/gofiber/fiber/v2 v2.24.0
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/vault/api v1.3.1
	github.com/jarcoal/httpmock v1.0.8
	github.com/joho/godotenv v1.4.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/jordan-wright/email v4.0.1-0.20210109023952-943e75fe5223+incompatible
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat-go/jwx v1.2.13
	github.com/matoous/go-nanoid/v2 v2.0.0
	github.com/mitchellh/mapstructure v1.4.3
	github.com/pkg/errors v0.9.1
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.etcd.io/etcd v3.3.27+incompatible
	go.uber.org/multierr v1.7.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/yaml v1.3.0 // indirect
)
