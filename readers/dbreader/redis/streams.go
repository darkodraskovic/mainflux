// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/readers/dbreader"
)

const (
	keyProtocol = "dbreader"
	keyDbType   = "dbtype"

	group  = "mainflux.dbreader"
	stream = "mainflux.things"

	thingPrefix = "thing."
	thingCreate = thingPrefix + "create"
	thingUpdate = thingPrefix + "update"
	thingRemove = thingPrefix + "remove"

	channelPrefix = "channel."
	channelCreate = channelPrefix + "create"
	channelUpdate = channelPrefix + "update"
	channelRemove = channelPrefix + "remove"

	exists = "BUSYGROUP Consumer Group name already exists"
)

var (
	errMetadataType       = errors.New("metadatada is not of type dbreader")
	errMetadataDBType     = errors.New("metadatada is not of type mssql dbreader")
	errMetadataFormat     = errors.New("malformed metadata")
	errMetadataNamespace  = errors.New("dbname not found in channel metadatada")
	errMetadataIdentifier = errors.New("dbname not found in thing metadatada")
)

var _ dbreader.EventStore = (*eventStore)(nil)

type eventStore struct {
	svc      dbreader.Service
	client   *redis.Client
	consumer string
	logger   logger.Logger
}

type dbReaderMetadata struct {
	Type         string                 `json:"type"`
	DbReaderData map[string]interface{} `json:"db_reader_data"`
	ChannelID    string                 `json:"channel_id"`
}

// NewEventStore returns new event store instance.
func NewEventStore(svc dbreader.Service, client *redis.Client, consumer string, log logger.Logger) dbreader.EventStore {
	return eventStore{
		svc:      svc,
		client:   client,
		consumer: consumer,
		logger:   log,
	}
}

func (es eventStore) Subscribe(subject string) error {
	err := es.client.XGroupCreateMkStream(stream, group, "$").Err()
	if err != nil && err.Error() != exists {
		return err
	}

	for {
		streams, err := es.client.XReadGroup(&redis.XReadGroupArgs{
			Group:    group,
			Consumer: es.consumer,
			Streams:  []string{stream, ">"},
			Count:    100,
		}).Result()
		if err != nil || len(streams) == 0 {
			continue
		}

		for _, msg := range streams[0].Messages {
			event := msg.Values
			var err error
			switch event["operation"] {
			case thingCreate:
				cte, e := decodeCreateThing(event)
				if e != nil {
					err = e
					break
				}
				err = es.handleCreateThing(cte)
			case thingUpdate:
				ute, e := decodeCreateThing(event)
				if e != nil {
					es.logger.Debug(e.Error())
					err = e
					break
				}
				err = es.handleCreateThing(ute)
			case thingRemove:
				rte := decodeRemoveThing(event)
				err = es.handleRemoveThing(rte)
			case channelRemove:
				rce := decodeRemoveChannel(event)
				err = es.handleRemoveChannel(rce)
			}
			if err != nil && err != errMetadataType {
				es.logger.Warn(fmt.Sprintf("Failed to handle event sourcing: %s", err.Error()))
				break
			}
			es.client.XAck(stream, group, msg.ID)
		}
	}
}

func decodeCreateThing(event map[string]interface{}) (createThingEvent, error) {
	// get metadata string from the event
	strmeta := read(event, "metadata", "{}")

	var metadata dbReaderMetadata
	if err := json.Unmarshal([]byte(strmeta), &metadata); err != nil {
		return createThingEvent{}, err
	}

	if metadata.Type != "dbReader" {
		return createThingEvent{}, errMetadataType
	}

	dbReaderDataJSON, err := json.Marshal(metadata.DbReaderData)

	if err != nil {
		return createThingEvent{}, err
	}

	cte := createThingEvent{
		id:           read(event, "id", ""),
		channelID:    metadata.ChannelID,
		dbReaderData: string(dbReaderDataJSON),
	}

	return cte, nil
}

func decodeRemoveThing(event map[string]interface{}) removeThingEvent {
	return removeThingEvent{
		id: read(event, "id", ""),
	}
}

func decodeRemoveChannel(event map[string]interface{}) removeChannelEvent {
	return removeChannelEvent{
		id: read(event, "id", ""),
	}
}

func (es eventStore) handleCreateThing(cte createThingEvent) error {
	return es.svc.CreateThing(cte.id, cte.channelID, cte.dbReaderData)
}

func (es eventStore) handleUpdateThing(cte createThingEvent) error {
	return es.svc.UpdateThing(cte.id, cte.channelID, cte.dbReaderData)
}

func (es eventStore) handleRemoveThing(rte removeThingEvent) error {
	return es.svc.RemoveThing(rte.id)
}

func (es eventStore) handleRemoveChannel(rce removeChannelEvent) error {
	return es.svc.RemoveChannel(rce.id)
}

func read(event map[string]interface{}, key, def string) string {
	val, ok := event[key].(string)
	if !ok {
		return def
	}

	return val
}
