//go:build linux
// +build linux

package epoll_server

//
//import (
//	errorServer "im-server/server/error"
//	"im-server/util"
//	"net"
//	"sync"
//	"syscall"
//	"time"
//)
//
//func NewEpollM() *EpollM {
//	return &EpollM{}
//}
//
//type EpollM struct {
//	conn sync.Map
//
//	socketFd int //监听socket的fd
//	epollFd  int //epoll的fd
//}
//
////关闭所有的链接
//func (e *EpollM) Close() {
//	e.conn.Range(func(k, v interface{}) bool {
//		util.LogPrintln("关闭所有的链接:", k, v)
//		conn := v.(*ServerConn)
//		e.CloseConn(conn.fd)
//		return true
//	})
//
//	syscall.Close(e.socketFd)
//	syscall.Close(e.epollFd)
//}
//
////获取符合条件的用户
//func (e *EpollM) GetUserConn(projectId int, projectUid int) (res *ServerConn) {
//
//	e.conn.Range(func(k, v interface{}) bool {
//		conn := v.(*ServerConn)
//		if conn.userInfo.ProjectId == projectId && conn.userInfo.ProjectUid == projectUid {
//			res = conn
//			return false
//		}
//		return true
//	})
//	return res
//
//}
//
////获取一个链接
//func (e *EpollM) GetConn(fd int) *ServerConn {
//	res, ok := e.conn.Load(fd)
//	if ok {
//		return res.(*ServerConn)
//	}
//	return nil
//}
//
////添加一个链接
//func (e *EpollM) AddConn(conn *ServerConn) {
//	e.conn.Store(conn.fd, conn)
//}
//
////删除一个链接
//func (e *EpollM) DelConn(fd int) {
//	e.conn.Delete(fd)
//}
//
////开启监听
//func (e *EpollM) Listen(ipAddr string, port int) error {
//	//使用系统调用,打开一个socket
//	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
//	if err != nil {
//		return err
//	}
//	//设置端口重复使用
//	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
//	if err != nil {
//		return err
//	}
//	//设置小包发送
//	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.TCP_NODELAY, 1)
//	if err != nil {
//		return err
//	}
//	//设置非阻塞io
//	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SOCK_NONBLOCK, 1)
//	if err != nil {
//		return err
//	}
//
//	//ip地址转换
//	var addr [4]byte
//	copy(addr[:], net.ParseIP(ipAddr).To4())
//	net.ParseIP(ipAddr).To4()
//	err = syscall.Bind(fd, &syscall.SockaddrInet4{
//		Port: port,
//		Addr: addr,
//	})
//	if err != nil {
//		return err
//	}
//
//	//开启监听
//	err = syscall.Listen(fd, 10)
//	if err != nil {
//		return err
//	}
//	e.socketFd = fd
//	return nil
//}
//
////在死循环中等待client发来的链接
//func (e *EpollM) Accept() error {
//
//	nfd, _, err := syscall.Accept(e.socketFd)
//	if err != nil {
//		return err
//	}
//	err = e.EpollAddEvent(nfd)
//	if err != nil {
//		return err
//	}
//	e.AddConn(&ServerConn{
//		fd:     nfd,
//		epollM: e,
//	})
//
//	return nil
//}
//
////关闭指定的链接
//func (e *EpollM) CloseConn(fd int) error {
//	conn := e.GetConn(fd)
//	util.LogPrintln("服务器主动关闭:", conn)
//
//	if conn == nil {
//		return errorServer.E("未找到要关闭的连接")
//	}
//	//conn.SendErrorMsg("服务器主动关闭")
//
//	err := e.EpollRemoveEvent(fd)
//	if err == nil {
//		if err = conn.Close(); err == nil {
//			e.DelConn(fd)
//			//修改在线状态
//			if conn.userInfo.ID > 0 {
//				UserIMOffline(conn.userInfo)
//			}
//			return nil
//		}
//	}
//
//	util.LogPrintln("连接关闭错误：", err)
//	return err
//
//}
//
////创建一个epoll
//func (e *EpollM) CreateEpoll() error {
//	//通过系统调用,创建一个epoll
//	fd, err := syscall.EpollCreate(1)
//	if err != nil {
//		return err
//	}
//	e.epollFd = fd
//	return nil
//}
//
//func (e *EpollM) CloseOvertimeConn() {
//	for {
//		overtime := int(time.Now().Unix()) - 30
//		e.conn.Range(func(k, v interface{}) bool {
//			conn := v.(*ServerConn)
//			if conn.lastTime > 0 && conn.lastTime < overtime {
//				e.CloseConn(conn.fd)
//				return false
//			}
//			return true
//		})
//
//	}
//}
//
////处理epoll
//func (e *EpollM) HandlerEpoll() error {
//	e.EpollAddEvent(e.socketFd) //把自己加入监听
//	events := make([]syscall.EpollEvent, 1024)
//	//在死循环中处理epoll
//	for {
//		//util.LogPrintln("epoll——run")
//		//msec -1,会一直阻塞,直到有事件可以处理才会返回, n 事件个数   ，这个和阻塞io不是一回事
//		//这里epoll 不支持接受accept 事件，有连接来时 还是一直阻塞的
//		//水平触发，要设置为 边缘触发。出现很多epoll_event——run 应该就是水平触发，没有read到socket里面的数据，下次又被触发了
//		//如果将上面的代码中的event.events = EPOLLIN | EPOLLET;改成event.events = EPOLLIN;，即使用LT模式，则运行程序后，会一直输出hello world。
//
//		//int n=read(fd,buf,size) 打算拷贝size，实际读了n   n=0,就说明读取完了，n=-1，errno= ewouldblock，就说明有错误，非阻塞io中也有可能表示读完了
//		//默认是阻塞的，阻塞io是绑定在fd的，这个需要要在accept的时候也去设置非阻塞io
//		//阻塞io   read,write,accept,listen  等等，如果用阻塞io，在epoll来的时候，可能会触发可读事件，但是校验和失败了，read就一直阻塞了
//
//		//检测io（io多路复用）   操作io
//		//write写数据的时候，如果缓冲区不够，要把链接放入io复用去监听，这时候只有有剩余空间，就有通知的写事件，这个时候就可以调用write，如果都写进去了，就会注销写事件
//		//链接断开是可以单向关闭的，一个是shutdown ,close,  epollrdhup  客户端写，服务器读端关闭，这时候服务器还能发送    epollhup 读写关都关闭了
//		n, err := syscall.EpollWait(e.epollFd, events, -1)
//		util.LogPrintln("epoll_event——run:", n)
//
//		if err != nil {
//			util.LogPrintln("epoll_event——run——err:", err)
//			return err
//		}
//		for i := 0; i < n; i++ {
//			//epollerr
//			//epollin
//			//epollout
//			//epollhup
//
//			util.LogPrintln("可读信息", events[i])
//			fd := int(events[i].Fd)
//
//			if fd == e.socketFd {
//				e.Accept()
//				continue
//			}
//
//			//先在map中是否有这个链接
//			conn := e.GetConn(int(events[i].Fd))
//			//socketAddr, _ := syscall.Getpeername(int(events[i].Fd))
//			//ip := socketAddr.(*syscall.SockaddrInet4).Addr
//			//port := socketAddr.(*syscall.SockaddrInet4).Port
//			//util.LogPrintln("ip：", ip, "port:", port)
//
//			if conn == nil { //没有这个链接,忽略
//				continue
//			}
//			if events[i].Events&syscall.EPOLLHUP == syscall.EPOLLHUP ||
//				events[i].Events&syscall.EPOLLRDHUP == syscall.EPOLLRDHUP ||
//				events[i].Events&syscall.EPOLLERR == syscall.EPOLLERR {
//				util.LogPrintln("error")
//				//断开||出错
//				if err = e.CloseConn(int(events[i].Fd)); err != nil {
//					util.LogPrintln("有错误:", err)
//					return err
//				}
//				//todo 如果错误是EPOLLRDHUP，说明客户端还有一句话发送出来了，可以走下面的解析，看看这个错误是啥
//			} else if events[i].Events == syscall.EPOLLIN {
//				//可读事件
//				//http://www.wetools.com/websocket/
//				//这里先不搞并发，否则，没读完，不是边缘触发的话，epoll会一直触发进入
//				util.LogPrintln("handle")
//				go conn.handle()
//			}
//		}
//	}
//}
//
////向epoll中加事件
//func (e *EpollM) EpollAddEvent(fd int) error {
//	//go net包
//	// 可读，可写，对端断开，边缘触发
//	//ev.events = _EPOLLIN | _EPOLLOUT | _EPOLLRDHUP | _EPOLLET
//	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
//		Events: syscall.EPOLLIN | syscall.EPOLLERR | syscall.EPOLLHUP | syscall.EPOLLRDHUP | -syscall.EPOLLET, //这里有几个，epoll就会检测这几个
//		Fd:     int32(fd),
//		Pad:    0,
//	})
//}
//
////从epoll中删除事件
//func (e *EpollM) EpollRemoveEvent(fd int) error {
//	return syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
//}
