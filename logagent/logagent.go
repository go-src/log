package main

import (
	"bufio"
	"code.google.com/p/gcfg"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

type Config struct {
	Section struct {
		Server  string
		Logfile []string
	}
}

var c = make(chan os.Signal, 1)

func main() {
	delay, retry := 1, 0

	cfg := Config{}
	err := gcfg.ReadFileInto(&cfg, "agent.conf")
	if err != nil {
		log.Fatalf("Failed to parse gcfg data: %s", err)
	}

Redial:
	raddr, _ := net.ResolveTCPAddr("tcp4", cfg.Section.Server)
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		retry++
		log.Printf("%s, retrying %d, delay %ds.\n", err, retry, delay)
		delay = delay + retry
		time.Sleep(time.Second * time.Duration(delay))
		goto Redial
	}

	log.Printf("Connected to server %s.\n", conn.RemoteAddr())
	delay = 1 // Reset the delay interval
	retry = 0
	go send(conn)
	go recv(conn)

	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Printf("\nCaught signal: %s\n", s)
	conn.Close()
	time.Sleep(time.Second * 1)
	goto Redial
}

func send(conn *net.TCPConn) {
	for {
		b := []byte("Hello World~\n")
		n, err := conn.Write(b)

		fmt.Printf("%d bytes written: %s", n, b)
		if err != nil {
			fmt.Println(err, ", Redialing...")
			redial(conn)
			return
		}
		time.Sleep(time.Second * 2)
	}
}

func recv(conn *net.TCPConn) {
	scanner := bufio.NewScanner(conn)
	i := 0
	for scanner.Scan() {
		fmt.Printf("[%d][%s]%s\n", i, conn.RemoteAddr(), scanner.Text())
		i++
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Reading input:", err)
		redial(conn)
		return
	}
}

func redial(conn *net.TCPConn) {
	fmt.Println("Cleanup elegantly...")
	c <- os.Interrupt
}
