package services_test

import (
	"testing"

	."github.com/onsi/gomega"
	."github.com/onsi/ginkgo"
)


func TestServiceSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}


