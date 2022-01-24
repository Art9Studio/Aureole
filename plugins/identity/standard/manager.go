package standard

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/internal/plugins"
	"aureole/plugins/identity/standard/migrations"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/tern/migrate"
	"github.com/mitchellh/mapstructure"
)

const pluginID = "4634"

type manager struct {
	pluginApi core.PluginAPI
	rawConf   *configs.IDManager
	conf      *config
	pool      *pgxpool.Pool
	features  map[string]bool
}

func (m *manager) Init(api core.PluginAPI) (err error) {
	m.pluginApi = api
	m.conf, err = initConfig(&m.rawConf.Config)
	if err != nil {
		return err
	}

	m.pool, err = pgxpool.Connect(context.Background(), m.conf.DBUrl)
	if err != nil {
		return fmt.Errorf("cannot connect to db: %v", err)
	}

	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		panic(err)
	}
	defer conn.Release()
	err = runDBMigrations(conn.Conn())
	if err != nil {
		return fmt.Errorf("cannot migrate db: %v", err)
	}

	m.features = map[string]bool{
		"Register":            true,
		"OnUserAuthenticated": true,
		"On2FA":               true,
		"GetData":             true,
		"Get2FAData":          true,
		"Update":              true,
	}

	return nil
}

func (m *manager) GetMetaData() plugins.Meta {
	return plugins.Meta{
		Type: adapterName,
		ID:   pluginID,
	}
}

func (m *manager) Register(c *plugins.Credential, i *plugins.Identity, _ string) (*plugins.Identity, error) {
	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, c)
	if err != nil {
		return nil, fmt.Errorf("cannot check user existence: %v", err)
	}

	if exists {
		return nil, errors.New("user already exists")
	} else {
		registeredIdent, err := registerIdentity(conn, i)
		if err != nil {
			return nil, fmt.Errorf("cannot register user: %v", err)
		}
		return registeredIdent, nil
	}
}

func (m *manager) OnUserAuthenticated(c *plugins.Credential, i *plugins.Identity, authnProvider string) (*plugins.Identity, error) {
	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, c)
	if err != nil {
		return nil, fmt.Errorf("cannot check user existence: %v", err)
	}

	var registeredIdent *plugins.Identity
	if exists {
		registeredIdent, err = getIdentity(conn, c)
		if err != nil {
			return nil, fmt.Errorf("cannot get user data: %v", err)
		}
	} else {
		if authnProvider != "password_based" && (i.EmailVerified || i.PhoneVerified) {
			if strings.HasPrefix(authnProvider, "social_provider$") {
				provider := strings.TrimPrefix(authnProvider, "social_provider$")
				registeredIdent, err = registerSocialProviderIdentity(conn, i, provider)
				if err != nil {
					return nil, fmt.Errorf("cannot register social provider user: %v", err)
				}
			} else {
				registeredIdent, err = registerIdentity(conn, i)
				if err != nil {
					return nil, fmt.Errorf("cannot register user: %v", err)
				}
			}
		} else {
			return nil, errors.New("user doesn't exists")
		}
	}
	return registeredIdent, nil
}

func (m *manager) On2FA(c *plugins.Credential, mfaData *plugins.MFAData) error {
	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, c)
	if err != nil {
		return fmt.Errorf("cannot check user existence: %v", err)
	}

	if exists {
		return save2FAData(conn, c, mfaData)
	} else {
		return fmt.Errorf("user doesn't exists: %v", err)
	}
}

func (m *manager) GetData(c *plugins.Credential, _, name string) (interface{}, error) {
	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, c)
	if err != nil {
		return nil, fmt.Errorf("cannot check user existence: %v", err)
	}

	if exists {
		var data interface{}
		sql := fmt.Sprintf("select %s from users where %s=$1", sanitize(name), sanitize(c.Name))
		err := conn.QueryRow(context.Background(), sql, c.Value).Scan(&data)
		if err != nil {
			return nil, fmt.Errorf("cannot get '%s' field from db: %v", name, err)
		}
		return data, nil
	} else {
		return nil, errors.New("user doesn't exists")
	}
}

func (m *manager) Get2FAData(c *plugins.Credential, mfaID string) (*plugins.MFAData, error) {
	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, c)
	if err != nil {
		return nil, fmt.Errorf("cannot check user existence: %v", err)
	}

	if exists {
		return get2FAData(conn, c, mfaID)
	} else {
		return nil, plugins.UserNotExistError
	}
}

