// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers/dbreader"
)

var _ dbreader.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger logger.Logger
	svc    dbreader.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc dbreader.Service, logger logger.Logger) dbreader.Service {
	return &loggingMiddleware{
		logger: logger,
		svc:    svc,
	}
}

func (lm loggingMiddleware) CreateThing(mfxThing string, channelID string, dbreaderID string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("create_thing mfx:dbreader:%s:%s took %s to complete", mfxThing, dbreaderID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateThing(mfxThing, channelID, dbreaderID)
}

func (lm loggingMiddleware) UpdateThing(mfxThing string, channelID string, dbreaderID string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("update_thing mfx:dbreader:%s:%s took %s to complete", mfxThing, dbreaderID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.UpdateThing(mfxThing, channelID, dbreaderID)
}

func (lm loggingMiddleware) RemoveThing(mfxThing string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("remove_thing mfx:dbreader:%s took %s to complete", mfxThing, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RemoveThing(mfxThing)
}

func (lm loggingMiddleware) CreateChannel(mfxChan string, DBName string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("create_channel mfx:dbreader:%s:%s took %s to complete", mfxChan, DBName, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateChannel(mfxChan, DBName)
}

func (lm loggingMiddleware) RemoveChannel(mfxChanID string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("remove_channel mfx_channel_%s took %s to complete", mfxChanID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RemoveChannel(mfxChanID)
}

func (lm loggingMiddleware) Publish(m dbreader.Message) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("publish of type %s for thing %s and channel %s took %s to complete", m.Type, m.ThingID, m.ChannelID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Publish(m)
}
