package workbenches_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkbenches(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workbenches Suite")
}
