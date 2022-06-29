// Copyright (c) 2022 Gobalsky Labs Limited
//
// Use of this software is governed by the Business Source License included
// in the LICENSE file and at https://www.mariadb.com/bsl11.
//
// Change Date: 18 months from the later of the date of the first publicly
// available Distribution of this version of the repository, and 25 June 2022.
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by version 3 or later of the GNU General
// Public License.

package sqlstore

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/shared/paths"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"google.golang.org/protobuf/proto"
)

var ErrBadID = errors.New("Bad ID (must be hex string)")

//go:embed migrations/*.sql
var embedMigrations embed.FS

func MigrateToLatestSchema(log *logging.Logger, config Config) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(log.Named("db migration").GooseLogger())

	poolConfig, err := config.ConnectionConfig.GetPoolConfig()
	if err != nil {
		return errors.Wrap(err, "migrating schema")
	}

	db := stdlib.OpenDB(*poolConfig.ConnConfig)
	defer db.Close()

	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		return err
	}

	if currentVersion > 0 && config.WipeOnStartup {
		if err := goose.Down(db, "migrations"); err != nil {
			return fmt.Errorf("error clearing sql schema: %w", err)
		}
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("error migrating sql schema: %w", err)
	}
	return nil
}

func ApplyDataRetentionPolicies(config Config) error {
	poolConfig, err := config.ConnectionConfig.GetPoolConfig()
	if err != nil {
		return errors.Wrap(err, "applying data retention policy")
	}

	db := stdlib.OpenDB(*poolConfig.ConnConfig)
	defer db.Close()

	for _, policy := range config.RetentionPolicies {
		if _, err := db.Exec(fmt.Sprintf("SELECT remove_retention_policy('%s', true);", policy.HypertableOrCaggName)); err != nil {
			return errors.Wrapf(err, "removing retention policy from %s", policy.HypertableOrCaggName)
		}

		if _, err := db.Exec(fmt.Sprintf("SELECT add_retention_policy('%s', INTERVAL '%s');", policy.HypertableOrCaggName, policy.DataRetentionPeriod)); err != nil {
			return errors.Wrapf(err, "adding retention policy to %s", policy.HypertableOrCaggName)
		}
	}

	return nil
}

func StartEmbeddedPostgres(log *logging.Logger, config Config, stateDir string) (*embeddedpostgres.EmbeddedPostgres, error) {
	embeddedPostgresRuntimePath := paths.JoinStatePath(paths.StatePath(stateDir), "sqlstore")
	embeddedPostgresDataPath := paths.JoinStatePath(paths.StatePath(stateDir), "sqlstore", "node-data")

	postgresLog := &bytes.Buffer{}

	embeddedPostgres := createEmbeddedPostgres(&embeddedPostgresRuntimePath, &embeddedPostgresDataPath,
		postgresLog, config.ConnectionConfig)

	if err := embeddedPostgres.Start(); err != nil {
		log.Errorf("postgres log: \n%s", postgresLog.String())
		return nil, fmt.Errorf("use embedded database was true, but failed to start: %w", err)
	}

	return embeddedPostgres, nil
}

func createEmbeddedPostgres(runtimePath *paths.StatePath, dataPath *paths.StatePath, writer io.Writer, conf ConnectionConfig) *embeddedpostgres.EmbeddedPostgres {
	dbConfig := embeddedpostgres.DefaultConfig().
		Username(conf.Username).
		Password(conf.Password).
		Database(conf.Database).
		Port(uint32(conf.Port)).
		Logger(writer)

	if runtimePath != nil {
		dbConfig = dbConfig.RuntimePath(runtimePath.String()).BinariesPath(runtimePath.String())
	}

	if dataPath != nil {
		dbConfig = dbConfig.DataPath(dataPath.String())
	}

	return embeddedpostgres.NewDatabase(dbConfig)
}

func executePaginationBatch[P proto.Message, T entities.PagedEntity[P]](ctx context.Context, batch *pgx.Batch, connection Connection, pagination entities.CursorPagination) entities.ConnectionData[P, T] {
	var connectionData entities.ConnectionData[P, T]

	results := connection.SendBatch(ctx, batch)
	defer results.Close()

	rowCountRows, err := results.Query()

	if err != nil {
		connectionData.Err = fmt.Errorf("error querying row count: %w", err)
		return connectionData
	}

	if err := pgxscan.ScanOne(&connectionData.TotalCount, rowCountRows); err != nil {
		connectionData.Err = fmt.Errorf("error scanning row count: %w", err)
		return connectionData
	}

	rows, err := results.Query()
	if err != nil {
		connectionData.Err = fmt.Errorf("error querying results: %w", err)
		return connectionData
	}

	var items []T

	if err := pgxscan.ScanAll(&items, rows); err != nil {
		connectionData.Err = fmt.Errorf("error scanning results: %w", err)
		return connectionData
	}

	connectionData.Entities, connectionData.PageInfo = entities.PageEntities[P](items, pagination)

	return connectionData
}
