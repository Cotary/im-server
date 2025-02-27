package rpc_controller

import (
	"golang.org/x/net/context"
	"im-server/grpc/proto_service"
	"im-server/service"
	"im-server/util"
)

type ImRpcController struct {
	proto_service.IMRpcServer
}

func (c *ImRpcController) UserAdd(ctx context.Context, in *proto_service.User) (*proto_service.HandleResponse, error) {
	util.LogPrintln("收到一个 UserAdd 请求，请求参数：", in)

	err := service.UserEdit(in)
	return handleResponse(err)
}

func (c *ImRpcController) UserEdit(ctx context.Context, in *proto_service.User) (*proto_service.HandleResponse, error) {
	util.LogPrintln("收到一个 UserEdit 请求，请求参数：", in)

	err := service.UserEdit(in)
	return handleResponse(err)
}
func (c *ImRpcController) UserDel(ctx context.Context, in *proto_service.User) (*proto_service.HandleResponse, error) {
	util.LogPrintln("收到一个 UserDel 请求，请求参数：", in)

	err := service.UserDel(in)
	return handleResponse(err)
}

// 查询方法
func (c *ImRpcController) QueryUserList(ctx context.Context, in *proto_service.QueryCommonReq) (*proto_service.IMUsers, error) {
	util.LogPrintln("收到一个 QueryUserList 请求，请求参数：", in)

	imUsers := service.QueryUserList(in)

	return &proto_service.IMUsers{
		User: imUsers,
	}, nil
}

//组

func (c *ImRpcController) GroupAdd(ctx context.Context, in *proto_service.Group) (*proto_service.HandleResponse, error) {
	util.LogPrintln("收到一个 GroupAdd 请求，请求参数：", in)

	err := service.GroupEdit(in)
	return handleResponse(err)
}

func (c *ImRpcController) GroupEdit(ctx context.Context, in *proto_service.Group) (*proto_service.HandleResponse, error) {
	util.LogPrintln("收到一个 GroupEdit 请求，请求参数：", in)

	err := service.GroupEdit(in)
	return handleResponse(err)
}

func (c *ImRpcController) GroupDel(ctx context.Context, in *proto_service.Group) (*proto_service.HandleResponse, error) {
	util.LogPrintln("收到一个 GroupDel 请求，请求参数：", in)

	err := service.GroupDel(in)
	return handleResponse(err)
}

// 查询方法
func (c *ImRpcController) QueryGroupList(ctx context.Context, in *proto_service.QueryCommonReq) (*proto_service.IMGroups, error) {
	util.LogPrintln("收到一个 QueryGroupList 请求，请求参数：", in)

	imGroups := service.QueryGroupList(in)

	return &proto_service.IMGroups{
		Group: imGroups,
	}, nil
}

func (c *ImRpcController) UserGroupEdit(ctx context.Context, in *proto_service.UserGroup) (*proto_service.HandleResponse, error) {
	util.LogPrintln("收到一个 UserGroupEdit 请求，请求参数：", in)

	err := service.UserGroupEdit(in)
	return handleResponse(err)
}

func (c *ImRpcController) QueryMessageList(ctx context.Context, in *proto_service.QueryMessageReq) (*proto_service.IMMessages, error) {
	util.LogPrintln("收到一个 QueryMessageList 请求，请求参数：", in)

	imMessages, err := service.QueryMessageList(in)

	return &proto_service.IMMessages{
		Message: imMessages,
	}, err
}

func handleResponse(err error) (*proto_service.HandleResponse, error) {

	response := &proto_service.HandleResponse{Ok: true, Msg: ""}
	if err != nil {
		response.Ok = false
		response.Msg = err.Error()
		return response, err
	}
	return response, nil
}
