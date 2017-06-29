package session

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSessionStore(t *testing.T) {
	Convey(`Testing the Session Store`, t, func() {
		s := NewStore()

		s.Delete("foo")

		s1, existed := s.GetOrNew("foo")
		So(existed, ShouldBeFalse)
		So(s1.Name(), ShouldEqual, "foo")

		var s1Deleted bool
		s1.SetOnDelete(func() { s1Deleted = true })

		s2, existed := s.GetOrNew("foo")
		So(existed, ShouldBeTrue)
		So(s2, ShouldEqual, s1)
		So(s1Deleted, ShouldBeFalse)

		s3 := s.New("foo")
		So(s3.Name(), ShouldEqual, "foo")
		So(s1Deleted, ShouldBeTrue)

		var s3Deleted bool
		s3.SetOnDelete(func() { s3Deleted = true })

		s.Delete("foo")
		So(s3Deleted, ShouldBeTrue)
	})
}
