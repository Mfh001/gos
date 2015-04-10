package leaderboard

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	. "gslib"
	"gslib/gen_server"
	"math"
	"net"
)

func Start() {
	addr := "tcp"
	port := "6379"
	gen_server.Start(LEADERBOARD_SERVER_ID, leaderboard, addr, port)
}

func MemberDataKey(name string) string {
	return fmt.Sprintf("%s:member_data", name)
}

func DeleteLeaderboard(name string) {
	gen_server.Call(LEADERBOARD_SERVER_ID, "deleteLeaderboard", name)
}

func MemberDataFor(name, member string) interface{} {
	return gen_server.Call(LEADERBOARD_SERVER_ID, "memberDataFor", name, member)
}

func (l *Leaderboard) DeleteLeaderboard(name string) {
	l.Conn.Send("MULTI")
	l.Conn.Send("DEL", name)
	l.Conn.Send("DEL", MemberDataKey(name))
	l.Conn.Send("EXEC")
}

func (l *Leaderboard) MemberDataFor(name, member string) (string, error) {
	return redis.String(l.Conn.Do("HGET", MemberDataKey(name), member))
}

func (l *Leaderboard) TotalMembers(name string) (int64, error) {
	return redis.Int64(l.Conn.Do("ZCARD", name))
}

func (l *Leaderboard) TotalPages(name string, pagesize int) (int64, error) {
	count, err := l.TotalMembers(name)
	if err != nil {
		return 0, err
	} else {
		return math.Ceil(l.TotalMembers(name) / float32(pagesize)), nil
	}
}

func (l *Leaderboard) TotalMembersInScoreRange(name string, minScore int, maxScore int) (int64, error) {
	return redis.Int64(l.Conn.Do("ZCOUNT", name, minScore, maxScore))
}

func (l *Leaderboard) ChangeScoreFor(name string, member string, delta int) (int64, error) {
	return redis.Int64(l.Conn.Do("ZINCRBY", name, delta, member))
}

func (l *Leaderboard) RangeFor(name string, member string) (int64, error) {
	return redis.Int64(l.Conn.Do("ZREVRANK", name, member))
}

func (l *Leaderboard) ScoreFor(name string, member string) (int64, error) {
	return redis.Int64(l.Conn.Do("ZSCORE", name, member))
}

func (l *Leaderboard) IsMemberRanked(name string, member string) (bool, error) {
	r, err := l.Conn.Do("ZSCORE", name, member)
	if err != nil {
		return r, err
	} else {
		return r.(type) == nil, nil
	}
}

func (l *Leaderboard) ScoreAndRankFor(name string, member string) (int64, int64, error) {
	l.Conn.Send("MULTI")
	l.Conn.Send("ZSCORE", name, member)
	l.Conn.Send("ZREVRANK", name, member)
	r, err := redis.Values(l.Conn.Do("EXEC"))
	if err != nil {
		return 0, 0, err
	}
	var score int
	var rank int
	if _, err := redis.Scan(r, &score, &rank); err != nil {
		return 0, 0, err
	} else {
		return score, rank + 1, nil
	}
}

func (l *Leaderboard) PageFor(name string, pagesize int, member string) (int, error) {
	r, err := l.Conn.Do("ZREVRANK", name, member)
	if err != nil {
		return 0, err
	} else if r == nil {
		return 0, nil
	} else {
		if rank, err := redis.Int(r); err != nil {
			return rank, err
		} else {
			return math.Ceil((rank + 1) / pagesize), nil
		}
	}

}

func (l *Leaderboard) ExpireLeaderboard(name string, seconds int) error {
	l.Conn.Send("MULTI")
	l.Conn.Send("EXPIRE", name, seconds)
	l.Conn.Send("EXPIRE", MemberDataKey(name), seconds)
	r, err := l.Conn.Do("EXEC")
	return err
}

func (l *Leaderboard) ExpireLeaderboardAt(name string, seconds int) error {
	l.Conn.Send("MULTI")
	l.Conn.Send("EXPIREAT", name, seconds)
	l.Conn.Send("EXPIREAT", MemberDataKey(name), seconds)
	r, err := l.Conn.Do("EXEC")
	return err
}

