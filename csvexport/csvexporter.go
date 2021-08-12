package csvexport

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/jszwec/csvutil"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/sparkle/entity"
	"go.uber.org/zap"
)

type Writer struct {
	writer     *io.PipeWriter
	done       chan struct{}
	csvEncoder *csvutil.Encoder
	csvWriter  *csv.Writer
	filename   string
	StopBlock  uint64
}

func New(ctx context.Context, store dstore.Store, filename string, stopBlock uint64, autoHeaders bool) (*Writer, error) {
	reader, writer := io.Pipe()
	csvWriter := csv.NewWriter(writer)
	csvEncoder := csvutil.NewEncoder(csvWriter)
	csvEncoder.Tag = "csv"
	csvEncoder.AutoHeader = autoHeaders

	ce := &Writer{
		filename:   filename,
		csvEncoder: csvEncoder,
		csvWriter:  csvWriter,
		writer:     writer,
		StopBlock:  stopBlock,
		done:       make(chan struct{}),
	}

	go func() {
		err := store.WriteObject(ctx, filename, reader)
		if err != nil {
			// todo: better handle error
			panic(fmt.Errorf("failed writting object in file object %q: %w", filename, err))
		}
		close(ce.done)
	}()

	return ce, nil
}

func (c *Writer) Encode(ent entity.Interface) error {
	return c.csvEncoder.Encode(ent)
}

func (c *Writer) Close() error {
	// FIXME: sync.Once ?

	c.csvWriter.Flush()
	if err := c.csvWriter.Error(); err != nil {
		return fmt.Errorf("error flushing csv encoder: %w", err)
	}

	if err := c.writer.Close(); err != nil {
		return fmt.Errorf("closing csv writer: %w", err)
	}
	zlog.Debug("waiting for the store write object to complete", zap.String("filename", c.filename))
	<-c.done
	return nil
}
