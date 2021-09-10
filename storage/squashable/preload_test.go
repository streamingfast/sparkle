package squashable

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getSnapshotInfo(t *testing.T) {
	test := []struct {
		name               string
		filename           string
		expectSnapshotFile *SnapshotFile
		expectError        bool
	}{
		{
			name:     "valid file",
			filename: "0006809700-0006829699.jsonl",
			expectSnapshotFile: &SnapshotFile{
				filename:      "0006809700-0006829699.jsonl",
				startBlockNum: 6809700,
				stopBlockNum:  6829699,
			},
			expectError: false,
		},
		{
			name:        "invalid file",
			filename:    "0006809700-.jsonl",
			expectError: true,
		},
		{
			name:        "invalid aggregate file, start block bigger then end block",
			filename:    "0006829699-0006809700.jsonl",
			expectError: true,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			sf, err := getSnapshotInfo(test.filename)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectSnapshotFile, sf)
			}
		})
	}

}

//
//41
//minStarblock = 0
//
//0 - 9
//10 - 19
//20 - 29
//0 - 29
//0 - 39
//
//
//0  - [{0 - 9}, {0 - 29}, {0 - 39}]
//10 - [{20 - 29}]
//40 - []
//
//

func testFilename(startBlock, endBlock uint64) string {
	return fmt.Sprintf("%010d-%010d.json", startBlock, endBlock)
}

func Test_pathFinderWrapper(t *testing.T) {

	test := []struct {
		name           string
		files          []string
		startBlockNum  uint64
		expectFilePath []*SnapshotFile
		expectError    error
	}{
		{
			name: "simple valid path",
			files: []string{
				testFilename(0, 9),
				testFilename(10, 19),
				testFilename(20, 29),
				testFilename(30, 39),
			},
			startBlockNum: 40,
			expectFilePath: []*SnapshotFile{
				{testFilename(0, 9), 0, 9},
				{testFilename(10, 19), 10, 19},
				{testFilename(20, 29), 20, 29},
				{testFilename(30, 39), 30, 39},
			},
		},
		{
			name: "valid path with only aggregate",
			files: []string{
				testFilename(0, 9),
				testFilename(0, 39),
				testFilename(10, 19),
				testFilename(10, 29),
				testFilename(20, 29),
				testFilename(30, 39),
			},
			startBlockNum: 40,
			expectFilePath: []*SnapshotFile{
				{testFilename(0, 39), 0, 39},
			},
		},
		{
			name: "valid path with aggregate help",
			files: []string{
				testFilename(0, 9),
				testFilename(10, 19),
				testFilename(10, 29),
				testFilename(20, 29),
				testFilename(30, 39),
			},
			startBlockNum: 40,
			expectFilePath: []*SnapshotFile{
				{testFilename(0, 9), 0, 9},
				{testFilename(10, 29), 10, 29},
				{testFilename(30, 39), 30, 39},
			},
		},
		{
			name: "valid path with aggregate help where some aggregate are too large",
			files: []string{
				testFilename(0, 49),
				testFilename(0, 29),
				testFilename(10, 19),
				testFilename(30, 39),
			},
			startBlockNum: 40,
			expectFilePath: []*SnapshotFile{
				{testFilename(0, 29), 0, 29},
				{testFilename(30, 39), 30, 39},
			},
		},
		{
			name: "valid path with aggregate help where there are more files then required",
			files: []string{
				testFilename(0, 29),
				testFilename(10, 19),
				testFilename(30, 39),
				testFilename(40, 49),
				testFilename(50, 59),
			},
			startBlockNum: 40,
			expectFilePath: []*SnapshotFile{
				{testFilename(0, 29), 0, 29},
				{testFilename(30, 39), 30, 39},
			},
		},
		{
			name: "invalid file path stops before desired start block",
			files: []string{
				testFilename(0, 9),
				testFilename(0, 29),
				testFilename(10, 19),
				testFilename(20, 29),
				testFilename(30, 39),
			},
			startBlockNum: 41,
			expectError:   fmt.Errorf("contiguous path is too short, expected end block 40 actual current end block 39"),
		},
		{
			name: "invalid file path stops has un-contiguous block ranges",
			files: []string{
				testFilename(0, 9),
				testFilename(0, 19),
				testFilename(10, 19),
				testFilename(30, 39),
			},
			startBlockNum: 40,
			expectError:   fmt.Errorf("unable to find a contiguous path expected block 20 actual current block 30"),
		},
		{
			name: "valid file path test",
			files: []string{
				testFilename(0, 9),
				testFilename(0, 19),
				testFilename(10, 19),
				testFilename(20, 29),
			},
			startBlockNum: 10,
			expectFilePath: []*SnapshotFile{
				{testFilename(0, 9), 0, 9},
			},
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {

			s := &store{
				StartBlock: test.startBlockNum,
			}
			out, err := s.pathFinder(test.files)
			if test.expectError != nil {
				assert.Equal(t, test.expectError, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectFilePath, out)
			}
		})
	}
}
