package leaderboard

import (
	"fmt"
	"github.com/go-redis/redis"
	"goslib/gen_server"
	"goslib/logger"
	"goslib/redisdb"
	"goslib/utils"
	"math"
	"reflect"
)

/*
   GenServer Callbacks
*/
type Server struct {
	name string
}

type MemberData map[string]string

type Member struct {
	Id    string
	Rank  int64
	Score int64
	Data  MemberData
}

// Start or create leaderboard
func Start(leaderboard string) {
	gen_server.Start(leaderboard, new(Server), leaderboard)
}

// Delete leaderboard
var deleteParams = &DeleteParams{}
func Delete(leaderboard string) {
	gen_server.Cast(leaderboard, deleteParams)
}

// Total members in leaderboard
var totalMembersParams = &TotalMembersParams{}
func TotalMembers(leaderboard string) (int, error) {
	result, err := gen_server.Call(leaderboard, totalMembersParams)
	if err != nil {
		return 0, err
	}
	return int(result.(int64)), err
}

// Total pages in leaderboard
func TotalPages(leaderboard string, pageSize int) (int, error) {
	result, err := gen_server.Call(leaderboard, &TotalPagesParams{pageSize})
	if err != nil {
		return 0, err
	}
	return result.(int), err
}

// Add member to leaderboard or update member
func RankMember(leaderboard string, member *Member) {
	gen_server.Cast(leaderboard, &RankMemberParams{member})
}

// Add members to leaderboard or update members
func RankMembers(leaderboard string, members []*Member) {
	gen_server.Cast(leaderboard, &RankMembersParams{members})
}

// Del member from leaderboard
func RemoveMember(leaderboard string, memberId string) {
	gen_server.Cast(leaderboard, &RemoveMembersParams{[]string{memberId}})
}

// Del members from leaderboard
func RemoveMembers(leaderboard string, memberIds []string) {
	gen_server.Cast(leaderboard, &RemoveMembersParams{memberIds})
}

// Change member's score
func ChangeScoreFor(leaderboard string, memberId string, score int64) {
	gen_server.Cast(leaderboard, &ChangeScoreForParams{memberId, score})
}

// Get member data
func MemberDataFor(leaderboard string, memberId string) (MemberData, error) {
	result, err := gen_server.Call(leaderboard, &MemberDataForParams{memberId})
	if err != nil {
		return nil, err
	}
	return result.(MemberData), err
}

// Update member data
func UpdateMemberData(leaderboard string, memberId string, memberData MemberData) error {
	_, err := gen_server.Call(leaderboard, &UpdateMemberDataParams{
		memberId,
		memberData,
	})
	if err != nil {
		return err
	}
	return err
}

// Get member's rank
func RankFor(leaderboard string, memberId string) (int64, error) {
	result, err := gen_server.Call(leaderboard, &RankForParams{memberId})
	if err != nil {
		return 0, err
	}
	return result.(int64), err
}

// Get member's score
func ScoreFor(leaderboard string, memberId string) (int64, error) {
	result, err := gen_server.Call(leaderboard, &ScoreForParams{memberId})
	if err != nil {
		return 0, err
	}
	return result.(int64), err
}

// Get member's rank and score
func RankAndScoreFor(leaderboard string, memberId string) (int64, int64, error) {
	rankAndScore, err := gen_server.Call(leaderboard, &RankAndScoreForParams{memberId})
	if err != nil {
		return 0, 0, err
	}
	member := rankAndScore.(*Member)
	return member.Rank, member.Score, err
}

// Get member's rank, score, member data
func MemberFor(leaderboard string, memberId string) (*Member, error) {
	result, err := gen_server.Call(leaderboard, &MemberForParams{memberId})
	if err != nil {
		return nil, err
	}
	member := result.(*Member)
	return member, err
}

