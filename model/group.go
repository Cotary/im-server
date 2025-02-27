package model

// Group 用户模型
type Group struct {
	BaseModel
	Name      string
	CreateUid int
	Des       string
	Avatar    string `gorm:"size:1000"`
}

func (Group) TableName() string {
	return "user_group"
}
