// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mssql

import (
	"fmt"
	"net/url"

	_ "github.com/denisenkom/go-mssqldb" // drivers for Microsoft SQL Server
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go/log"
	"gitlab.com/mainflux/takeda/dbreader/fieldbinding"
	"gitlab.com/mainflux/takeda/dbreader/reader"
)

type mssqlReader struct {
	DBType     string `json:"dbtype"`
	Server     string `json:"server"`
	Instance   string `json:"instance"`
	Port       int    `json:"port"`
	User       string `json:"dbuser"`
	Pass       string `json:"dbpass"`
	Database   string `json:"database"`
	appName    string
	TableName  string  `json:"table"`
	Columns    string  `json:"columns"`
	Where      string  `json:"where"`
	Interval64 float64 `json:"interval"`
}

var _ reader.Reader = (*mssqlReader)(nil)

// New returns new dbreader client instance.
func New() reader.Reader {
	return &mssqlReader{}
}

// Validates - check whether mandatory config items are set
func (r *mssqlReader) Init(params map[string]string) {
	if r.Where == "" {
		// Get all rows "WHERE 1=1"
		r.Where = "1=1"
	}
	if r.Columns == "" {
		// Retrieve all columns with "SELECT *"
		r.Columns = "*"
	}

	r.appName = "dbreader"
}

func (r *mssqlReader) Interval() float64 {
	return r.Interval64
}

func (r *mssqlReader) String() string {
	return fmt.Sprintf("%s:%d %s:%s", r.Server, r.Port, r.Database, r.TableName)
}

// Run Connect creates a connection to the Microsoft SQL Server
func (r *mssqlReader) Read() (out []map[string]interface{}, err error) {
	out = []map[string]interface{}{}

	query := url.Values{}
	query.Add("app name", r.appName)
	query.Add("database", r.Database)
	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(r.User, r.Pass),
		Host:     fmt.Sprintf("%s:%d", r.Server, r.Port),
		RawQuery: query.Encode(),
	}
	if len(r.Instance) > 0 {
		u.Path = r.Instance
		u.Host = r.Server
	}

	db, err := sqlx.Open("sqlserver", u.String())
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Error(err)
		return out, err
	}

	// Construct SQL Statement for data retrieving
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s", r.Columns, r.TableName, r.Where)

	rows, err := db.Queryx(q)
	if err != nil {
		return out, err
	}

	var fields []string
	fb := fieldbinding.NewFieldBinding()
	if fields, err = rows.Columns(); err != nil {
		return out, err
	}
	fb.PutFields(fields)

	// Get data from recordset and append to output array
	for rows.Next() {
		if err := rows.Scan(fb.GetFieldPtrArr()...); err != nil {
			return out, err
		}
		out = append(out, fb.GetFieldArr())
	}
	if err := rows.Err(); err != nil {
		return out, err
	}

	return out, nil
}
