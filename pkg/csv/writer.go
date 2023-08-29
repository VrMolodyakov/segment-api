package csv

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

type Writer[T CSVWritable] func(w io.Writer, args []T) error

type csvWriter[T CSVWritable] struct {
	writer Writer[T]
}

func NewCSVWriter[T CSVWritable](writer Writer[T]) csvWriter[T] {
	return csvWriter[T]{
		writer: writer,
	}
}

func (csvw *csvWriter[T]) Write(w io.Writer, args []T) error {
	return csvw.writer(w, args)
}

func Write[T CSVWritable](w io.Writer, args []T) error {
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