func (l *Leaderboard) Members(name string, pagesize int, currentPage int) ([]*LeaderboardData, error) {
	totalPage, err := l.TotalPages(name, pagesize)
	if err != nil {
		return nil, err
	}
	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPage {
		currentPage = totalPage
	}
	startOffset := (currentPage - 1) * pagesize
	endOffset := startOffset + pagesize - 1
	datas, err1 := redis.Values(l.Conn.Do("ZREVRANGE", name, startOffset, endOffset))
	if err1 != nil {
		return nil, err
	}
	return rankedInList(name, datas)
}

func (l *Leaderboard) AllMembers(name string) ([]*LeaderboardData, error) {
	datas, err := redis.Values(l.Conn.Do("ZREVRANGE", name, 0, -1))
	if err != nil {
		return nil, err
	}
	return rankedInList(name, datas)
}

func (l *Leaderboard) MembersFromScoreRange(name string, minScore int, maxScore int) ([]*LeaderboardData, error) {
	datas, err := redis.Values(l.Conn.Do("ZREVRANGEBYSCORE", maxScore, minScore))
	if err != nil {
		return nil, err
	}
	return rankedInList(name, datas)
}

func (l *Leaderboard) MembersFromRankRange(name string, minRank int, maxRank int) ([]*LeaderboardData, error) {
	if minRank < 1 {
		minRank = 1
	}
	totalMember, err = l.TotalMembers(name)
	if err != nil {
		return nil, err
	}
	if maxRank > totalMember {
		maxRank = totalMember
	}
	datas, err1 := redis.Values(l.Conn.Do("ZREVRANGEBYRANK", maxRank, minRank))
	if err != nil {
		return nil, err
	}
	return rankedInList(name, datas)
}

func (l *Leaderboard) TopMember(name string) (*LeaderboardData, error) {
	topMember, err := l.MembersFromRankRange(name, 1, 1)
	if err != nil {
		return nil, err
	}
	if len(topMember) == 0 {
		return nil, err
	}
	return topMember[0], nil
}

func (l *Leaderboard) MemberAt(name string, pagesize int, position int) (*LeaderboardData, error) {
	totalMember, err = l.TotalMembers(name)
	if err != nil {
		return nil, err
	}
	if position <= totalMember {
		currentPage = math.Ceil(position / pagesize)
		members, err = l.Members(name, pagesize, currentPage)
		if len(members) == 0 {
			return nil, nil
		} else {
			return members[0], nil
		}
	} else {
		return nil, nil
	}
}

func (l *Leaderboard) AroundMe(name string, pagesize int, member string) ([]*LeaderboardData, error) {
	rank, err := redis.Int(l.Conn.Do("ZREVRANK", name, member))
	if err != nil {
		return 0, err
	}
	startOffset := rank - math.Ceil(pagesize/2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset = startOffset + pagesize - 1
	datas, err1 := redis.Values(l.Conn.Do("ZREVRANGE", startOffset, endOffset))
	if err1 != nil {
		return nil, err1
	}
	return rankedInList(datas)
}

func rankedInList(name string, datas []interface{}) ([]*LeaderboardData, error) {
	var members []*LeaderboardData
	for i, data := range datas {
		key := data.(string)
		score, rank, err := l.ScoreAndRankFor(name, key)
		memberData, err := l.MemberDataFor(name, key)
		if err != nil {
			return nil, err
		}
		members = append(members, &LeaderboardData{key, score, rank, memberData})
	}
	return members
}

type Leaderboard struct {
	Conn net.Conn
}

func (l *Leaderboard) Init(args []interface{}) (err error) {
	addr := args[0].(string)
	port := args[1].(string)
	conn, err := redis.Dial(addr, port)
	if err != nil {
		panic(err)
	}
	l.Conn = conn
}

func (l *Leaderboard) HandleCall(args []interface{}) interface{} {
}

func (l *Leaderboard) HandleCast(args []interface{}) {
}

func (l *Leaderboard) Terminate() {
	l.Conn.Close()
}
