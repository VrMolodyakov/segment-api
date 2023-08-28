package converter

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
)

type CSVWritable interface {
	Row() []string
	Headers() []string
}

type Writer func(w io.Writer, args ...CSVWritable) error

type csvWriter struct {
	writer Writer
}

func NewCSVWriter(writer Writer) csvWriter {
	return csvWriter{
		writer: writer,
	}
}

func (csvw *csvWriter) Write(w io.Writer, args ...CSVWritable) error {
	return csvw.writer(w, args...)
}

func Write(w io.Writer, args ...CSVWritable) error {
	writer := csv.NewWriter(w)
	if len(args) == 0 {
		return errors.New("no arguments provided")
	}
	if err := writer.Write(args[0].Headers()); err != nil {
		return fmt.Errorf("couldn't write headers : %w", err)
	}
	for _, row := range args {
		if err := writer.Write(row.Row()); err != nil {
			return fmt.Errorf("couldn't write row : %w", err)
		}
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Errorf("couldn't flush writer : %w", err)
	}
	return nil
}
