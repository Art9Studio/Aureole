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
	"strconv"
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

	return nil
}

func (m standart) GetMetadata() core.Metadata {
	return meta
}

func (m *standart) GetCustomAppRoutes() []*core.Route {
	return []*core.Route{}
}

func (s *standart) RegisterOrUpdate(authResp *core.AuthResult) (*core.AuthResult, error) {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	if authResp.Cred == nil {
		return nil, errors.New("nil cred is found")
	}

	user, err := getUser(conn, authResp.Cred)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) && authResp.User != nil {
			return registerOrUpdateUser(conn, authResp)
		} else {
			return nil, err
		}
	}

	var (
		cred *core.Credential
	)
	if authResp.User != nil {
		storedDataMap := user.AsMap()
		inDataMap := authResp.User.AsMap()
		if !isMapsEqual(inDataMap, storedDataMap) {
			return registerOrUpdateUser(conn, authResp)
		}
	} else if user != nil {
		authResp.User = &core.User{ID: user.ID}
	}

	if authResp.ImportedUser != nil {
		cred = &core.Credential{Name: "id", Value: user.ID}
		storedData, err := getImportedUser(conn, cred, authResp.ProviderId)
		if err != nil {
			return nil, err
		}
		storedDataMap := storedData.AsMap()
		inDataMap := authResp.ImportedUser.AsMap()
		if !isMapsEqual(inDataMap, storedDataMap) {
			return registerOrUpdateUser(conn, authResp)
		}
	}

	if authResp.Secrets != nil {
		if err = setSecrets(conn, user.ID, authResp.ProviderId, authResp.Secrets); err != nil {
			return nil, err
		}
	}

	return authResp, nil
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

func (s *standart) GetUser(cred *core.Credential) (*core.User, error) {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	user, err := getUser(conn, cred)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNoUser
		}
		return nil, err
	}
	return user, nil
}

func (s *standart) SetSecrets(cred *core.Credential, pluginId string, payload *core.Secrets) error {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	user, err := getUser(conn, cred)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.ErrNoUser
		}
		return err
	}

	return setSecrets(conn, user.ID, pluginId, payload)
}