func (m *manager) Update(c *plugins.Credential, i *plugins.Identity, _ string) (*plugins.Identity, error) {
	conn, err := m.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, c)
	if err != nil {
		return nil, fmt.Errorf("cannot check user existence: %v", err)
	}

	if exists {
		registeredIdent, err := updateIdentity(conn, c, i)
		if err != nil {
			return nil, fmt.Errorf("cannot update user data: %v", err)
		}
		return registeredIdent, nil
	} else {
		return nil, errors.New("user doesn't exists")
	}
}

func (m *manager) CheckFeaturesAvailable(requiredFeatures []string) error {
	for _, f := range requiredFeatures {
		if available, ok := m.features[f]; !ok || !available {
			return fmt.Errorf("feature %s hasn't implemented", f)
		}
	}
	return nil
}

func initConfig(rawConf *configs.RawConfig) (*config, error) {
	adapterConf := &config{}
	if err := mapstructure.Decode(rawConf, adapterConf); err != nil {
		return nil, err
	}
	return adapterConf, nil
}

func runDBMigrations(conn *pgx.Conn) error {
	migrator, err := migrate.NewMigrator(context.Background(), conn, "schema_version")
	if err != nil {
		return fmt.Errorf("unable to create a migrator: %v", err)
	}

	for name, migration := range migrations.Migrations {
		migrator.AppendMigration(name, migration.UpSQL, migration.DownSQL)
	}
	return migrator.Migrate(context.Background())
}

func isUserExists(conn *pgxpool.Conn, cred *plugins.Credential) (exists bool, err error) {
	sql := fmt.Sprintf("select exists(select 1 from users where %s=$1)", sanitize(cred.Name))
	err = conn.QueryRow(context.Background(), sql, cred.Value).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func getIdentity(conn *pgxpool.Conn, cred *plugins.Credential) (*plugins.Identity, error) {
	sql := fmt.Sprintf(`SELECT u.*, provider_name, payload FROM users u
							  LEFT JOIN social_providers sp ON u.id = sp.user_id
                              WHERE u.%s = $1;`,
		sanitize(cred.Name))
	rows, err := conn.Query(context.Background(), sql, cred.Value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		ident               plugins.Identity
		userSocialProviders []map[string]interface{}
	)
	for rows.Next() {
		var (
			providerName *string
			payload      map[string]interface{}
		)
		err = rows.Scan(&ident.ID, &ident.Username, &ident.Phone, &ident.Email, &ident.EmailVerified,
			&ident.PhoneVerified, &ident.Additional, &providerName, &payload)
		if err != nil {
			return nil, err
		}

		if providerName != nil && payload != nil {
			userSocialProviders = append(userSocialProviders, map[string]interface{}{
				"provider_name": providerName,
				"payload":       payload,
			})
		}
	}

	if rows.Err() != nil {
		return nil, err
	}
	if userSocialProviders != nil {
		if ident.Additional == nil {
			ident.Additional = make(map[string]interface{})
			ident.Additional["social_providers"] = userSocialProviders
		}
	}
	return &ident, nil
}

func registerIdentity(conn *pgxpool.Conn, newIdent *plugins.Identity) (*plugins.Identity, error) {
	var ident plugins.Identity
	sql, values, err := getCreateQuery(newIdent)
	if err != nil {
		return nil, err
	}
	err = conn.QueryRow(context.Background(), sql, values...).Scan(&ident.ID, &ident.Username, &ident.Phone,
		&ident.Email, &ident.EmailVerified, &ident.PhoneVerified, &ident.Additional)
	if err != nil {
		return nil, err
	}
	return &ident, nil
}

func registerSocialProviderIdentity(conn *pgxpool.Conn, newIdent *plugins.Identity, provider string) (*plugins.Identity, error) {
	socialProviderData, ok := newIdent.Additional["social_provider_data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("cannot get social provider data")
	}
	oauth2PayloadBytes, err := json.Marshal(socialProviderData["payload"])
	if err != nil {
		return nil, err
	}
	pluginID := socialProviderData["plugin_id"].(string)
	delete(newIdent.Additional, "social_provider_data")

	tx, err := conn.Begin(context.Background())
	if err != nil {
		return nil, err
	}

	var ident plugins.Identity
	createUserSql, values, err := getCreateQuery(newIdent)
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(context.Background(), createUserSql, values...).Scan(&ident.ID, &ident.Username, &ident.Phone,
		&ident.Email, &ident.EmailVerified, &ident.PhoneVerified, &ident.Additional)
	if err != nil {
		return nil, err
	}

	saveProviderSql := "insert into social_providers(user_id, plugin_id, provider_name, payload) values ($1, $2, $3, $4);"
	_, err = tx.Exec(context.Background(), saveProviderSql, ident.ID, pluginID, provider, string(oauth2PayloadBytes))
	if err != nil {
		return nil, err
	}

	getSocialProvidersSql := "select provider_name, payload from social_providers where user_id=$1"
	rows, err := tx.Query(context.Background(), getSocialProvidersSql, ident.ID)
	if err != nil {
		return nil, err
	}

	var userSocialProviders []map[string]interface{}
	for rows.Next() {
		var (
			providerName string
			payload      map[string]interface{}
		)
		err = rows.Scan(&providerName, &payload)
		if err != nil {
			return nil, err
		}

		if providerName != "" && payload != nil {
			userSocialProviders = append(userSocialProviders, map[string]interface{}{
				"provider_name": providerName,
				"payload":       payload,
			})
		}
	}

	if rows.Err() != nil {
		return nil, err
	}
	ident.Additional["social_providers"] = userSocialProviders

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, tx.Rollback(context.Background())
	}
	return &ident, nil
}

