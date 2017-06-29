package server

import (
	"fmt"
	"io"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"go.uber.org/zap"
)

var errProtocolViolation = packets.ConnErrors[packets.ErrProtocolViolation]

func (c *Client) sendRoutine(w io.Writer) {
	for msg := range c.sendCh {
		if err := msg.Write(w); err != nil {
			c.setError(err)
			return
		}
	}
}

func (c *Client) send(packet packets.ControlPacket) (err error) {
	defer func() {
		if r := recover(); r != nil {
			c.log.Error(fmt.Sprintf("unexpected error: %v", r))
			err = fmt.Errorf("unexpected error: %v", r)
		}
	}()

	if err = c.getError(); err != nil {
		return err
	}

	switch packet := packet.(type) {
	case *packets.ConnackPacket:
		c.log.Debug(
			"send connack",
			zap.String("return-code", packets.ConnackReturnCodes[packet.ReturnCode]),
		)
	case *packets.PublishPacket:
		c.log.Debug(
			"send publish",
			zap.Bool("dup", packet.Dup),
			zap.Bool("ret", packet.Retain),
			zap.Int("len", len(packet.Payload)),
			zap.String("topic", packet.TopicName),
			zap.Uint16("mid", packet.MessageID),
			zap.Uint8("qos", packet.Qos),
		)
	case *packets.PubackPacket:
		c.log.Debug(
			"send puback",
			zap.Uint16("mid", packet.MessageID),
		)
	case *packets.PubrecPacket:
		c.log.Debug(
			"send pubrec",
			zap.Uint16("mid", packet.MessageID),
		)
	case *packets.PubrelPacket:
		c.log.Debug(
			"send pubrel",
			zap.Uint16("mid", packet.MessageID),
		)
	case *packets.PubcompPacket:
		c.log.Debug(
			"send pubcomp",
			zap.Uint16("mid", packet.MessageID),
		)
	case *packets.SubackPacket:
		c.log.Debug(
			"send suback",
			zap.Uint16("mid", packet.MessageID),
		)
	case *packets.UnsubackPacket:
		c.log.Debug(
			"send unsuback",
			zap.Uint16("mid", packet.MessageID),
		)
	case *packets.PingrespPacket:
		c.log.Debug("send pingresp")
	}

	c.sendCh <- packet

	return
}

func (c *Client) receiveRoutine(r io.Reader) {
	for {
		msg, err := packets.ReadPacket(r)
		if err != nil {
			c.setError(err)
			return
		}
		err = c.receive(msg)
		if err != nil {
			c.setError(err)
			return
		}
	}
}

func packetType(packet packets.ControlPacket) byte {
	if packet == nil {
		return 0
	}
	switch packet := packet.(type) {
	case *packets.ConnectPacket:
		return packet.FixedHeader.MessageType
	case *packets.ConnackPacket:
		return packet.FixedHeader.MessageType
	case *packets.PublishPacket:
		return packet.FixedHeader.MessageType
	case *packets.PubackPacket:
		return packet.FixedHeader.MessageType
	case *packets.PubrecPacket:
		return packet.FixedHeader.MessageType
	case *packets.PubrelPacket:
		return packet.FixedHeader.MessageType
	case *packets.PubcompPacket:
		return packet.FixedHeader.MessageType
	case *packets.SubscribePacket:
		return packet.FixedHeader.MessageType
	case *packets.SubackPacket:
		return packet.FixedHeader.MessageType
	case *packets.UnsubscribePacket:
		return packet.FixedHeader.MessageType
	case *packets.UnsubackPacket:
		return packet.FixedHeader.MessageType
	case *packets.PingreqPacket:
		return packet.FixedHeader.MessageType
	case *packets.PingrespPacket:
		return packet.FixedHeader.MessageType
	case *packets.DisconnectPacket:
		return packet.FixedHeader.MessageType
	}
	return 0
}

func (c *Client) receive(packet packets.ControlPacket) (err error) {
	defer func() {
		if r := recover(); r != nil {
			c.log.Error(fmt.Sprintf("unexpected error: %v", r))
			err = fmt.Errorf("unexpected error: %v", r)
		}
	}()

	// Validation
	if packet == nil {
		return errProtocolViolation
	}
	packetType := packetType(packet)
	if packetType == 0 {
		return errProtocolViolation
	}
	if (packetType == packets.Connect) != (c.session == nil) {
		return errProtocolViolation
	}
	if packet.Details().Qos == 0x03 {
		return errProtocolViolation
	}

	// KeepAlive
	c.keepAlive.Kick()

	// Handle
	switch packet := packet.(type) {
	case *packets.ConnectPacket:
		c.log.Debug(
			"receive connect",
			zap.String("cid", packet.ClientIdentifier),
			zap.String("username", packet.Username),
		)
		return c.handleConnect(packet)
	case *packets.PublishPacket:
		c.log.Debug(
			"receive publish",
			zap.Bool("dup", packet.Dup),
			zap.Bool("ret", packet.Retain),
			zap.Int("len", len(packet.Payload)),
			zap.String("topic", packet.TopicName),
			zap.Uint16("mid", packet.MessageID),
			zap.Uint8("qos", packet.Qos),
		)
		return c.handlePublish(packet)
	case *packets.PubackPacket:
		c.log.Debug(
			"receive puback",
			zap.Uint16("mid", packet.MessageID),
		)
		return c.handlePuback(packet)
	case *packets.PubrecPacket:
		c.log.Debug(
			"receive pubrec",
			zap.Uint16("mid", packet.MessageID),
		)
		return c.handlePubrec(packet)
	case *packets.PubrelPacket:
		c.log.Debug(
			"receive pubrel",
			zap.Uint16("mid", packet.MessageID),
		)
		return c.handlePubrel(packet)
	case *packets.PubcompPacket:
		c.log.Debug(
			"receive pubcomp",
			zap.Uint16("mid", packet.MessageID),
		)
		return c.handlePubcomp(packet)
	case *packets.SubscribePacket:
		c.log.Debug(
			"receive subscribe",
			zap.Uint16("mid", packet.MessageID),
			zap.Strings("topics", packet.Topics),
		)
		return c.handleSubscribe(packet)
	case *packets.UnsubscribePacket:
		c.log.Debug(
			"receive unsubscribe",
			zap.Uint16("mid", packet.MessageID),
			zap.Strings("topics", packet.Topics),
		)
		return c.handleUnsubscribe(packet)
	case *packets.PingreqPacket:
		c.log.Debug("receive pingreq")
		return c.handlePingreq(packet)
	case *packets.DisconnectPacket:
		c.log.Debug("receive disconnect")
		return c.handleDisconnect(packet)
	default:
		return errProtocolViolation
	}
}
