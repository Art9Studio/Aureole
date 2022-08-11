package standard

import (
	"aureole/internal/configs"
	"aureole/internal/core"
	"aureole/plugins/identity-manager/standard/migrations"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/tern/migrate"
	"github.com/mitchellh/mapstructure"
)

// const pluginID = "4634"
//go:embed meta.yaml
var rawMeta []byte

var meta core.Metadata

func init() {
	meta = core.IDManagerRepo.Register(rawMeta, Create)
}

type standart struct {
	pluginApi core.PluginAPI
	rawConf   configs.PluginConfig
	conf      *config
	pool      *pgxpool.Pool
	features  map[string]bool
}

func Create(conf configs.PluginConfig) core.IDManager {
	return &standart{rawConf: conf}
}

func (m *standart) Init(api core.PluginAPI) (err error) {
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
		"OnMFA":               true,
		"GetData":             true,
		"GetMFAData":          true,
		"Update":              true,
	}

	return nil
}

func (m standart) GetMetadata() core.Metadata {
	return meta

}

func (m *standart) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{}
}

// Register todo(Talgat) add Secrets
func (s *standart) Register(c *core.Credential, i *core.Identity, u *core.User, _ string) (*core.User, error) {
	conn, err := s.pool.Acquire(context.Background())
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
		registeredIdent, err := registerUser(conn, u)
		if err != nil {
			return nil, fmt.Errorf("cannot register user: %v", err)
		}
		return registeredIdent, nil
	}
}

func (s *standart) OnUserAuthenticated(authRes *core.AuthResult) (*core.AuthResult, error) {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, authRes.Cred)
	if err != nil {
		return nil, fmt.Errorf("cannot check user existence: %v", err)
	}

	var (
		registeredUser *core.User
		importedUser   *core.ImportedUser
	)

	if exists {
		registeredUser, err = getUser(conn, authRes.Cred)
		if err != nil {
			return nil, fmt.Errorf("cannot get user data: %v", err)
		}
		importedUser, err = getImportedUser(conn, *registeredUser.ID, *authRes.ImportedUser.PluginID)
		if err != nil {
			return nil, fmt.Errorf("cannot get imported user: %v", err)
		}
		authRes.User = registeredUser
		authRes.ImportedUser = importedUser
	} else {
		if authRes.Provider != "password_based" && (authRes.Identity.EmailVerified || authRes.Identity.PhoneVerified) {
			if strings.HasPrefix(authRes.Provider, "social_provider$") {
				authRes, err = registerOauth2User(conn, authRes)
				if err != nil {
					return nil, fmt.Errorf("cannot register social provider user: %v", err)
				}
			} else {
				registeredUser, err = registerUser(conn, authRes.User)
				if err != nil {
					return nil, fmt.Errorf("cannot register user: %v", err)
				}
				authRes.User = registeredUser
			}
		} else {
			return nil, errors.New("user doesn't exists")
		}
	}
	return authRes, nil
}

func (s *standart) OnMFA(c *core.Credential, mfaData *core.MFAData) error {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, c)
	if err != nil {
		return fmt.Errorf("cannot check user existence: %v", err)
	}

	if exists {
		return saveMFAData(conn, c, mfaData)
	} else {
		return fmt.Errorf("user doesn't exists: %w", pgx.ErrNoRows)
	}
}

func (s *standart) GetData(c *core.Credential, _, name string) (interface{}, error) {
	conn, err := s.pool.Acquire(context.Background())
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
		sql := fmt.Sprintf("select payload->'%s' from mfa where user_id=(select id from users where %s=$1)", strings.ReplaceAll(name, "\"", ""), strings.ReplaceAll(c.Name, "\"", ""))
		err := conn.QueryRow(context.Background(), sql, c.Value).Scan(&data)
		if err != nil {
			return nil, fmt.Errorf("cannot get '%s' field from db: %v", name, err)
		}
		return data, nil
	} else {
		return nil, errors.New("user doesn't exists")
	}
}

func (s *standart) GetMFAData(c *core.Credential, mfaID string) (*core.MFAData, error) {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, c)
	if err != nil {
		return nil, fmt.Errorf("cannot check user existence: %v", err)
	}

	if exists {
		return getMFAData(conn, c, mfaID)
	}

	return nil, core.UserNotExistError
}