// Get members around me
func MembersAroundMe(leaderboard string, memberId string, pageSize int) ([]*Member, error) {
	result, err := gen_server.Call(leaderboard, &MembersAroundMeParams{
		memberId,
		pageSize,
	})
	if err != nil {
		return nil, err
	}
	members := result.([]*Member)
	return members, err
}

// Get members in page
func MembersInPage(leaderboard string, page int, pageSize int) ([]*Member, error) {
	result, err := gen_server.Call(leaderboard, &MembersInPageParams{
		page,
		pageSize,})
	if err != nil {
		return nil, err
	}
	members := result.([]*Member)
	return members, err
}

func (self *Server) Init(args []interface{}) (err error) {
	self.name = args[0].(string)
	return nil
}

func (self *Server) HandleCast(msg interface{}) {
	_, err := self.handleCallAndCast(msg)
	if err != nil {
		logger.ERR("leaderboard ", reflect.TypeOf(msg).String(), " err: ", err)
	}
}

func (self *Server) HandleCall(msg interface{}) (interface{}, error) {
	result, err := self.handleCallAndCast(msg)
	if err != nil {
		logger.ERR("leaderboard ", reflect.TypeOf(msg).String(), " err: ", err)
	}
	return result, err
}

type DeleteParams struct {}
type TotalMembersParams struct {}
type RankMembersParams struct { members []*Member }
type ChangeScoreForParams struct {
	memberId string
	score int64
}
type MemberDataForParams struct {memberId string}
type UpdateMemberDataParams struct {
	memberId string
	memberData MemberData
}
type RankForParams struct { memberId string }
type ScoreForParams struct { memberId string }
type RankAndScoreForParams struct { memberId string }
type MemberForParams struct { memberId string }
type MembersAroundMeParams struct {
	memberId string
	pageSize int
}
type MembersInPageParams struct {
	page int
	pageSize int
}

