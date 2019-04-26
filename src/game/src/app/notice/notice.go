package notice

import "gen/api/pt"

const (
	BUILDING_NOT_UNLOCK       = "error_building_not_unlock"
	NOT_ENOUGH_RESOURCE       = "error_reach_max_building_amount"
	REACH_MAX_BUILDING_AMOUNT = "error_reach_max_building_amount"
	BUILDING_QUEUE_FULL       = "error_building_queue_full"
	BUILDING_NOT_EXIST        = "error_building_not_exist"
	BUILDING_REACH_MAX_LEVEL  = "error_reach_max_level"
	NOT_ENOUGH_GEMS           = "error_not_enough_gems"
)

func Fail(notice string) (string, *pt.Fail) {
	return pt.PT_Fail, &pt.Fail{Fail: notice}
}

func OK() (string, *pt.Ok) {
	return pt.PT_Ok, &pt.Ok{OK: true}
}
