// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

type createThingEvent struct {
	id           string
	channelID    string
	dbReaderData string
}

type removeThingEvent struct {
	id string
}

type removeChannelEvent struct {
	id string
}
