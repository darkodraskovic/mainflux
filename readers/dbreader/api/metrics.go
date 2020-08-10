// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"gitlab.com/mainflux/takeda/dbreader"
)

var _ dbreader.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     dbreader.Service
}

// MetricsMiddleware instruments core service by tracking request count and latency.
func MetricsMiddleware(svc dbreader.Service, counter metrics.Counter, latency metrics.Histogram) dbreader.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) CreateThing(mfxDevID string, channelID string, dbreaderID string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "create_thing").Add(1)
		mm.latency.With("method", "create_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.CreateThing(mfxDevID, channelID, dbreaderID)
}

func (mm *metricsMiddleware) UpdateThing(mfxDevID string, channelID, dbreaderID string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "update_thing").Add(1)
		mm.latency.With("method", "update_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.UpdateThing(mfxDevID, channelID, dbreaderID)
}

func (mm *metricsMiddleware) RemoveThing(mfxDevID string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "remove_thing").Add(1)
		mm.latency.With("method", "remove_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.RemoveThing(mfxDevID)
}

func (mm *metricsMiddleware) CreateChannel(mfxChanID string, DBName string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "create_channel").Add(1)
		mm.latency.With("method", "create_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.CreateChannel(mfxChanID, DBName)
}

func (mm *metricsMiddleware) RemoveChannel(mfxChanID string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "remove_channel").Add(1)
		mm.latency.With("method", "remove_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.RemoveChannel(mfxChanID)
}

func (mm *metricsMiddleware) Publish(m dbreader.Message) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "publish").Add(1)
		mm.latency.With("method", "publish").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Publish(m)
}
