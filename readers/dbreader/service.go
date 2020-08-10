// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package dbreader

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/readers/dbreader/reader"
)

const (
	protocol      = "dbreader"
	thingSuffix   = "thing"
	channelSuffix = "channel"
)

var (
	// ErrMalformedIdentity indicates malformed identity received (e.g.
	// invalid namespace or ID).
	ErrMalformedIdentity = errors.New("malformed identity received")

	// ErrMalformedMessage indicates malformed dbreader message.
	ErrMalformedMessage = errors.New("malformed message received")

	// ErrNotFoundIdentifier indicates a non-existent route map for a database identifier.
	ErrNotFoundIdentifier = errors.New("route map not found for this dbreader identifier")

	// ErrNotFoundNamespace indicates a non-existent route map for an database identifier.
	ErrNotFoundNamespace = errors.New("route map not found for this dbreader identifier")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateThing creates thing  mfx:dbreader
	CreateThing(string, string, string) error
	// UpdateThing updates thing mfx:dbreader
	UpdateThing(string, string, string) error
	// RemoveThing removes thing mfx:dbreader
	RemoveThing(string) error
	// CreateChannel creates channel mfx:dbreader
	CreateChannel(string, string) error
	// RemoveChannel removes channel mfx:dbreader
	RemoveChannel(string) error
	// Publish forwards messages from the dbreader to Mainflux NATS broker
	Publish(Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	publisher     messaging.Publisher
	ReaderManager ReaderManager
	logger        logger.Logger
	batch         bool
}

// New instantiates the dbreader adapter implementation.
func New(pub messaging.Publisher, ReaderManager ReaderManager, logger logger.Logger) Service {
	return &adapterService{
		publisher:     pub,
		ReaderManager: ReaderManager,
		logger:        logger,
	}
}

// Publish forwards messages from dbreader to Mainflux NATS broker
func (as *adapterService) Publish(m Message) error {
	for _, row := range m.Table {
		rowID := reader.MFX_ID
		if val, ok := row[reader.MFX_ID]; ok {
			rowID = fmt.Sprint(val)
		}
		for col, val := range row {
			if val == nil || col == reader.MFX_ID {
				continue
			}

			v := strings.Trim(fmt.Sprintf("%v", val), " ")

			if v == "" {
				continue
			}

			msg := messaging.Message{
				Publisher: m.ThingID,
				Protocol:  protocol,
				Channel:   m.ChannelID,
				Subtopic:  col,
				Payload:   genSenML(rowID, v),
			}
			topic := "channels." + m.ChannelID + "." + col
			if err := as.publisher.Publish(topic, msg); err != nil {
				return err
			}
		}
	}
	return nil
}

func (as *adapterService) CreateThing(thingID string, channelID string, metadata string) error {
	if len(metadata) <= 0 || metadata == "null" {
		return ErrMalformedIdentity
	}

	id, err := as.ReaderManager.Create(thingID, channelID, metadata)
	if err != nil {
		return err
	}

	as.ReaderManager.SaveAll()
	as.ReaderManager.Schedule(as, id)

	return nil
}

func (as *adapterService) UpdateThing(thingID string, channelID string, metadata string) error {
	return nil
}

func (as *adapterService) RemoveThing(thingID string) error {
	if err := as.ReaderManager.Delete(thingID); err != nil {
		return fmt.Errorf("Reader %s not found: %s", thingID, err.Error())

	}
	as.ReaderManager.SaveAll()
	as.logger.Info(fmt.Sprintf("Successfully removed reader for %s", thingID))
	return nil
}

func (as *adapterService) CreateChannel(channelID string, dbreaderDBName string) error {
	return nil
}

func (as *adapterService) RemoveChannel(channelID string) error {
	// 1. find whether thing exists which has this channel assigned
	// if found....find if dbreader is running
	// if running ... stop it
	// 2. remove from CSV
	// if not running..nothing
	// if not found...nothing

	return nil
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
func isBoolean(s string) bool {
	_, err := strconv.ParseBool(s)
	return err == nil
}

var timestamps = map[string]string{
	"ANSIC":       "Mon Jan _2 15:04:05 2006",
	"UnixDate":    "Mon Jan _2 15:04:05 MST 2006",
	"RubyDate":    "Mon Jan 02 15:04:05 -0700 2006",
	"RFC822":      "02 Jan 06 15:04 MST",
	"RFC822Z":     "02 Jan 06 15:04 -0700",
	"RFC850":      "Monday, 02-Jan-06 15:04:05 MST",
	"RFC1123":     "Mon, 02 Jan 2006 15:04:05 MST",
	"RFC1123Z":    "Mon, 02 Jan 2006 15:04:05 -0700",
	"RFC3339":     "2006-01-02T15:04:05Z07:00",
	"RFC3339Nano": "2006-01-02T15:04:05.999999999Z07:00",
	"Stamp":       "Jan _2 15:04:05",
	"StampMilli":  "Jan _2 15:04:05.000",
	"StampMicro":  "Jan _2 15:04:05.000000",
	"StampNano":   "Jan _2 15:04:05.000000000",
}

func genSenML(rowID, v string) []byte {
	vKey, vLim := "vs", `"`
	if isNumeric(v) {
		vKey, vLim = "v", ""
	} else if isBoolean(v) {
		vKey, vLim = "vb", ""
	} else if isDate(v) {
		vKey, vLim = "vd", ""
	}
	senml := fmt.Sprintf(`[{"n":"%s","%s":%s%v%s}]`, rowID, vKey, vLim, v, vLim)

	return []byte(senml)
}

func isDate(s string) bool {
	for _, ts := range timestamps {
		_, err := time.Parse(ts, s)
		if err == nil {
			return true
		}
	}
	return false
}
