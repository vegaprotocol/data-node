package sqlstore

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/shared/paths"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	shopspring "github.com/jackc/pgtype/ext/shopspring-numeric"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
	uuid "github.com/satori/go.uuid"
)

var (
	ErrBadID   = errors.New("Bad ID (must be hex string)")
	tableNames = [...]string{"ledger", "accounts", "parties", "assets", "blocks"}
)

var Pool *pgxpool.Pool
var GlobalTx pgx.Tx
var GlobalConn *pgx.Conn

type TxDelegator struct {
}

func (t TxDelegator) Begin(ctx context.Context) (pgx.Tx, error) {
	if GlobalTx != nil {
		return GlobalTx.Begin(ctx)
	} else {
		return GlobalConn.Begin(ctx)
	}
}

func (t TxDelegator) BeginFunc(ctx context.Context, f func(pgx.Tx) error) (err error) {
	if GlobalTx != nil {
		return GlobalTx.BeginFunc(ctx, f)
	} else {
		return GlobalConn.BeginFunc(ctx, f)
	}
}

func (t TxDelegator) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {

	if GlobalTx != nil {
		return GlobalTx.CopyFrom(ctx, tableName, columnNames, rowSrc)
	} else {
		return GlobalConn.CopyFrom(ctx, tableName, columnNames, rowSrc)
	}

}

func (t TxDelegator) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	if GlobalTx != nil {
		return GlobalTx.SendBatch(ctx, b)
	} else {
		return GlobalConn.SendBatch(ctx, b)
	}

}

func (t TxDelegator) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	if GlobalTx != nil {
		return GlobalTx.Prepare(ctx, name, sql)
	} else {
		return GlobalConn.Prepare(ctx, name, sql)
	}
}

func (t TxDelegator) Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error) {
	if GlobalTx != nil {
		return GlobalTx.Exec(ctx, sql, arguments...)
	} else {
		return GlobalConn.Exec(ctx, sql, arguments...)
	}
}

func (t TxDelegator) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if GlobalTx != nil {
		return GlobalTx.Query(ctx, sql, args...)
	} else {
		return GlobalConn.Query(ctx, sql, args...)
	}
}

func (t TxDelegator) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if GlobalTx != nil {
		return GlobalTx.QueryRow(ctx, sql, args...)
	} else {
		return GlobalConn.QueryRow(ctx, sql, args...)
	}
}

func (t TxDelegator) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	if GlobalTx != nil {
		return GlobalTx.QueryFunc(ctx, sql, args, scans, f)
	} else {
		return GlobalConn.QueryFunc(ctx, sql, args, scans, f)
	}
}

func init() {
	var err error
	GlobalConn, err = pgx.Connect(context.Background(), "postgresql://vega:vega@127.0.0.1:5432/vega")
	if err != nil {
		panic(err)
	}
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

type SQLStore struct {
	conf Config
	log  *logging.Logger
	pool TxDelegator
	db   *embeddedpostgres.EmbeddedPostgres
}

func (s *SQLStore) makeConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		s.conf.Username,
		s.conf.Password,
		s.conf.Host,
		s.conf.Port,
		s.conf.Database)
}

func (s *SQLStore) makePoolConfig() (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(s.makeConnectionString())
	if err != nil {
		return nil, err
	}
	cfg.ConnConfig.RuntimeParams["application_name"] = "Vega Data Node"
	return cfg, nil
}

func (s *SQLStore) migrateToLatestSchema() error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(s.log.Named("db migration").GooseLogger())

	db := stdlib.OpenDB(*Pool.Config().ConnConfig)

	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return err
	}

	if currentVersion > 0 && s.conf.WipeOnStartup {
		if err := goose.Down(db, "migrations"); err != nil {
			return fmt.Errorf("error clearing sql schema: %w", err)
		}
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("error migrating sql schema: %w", err)
	}
	return nil
}

