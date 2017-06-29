package topic

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTopicStore(t *testing.T) {
	Convey(`Testing the Topic Store`, t, func() {
		s := NewStore()

		foo := s.Get("foo")
		So(foo.Name(), ShouldEqual, "foo")

		So(s.Get("foo"), ShouldEqual, foo)

		So(s.Match("foo"), ShouldContain, foo)
		So(s.Match("bar"), ShouldNotContain, foo)
		So(s.Match("#"), ShouldContain, foo)
		So(s.Match("+"), ShouldContain, foo)

	})
}
