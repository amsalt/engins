package conf

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"

	"github.com/amsalt/engins/errs"
)

func init() {
	Register(&CSVParser{})
}

type CSVParser struct {
}

func (c *CSVParser) Name() string {
	return "csv"
}

func (c *CSVParser) Parse(path string, target interface{}) error {
	csvFile, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := csv.NewReader(bufio.NewReader(csvFile))

	resultv := reflect.ValueOf(target)
	if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		panic("target must be a slice address")
	}

	// ignore header
	_, err = reader.Read()
	if err != nil {
		return err
	}

	slicev := resultv.Elem()
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()

	var i int
	for {
		elemp := reflect.New(elemt)
		err := c.parseOneLine(reader, elemp.Interface())

		{
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("parse csv failed: %+v", err)
				return err
			}
		}

		slicev = reflect.Append(slicev, elemp.Elem())
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))

	return err
}

func (c *CSVParser) parseOneLine(reader *csv.Reader, v interface{}) error {
	record, err := reader.Read()
	if err != nil {
		return err
	}
	s := reflect.ValueOf(v).Elem()
	if s.NumField() != len(record) {
		return errs.NewFieldMismatch(s.NumField(), len(record))
	}

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		switch f.Type().String() {
		case "string":
			f.SetString(record[i])
		case "int":
			ival, err := strconv.ParseInt(record[i], 10, 0)
			if err != nil {
				return err
			}
			f.SetInt(ival)
		case "float32":
			ival, err := strconv.ParseFloat(record[i], 32)
			if err != nil {
				return err
			}
			f.SetFloat(ival)
		case "float64":
			ival, err := strconv.ParseFloat(record[i], 64)
			if err != nil {
				return err
			}
			f.SetFloat(ival)

		default:
			return errs.NewUnsupportedType(f.Type().String())
		}
	}

	return nil
}
