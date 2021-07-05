package app

import (
	"testing"

	"github.com/onsi/ginkgo/reporters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestJWTCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("jwt-app.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "app", []Reporter{junitReporter})
}
