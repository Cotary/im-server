package service

import (
	"im-server/grpc/proto_service"
	"im-server/model"
	"im-server/server/error"
	"im-server/util"
)

func UserGroupEdit(r *proto_service.UserGroup) error {

	tx := model.DB.Begin()
	defer tx.Rollback()

	//todo 有错误的时候需不需要tx.Rollback() ？

	//ProjectId    int64   `protobuf:"varint,1,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	//ProjectUid   int64   `protobuf:"varint,2,opt,name=project_uid,json=projectUid,proto3" json:"project_uid,omitempty"`
	//ProjectUsers []*User `protobuf:"bytes,3,rep,name=project_users,json=projectUsers,proto3" json:"project_users,omitempty"`
	//GroupId      int64   `protobuf:"varint,4,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`

	if r.GroupId <= 0 {
		return errorServer.E("用户组id不合法")
	}
	if len(r.ProjectUsers) <= 0 {
		return errorServer.E("用户ids不合法")
	}

	// 判断group_id是否存在
	group := model.Group{}
	tx.Where(map[string]interface{}{"id": r.GroupId}).First(&group)
	if group == (model.Group{}) {
		return errorServer.E("用户组不存在")
	}

	//删除
	tx.Where(map[string]interface{}{"group_id": r.GroupId}).Delete(model.UserGroup{})
	//使用in 必须使用[]string 的格式，[]int应该也可以，不能是[]interface
	dbUids, _ := util.ArrayColumn(r.ProjectUsers, "ProjectUid")

	addIMUsers := []model.User{}
	tx.Where(map[string]interface{}{"project_id": r.ProjectId}).Where("project_uid in (?)", dbUids).Find(&addIMUsers)
	//把这些人新增
	//addUserGroup := []*model.UserGroup{}
	for _, v := range addIMUsers {

		//addUserGroup = append(addUserGroup, &model.UserGroup{
		//	Uid:     int(v.ID),
		//	GroupId: int(r.GroupId),
		//})

		if err := tx.Create(&model.UserGroup{
			Uid:     int(v.ID),
			GroupId: int(r.GroupId),
		}).Error; err != nil {
			return errorServer.E("UserGroupEdit失败")
		}
	}
	//todo 使用批量创建新数据
	//util.LogPrintln(111, addUserGroup)
	//if len(addUserGroup) > 0 {
	//	if err := tx.Create(&addUserGroup).Error; err != nil {
	//		return errorServer.E("UserGroupEdit")
	//	}
	//}

	tx.Commit()
	return nil

}
