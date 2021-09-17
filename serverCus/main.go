package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

const timeout = 10 * time.Second

type client struct {
	name string
	mess chan string
}
type message struct {
	from    string
	content string
}

var (
	mess    = make(chan message)
	clients = make(map[client]bool)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println(err)
	}
	go broadcast()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnect(conn)
	}
}

func broadcast() {
	for {
		msg := <-mess
		fmt.Println("mess: " + msg.content)
		for cli := range clients {
			if cli.name != msg.from {
				cli.mess <- msg.content
			}
		}
	}
}

func handleConnect(conn net.Conn) {
	var b []byte
	conn.Read(b)
	ch := make(chan string)
	go clientWritter(conn, ch)
	who := conn.RemoteAddr().String()

	welcomeClient(ch, who)
	m := message{from: who, content: who + " has arrived"}
	mess <- m
	cli := client{
		name: who,
		mess: ch,
	}

	timer := time.NewTimer(timeout)
	go func() {
		<-timer.C
		conn.Close()
	}()

	clients[cli] = true
	input := bufio.NewScanner(conn)
	for input.Scan() {
		m.content = who + ": " + input.Text()
		timer.Reset(timeout)
		mess <- m
	}
	clientLeft(conn, cli, m, who+"has left")
}

func clientWritter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func welcomeClient(ch chan string, who string) {
	fmt.Println("client " + who)
	ch <- "wellcome " + who
	ch <- "----------List client online----------"
	for cli := range clients {
		ch <- cli.name
	}
	ch <- "----------End----------"
}

func clientLeft(conn net.Conn, cli client, m message, content string) {
	m.content = content
	mess <- m
	delete(clients, cli)
	conn.Close()
}
