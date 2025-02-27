package service

import (
	"im-server/grpc/proto_service"
	"im-server/model"
	"im-server/server/error"
	"im-server/util"
)

func GroupEdit(r *proto_service.Group) error {

	//check
	uid, err := model.GetUidByProjectInfo(int(r.ProjectId), int(r.ProjectUid))
	if err != nil {
		return errorServer.E("查询不到信息")
	}

	tx := model.DB.Begin()

	groupModel := model.Group{
		Name:      r.Name,
		Avatar:    r.Avatar,
		CreateUid: uid,
		Des:       r.Des,
		BaseModel: model.BaseModel{
			ID: uint(r.GroupId),
		},
	}

	if groupModel.ID > 0 {
		//修改

		dbCount := 0
		dbItem := &model.Group{}
		tx.Where(map[string]interface{}{"id": groupModel.ID}).Find(dbItem).Count(&dbCount)
		dbItem.Name = groupModel.Name
		dbItem.Avatar = groupModel.Avatar
		if err = tx.Save(&dbItem).Error; err != nil {
			return errorServer.E("GroupEdit更新失败")
		}
	} else {
		//新增
		dbCount := 0
		dbItem := &model.Group{}
		tx.Where(map[string]interface{}{"name": groupModel.Name}).Find(dbItem).Count(&dbCount)
		if dbCount > 0 {
			return errorServer.E("群名称已被使用")
		}

		if err = tx.Create(&groupModel).Error; err != nil {
			return errorServer.E("GroupEdit新增失败")
		}
	}

	tx.Commit()
	return nil

}

func GroupDel(r *proto_service.Group) error {

	tx := model.DB.Begin()
	if err := tx.Where(map[string]interface{}{"id": r.GroupId}).Delete(model.Group{}); err != nil {
		return errorServer.E("GroupDel失败")
	}

	tx.Commit()
	return nil

}

// 查询这个人有那些组
func QueryGroupList(req *proto_service.QueryCommonReq) []*proto_service.IMGroup {

	tx := model.DB.Begin()
	uid, _ := model.GetUidByProjectInfo(int(req.ProjectId), int(req.ProjectUid))
	groupUserList := []model.UserGroup{}
	tx.Where(model.UserGroup{Uid: uid}).Find(&groupUserList)
	groupIds, _ := util.ArrayColumn(groupUserList, "GroupId")

	groupList := []model.Group{}
	tx.Where("id in (?)", groupIds).
		Or(map[string]interface{}{"create_uid": int(req.ProjectUid)}).
		Find(&groupList)

	IMGroups := []*proto_service.IMGroup{}
	for _, v := range groupList {
		IMGroups = append(IMGroups, &proto_service.IMGroup{
			Id:     int64(v.ID),
			Name:   v.Name,
			Avatar: v.Avatar,
			Des:    v.Des,
		})
	}
	tx.Commit()
	return IMGroups

}
