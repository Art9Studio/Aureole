module aureole

go 1.16

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	go.uber.org/atomic => github.com/uber-go/atomic v1.9.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/arsmn/fiber-swagger/v2 v2.31.1
	github.com/avast/retry-go/v4 v4.0.3
	github.com/coocood/freecache v1.2.1
	github.com/coreos/etcd v3.3.27+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fatih/structs v1.1.0
	github.com/go-openapi/spec v0.20.4
	github.com/go-playground/locales v0.14.0
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-test/deep v1.0.8
	github.com/gofiber/fiber/v2 v2.32.0
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/vault/api v1.5.0
	github.com/jackc/pgx/v4 v4.16.1
	github.com/jackc/tern v1.12.5
	github.com/jarcoal/httpmock v1.1.0
	github.com/joho/godotenv v1.4.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/jordan-wright/email v4.0.1-0.20210109023952-943e75fe5223+incompatible
	github.com/jpillora/overseer v1.1.6
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat-go/jwx v1.2.21
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matoous/go-nanoid/v2 v2.0.0
	github.com/mitchellh/mapstructure v1.4.3
	github.com/pkg/errors v0.9.1
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.1
	github.com/swaggo/swag v1.8.1
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	go.etcd.io/etcd v3.3.27+incompatible
	go.uber.org/multierr v1.8.0
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	golang.org/x/tools v0.1.8 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/yaml v1.3.0 // indirect
)