func save2FAData(conn *pgxpool.Conn, cred *plugins.Credential, data *plugins.MFAData) error {
	bytesPayload, err := json.Marshal(data.Payload)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`insert into mfa(user_id, plugin_id, provider_name, payload) 
		                      (select id, $1, $2, $3::json from users where %s=$3);`,
		sanitize(cred.Name))
	_, err = conn.Exec(context.Background(), sql, data.PluginID, data.ProviderName, string(bytesPayload), cred.Value)
	if err != nil {
		return err
	}
	return nil
}

func updateIdentity(conn *pgxpool.Conn, cred *plugins.Credential, newIdent *plugins.Identity) (*plugins.Identity, error) {
	var ident plugins.Identity
	sql, values, err := getUpdateQuery(cred, newIdent)
	if err != nil {
		return nil, err
	}
	err = conn.QueryRow(context.Background(), sql, values...).Scan(&ident.ID, &ident.Username, &ident.Phone,
		&ident.Email, &ident.EmailVerified, &ident.PhoneVerified, &ident.Additional)
	if err != nil {
		return nil, err
	}
	return &ident, nil
}

func get2FAData(conn *pgxpool.Conn, cred *plugins.Credential, mfaID string) (*plugins.MFAData, error) {
	var data plugins.MFAData
	sql := fmt.Sprintf(`select plugin_id, provider_name, payload from mfa 
		                      where plugin_id=$1 and user_id=(select id from users where %s=$2);`,
		sanitize(cred.Name))
	err := conn.QueryRow(context.Background(), sql, mfaID, cred.Value).Scan(&data.PluginID, &data.ProviderName, &data.Payload)
	if err != nil {
		return nil, fmt.Errorf("cannot get 2fa data from db: %v", err)
	}
	return &data, nil
}

func getCreateQuery(ident *plugins.Identity) (string, []interface{}, error) {
	identMap := ident.AsMap()
	if ident.Additional != nil && len(ident.Additional) != 0 {
		bytesAdditionalData, err := json.Marshal(ident.Additional)
		if err != nil {
			return "", nil, err
		}
		identMap["additional"] = string(bytesAdditionalData)
	}

	var (
		values   []interface{}
		colsStmt string
		valsStmt string
		n        = 1
	)
	for k, v := range identMap {
		colsStmt += sanitize(k) + ","
		valsStmt += fmt.Sprintf("$%d,", n)
		values = append(values, v)
		n++
	}

	colsStmt = colsStmt[:len(colsStmt)-1]
	valsStmt = valsStmt[:len(valsStmt)-1]

	return fmt.Sprintf("insert into users(%s) values (%s) returning *;", colsStmt, valsStmt), values, nil
}

func getUpdateQuery(cred *plugins.Credential, ident *plugins.Identity) (string, []interface{}, error) {
	identMap := ident.AsMap()
	if ident.Additional != nil && len(ident.Additional) != 0 {
		bytesAdditionalData, err := json.Marshal(ident.Additional)
		if err != nil {
			return "", nil, err
		}
		identMap["additional"] = string(bytesAdditionalData)
	}

	var (
		colsStmt string
		valsStmt string
		values   = []interface{}{cred.Value}
		n        = 2
	)
	for k, v := range identMap {
		colsStmt += sanitize(k) + ","
		valsStmt += fmt.Sprintf("$%d,", n)
		values = append(values, v)
		n++
	}

	colsStmt = colsStmt[:len(colsStmt)-1]
	valsStmt = valsStmt[:len(valsStmt)-1]

	sql := fmt.Sprintf("update users set (%s)=(%s) where %s=$1;", colsStmt, valsStmt, sanitize(cred.Name))
	return sql, values, nil
}

func sanitize(ident string) string {
	return pgx.Identifier.Sanitize([]string{ident})
}
