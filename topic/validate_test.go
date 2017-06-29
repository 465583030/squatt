package topic

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidate(t *testing.T) {
	Convey(`Testing the Topic validation`, t, func() {
		So(Validate("", true), ShouldEqual, errInvalidLength) // All Topic Names and Topic Filters MUST be at least one character long [MQTT-4.7.3-1]

		So(Validate("\U00000000", true), ShouldEqual, errInvalidUTF8) // Topic Names and Topic Filters MUST NOT include the null character (Unicode U+0000) [Unicode] [MQTT-4.7.3-2]

		So(Validate("#", true), ShouldBeNil)
		So(Validate("+", true), ShouldBeNil)

		So(Validate("foo", true), ShouldBeNil)
		So(Validate("foo/bar", true), ShouldBeNil)

		So(Validate("foo/#", true), ShouldBeNil)
		So(Validate("foo/+", true), ShouldBeNil)
		So(Validate("foo/+/bar", true), ShouldBeNil)

		So(Validate("foo/#", false), ShouldEqual, errWildcardNotAllowed)
		So(Validate("foo/+", false), ShouldEqual, errWildcardNotAllowed)
		So(Validate("foo/+/bar", false), ShouldEqual, errWildcardNotAllowed)

		So(Validate("foo/#bar", true), ShouldEqual, errInvalidWildcardLocation)
		So(Validate("foo/+bar", true), ShouldEqual, errInvalidWildcardLocation)
		So(Validate("foo/#/bar", true), ShouldEqual, errInvalidWildcardLocation)
	})
}
