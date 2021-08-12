package cli

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
	"github.com/streamingfast/cli"
	"github.com/streamingfast/sparkle/storage/postgres"
	"go.uber.org/zap"
)

var blockRangeRegex = regexp.MustCompile(`(\d{10})-(\d{10})`)

type VersionedSubgraph struct {
	name    string
	version string
}

func parseSubgraphVersionedName(in string) (*VersionedSubgraph, error) {
	atIndex := strings.IndexByte(in, '@')
	if atIndex == -1 {
		return nil, fmt.Errorf("invalid syntax,  should be o the form <name>@<version>")
	}

	return &VersionedSubgraph{
		name:    in[0:atIndex],
		version: in[atIndex+1:],
	}, nil
}

func getBlockRange(filename string) (uint64, uint64, error) {
	match := blockRangeRegex.FindStringSubmatch(filename)
	if match == nil {
		return 0, 0, fmt.Errorf("no block range in filename: %s", filename)
	}

	startBlock, _ := strconv.ParseUint(match[1], 10, 64)
	stopBlock, _ := strconv.ParseUint(match[2], 10, 64)
	return startBlock, stopBlock, nil
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

	db.SetMaxOpenConns(250)

	zlog.Info("database connections created")
	return db, nil
}

func ExampleSparkle(in string) string {
	return string(cli.ExamplePrefixed("sparkle", in))
}

type BlockRange struct {
	Start uint64
	Stop  uint64
}

func (b BlockRange) Unbounded() bool {
	return b.Start == 0 && b.Stop == 0
}

func (b BlockRange) ReprocRange() string {
	return fmt.Sprintf("%d:%d", b.Start, b.Stop+1)
}

func (b BlockRange) String() string {
	return fmt.Sprintf("%s - %s", blockNum(b.Start), blockNum(b.Stop))
}

type blockNum uint64

func (b blockNum) String() string {
	return "#" + strings.ReplaceAll(humanize.Comma(int64(b)), ",", " ")
}
