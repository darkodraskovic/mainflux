//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// +build !test

package api

import (
	"fmt"
	"time"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/re"
)

var _ re.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    re.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc re.Service, logger log.Logger) re.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Info() (info re.Info, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method info took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Info()
}

func (lm *loggingMiddleware) CreateStream(name, topic, row string) (result string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_string for stream %s with topic %s took %s to complete", name, topic, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateStream(name, topic, row)
}

func (lm *loggingMiddleware) UpdateStream(name, topic, row string) (result string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_string for stream %s and topic %s took %s to complete", name, topic, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.UpdateStream(name, topic, row)
}

func (lm *loggingMiddleware) ListStreams() (streams []string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListStreams()
}

func (lm *loggingMiddleware) ViewStream(id string) (stream re.Stream, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view with id %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ViewStream(id)
}
