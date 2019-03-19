package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"goslib/leaderboard"
)

var _ = Describe("Game", func() {
	It("Leaderboard should workd", func() {
		name := "__fake_leaderboard"
		memberId := "fake_member_id"
		memberScore := int64(10)
		memberData := map[string]string{
			"name": "fake_member",
			"age":  "13",
		}

		leaderboard.Start(name)
		leaderboard.Delete(name)

		leaderboard.RankMember(name, &leaderboard.Member{
			Id:    memberId,
			Score: memberScore,
			Data:  memberData,
		})

		members := make([]*leaderboard.Member, 2)
		members[0] = &leaderboard.Member{
			Id:    "fake_1",
			Score: 11,
			Data: map[string]string{
				"name": "fake_1",
				"age":  "14",
			},
		}
		members[1] = &leaderboard.Member{
			Id:    "fake_2",
			Score: 12,
			Data: map[string]string{
				"name": "fake_2",
				"age":  "15",
			},
		}
		leaderboard.RankMembers(name, members)

		count, err := leaderboard.TotalMembers(name)
		Expect(err).Should(BeNil())
		Expect(count).Should(Equal(3))

		pages, err := leaderboard.TotalPages(name, 10)
		Expect(err).Should(BeNil())
		Expect(pages).Should(Equal(1))

		member, err := leaderboard.MemberFor(name, memberId)
		Expect(err).Should(BeNil())
		Expect(member.Data["name"]).Should(Equal(memberData["name"]))

		leaderboard.ChangeScoreFor(name, memberId, 100)
		leaderboard.UpdateMemberData(name, memberId, map[string]string{
			"name": "changed_fake_member",
		})
		member, err = leaderboard.MemberFor(name, memberId)
		Expect(err).Should(BeNil())
		Expect(member.Score).Should(Equal(int64(100)))
		Expect(member.Data["name"]).Should(Equal("changed_fake_member"))

		data, err := leaderboard.MemberDataFor(name, memberId)
		Expect(err).Should(BeNil())
		Expect(data["name"]).Should(Equal("changed_fake_member"))

		rank, err := leaderboard.RankFor(name, memberId)
		Expect(err).Should(BeNil())
		Expect(rank).Should(Equal(int64(1)))

		score, err := leaderboard.ScoreFor(name, memberId)
		Expect(err).Should(BeNil())
		Expect(score).Should(Equal(int64(100)))

		rank, score, err = leaderboard.RankAndScoreFor(name, memberId)
		Expect(err).Should(BeNil())
		Expect(rank).Should(Equal(int64(1)))
		Expect(score).Should(Equal(int64(100)))

		members, err = leaderboard.MembersAroundMe(name, memberId, 10)
		Expect(err).Should(BeNil())
		Expect(len(members)).Should(Equal(3))

		members, err = leaderboard.MembersInPage(name, 1, 10)
		Expect(err).Should(BeNil())
		Expect(len(members)).Should(Equal(3))

		leaderboard.RemoveMember(name, memberId)
		count, err = leaderboard.TotalMembers(name)
		Expect(err).Should(BeNil())
		Expect(count).Should(Equal(2))

		leaderboard.RemoveMembers(name, []string{"fake_1", "fake_2"})
		count, err = leaderboard.TotalMembers(name)
		Expect(err).Should(BeNil())
		Expect(count).Should(Equal(0))
	})
})
