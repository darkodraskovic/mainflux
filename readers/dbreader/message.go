// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package dbreader

// Message represent an tableadapter message
type Message struct {
	Type      string
	ChannelID string
	ThingID   string
	Table     []map[string]interface{}
}
