package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip            string
	Port          int
	OnlineUserMap map[string]*User
	mapLock       sync.RWMutex
	C             chan string
}

func NewServer(ip string, port int) *Server {

	server := &Server{
		Ip:            ip,
		Port:          port,
		OnlineUserMap: make(map[string]*User),
		C:             make(chan string),
	}

	// 服务器消息管道监听
	go server.listenChannel()

	return server
}

// 监听服务器的消息管道 一旦管道中有消息 就发送给所有用户
func (this *Server) listenChannel() {

	for {
		msg := <-this.C
		this.mapLock.Lock()
		for _, v := range this.OnlineUserMap {
			v.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播
func (this *Server) Broadcast(user *User, msg string) {

	this.C <- fmt.Sprintf("["+"%s"+"]: "+"%s\n", user.Name, msg)
}

// 监听用户发来的消息
func (this *Server) listenUserMsg(user *User, isLive chan bool) {

	b := make([]byte, 4096)
	for {

		n, err := user.Conn.Read(b)
		if err != nil && err != io.EOF {
			fmt.Println("服务器接受数据报错")
			return
		}
		if n == 0 {
			fmt.Println("用户下线")
			user.Offline()
			// 下线了 就需要结束 监听用户发来消息的go程
			return
		}

		// 证明用户是活动的 不会被强踢掉
		isLive <- true
		// fmt.Println("长度", len(b))
		// n-1 是因为用户消息是以 换行符结束的  需要剔除换行符
		userMsg := string(b[:n-1])
		fmt.Println("服务端收到的消息:", userMsg)
		// 用户消息处理
		user.DoMessage(userMsg)
	}
}

// 与服务器建立连接之后 对客户端的后续处理
func (this *Server) handle(conn net.Conn) {

	isLive := make(chan bool)
	user := NewUser(conn, this)

	user.Online()
	// 监听用户消息
	go this.listenUserMsg(user, isLive)
	// 阻塞 以免handle go程 执行结束
	for {

		select {
		case <-isLive:

		case <-time.After(300 * time.Second):
			this.forceOffline(user)
			return

		}
	}
}

// 服务端强踢下线
func (this *Server) forceOffline(user *User) {
	user.SendToClient("您被踢了")
	user.Offline()
	close(user.C)
	user.Conn.Close()
}

// 启动服务器
func (this *Server) Start() {

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	fmt.Println("服务器已经启动, 工作端口为:", this.Port)
	for {

		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		// 处理每个客户端的连接
		go this.handle(conn)
	}
}