func registerNumericType(poolConfig *pgxpool.Config) {
	// Cause postgres numeric types to be loaded as shopspring decimals and vice-versa
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &shopspring.Numeric{},
			Name:  "numeric",
			OID:   pgtype.NumericOID,
		})
		return nil
	}
}

func InitialiseStorage(log *logging.Logger, config Config, vegapaths paths.Paths) (*SQLStore, error) {
	s := SQLStore{
		conf: config,
		log:  log.Named("sql_store"),
	}

	if s.conf.UseEmbedded {
		embeddedPostgresRuntimePath := paths.JoinStatePath(paths.StatePath(vegapaths.StatePathFor(paths.DataNodeStorageHome)), "sqlstore")
		embeddedPostgresDataPath := paths.JoinStatePath(paths.StatePath(vegapaths.StatePathFor(paths.DataNodeStorageHome)), "node-data")

		if err := s.initializeEmbeddedPostgres(&embeddedPostgresRuntimePath, &embeddedPostgresDataPath, io.Discard); err != nil {
			return nil, fmt.Errorf("use embedded database was true, but failed to start: %w", err)
		}
	}

	return setupStorage(&s)
}

func InitialiseTestStorage(log *logging.Logger, config Config) (*SQLStore, error) {
	s := SQLStore{
		conf: config,
		log:  log.Named("sql_store_test"),
	}

	if s.conf.UseEmbedded {
		testID := uuid.NewV4().String()

		tempDir, err := ioutil.TempDir("", testID)

		if err != nil {
			return nil, err
		}

		embeddedPostgresRuntimePath := paths.JoinStatePath(paths.StatePath(tempDir), "sqlstore")
		embeddedPostgresDataPath := paths.JoinStatePath(paths.StatePath(tempDir), "sqlstore", "node-data")

		postgresLog := &bytes.Buffer{}

		if err := s.initializeEmbeddedPostgres(&embeddedPostgresRuntimePath, &embeddedPostgresDataPath, postgresLog); err != nil {
			log.Errorf("postgres log: \n%s", postgresLog.String())
			return nil, fmt.Errorf("use embedded database was true, but failed to start: %w", err)
		}
	}

	return setupStorage(&s)
}

func setupStorage(store *SQLStore) (*SQLStore, error) {
	poolConfig, err := store.makePoolConfig()
	if err != nil {
		return nil, fmt.Errorf("error configuring database: %w", err)
	}

	registerNumericType(poolConfig)

	if Pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig); err != nil {
		store.Stop()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err = store.migrateToLatestSchema(); err != nil {
		store.Stop()
		return nil, fmt.Errorf("error migrating schema: %w", err)
	}

	return store, nil
}

func (s *SQLStore) DeleteEverything() error {
	for _, table := range tableNames {
		if _, err := Pool.Exec(context.Background(), "truncate table "+table+" CASCADE"); err != nil {
			return fmt.Errorf("error truncating table: %s %w", table, err)
		}
	}
	return nil
}

func (s *SQLStore) initializeEmbeddedPostgres(runtimePath *paths.StatePath, dataPath *paths.StatePath, writer io.Writer) error {
	dbConfig := embeddedpostgres.DefaultConfig().
		Username(s.conf.Username).
		Password(s.conf.Password).
		Database(s.conf.Database).
		Port(uint32(s.conf.Port)).
		Logger(writer)

	if runtimePath != nil {
		dbConfig = dbConfig.RuntimePath(runtimePath.String()).BinariesPath(runtimePath.String())
	}

	if dataPath != nil {
		dbConfig = dbConfig.DataPath(dataPath.String())
	}

	s.db = embeddedpostgres.NewDatabase(dbConfig)
	return s.db.Start()
}

func (s *SQLStore) Stop() error {
	if !s.conf.Enabled || !s.conf.UseEmbedded || s.db == nil {
		return nil
	}

	return s.db.Stop()
}
