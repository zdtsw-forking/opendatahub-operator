package datasciencecluster_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatasciencecluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datasciencecluster Suite")
}
