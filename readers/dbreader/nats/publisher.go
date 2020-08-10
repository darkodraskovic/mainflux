// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	broker "github.com/nats-io/go-nats"
)

var _ messaging.Publisher = (*natsPublisher)(nil)

type natsPublisher struct {
	nc *broker.Conn
}

// NewMessagePublisher instantiates NATS message publisher.
func NewMessagePublisher(nc *broker.Conn) messaging.Publisher {
	return &natsPublisher{nc}
}

func (pub *natsPublisher) Publish(topic string, msg messaging.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	return pub.nc.Publish(topic, data)
}
