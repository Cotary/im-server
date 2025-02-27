package service

import (
	"im-server/grpc/proto_service"
	"im-server/model"
	"im-server/server/error"
	"im-server/util"
)

func UserEdit(userRpc *proto_service.User) error {

	tx := model.DB.Begin()
	defer tx.Rollback()
	//check ,这里有可能是新增
	//if _, err := model.GetUidByProjectInfo(int(userRpc.ProjectId), int(userRpc.ProjectUid)); err != nil {
	//	return errorServer.E("查询不到信息")
	//}
	if len(userRpc.Name) == 0 {
		return errorServer.E("名称不能为空")
	}

	formatUserModel := model.User{
		Name:       userRpc.Name,
		Avatar:     userRpc.Avatar,
		ProjectId:  int(userRpc.ProjectId),
		ProjectUid: int(userRpc.ProjectUid),
		//todo 使用pwd来鉴权
	}

	dbUser := &model.User{}
	dbUserCount := 0
	tx.Where(
		model.User{
			ProjectId:  formatUserModel.ProjectId,
			ProjectUid: formatUserModel.ProjectUid,
		}).First(dbUser).Count(&dbUserCount)

	if dbUserCount > 0 {
		//编辑
		dbUser.Name = formatUserModel.Name
		dbUser.Avatar = formatUserModel.Avatar
		util.LogPrintln("编辑", dbUser)
		if err := tx.Save(&dbUser).Error; err != nil {
			return errorServer.E("更新失败")
		}
	} else {
		//新增
		util.LogPrintln("新增", formatUserModel)
		if err := tx.Create(&formatUserModel).Error; err != nil {
			return errorServer.E("注册失败")
		}
	}
	tx.Commit()
	return nil

}

func UserDel(userRpc *proto_service.User) error {

	//check
	if _, err := model.GetUidByProjectInfo(int(userRpc.ProjectId), int(userRpc.ProjectUid)); err != nil {
		return errorServer.E("查询不到信息")
	}

	tx := model.DB.Begin()
	formatUserModel := model.User{
		ProjectId:  int(userRpc.ProjectId),
		ProjectUid: int(userRpc.ProjectUid),
	}
	dbUser := model.User{}
	dbUserCount := 0
	tx.Where(model.User{ProjectId: formatUserModel.ProjectId, ProjectUid: formatUserModel.ProjectUid}).Find(&dbUser).Count(&dbUserCount)
	if dbUserCount > 0 {
		//删除
		tx.Delete(&dbUser)
	}
	tx.Commit()
	return nil

}

func QueryUserList(req *proto_service.QueryCommonReq) []*proto_service.IMUser {

	//check
	if _, err := model.GetUidByProjectInfo(int(req.ProjectId), int(req.ProjectUid)); err != nil {
		return nil
	}

	tx := model.DB.Begin()
	userList := []model.User{}
	if req.GroupId == 0 {
		tx.Where(model.User{ProjectId: int(req.ProjectId)}).Find(&userList)

	} else {
		//todo 尝试下left join
		groupUserList := []model.UserGroup{}
		tx.Where(model.UserGroup{GroupId: int(req.GroupId)}).Find(&groupUserList)
		userIds, _ := util.ArrayColumn(groupUserList, "Uid")
		tx.Where(map[string]interface{}{"project_id": req.ProjectId}).Where("id in (?)", userIds).Find(&userList)

	}
	IMUsers := []*proto_service.IMUser{}
	for _, v := range userList {
		IMUsers = append(IMUsers, &proto_service.IMUser{
			Id:             int64(v.ID),
			ProjectId:      int64(v.ProjectId),
			ProjectUid:     int64(v.ProjectUid),
			Name:           v.Name,
			IsOnline:       int64(v.IsOnline),
			Avatar:         v.Avatar,
			Pwd:            v.Pwd,
			LastOnlineTime: int64(v.LastOnlineTime),
		})
	}
	tx.Commit()
	return IMUsers

}
