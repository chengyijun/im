package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	Ip     string
	Port   int
	Name   string
	Conn   net.Conn
	chioce int
}

func NewClient(ip string, port int) *Client {

	client := &Client{

		Ip:     ip,
		Port:   port,
		chioce: 999,
	}
	return client
}
func (this *Client) Start() {

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		panic("连接服务器报错")
	}
	this.Conn = conn
	go this.listenServerResponse()
	fmt.Println("连接上了")
	this.Run()
}
func (this *Client) Rename() {
	fmt.Println("请输入新的用户名:")
	var newName string
	fmt.Scanln(&newName)
	msg := "rename|" + newName + "\n"
	this.Conn.Write([]byte(msg))

}
func (this *Client) Run() {
	for this.chioce != 0 {
		for this.Menu() != true {
		}

		switch this.chioce {
		case 1:
			fmt.Println("公聊模式")
			this.PublicChat()
		case 2:
			fmt.Println("私聊模式")
			this.PrivateChat()
		case 3:
			fmt.Println("更新用户名")
			this.Rename()
		case 0:
			return
		}
	}

}
func (this *Client) WhoOnline() {
	msg := "who\n"
	this.Conn.Write([]byte(msg))
}
func (this *Client) PrivateChat() {

	this.WhoOnline()
	var targetUsername string
	var msg string
	fmt.Println("请输入要私聊的用户名:(exit退出)")
	fmt.Scanln(&targetUsername)
	for targetUsername != "exit" {

		fmt.Println("请输入私聊内容:(exit退出)")
		fmt.Scanln(&msg)
		for msg != "exit" {

			targetMsg := "to|" + targetUsername + "|" + msg + "\n"
			this.Conn.Write([]byte(targetMsg))
			fmt.Println("请输入私聊内容:(exit退出)")
			fmt.Scanln(&msg)
		}
		fmt.Println("请输入要私聊的用户名:(exit退出)")
		fmt.Scanln(&targetUsername)
	}

}
func (this *Client) PublicChat() {

	var msg string

	fmt.Println("请输入内容：(exit退出)")
	fmt.Scanln(&msg)
	for msg != "exit" {

		this.Conn.Write([]byte(msg + "\n"))
		fmt.Println("请输入内容：(exit退出)")
		fmt.Scanln(&msg)
	}
}
func (this *Client) listenServerResponse() {

	io.Copy(os.Stdout, this.Conn)
}
func (this *Client) Menu() bool {

	fmt.Println("======菜单======")
	fmt.Println("======1-公聊======")
	fmt.Println("======2-私聊======")
	fmt.Println("======3-更新用户名======")
	fmt.Println("======0-退出======")
	fmt.Println("请输入您的选择:")
	var chioce int
	fmt.Scanln(&chioce)

	if chioce >= 0 && chioce <= 3 {
		this.chioce = chioce
		return true
	} else {
		fmt.Println("输入的数字不合法")
		return false
	}

}

var serverIp string
var serverPort int

func init() {

	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址 默认是127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口 默认是8888")
}
func main() {

	flag.Parse()
	client := NewClient(serverIp, serverPort)
	client.Start()
}
