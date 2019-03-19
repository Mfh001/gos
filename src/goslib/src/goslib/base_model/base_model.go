package base_model

type ModelInterface interface {
	GetUuid() string
	GetTableName() string
	Save()
	Delete()
	SqlForRec(status int8) string
}

/*
 * Memory data status
 */
const (
	STATUS_EMPTY  = 0
	STATUS_ORIGIN = 1
	STATUS_CREATE = 2
	STATUS_UPDATE = 3
	STATUS_DELETE = 4
)
