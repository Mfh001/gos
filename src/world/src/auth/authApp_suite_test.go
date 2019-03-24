package auth

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuthApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AuthApp Suite")
}
