# im-server

- 使用epoll实现的多reactor模式的io模型
- 主reactor负责监听新链接，其次负责读取消息
- 手动解析ws协议
- 封装简易的用户应用层IM
