package service

import (
	"im-server/grpc/proto_service"
	"im-server/model"
	errorServer "im-server/server/error"
)

// 查询聊天消息
func QueryMessageList(req *proto_service.QueryMessageReq) (res []*proto_service.IMMessage, err error) {
	if req.StartTime == 0 || req.EndTime == 0 {
		err = errorServer.E("时间参数不能为空")
		return
	}
	uid := 0
	uid, err = model.GetUidByProjectInfo(int(req.QueryCommonRes.ProjectId), int(req.QueryCommonRes.ProjectUid))
	if err != nil {
		return
	}

	messageList := []model.Message{}
	if req.QueryCommonRes.GroupId == 0 {
		//用户

		if req.ToProjectUid == 0 || req.ToProjectId == 0 {
			err = errorServer.E("参数错误")
			return
		}
		toUid := 0
		toUid, err = model.GetUidByProjectInfo(int(req.ToProjectId), int(req.ToProjectUid))
		if err != nil {
			return
		}

		model.DB.Where("time >= ? and time <= ? and uid=? and to_uid = ?", req.StartTime, req.EndTime, uid, toUid).Find(&messageList)

	} else {
		//group
		model.DB.Where("time >= ? and time <= ? and uid=? and group_id = ?", req.StartTime, req.EndTime, uid, req.QueryCommonRes.GroupId).Find(&messageList)

	}

	for _, v := range messageList {
		res = append(res, &proto_service.IMMessage{
			Id:          int64(v.ID),
			Uid:         int64(v.Uid),
			GroupId:     int64(v.GroupId),
			ToUid:       int64(v.ToUid),
			MessageType: int64(v.MessageType),
			Message:     v.Message,
			Time:        int64(v.Time),
		})
	}

	return
}
