package epoll_server

import (
	"encoding/json"
	"im-server/model"
	"im-server/server/error"
	"im-server/util"
	"time"
)

/*
*
用户上线：
1.修改online字段
*/
func UserIMOnline(request IMMessage) model.User {

	tx := model.DB.Begin()
	userInfo := model.User{}
	tx.Where(map[string]interface{}{"project_id": request.ProjectId, "project_uid": request.ProjectUid}).Find(&userInfo)
	//上线
	userInfo.IsOnline = 1
	if err := tx.Save(&userInfo).Error; err != nil {
		util.LogPrintln("上线更新失败")
	}
	tx.Commit()
	return userInfo
}

// 用户离线
func UserIMOffline(user model.User) {
	userQuery := model.User{}
	model.DB.Where(map[string]interface{}{"id": user.ID}).Find(&userQuery)
	userQuery.IsOnline = 0
	userQuery.LastOnlineTime = int(time.Now().Unix())
	if err := model.DB.Save(userQuery).Error; err != nil {
		//记录严重日志，数据没统一
	}
}

// 保存消息记录
func UserIMSaveMessage(request IMMessage, isSend bool) error {

	sendUid, _ := model.GetUidByProjectInfo(request.ProjectId, request.ProjectUid)
	sendToUid, _ := model.GetUidByProjectInfo(request.ToProjectId, request.ToProjectUid)

	messageModel := model.Message{
		Uid:         sendUid,
		GroupId:     request.GroupId,
		ToUid:       sendToUid,
		Message:     request.Data,
		MessageType: request.Type,
		Time:        request.Time,
		IsSend:      util.BoolToInt(isSend),
	}

	util.LogPrintln("插入消息数据：", messageModel)
	if err := model.DB.Create(&messageModel).Error; err != nil {
		return errorServer.E("UserIMSaveMessageError")
	}

	return nil
}

// 发送离线消息
func UserIMSendOfflineMessage(request IMMessage, e *event) error {

	curUid, _ := model.GetUidByProjectInfo(request.ProjectId, request.ProjectUid)

	offLineDbMessage := []model.Message{}
	model.DB.Where(map[string]interface{}{"to_uid": curUid, "is_send": 0}).Find(&offLineDbMessage)

	for _, v := range offLineDbMessage {
		sendProjectId, sendProjectUid, err := model.GetProjectByUidInfo(v.Uid)
		if err == nil {
			imMessage := IMMessage{
				ProjectId:    sendProjectId,
				ProjectUid:   sendProjectUid,
				Time:         v.Time,
				Data:         v.Message,
				Type:         v.MessageType,
				ToProjectId:  request.ToProjectId,
				ToProjectUid: request.ToProjectUid,
			}
			//发送离线消息
			responseJson, _ := json.Marshal(imMessage)
			e.Send([]byte(responseJson))
			//更新状态
			v.IsSend = 1
			if err = model.DB.Save(&v).Error; err != nil {
				return err
			}
		}

	}
	return nil
}
