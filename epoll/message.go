package epoll_server

import (
	"encoding/json"
	"im-server/model"
	"im-server/server/error"
	"im-server/util"
	"time"
)

const (
	NormalMsg    = 1 //普通消息
	ReceiptMsg   = 2 //回执消息
	HeartbeatMsg = 3 //心跳消息
	ErrorMsg     = 4 //错误消息
)

type IMMessage struct {
	ProjectId  int `json:"project_id"`
	ProjectUid int `json:"project_uid"`
	// Uid int  //为0是服务器发送的
	MessageId    string `json:"message_id"`     //雪花算法，唯一id
	Time         int    `json:"time,omitempty"` //时间戳
	Data         string `json:"data,omitempty"` //数据
	Type         int    `json:"type,omitempty"` //1普通消息2回执消息,表示已经收到 3心跳[ProjectId  ProjectUid]4.错误提示
	ToProjectId  int    `json:"to_project_id,omitempty"`
	ToProjectUid int    `json:"to_project_uid,omitempty"`
	GroupId      int    `json:"group_id,omitempty"`
}

func PreHandleMessage(data []byte) (IMMessage, error) {
	request := IMMessage{}

	jsonDecodeErr := json.Unmarshal(data, &request)
	if len(data) == 0 || jsonDecodeErr != nil {
		return request, errorServer.E("解析失败:" + jsonDecodeErr.Error())
	}

	//必须
	if request.ProjectUid < 0 || request.ProjectId < 0 {
		return request, errorServer.E("缺少必要身份信息参数")
	}

	_, err := model.GetUidByProjectInfo(int(request.ProjectId), int(request.ProjectUid))
	if err != nil {
		return request, errorServer.E("身份信息验证失败")
	}

	if request.Type == NormalMsg {
		if request.GroupId == 0 {
			_, err = model.GetUidByProjectInfo(int(request.ToProjectId), int(request.ToProjectUid))
			if err != nil {
				return request, errorServer.E("接收身份信息验证失败")
			}
		} else {
			//todo 校验group_id 是否正确
		}
	}

	//补充时间
	request.Time = int(time.Now().Unix())

	return request, nil
}

func HandleBusiness(e *event, content []byte) {
	//更新最后登录时间
	e.lastTime = int(time.Now().Unix())

	//读取消息

	//预处理数据 必须符合requestMessage 格式
	//content[:] 注意这个地方要这么写，不然解析不出来,末尾会多个结束符
	requestMessage, requestMessageErr := PreHandleMessage(content)
	if requestMessageErr != nil {
		e.SendErrorMsg(requestMessageErr.Error())
		return
	}

	//上线操作
	if e.userInfo == (model.User{}) {
		userInfo := UserIMOnline(requestMessage)
		e.userInfo = userInfo

		util.LogPrintln(e.userInfo.Name, " 上线成功!")
		//上线成功后发送离线消息
		offlineErr := UserIMSendOfflineMessage(requestMessage, e)
		if offlineErr != nil {
			util.LogPrintln("发送离线消息失败：", offlineErr.Error())
		}
	} else {
		//已在线，处理心跳
		HandleHeartbeat(requestMessage)
	}

	//读取消息
	if e.userInfo != (model.User{}) && requestMessage.Type == NormalMsg {

		//发送消息socket
		isSend := false
		ToConn := e.GetUserConn(requestMessage.ToProjectId, requestMessage.ToProjectUid)
		if ToConn != nil {
			util.LogPrintln(e.userInfo.Name, " 给 ", ToConn.userInfo.Name, " 发送消息：", requestMessage.Data)

			ToConn.SendMsg(requestMessage)
			isSend = true
		}

		//存下消息
		err := UserIMSaveMessage(requestMessage, isSend)
		if err != nil {
			util.LogPrintln("消息存储失败")
		}

	}

	//回执消息
	e.SendReceiptMsg(requestMessage)
}

// 处理心跳消息
func HandleHeartbeat(msg IMMessage) {
	UserIMOnline(msg)
}
