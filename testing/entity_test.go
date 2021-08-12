package testing

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jmoiron/sqlx"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/sparkle/entity"
	"github.com/streamingfast/sparkle/storage/postgres"
	"github.com/test-go/testify/require"
	"go.uber.org/zap"
)

//sparkle_test

type Test struct {
	ID            uint64       `db:"id"`
	BoolPtr       *bool        `db:"native_bool_ptr"`
	Bool          bool         `db:"native_bool"`
	EntityBool    entity.Bool  `db:"entity_bool"`
	EntityBoolPtr *entity.Bool `db:"entity_bool_ptr"`
	Bytes         []byte       `db:"native_bytes"`
	EntityBytes   entity.Bytes `db:"entity_bytes"`
}

var testSchemaStmt = `
CREATE TABLE public.test (
	id serial PRIMARY KEY,
    native_bool_ptr boolean,
	native_bool boolean NOT NULL,
	entity_bool boolean NOT NULL,
	entity_bool_ptr boolean,
	native_bytes bytea, 
	entity_bytes bytea
);
`

var insertStmt = `
INSERT INTO public.test (
	"native_bool_ptr",
	"native_bool",
	"entity_bool",
	"entity_bool_ptr",
	"native_bytes",
	"entity_bytes"
) VALUES (
	:native_bool_ptr,
	:native_bool,
	:entity_bool,
	:entity_bool_ptr,
	:native_bytes,
	:entity_bytes
)
`

func init() {
	logging.TestingOverride()
}

func Test_Type(t *testing.T) {
	ctx := context.Background()
	testEntityDBDSN := os.Getenv("TEST_ENTITY_DB")
	if testEntityDBDSN == "" {
		t.Skip("skipping entity test set  'TEST_ENTITY_DB' to you psql DSB to test entity encoding")
		return
	}
	postgresDSN, err := postgres.ParseDSN(testEntityDBDSN)
	require.NoError(t, err)

	db, err := createPostgresDB(ctx, postgresDSN)
	require.NoError(t, err)

	cleanTestDB(t, ctx, db)

	entityBoolTrue := entity.NewBool(true)
	tests := []struct {
		name        string
		in          *Test
		expectedOut *Test
	}{
		{
			name: "empty struct",
			in:   nil,
		},
		{
			name: "struct with values",
			in: &Test{
				BoolPtr:       b(true),
				Bool:          true,
				EntityBool:    entity.NewBool(true),
				EntityBoolPtr: &entityBoolTrue,
				Bytes:         []byte{0xaa},
				EntityBytes:   entity.Bytes([]byte{0xaa}),
			},
		},
	}

	for idx, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := db.NamedExecContext(ctx, insertStmt, test.in)
			require.NoError(t, err)
			rowCount, err := resp.RowsAffected()
			require.NoError(t, err)
			assert.Equal(t, int64(1), rowCount)
			out := &Test{}
			expectedDBID := (idx + 1)
			query := fmt.Sprintf("SELECT * FROM public.test WHERE id = %d", expectedDBID)
			err = db.GetContext(ctx, out, query)
			require.NoError(t, err)
			// this is a small hack to make sure that teh ID cange after insertion doesn't break the test
			test.in.ID = uint64(expectedDBID)
			assert.Equal(t, test.in, out)
		})
	}

}

func cleanTestDB(t *testing.T, ctx context.Context, db *sqlx.DB) {
	query := "DROP SCHEMA public CASCADE;"
	_, err := db.ExecContext(ctx, query)
	require.NoError(t, err)

	query = "CREATE SCHEMA public;"
	_, err = db.ExecContext(ctx, query)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, testSchemaStmt)
	require.NoError(t, err)
}

func createPostgresDB(ctx context.Context, connectionInfo *postgres.DSN) (*sqlx.DB, error) {
	dsn := connectionInfo.DSN()
	zlog.Info("connecting to postgres", zap.String("data_source", dsn))
	dbConnecCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dbCancel()

	db, err := sqlx.ConnectContext(dbConnecCtx, "postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(100)

	return db, nil
}

func b(val bool) *bool {
	return &val
}