func (s *standart) GetSecret(cred *core.Credential, pluginId string, secret string) (core.Secret, error) {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	user, err := getUser(conn, cred)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNoUser
		}
		return nil, core.WrapErrDB(err.Error())
	}

	out, err := getSecret(conn, user.ID, pluginId, secret)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *standart) GetSecrets(userId, pluginId string) (*core.Secrets, error) {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	secrets, err := getSecrets(conn, userId, pluginId)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (s *standart) UseScratchCode(cred *core.Credential, code string) error {
	ctx := context.Background()
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	user, err := getUser(conn, cred)
	if err != nil {
		return err
	}

	codes, err := getSecret(conn, user.ID, "0", "scratch_codes")
	if err != nil {
		return err
	}

	codesArr := strings.Split(*codes, ",")
	var (
		ok bool
		sb strings.Builder
	)
	for _, c := range codesArr {
		if c == code {
			ok = true
		} else {
			sb.WriteString(c)
			sb.WriteByte(',')
		}
	}
	if ok {
		res := sb.String()[:len(sb.String())-1]
		return setSecrets(conn, user.ID, "0", &core.Secrets{"scratch_codes": &res})
	}
	return errors.New("code not found")
}

//todo(Talgat) delete
func (s *standart) SetSecret(cred *core.Credential, pluginId string, secret core.Secret) error {
	ctx := context.Background()
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	exists, err := isUserExists(conn, cred)
	if err != nil {
		return fmt.Errorf("cannot check user existence: %v", err)
	}
	if !exists {
		return errors.New("user doesn't exists")
	}

	var user *core.User
	if user, err = getUser(conn, cred); err != nil {
		return fmt.Errorf("cannot get user: %w", err)
	}

	payload, err := json.Marshal(secret)
	if err != nil {
		return err
	}

	sql := `update secrets set payload=$1 where user_id=$2 and plugin_id=3;`
	_, err = conn.Exec(ctx, sql, payload, user.ID, pluginId)
	if err != nil {
		return core.WrapErrDB(err.Error())
	}
	return nil
}

func (s *standart) IsMFAEnabled(cred *core.Credential) (bool, error) {
	conn, err := s.pool.Acquire(context.Background())
	if err != nil {
		return false, fmt.Errorf("cannot acquire connection: %v", err)
	}
	defer conn.Release()

	return isMFAEnabled(conn, cred)
}

func isMFAEnabled(conn *pgxpool.Conn, cred *core.Credential) (bool, error) {
	var ret bool
	qry := fmt.Sprintf("select is_mfa_enabled from users where %s=$1;", cred.Name)
	if err := conn.QueryRow(context.Background(), qry, cred.Value).Scan(&ret); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return ret, nil
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

func isUserExists(conn *pgxpool.Conn, cred *core.Credential) (bool, error) {
	var ret bool
	sql := fmt.Sprintf("select exists(select 1 from users where %s=$1)", sanitize(cred.Name))
	if err := conn.QueryRow(context.Background(), sql, cred.Value).Scan(&ret); err != nil {
		return false, core.WrapErrDB(err.Error())
	}
	return ret, nil
}

func isImportedUserExists(conn *pgxpool.Conn, userId, pluginId string) (bool, error) {
	var ret bool
	sql := "select exists(select 1 from imported_users where user_id=$1 and plugin_id=$2)"
	if err := conn.QueryRow(context.Background(), sql, userId, pluginId).Scan(&ret); err != nil {
		return false, err
	}
	return ret, nil
}

func getImportedUser(conn *pgxpool.Conn, cred *core.Credential, pluginId string) (*core.ImportedUser, error) {
	//todo(Talgat) replace with sql constructors
	sql := fmt.Sprintf("SELECT * FROM imported_users WHERE %s = $1 and plugin_id = $2;", cred.Name)

	row := conn.QueryRow(context.Background(), sql, cred.Value, pluginId)

	var importedUser core.ImportedUser
	if err := row.Scan(
		&importedUser.PluginID,
		&importedUser.ProviderName,
		&importedUser.ProviderId,
		&importedUser.UserId,
		&importedUser.Additional,
	); err != nil {
		return nil, err
	}

	return &importedUser, nil
}

func getUser(conn *pgxpool.Conn, cred *core.Credential) (*core.User, error) {
	sql := fmt.Sprintf(`SELECT * FROM users WHERE %s = $1;`, cred.Name)
	row := conn.QueryRow(context.Background(), sql, cred.Value)

	var user core.User

	var (
		userId    int
		userIdStr string
	)

	if err := row.Scan(
		&userId,
		&user.Username,
		&user.Phone, &user.Email,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.IsMFAEnabled,
		&user.EnabledMFAs,
	); err != nil {
		return nil, err
	}

	userIdStr = strconv.Itoa(userId)
	user.ID = userIdStr

	return &user, nil
}

func getSecret(conn *pgxpool.Conn, userId, pluginId string, secret string) (core.Secret, error) {
	var ret core.Secret
	qry := fmt.Sprintf("SELECT payload ->> '%s' FROM secrets WHERE user_id=$1 AND plugin_id=$2;", secret)
	row := conn.QueryRow(context.Background(), qry, userId, pluginId)
	if err := row.Scan(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func getSecrets(conn *pgxpool.Conn, userId, pluginId string) (*core.Secrets, error) {
	ret := &core.Secrets{}
	qry := "SELECT payload FROM secrets WHERE user_id=$1 AND plugin_id=$2;"
	row := conn.QueryRow(context.Background(), qry, userId, pluginId)
	if err := row.Scan(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func setSecrets(conn *pgxpool.Conn, userId, pluginId string, payload *core.Secrets) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	saveSecretsQry := `insert into secrets(user_id, plugin_id, payload) values ($1, $2, $3)
						on conflict (user_id, plugin_id) do update set payload = $3;`
	_, err = conn.Exec(context.Background(), saveSecretsQry, userId, pluginId, payloadBytes)
	if err != nil {
		return err
	}
	return nil
}

func setSecretsTx(tx pgx.Tx, userId, pluginId string, payload core.Secrets) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	saveSecretsQry := `insert into secrets(user_id, plugin_id, payload) values ($1, $2, $3)
						on conflict (user_id, plugin_id) do update set payload = $3;`
	_, err = tx.Exec(context.Background(), saveSecretsQry, userId, pluginId, payloadBytes)
	if err != nil {
		return err
	}
	return nil
}

func registerOrUpdateUser(conn *pgxpool.Conn, authRes *core.AuthResult) (*core.AuthResult, error) {
	ctx := context.Background()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	var (
		userId    int
		userIdStr string
	)
	if authRes.User != nil {
		sql, values, err := getUpsertUserQry(authRes.User)
		if err != nil {
			return nil, err
		}
		err = tx.QueryRow(ctx, sql, values...).Scan(&userId)
		if err != nil {
			return nil, err
		}
		userIdStr = strconv.Itoa(userId)
		authRes.User.ID = userIdStr
	}

	if authRes.ImportedUser != nil {
		authRes.ImportedUser.UserId = userIdStr
		sql, values, err := getUpsertImportedUserQry(authRes.ImportedUser)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(context.Background(), sql, values...)
		if err != nil {
			if err = tx.Rollback(ctx); err != nil {
				return nil, fmt.Errorf("cannot rollback in registerOrUpdateUser: %w", err)
			}
			return nil, err
		}
	}
	if authRes.Secrets != nil {
		if err = setSecretsTx(tx, userIdStr, authRes.ProviderId, *authRes.Secrets); err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				return nil, fmt.Errorf("cannot rollback in registerOrUpdateUser: %w", err)
			}
			return nil, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, tx.Rollback(ctx)
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

func getUpsertUserQry(user *core.User) (string, []interface{}, error) {
	userMap := user.AsMap()

	var constraint string

	if _, ok := userMap["id"]; ok {
		constraint = "id"
	} else if _, ok = userMap["email"]; ok {
		constraint = "email"
	} else if _, ok = userMap["phone"]; ok {
		constraint = "phone"
	}

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

	return fmt.Sprintf(
		`insert into users(%s) values (%s)
				on conflict (%s) do update
				set (%s) = (%s) returning id;`,
		colsStmt, valsStmt, constraint, colsStmt, valsStmt,
	), values, nil
}

func getUpsertImportedUserQry(iu *core.ImportedUser) (string, []interface{}, error) {
	iuMap := iu.AsMap()

	var (
		values   []interface{}
		colsStmt string
		valsStmt string
		n        = 1
	)

	for k, v := range iuMap {
		colsStmt += sanitize(k) + ","
		valsStmt += fmt.Sprintf("$%d,", n)
		values = append(values, v)
		n++
	}

	colsStmt = colsStmt[:len(colsStmt)-1]
	valsStmt = valsStmt[:len(valsStmt)-1]

	return fmt.Sprintf(
		`insert into imported_users (%s) values (%s)
				on conflict (user_id, plugin_id) do update
				set (%s) = (%s);`,
		colsStmt, valsStmt, colsStmt, valsStmt,
	), values, nil
}

func getUpdateQuery(cred *core.Credential, user *core.User) (string, []interface{}, error) {
	userMap := user.AsMap()

	var (
		colsStmt string
		valsStmt string
		values   = []interface{}{cred.Value}
		n        = 2
	)
	for k, v := range userMap {
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

func getImportedUserUpdateQuery(userId string, newImported *core.ImportedUser) (string, []interface{}, error) {
	importedMap := newImported.AsMap()

	var (
		colsStmt string
		valsStmt string
		values   []interface{}
		n        = 1
	)
	for k, v := range importedMap {
		colsStmt += sanitize(k) + ","
		valsStmt += fmt.Sprintf("$%d,", n)
		values = append(values, v)
		n++
	}

	colsStmt = colsStmt[:len(colsStmt)-1]
	valsStmt = valsStmt[:len(valsStmt)-1]

	sql := fmt.Sprintf("update imported_users set (%s)=(%s) where %s=$1 returning *;", colsStmt, valsStmt, userId)
	return sql, values, nil
}

func sanitize(ident string) string {
	return pgx.Identifier.Sanitize([]string{ident})
}

func isMapsEqual(main, compared map[string]interface{}) bool {
	for k, v := range main {
		if v != compared[k] {
			return false
		}
	}
	return true
}
