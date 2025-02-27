package model

// UserGroup 用户模型
type UserGroup struct {
	BaseModel
	Uid     int
	GroupId int
}

func (UserGroup) TableName() string {
	return "user_group_relation"
}
