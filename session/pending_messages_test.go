package session

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPendingMessages(t *testing.T) {
	Convey(`Testing pendingMessages`, t, func() {
		var p pendingMessages
		So(p.Index(1), ShouldEqual, -1)
		p = p.Insert(&packets.PublishPacket{MessageID: 2})
		p = p.Insert(&packets.PublishPacket{MessageID: 1})
		So(p.Index(1), ShouldEqual, 0)
		p = p.Remove(1)
		So(p.Index(1), ShouldEqual, -1)
		p = p.Remove(1)
		p = p.Remove(2)
		So(p, ShouldBeEmpty)
	})
}
