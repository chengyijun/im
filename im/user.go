package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	Conn   net.Conn
	C      chan string
	Server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	remoteAddr := conn.RemoteAddr().String()
	user := &User{
		Name: remoteAddr,
		Addr: remoteAddr,
		Conn: conn,

		C:      make(chan string),
		Server: server,
	}

	// 启动管道监听
	go user.listenChannel()
	return user
}
func (this *User) listenChannel() {

	for {

		msg := <-this.C
		this.SendToClient(msg)
	}
}
func (this *User) SendToClient(msg string) {

	this.Conn.Write([]byte(msg))
}

func (this *User) Online() {

	this.Server.mapLock.Lock()
	this.Server.OnlineUserMap[this.Name] = this
	this.Server.mapLock.Unlock()

	this.Server.Broadcast(this, "上线了")
}

func (this *User) Offline() {

	this.Server.mapLock.Lock()
	delete(this.Server.OnlineUserMap, this.Name)
	this.Server.mapLock.Unlock()

	this.Server.Broadcast(this, "下线了")
}
func (this *User) SendMsgToAll(msg string) {

	this.Server.Broadcast(this, msg)
}
func (this *User) DoMessage(userMsg string) {

	if userMsg == "who" {

		this.Server.mapLock.Lock()
		for _, v := range this.Server.OnlineUserMap {
			onlineUserStr := "[" + v.Name + "]: 在线..." + "\n"
			this.SendToClient(onlineUserStr)
		}
		this.Server.mapLock.Unlock()
	} else if len(userMsg) > 7 && userMsg[:7] == "rename|" {

		newName := userMsg[7:]
		this.Server.mapLock.Lock()
		delete(this.Server.OnlineUserMap, this.Name)
		this.Server.OnlineUserMap[newName] = this
		this.Server.mapLock.Unlock()
		this.Name = newName
		this.SendToClient("用户名修改成功\n")
	} else if strings.HasPrefix(userMsg, "to|") {

		args := strings.Split(userMsg, "|")
		targetUsername := args[1]
		chatMsg := args[2]
		this.Server.mapLock.Lock()
		targetUser := this.Server.OnlineUserMap[targetUsername]
		newMsg := "[" + targetUsername + "](私聊): " + chatMsg + "\n"
		targetUser.SendToClient(newMsg)
		this.Server.mapLock.Unlock()
	} else {
		this.SendMsgToAll(userMsg)
	}
}
