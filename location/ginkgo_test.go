package location_test

import (
	"reflect"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestSuite(t *testing.T) {
	type tag struct{}
	format.MaxLength = 0
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, reflect.TypeOf(tag{}).PkgPath())
}
