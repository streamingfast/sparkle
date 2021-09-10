package squashable

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/streamingfast/sparkle/entity"

	"go.uber.org/zap"

	"github.com/streamingfast/dstore"
)

type SnapshotFile struct {
	filename      string
	startBlockNum uint64
	stopBlockNum  uint64
}

type BlockRange struct {
	startBlock uint64
	stopBlock  uint64
}

func (s *store) Preload(ctx context.Context, in dstore.Store) (uint64, uint64, error) {
	s.logger.Info("prelaoding snapshots")
	files := []string{}
	err := in.Walk(context.Background(), "", "", func(filename string) (err error) {
		files = append(files, filename)
		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("unable to walk input store: %w", err)
	}

	snapshotPath, err := s.pathFinder(files)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to find a valid snapshot path: %w", err)
	}

	s.logger.Info("loading snapshots", zap.Int("file_count", len(snapshotPath)))

	if len(snapshotPath) == 0 {
		return 0, 0, nil
	}

	for _, snapshot := range snapshotPath {
		if err := s.loadSnapshotFile(ctx, in, snapshot.filename); err != nil {
			return 0, 0, fmt.Errorf("unable to load snapshot file: %w", err)
		}
	}

	return snapshotPath[0].startBlockNum, snapshotPath[len(snapshotPath)-1].stopBlockNum, nil

}

type snapshotRawMessage struct {
	TableIdx int             `json:"t"`
	Entity   json.RawMessage `json:"d"`
}

func (s *store) loadSnapshotFile(ctx context.Context, in dstore.Store, snapshotsfilePath string) error {

	s.logger.Info("decoding filepath", zap.String("filepath", snapshotsfilePath))
	reader, err := in.OpenObject(ctx, snapshotsfilePath)
	if err != nil {
		return fmt.Errorf("unable to load input file %q: %w", snapshotsfilePath, err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	//how big can entities be ?
	//buf := make([]byte, ScannerMaxCapacity)
	//scanner.Buffer(buf, ScannerMaxCapacity)

	scanner.Scan()
	tableIdx := make(map[int]string)
	if err := json.Unmarshal(scanner.Bytes(), &tableIdx); err != nil {
		return err
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for scanner.Scan() {
		sr := snapshotRawMessage{}
		if err := json.Unmarshal(scanner.Bytes(), &sr); err != nil {
			return err
		}
		tableName := tableIdx[sr.TableIdx]
		reflectType, ok := s.subgraph.Entities.GetType(tableName)
		if !ok {
			return fmt.Errorf("no entity registered for table name %q", tableName)
		}
		cachedTable := s.cache[tableName]
		if cachedTable == nil {
			cachedTable = make(map[string]entity.Interface)
			s.cache[tableName] = cachedTable
		}

		el := reflect.New(reflectType).Interface()
		if err := json.Unmarshal(sr.Entity, el); err != nil {
			return fmt.Errorf("unmarshal raw entity: %w", err)
		}

		modifier := el.(entity.Interface)
		modifier.SetExists(true)

		id := modifier.GetID()
		cachedTable[id] = s.subgraph.MergeFunc(s.step, cachedTable[id], modifier)

	}

	return scanner.Err()
}

func (s *store) pathFinder(filenames []string) (out []*SnapshotFile, err error) {
	var current *SnapshotFile
	for _, filename := range filenames {
		if err != nil {
			return nil, fmt.Errorf("unable to find a valid snaphost file path: %w", err)
		}

		snapshotFile, err := getSnapshotInfo(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to parse snapshot filename %q: %w", filename, err)
		}

		if current == nil {
			if snapshotFile.stopBlockNum >= s.StartBlock {
				// the files does not fit the desired startblock
				continue
			}
			// this is the initial case we simply want to append the file and setup up
			current = snapshotFile
			continue
		}

		if snapshotFile.startBlockNum == current.startBlockNum {
			// we need to look at the end block to determine which snapshotFile we want
			if snapshotFile.stopBlockNum > current.stopBlockNum && snapshotFile.stopBlockNum < s.StartBlock {
				current = snapshotFile
			}
			continue
		}

		// at this point we are at a "start block boundary"
		if snapshotFile.startBlockNum < current.stopBlockNum {
			// in this case we have already passed this start block we should just
			// skip it
			continue
		}

		// the start block is a valid one we should check for contiguous values
		if snapshotFile.startBlockNum != (current.stopBlockNum + 1) {
			return nil, fmt.Errorf("unable to find a contiguous path expected block %d actual current block %d", current.stopBlockNum+1, snapshotFile.startBlockNum)
		}

		if current.stopBlockNum == (s.StartBlock - 1) {
			// path found! the current will be once the loop is done
			break
		}

		out = append(out, current)
		current = snapshotFile
	}
	if current == nil {
		return out, nil
	}
	if current.stopBlockNum != (s.StartBlock - 1) {
		return nil, fmt.Errorf("contiguous path is too short, expected end block %d actual current end block %d", s.StartBlock-1, current.stopBlockNum)
	}
	out = append(out, current)
	return out, nil
}

func getSnapshotInfo(filename string) (*SnapshotFile, error) {
	number := regexp.MustCompile(`(\d{10})-(\d{10})`)
	match := number.FindStringSubmatch(filename)
	if match == nil {
		return nil, fmt.Errorf("no block range in filename: %s", filename)
	}

	startBlock, _ := strconv.ParseUint(match[1], 10, 64)
	stopBlock, _ := strconv.ParseUint(match[2], 10, 64)
	if startBlock >= stopBlock {
		return nil, fmt.Errorf("invalid block range for file %q start block is greater or equal to end block", filename)
	}
	return &SnapshotFile{
		filename:      filename,
		startBlockNum: startBlock,
		stopBlockNum:  stopBlock,
	}, nil
}
