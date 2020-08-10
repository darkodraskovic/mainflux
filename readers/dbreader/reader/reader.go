// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package reader

const MFX_ID = "mfx_id"

// Reader represents the csv dbreader client.
type Reader interface {
	// Init sets reader specific params
	Init(map[string]string)
	// Read reads data tabular subset and stores it in []map[string]interface{}
	Read() ([]map[string]interface{}, error)
	// String returns string representation of reader
	String() string
	// Interval used for reader scheduler
	Interval() float64
}
