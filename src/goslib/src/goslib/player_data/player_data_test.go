package player_data

import (
	"gen/db"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("World", func() {
	It("should startup", func() {
		content, err := Compress(&db.PlayerData{})
		Expect(err).To(BeNil())

		playerData, err := Decompress(content)
		Expect(err).To(BeNil())
		Expect(playerData).Should(Equal(&db.PlayerData{}))
	})
})
