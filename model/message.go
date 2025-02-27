package model

// Message 用户模型
type Message struct {
	BaseModel
	Uid         int
	GroupId     int
	ToUid       int
	IsRead      int
	IsSend      int
	MessageType int
	Message     string `gorm:"size:1000"`
	Time        int
}

func (Message) TableName() string {
	return "message"
}
