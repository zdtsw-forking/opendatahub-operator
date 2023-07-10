package modelmeshserving_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestModelmeshserving(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Modelmeshserving Suite")
}
