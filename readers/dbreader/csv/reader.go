// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/readers/dbreader/reader"
)

const (
	interval = 60 // seconds
	// FSUser stores reader manager params user key
	FSUser = "FSUser"
	// FSPass stores reader manager params pass key
	FSPass = "FSPass"
)

type csvReader struct {
	FileName   string `json:"filename"`
	ColsStr    string `json:"columns,omitempty"`
	columns    []string
	Interval64 float64 `json:"interval"`
	DBType     string  `json:"dbtype,omitempty"`
	HeadPos    int     `json:"header_pos"`
}

var (
	errEmptyCSV    = errors.New("the CSV file is empty ")
	errOpenCSVFile = errors.New("failed to open CSV file")
)

var _ reader.Reader = (*csvReader)(nil)

// New returns new reader client instance.
func New() reader.Reader {
	return &csvReader{}
}

// Validates - check whether mandatory config items are set
func (r *csvReader) Init(params map[string]string) {
	if len(r.ColsStr) > 0 {
		r.columns = strings.Split(r.ColsStr, ",")
	}
	if r.Interval64 <= 0 {
		r.Interval64 = interval
	}

	// Used for network file sharing on Windows
	usr := params[FSUser]
	pass := params[FSPass]
	if len(usr) < 1 || len(pass) < 1 {
		return
	}
	idx := strings.LastIndex(r.FileName, `\`)
	if idx < 0 {
		return
	}
	exec.Command("net", "use", r.FileName[:idx], fmt.Sprintf("/user:%s", usr), pass).CombinedOutput()
}

func (r *csvReader) Interval() float64 {
	return r.Interval64
}

func (r *csvReader) String() string {
	return fmt.Sprintf("%s", r.FileName)
}

// Read opens the file and reads out the data
func (r *csvReader) Read() (out []map[string]interface{}, err error) {
	// Open file
	file, err := os.Open(r.FileName)
	if err != nil {
		return out, fmt.Errorf(fmt.Sprintf("%s %s", errOpenCSVFile.Error(), r.FileName))
	}
	defer file.Close()

	// Parse file
	cr := csv.NewReader(file)
	cr.TrimLeadingSpace = true
	cr.LazyQuotes = true

	// Get headers
	record, err := cr.Read()
	if err == io.EOF {
		return out, errEmptyCSV
	}
	if err != nil {
		return out, err
	}

	// Set columns
	headers := make(map[int]string)
	if len(r.columns) > 0 {
		for i, header := range record {
			if contains(r.columns, header) {
				headers[i] = header
			}
		}
	} else {
		for i, header := range record {
			headers[i] = header
		}
	}

	if err := r.rewindHead(cr); err != nil {
		if err == io.EOF {
			return out, nil
		}
		return out, err
	}

	for {
		record, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return out, err
		}

		r.HeadPos++
		row := make(map[string]interface{}, len(headers))
		row[reader.MFX_ID] = r.HeadPos
		for i, y := range record {
			if header, ok := headers[i]; ok {
				value := parseValue(y)
				row[header] = value
			}
		}
		out = append(out, row)
	}

	return out, nil
}

func (r *csvReader) rewindHead(cr *csv.Reader) error {
	for i := 0; i < r.HeadPos; i++ {
		if _, err := cr.Read(); err != nil {
			return err
		}
	}
	return nil
}

func parseValue(value string) interface{} {
	if v, err := strconv.ParseInt(value, 32, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseUint(value, 32, 64); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(value, 32); err == nil {
		return v
	}
	if v, err := strconv.ParseBool(value); err == nil {
		return v
	}
	return value
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