func (self *Server) handleCallAndCast(msg interface{}) (interface{}, error) {
	switch params := msg.(type) {
	case *DeleteParams:
		memberIds, err := redisdb.Instance().ZRange(self.name, 0, -1).Result()
		if err == redis.Nil {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
		return self.removeMembers(memberIds)
	case *TotalMembersParams:
		count, err := redisdb.Instance().ZCard(self.name).Result()
		if err == redis.Nil {
			return 0, nil
		}
		return count, err
	case *TotalPagesParams:
		return self.totalPage(params.pageSize)
	case *RankMemberParams:
		return self.rankMember(params.member)
	case *RankMembersParams:
		for _, member := range params.members {
			if _, err := self.rankMember(member); err != nil {
				return nil, err
			}
		}
		return len(params.members), nil
	case *RemoveMembersParams:
		return self.removeMembers(params.memberIds)
	case *ChangeScoreForParams:
		return self.rankMember(&Member{
			Id:    params.memberId,
			Score: params.score,
		})
	case *MemberDataForParams:
		data, err := self.getMemberData(params.memberId)
		if err != nil {
			return nil, err
		}
		return data, err
	case *UpdateMemberDataParams:
		return self.setMemberData(params.memberId, params.memberData)
	case *RankForParams:
		return self.getRank(params.memberId)
	case *ScoreForParams:
		return self.getScore(params.memberId)
	case *RankAndScoreForParams:
		rank, err := self.getRank(params.memberId)
		if err != nil {
			return nil, err
		}
		score, err := self.getScore(params.memberId)
		if err != nil {
			return nil, err
		}
		return &Member{
			Id:    params.memberId,
			Rank:  rank,
			Score: int64(score),
		}, nil
	case *MemberForParams:
		return self.getMember(params.memberId)
	case *MembersAroundMeParams:
		rank, err := self.getRank(params.memberId)
		if err != nil {
			return nil, err
		}
		startOffset := utils.Max(int(rank)-params.pageSize, 0)
		endOffset := startOffset + params.pageSize - 1
		memberIds, err := redisdb.Instance().ZRevRange(self.name, int64(startOffset), int64(endOffset)).Result()
		if err == redis.Nil {
			return self.getMembers([]string{})
		}
		if err != nil {
			return nil, err
		}
		return self.getMembers(memberIds)
	case *MembersInPageParams:
		totalPage, err := self.totalPage(params.pageSize)
		if err != nil {
			return nil, err
		}
		currentPage := utils.Min(utils.Max(params.page, 1), totalPage)
		indexForRedis := currentPage - 1
		startOffset := utils.Max(indexForRedis*params.pageSize, 0)
		endOffset := startOffset + params.pageSize - 1
		memberIds, err := redisdb.Instance().ZRevRange(self.name, int64(startOffset), int64(endOffset)).Result()
		if err == redis.Nil {
			return self.getMembers([]string{})
		}
		if err != nil {
			return nil, err
		}
		return self.getMembers(memberIds)
	}
	return nil, nil
}

func (self *Server) Terminate(reason string) (err error) {
	return nil
}

type RankMemberParams struct { member *Member }
func (self *Server) rankMember(member *Member) (int64, error) {
	count, err := redisdb.Instance().ZAdd(self.name, redis.Z{
		Member: member.Id,
		Score:  float64(member.Score),
	}).Result()
	if err != nil {
		return 0, err
	}
	if member.Data != nil {
		_, err := self.setMemberData(member.Id, member.Data)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

type RemoveMembersParams struct { memberIds []string }
func (self *Server) removeMembers(memberIds []string) (int64, error) {
	if len(memberIds) == 0 {
		return 0, nil
	}
	ids := make([]interface{}, len(memberIds))
	memberDataKeys := make([]string, len(memberIds))
	for i, v := range memberIds {
		memberDataKeys[i] = memberDataKey(self.name, v)
		ids[i] = v
	}
	redisdb.Instance().Del(memberDataKeys...)
	return redisdb.Instance().ZRem(self.name, ids...).Result()
}

func (self *Server) setMemberData(memberId string, data MemberData) (string, error) {
	memberData := make(map[string]interface{})
	for k, v := range data {
		memberData[k] = v
	}
	return redisdb.Instance().HMSet(memberDataKey(self.name, memberId), memberData).Result()
}

func (self *Server) getMemberData(memberId string) (MemberData, error) {
	data, err := redisdb.Instance().HGetAll(memberDataKey(self.name, memberId)).Result()
	if err == redis.Nil || len(data) == 0 {
		return data, nil
	}
	return data, err
}

func memberDataKey(leaderboard string, memberId string) string {
	return fmt.Sprintf("%s:%s", leaderboard, memberId)
}

func (self *Server) getMember(memberId string) (*Member, error) {
	rank, err := self.getRank(memberId)
	if err != nil {
		return nil, err
	}
	score, err := self.getScore(memberId)
	if err != nil {
		return nil, err
	}
	data, err := self.getMemberData(memberId)
	if err != nil {
		return nil, err
	}
	return &Member{
		Id:    memberId,
		Rank:  rank,
		Score: int64(score),
		Data:  data,
	}, nil
}

func (self *Server) getMembers(memberIds []string) ([]*Member, error) {
	members := make([]*Member, len(memberIds))
	for idx, memberId := range memberIds {
		member, err := self.getMember(memberId)
		if err != nil {
			return nil, err
		}
		members[idx] = member
	}
	return members, nil
}

type TotalPagesParams struct { pageSize int }
func (self *Server) totalPage(pageSize int) (int, error) {
	count, err := redisdb.Instance().ZCard(self.name).Result()
	if err != nil {
		return 0, err
	}
	return int(math.Ceil(float64(count) / float64(pageSize))), nil
}

func (self *Server) getRank(memberId string) (int64, error) {
	rank, err := redisdb.Instance().ZRevRank(self.name, memberId).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return rank + 1, nil
}

func (self *Server) getScore(memberId string) (int64, error) {
	score, err := redisdb.Instance().ZScore(self.name, memberId).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int64(score), nil
}
