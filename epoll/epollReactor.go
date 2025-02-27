package epoll_server

import (
	"fmt"
	"im-server/util"
	"net"
	"sync"
	"syscall"
	"time"
)

type epollReactor struct {
	code              int
	listenFd          int
	epollFd           int
	stop              bool
	isMain            bool //是否是主监听
	multiEpollReactor *MultiEpollReactor

	events sync.Map //fd->event
}

func (e *epollReactor) accept() error {
	//accept
	workFd, _, err := syscall.Accept(syscall.Handle(e.listenFd))
	if err != nil {
		return err
	}
	//设置非阻塞io
	err = syscall.SetNonblock(workFd, true)
	if err != nil {
		return err
	}
	//分发给subReactor

	//加入epoll监听
	subReactor := e.multiEpollReactor.getSubReactor()
	err = subReactor.addEvent(int(workFd), syscall.EPOLLIN|-syscall.EPOLLET)
	if err != nil {
		return err
	}

	//把event加入subReactor
	subReactor.addWorkConn(newEvent(subReactor, workFd))
	fmt.Println("accept over")
	return nil

}

// 开启监听
func (e *epollReactor) Listen(ipAddr string, port int) error {
	//使用系统调用,打开一个socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return err
	}
	//设置端口重复使用
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return err
	}
	//设置小包发送
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.TCP_NODELAY, 1)
	if err != nil {
		return err
	}
	//设置非阻塞io
	err = syscall.SetNonblock(fd, true)
	if err != nil {
		return err
	}

	//ip地址转换
	var addr [4]byte
	copy(addr[:], net.ParseIP(ipAddr).To4())
	net.ParseIP(ipAddr).To4()
	err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Port: port,
		Addr: addr,
	})
	if err != nil {
		return err
	}

	//开启监听
	err = syscall.Listen(fd, 10)
	if err != nil {
		return err
	}
	e.listenFd = int(fd)

	//这里把监听socket放在epoll
	err = e.addEvent(e.listenFd, syscall.EPOLLIN|-syscall.EPOLLET)
	if err != nil {
		return err
	}
	e.addWorkConn(newEvent(e, e.listenFd))

	return nil
}

func (e *epollReactor) getEvent(fd int) *event {
	res, ok := e.events.Load(fd)
	if ok {
		return res.(*event)
	}
	return nil
}
func (e *epollReactor) closeEvent(event *event) {
	e.delEvent(event)
	e.delWorkConn(event)

}

func (e *epollReactor) addWorkConn(event *event) {
	e.events.Store(event.fd, event)
}
func (e *epollReactor) delWorkConn(event *event) {
	e.events.Delete(event.fd)
}
func (e *epollReactor) addEvent(fd int, events uint32) error {
	//syscall.EPOLLIN | syscall.EPOLLERR | syscall.EPOLLHUP | syscall.POLLUTED | -syscall.EPOLLET, //这里有几个，epoll就会检测这几个
	err := syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
		Pad:    0,
	})
	return err
}

func (e *epollReactor) editEvent(fd int, events uint32) error {
	err := syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
		Pad:    0,
	})
	return err
}

func (e *epollReactor) delEvent(event *event) error {

	//从epoll中去除
	err := syscall.EpollCtl(e.epollFd, syscall.EPOLL_CTL_DEL, event.fd, nil)
	//去除reactor的

	return err
}
func eventLoop(e *epollReactor) {
	for !e.stop {
		//定时器
		e.eventLoopOnce(-1)
	}
}

func (e *epollReactor) CloseOvertimeConn() {
	for {
		overtime := int(time.Now().Unix()) - 30
		e.events.Range(func(k, v interface{}) bool {
			conn := v.(*event)
			if conn.lastTime > 0 && conn.lastTime < overtime {
				e.closeEvent(conn)
				return false
			}
			return true
		})

	}
}

func (e *epollReactor) eventLoopOnce(timeout int) error {

	println("启动:", e.code)
	events := make([]syscall.EpollEvent, 1024)
	n, err := syscall.EpollWait(e.epollFd, events, timeout)
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		//epollerr
		//epollin
		//epollout
		//epollhup

		util.LogPrintln("epoll 触发", events[i])
		event := e.getEvent(int(events[i].Fd))
		if events[i].Events&syscall.EPOLLHUP == syscall.EPOLLHUP ||
			events[i].Events&syscall.EPOLLRDHUP == syscall.EPOLLRDHUP {
			fmt.Println("epollhup")
			event.handleHup()
		}
		if events[i].Events&syscall.EPOLLERR == syscall.EPOLLERR {
			fmt.Println("epollerr")
			//1.直接处理
			event.handleErr()
			//2.交给io函数处理，就是在读或者写的时候触发
			//events[i].Events=(syscall.EPOLLIN|syscall.EPOLLOUT)|events[i].Events

		}
		//epollin
		if events[i].Events&syscall.EPOLLIN == syscall.EPOLLIN {
			fmt.Println("epollin")
			event.handleIn()
		}
		//epollout
		if events[i].Events&syscall.EPOLLOUT == syscall.EPOLLOUT {
			fmt.Println("epollout")
			event.handleOut()
		}

	}
	return nil
}
