package dscinitialization_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDscinitialization(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dscinitialization Suite")
}
