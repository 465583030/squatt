package session

import (
	"sort"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/htdvisser/pkg/sortutil"
)

type pendingMessages []packets.ControlPacket

func (p pendingMessages) Len() int { return len(p) }
func (p pendingMessages) Less(i, j int) bool {
	return p[i].Details().MessageID < p[j].Details().MessageID
}
func (p pendingMessages) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p pendingMessages) Search(messageID uint16) int {
	return sort.Search(len(p), func(i int) bool {
		return p[i].Details().MessageID >= messageID
	})
}
func (p pendingMessages) Index(messageID uint16) int {
	pos := p.Search(messageID)
	if pos < len(p) && p[pos].Details().MessageID == messageID {
		return pos
	}
	return -1
}
func (p pendingMessages) Insert(msg packets.ControlPacket) pendingMessages {
	updated := append(p, msg)
	sortutil.SwapLast(updated)
	return updated
}
func (p pendingMessages) Remove(messageID uint16) pendingMessages {
	pos := p.Search(messageID)
	if pos < len(p) && p[pos].Details().MessageID == messageID {
		return append(p[:pos], p[pos+1:]...)
	}
	return p
}

func (s *Session) inFlight() int {
	return s.pendingAck.Len() + s.pendingRec.Len() + s.pendingRel.Len() + s.pendingComp.Len()
}
