package auth

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuth(t *testing.T) {
	Convey(`When calling NoAuth`, t, func() {
		auth, err := NoAuth("id", "user", []byte("foo"))
		Convey(`Then There should be no error`, func() { So(err, ShouldBeNil) })
		Convey(`Then it should return the username`, func() { So(auth.Username(), ShouldEqual, "user") })
		Convey(`Then Connecting should be allowed`, func() { So(auth.CanConnect(), ShouldBeTrue) })
		Convey(`Then Publishing should be allowed`, func() { So(auth.CanPublishTo(""), ShouldBeTrue) })
		Convey(`Then Subscribing should be allowed`, func() { So(auth.CanSubscribeTo(""), ShouldBeTrue) })
	})
}