func (s *standart) Update(c *core.Credential, i *core.Identity, _ string) (*core.Identity, error) {
	conn, err := s.pool.Acquire(context.Background())
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

func (s *standart) CheckFeaturesAvailable(requiredFeatures []string) error {
	for _, f := range requiredFeatures {
		if available, ok := s.features[f]; !ok || !available {
			return fmt.Errorf("feature %s hasn't implemented", f)
		}
	}
	return nil
}

func initConfig(conf *configs.RawConfig) (*config, error) {
	PluginConf := &config{}
	if err := mapstructure.Decode(conf, PluginConf); err != nil {
		return nil, err
	}
	return PluginConf, nil
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

func isUserExists(conn *pgxpool.Conn, cred *core.Credential) (exists bool, err error) {
	sql := fmt.Sprintf("select exists(select 1 from users where %s=$1)", sanitize(cred.Name))
	err = conn.QueryRow(context.Background(), sql, cred.Value).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func getImportedUser(conn *pgxpool.Conn, userId, pluginId string) (*core.ImportedUser, error) {
	//todo(Talgat) replace with sql constructors
	sql := `SELECT provider_name, payload FROM imported_users WHERE user_id = $1 and plugin_id = $2;`

	rows, err := conn.Query(context.Background(), sql, userId, pluginId)
	if err != nil {
		return nil, err
	}

	var importedUser *core.ImportedUser

	for rows.Next() {
		var (
			payload      map[string]interface{}
			providerName string
		)
		if err = rows.Scan(&providerName, &payload); err != nil {
			return nil, err
		}

		if providerName != "" && payload != nil {
			importedUser.ProviderName = &providerName
			importedUser.Additional = payload
		}
	}
	if rows.Err() != nil {
		return nil, err
	}
	return importedUser, nil
}

func getUser(conn *pgxpool.Conn, cred *core.Credential) (*core.User, error) {
	sql := fmt.Sprintf(`SELECT * FROM users WHERE %s = $1;`, cred.Name)
	row := conn.QueryRow(context.Background(), sql, cred.Value)

	var user core.User

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Phone, &user.Email,
		&user.EmailVerified,
		&user.PhoneVerified,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func registerUser(conn *pgxpool.Conn, newUser *core.User) (*core.User, error) {
	sql, values, err := getCreateQuery(newUser)
	if err != nil {
		return nil, err
	}
	err = conn.QueryRow(context.Background(), sql, values...).Scan(&newUser.ID)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func registerOauth2User(conn *pgxpool.Conn, authRes *core.AuthResult) (*core.AuthResult, error) {
	importedUserData := authRes.ImportedUser.Additional
	oauth2PayloadBytes, err := json.Marshal(importedUserData)
	if err != nil {
		return nil, err
	}

	tx, err := conn.Begin(context.Background())
	if err != nil {
		return nil, err
	}

	createUserSql, values, err := getCreateQuery(authRes.User)

	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(context.Background(), createUserSql, values...).Scan(&authRes.User.ID)
	if err != nil {
		return nil, err
	}

	saveImportedUser := "insert into imported_users(user_id, plugin_id, provider_id, provider_name, additional) values ($1, $2, $3, $4, $5);"
	_, err = tx.Exec(context.Background(), saveImportedUser,
		authRes.User.ID,
		authRes.ImportedUser.PluginID,
		authRes.ImportedUser.ProviderId,
		authRes.Provider,
		string(oauth2PayloadBytes))
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, tx.Rollback(context.Background())
	}
	return authRes, nil
}

func saveMFAData(conn *pgxpool.Conn, cred *core.Credential, data *core.MFAData) error {
	bytesPayload, err := json.Marshal(data.Payload)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`insert into mfa (user_id, plugin_id, provider_name, payload) 
		                      values ((select id from users where %s=$4), $1, $2, $3::json);`,
		sanitize(cred.Name))
	_, err = conn.Exec(context.Background(), sql, data.PluginID, data.ProviderName, string(bytesPayload), cred.Value)
	if err != nil {
		return err
	}
	return nil
}

func updateIdentity(conn *pgxpool.Conn, cred *core.Credential, newIdent *core.Identity) (*core.Identity, error) {
	var ident core.Identity
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

func getMFAData(conn *pgxpool.Conn, cred *core.Credential, mfaID string) (*core.MFAData, error) {
	var data core.MFAData
	qry := fmt.Sprintf(`select plugin_id, provider_name, payload from mfa 
		                      where plugin_id=$1 and user_id=(select id from users where %s=$2);`,
		sanitize(cred.Name))
	err := conn.QueryRow(context.Background(), qry, mfaID, cred.Value).Scan(&data.PluginID, &data.ProviderName, &data.Payload)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("cannot get 2fa data from db: %v", err)
	}
	return &data, nil
}

//todo(Talgat) create for ImportedUser
func getCreateQuery(user *core.User) (string, []interface{}, error) {
	userMap := user.AsMap()

	var (
		values   []interface{}
		colsStmt string
		valsStmt string
		n        = 1
	)
	for k, v := range userMap {
		colsStmt += sanitize(k) + ","
		valsStmt += fmt.Sprintf("$%d,", n)
		values = append(values, v)
		n++
	}

	colsStmt = colsStmt[:len(colsStmt)-1]
	valsStmt = valsStmt[:len(valsStmt)-1]

	return fmt.Sprintf("insert into users(%s) values (%s) returning id;", colsStmt, valsStmt), values, nil
}

func getUpdateQuery(cred *core.Credential, ident *core.Identity) (string, []interface{}, error) {
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

	sql := fmt.Sprintf("update users set (%s)=(%s) where %s=$1 returning *;", colsStmt, valsStmt, sanitize(cred.Name))
	return sql, values, nil
}

func sanitize(ident string) string {
	return pgx.Identifier.Sanitize([]string{ident})
}
