package pipelinerun

import (
	"testing"

	"github.com/onsi/ginkgo/reporters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPipelineRun(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("pipelinerun-test.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "test PipelineRun controller", []Reporter{junitReporter})
}
