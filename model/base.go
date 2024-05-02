package model

type Base struct {
	Id         uint64 `json:"id,omitempty" db:"id"`
	CreateTime int64  `json:"create_time" db:"create_time"`
	UpdateTime int64  `json:"update_time" db:"update_time"`
}

func (t *Base) IsEmpty() bool {
	return t.Id == 0
}
