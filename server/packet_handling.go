package server

import (
	"io"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/squatt/topic"
	"github.com/segmentio/ksuid"
)

func (c *Client) handleConnect(packet *packets.ConnectPacket) (err error) {
	connack := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
	defer func() { err = packets.ConnErrors[connack.ReturnCode] }()
	if errCode := packet.Validate(); errCode != 0 {
		connack.ReturnCode = errCode
		c.send(connack)
		return
	}
	if packet.ClientIdentifier == "" {
		packet.ClientIdentifier = ksuid.New().String()
	}
	auth, err := c.server.auth(packet.ClientIdentifier, packet.Username, packet.Password)
	if err != nil || !auth.CanConnect() {
		connack.ReturnCode = packets.ErrRefusedNotAuthorised
		c.send(connack)
		return
	}
	if packet.CleanSession {
		c.session = c.server.sessions.New(packet.ClientIdentifier)
	} else {
		session, ok := c.server.sessions.GetOrNew(packet.ClientIdentifier)
		if ok {
			if session.Persistent() {
				connack.SessionPresent = true
			}
			session.Disconnect()
		}
		c.session = session
		c.session.SetPersistent()
	}

	c.session.SetAuth(auth)
	c.session.SetLogger(c.log)

	c.session.SetOnDisconnect(func() {
		if !c.session.Persistent() {
			c.session.Delete()
		}
	})
	c.session.SetOnDelete(func() {
		c.server.Unsubscribe(c.session)
	})

	if packet.WillFlag {
		c.session.SetWill(packet.WillTopic, packet.WillMessage, packet.WillQos, packet.WillRetain)
	}

	if packet.Keepalive != 0 {
		c.keepAlive = newWatchdog(time.Duration(packet.Keepalive)*time.Millisecond*1500, func() {
			c.setError(errKeepAliveTimeout)
		})
	}

	c.session.DeliverTo(c.server.Publish())

	if err := c.send(connack); err != nil {
		return err
	}

	sendCh := make(chan packets.ControlPacket, ClientSendBufferSize)
	go func() {
		for msg := range sendCh {
			if err := c.send(msg); err != nil {
				c.setError(err)
				return
			}
		}
		c.setError(io.EOF)
	}()

	c.session.Connect(sendCh)
	c.session.ResendPending()

	return
}

func (c *Client) handlePublish(packet *packets.PublishPacket) error {
	if err := topic.Validate(packet.TopicName, true); err != nil {
		return err
	}
	c.session.ReceivePublish(packet)
	return nil
}

func (c *Client) handlePuback(packet *packets.PubackPacket) error {
	c.session.ReceivePuback(packet)
	return nil
}

func (c *Client) handlePubrec(packet *packets.PubrecPacket) error {
	c.session.ReceivePubrec(packet)
	return nil
}

func (c *Client) handlePubrel(packet *packets.PubrelPacket) error {
	c.session.ReceivePubrel(packet)
	return nil
}

func (c *Client) handlePubcomp(packet *packets.PubcompPacket) error {
	c.session.ReceivePubcomp(packet)
	return nil
}

func (c *Client) handleSubscribe(packet *packets.SubscribePacket) error {
	if len(packet.Topics) != len(packet.Qoss) {
		return packets.ConnErrors[packets.ErrProtocolViolation]
	}
	suback := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
	suback.MessageID = packet.MessageID
	suback.ReturnCodes = make([]byte, len(packet.Topics))
	for i, topicName := range packet.Topics {
		if err := topic.Validate(topicName, true); err != nil {
			return err
		}
		if c.session.CanSubscribeTo(topicName) {
			c.server.Subscribe(c.session, c.server.topics.Get(topicName), packet.Qoss[i])
			suback.ReturnCodes[i] = packet.Qoss[i]
		} else {
			suback.ReturnCodes[i] = 0x80
		}
	}
	c.send(suback)
	return nil
}

func (c *Client) handleUnsubscribe(packet *packets.UnsubscribePacket) error {
	topics := make([]*topic.Topic, len(packet.Topics))
	for i, topic := range packet.Topics {
		topics[i] = c.server.topics.Get(topic)
	}
	c.server.Unsubscribe(c.session, topics...)
	unsuback := packets.NewControlPacket(packets.Unsuback).(*packets.UnsubackPacket)
	unsuback.MessageID = packet.MessageID
	c.send(unsuback)
	return nil
}

func (c *Client) handlePingreq(packet *packets.PingreqPacket) error {
	c.send(packets.NewControlPacket(packets.Pingresp).(*packets.PingrespPacket))
	return nil
}

func (c *Client) handleDisconnect(packet *packets.DisconnectPacket) error {
	c.session.ClearWill()
	c.session.Disconnect()
	return nil
}
