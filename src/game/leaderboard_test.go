/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Game", func() {
	//It("Leaderboard should workd", func() {
	//	name := "__fake_leaderboard"
	//	memberId := "fake_member_id"
	//	memberScore := int64(10)
	//	memberData := map[string]string{
	//		"name": "fake_member",
	//		"age":  "13",
	//	}
	//
	//	leaderboard.Start(name)
	//	leaderboard.Delete(name)
	//
	//	leaderboard.RankMember(name, &leaderboard.Member{
	//		Id:    memberId,
	//		Score: memberScore,
	//		Data:  memberData,
	//	})
	//
	//	members := make([]*leaderboard.Member, 2)
	//	members[0] = &leaderboard.Member{
	//		Id:    "fake_1",
	//		Score: 11,
	//		Data: map[string]string{
	//			"name": "fake_1",
	//			"age":  "14",
	//		},
	//	}
	//	members[1] = &leaderboard.Member{
	//		Id:    "fake_2",
	//		Score: 12,
	//		Data: map[string]string{
	//			"name": "fake_2",
	//			"age":  "15",
	//		},
	//	}
	//	leaderboard.RankMembers(name, members)
	//
	//	count, err := leaderboard.TotalMembers(name)
	//	Expect(err).Should(BeNil())
	//	Expect(count).Should(Equal(3))
	//
	//	pages, err := leaderboard.TotalPages(name, 10)
	//	Expect(err).Should(BeNil())
	//	Expect(pages).Should(Equal(1))
	//
	//	member, err := leaderboard.MemberFor(name, memberId)
	//	Expect(err).Should(BeNil())
	//	Expect(member.Data["name"]).Should(Equal(memberData["name"]))
	//
	//	leaderboard.ChangeScoreFor(name, memberId, 100)
	//	leaderboard.UpdateMemberData(name, memberId, map[string]string{
	//		"name": "changed_fake_member",
	//	})
	//	member, err = leaderboard.MemberFor(name, memberId)
	//	Expect(err).Should(BeNil())
	//	Expect(member.Score).Should(Equal(int64(100)))
	//	Expect(member.Data["name"]).Should(Equal("changed_fake_member"))
	//
	//	data, err := leaderboard.MemberDataFor(name, memberId)
	//	Expect(err).Should(BeNil())
	//	Expect(data["name"]).Should(Equal("changed_fake_member"))
	//
	//	rank, err := leaderboard.RankFor(name, memberId)
	//	Expect(err).Should(BeNil())
	//	Expect(rank).Should(Equal(int64(1)))
	//
	//	score, err := leaderboard.ScoreFor(name, memberId)
	//	Expect(err).Should(BeNil())
	//	Expect(score).Should(Equal(int64(100)))
	//
	//	rank, score, err = leaderboard.RankAndScoreFor(name, memberId)
	//	Expect(err).Should(BeNil())
	//	Expect(rank).Should(Equal(int64(1)))
	//	Expect(score).Should(Equal(int64(100)))
	//
	//	members, err = leaderboard.MembersAroundMe(name, memberId, 10)
	//	Expect(err).Should(BeNil())
	//	Expect(len(members)).Should(Equal(3))
	//
	//	members, err = leaderboard.MembersInPage(name, 1, 10)
	//	Expect(err).Should(BeNil())
	//	Expect(len(members)).Should(Equal(3))
	//
	//	leaderboard.RemoveMember(name, memberId)
	//	count, err = leaderboard.TotalMembers(name)
	//	Expect(err).Should(BeNil())
	//	Expect(count).Should(Equal(2))
	//
	//	leaderboard.RemoveMembers(name, []string{"fake_1", "fake_2"})
	//	count, err = leaderboard.TotalMembers(name)
	//	Expect(err).Should(BeNil())
	//	Expect(count).Should(Equal(0))
	//})
})
