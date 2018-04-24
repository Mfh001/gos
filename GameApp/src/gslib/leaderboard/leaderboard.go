package leaderboard

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	. "gslib"
	"goslib/gen_server"
	"math"
)

type Leaderboard struct {
	Conn redis.Conn
}

func Start() {
	addr := "tcp"
	port := ":6379"
	gen_server.Start(LEADERBOARD_SERVER_ID, new(Leaderboard), addr, port)
}

func MemberDataKey(name string) string {
	return fmt.Sprintf("%s:member_data", name)
}

func DeleteLeaderboard(name string) {
	gen_server.Call(LEADERBOARD_SERVER_ID, "deleteLeaderboard", name)
}

func MemberDataFor(name, member string) (interface{}, error) {
	return gen_server.Call(LEADERBOARD_SERVER_ID, "memberDataFor", name, member)
}

func (l *Leaderboard) HandleCall(args []interface{}) interface{} {
	return handleCallAndCast(args)
}

func (l *Leaderboard) HandleCast(args []interface{}) {
	handleCallAndCast(args)
}

func handleCallAndCast(args []interface{}) error {
	//method := args[0].(string)
	//if method == "add" {
	//	return t.DoAdd(args[1].(string), args[2].(int), args[3].(func()))
	//} else if method == "update" {
	//	return t.DoUpdate(args[1].(string), args[2].(int))
	//} else if method == "finish" {
	//	return t.DoFinish(args[1].(string))
	//} else if method == "del" {
	//	return t.DoDel(args[1].(string))
	//}
	return nil
}

// Private Methods
func (l *Leaderboard) DeleteLeaderboard(name string) {
	l.Conn.Send("MULTI")
	l.Conn.Send("DEL", name)
	l.Conn.Send("DEL", MemberDataKey(name))
	l.Conn.Send("EXEC")
}

func (l *Leaderboard) MemberDataFor(name, member string) (string, error) {
	return redis.String(l.Conn.Do("HGET", MemberDataKey(name), member))
}

func (l *Leaderboard) TotalMembers(name string) (int, error) {
	return redis.Int(l.Conn.Do("ZCARD", name))
}

func (l *Leaderboard) TotalPages(name string, pagesize int) (int, error) {
	count, err := l.TotalMembers(name)
	if err != nil {
		return 0, err
	} else {
		return int(math.Ceil(float64(count) / float64(pagesize))), nil
	}
}

func (l *Leaderboard) TotalMembersInScoreRange(name string, minScore int, maxScore int) (int, error) {
	return redis.Int(l.Conn.Do("ZCOUNT", name, minScore, maxScore))
}

func (l *Leaderboard) ChangeScoreFor(name string, member string, delta int) (int, error) {
	return redis.Int(l.Conn.Do("ZINCRBY", name, delta, member))
}

func (l *Leaderboard) RangeFor(name string, member string) (int, error) {
	return redis.Int(l.Conn.Do("ZREVRANK", name, member))
}

func (l *Leaderboard) ScoreFor(name string, member string) (int, error) {
	return redis.Int(l.Conn.Do("ZSCORE", name, member))
}

func (l *Leaderboard) IsMemberRanked(name string, member string) (bool, error) {
	r, err := l.Conn.Do("ZSCORE", name, member)
	if err != nil {
		return false, err
	} else {
		switch r.(type) {
		case nil:
			return false, nil
		default:
			return true, nil
		}
	}
}

func (l *Leaderboard) ScoreAndRankFor(name string, member string) (int, int, error) {
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
	rank, err := redis.Int(l.Conn.Do("ZREVRANK", name, member))
	if err != nil {
		return 0, err
	} else {
		return int(math.Ceil((float64(rank) + 1) / float64(pagesize))), nil
	}
}

func (l *Leaderboard) ExpireLeaderboard(name string, seconds int) error {
	l.Conn.Send("MULTI")
	l.Conn.Send("EXPIRE", name, seconds)
	l.Conn.Send("EXPIRE", MemberDataKey(name), seconds)
	_, err := l.Conn.Do("EXEC")
	return err
}

func (l *Leaderboard) ExpireLeaderboardAt(name string, seconds int) error {
	l.Conn.Send("MULTI")
	l.Conn.Send("EXPIREAT", name, seconds)
	l.Conn.Send("EXPIREAT", MemberDataKey(name), seconds)
	_, err := l.Conn.Do("EXEC")
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
	return l.rankedInList(name, datas)
}

func (l *Leaderboard) AllMembers(name string) ([]*LeaderboardData, error) {
	datas, err := redis.Values(l.Conn.Do("ZREVRANGE", name, 0, -1))
	if err != nil {
		return nil, err
	}
	return l.rankedInList(name, datas)
}

func (l *Leaderboard) MembersFromScoreRange(name string, minScore int, maxScore int) ([]*LeaderboardData, error) {
	datas, err := redis.Values(l.Conn.Do("ZREVRANGEBYSCORE", maxScore, minScore))
	if err != nil {
		return nil, err
	}
	return l.rankedInList(name, datas)
}

func (l *Leaderboard) MembersFromRankRange(name string, minRank int, maxRank int) ([]*LeaderboardData, error) {
	if minRank < 1 {
		minRank = 1
	}
	totalMember, err := l.TotalMembers(name)
	if err != nil {
		return nil, err
	}
	if maxRank > totalMember {
		maxRank = totalMember
	}
	datas, err := redis.Values(l.Conn.Do("ZREVRANGEBYRANK", maxRank, minRank))
	if err != nil {
		return nil, err
	}
	return l.rankedInList(name, datas)
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
	totalMember, err := l.TotalMembers(name)
	if err != nil {
		return nil, err
	}
	if position <= totalMember {
		currentPage := int(math.Ceil(float64(position) / float64(pagesize)))
		members, _ := l.Members(name, pagesize, currentPage)
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
		return nil, err
	}
	startOffset := rank - int(math.Ceil(float64(pagesize/2)))
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := startOffset + pagesize - 1
	datas, err1 := redis.Values(l.Conn.Do("ZREVRANGE", startOffset, endOffset))
	if err1 != nil {
		return nil, err1
	}
	return l.rankedInList(name, datas)
}

func (l *Leaderboard) rankedInList(name string, datas []interface{}) ([]*LeaderboardData, error) {
	var members []*LeaderboardData
	for _, data := range datas {
		key := data.(string)
		score, rank, err := l.ScoreAndRankFor(name, key)
		memberData, err := l.MemberDataFor(name, key)
		if err != nil {
			return nil, err
		}
		members = append(members, &LeaderboardData{key, score, rank, memberData})
	}
	return members, nil
}

func (l *Leaderboard) Init(args []interface{}) (err error) {
	addr := args[0].(string)
	port := args[1].(string)
	conn, err := redis.Dial(addr, port)
	if err != nil {
		panic(err)
	}
	l.Conn = conn
	return err
}

func (l *Leaderboard) Terminate(reason string) (err error) {
	l.Conn.Close()
	return nil
}
