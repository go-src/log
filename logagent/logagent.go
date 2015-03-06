package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	delay, retry := 1, 0

Redial:
	raddr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:46714")
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
	go send(conn)
	go recv(conn)

	signal.Notify(c, syscall.SIGUSR2)
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
			syscall.Kill(os.Getpid(), syscall.SIGUSR2)
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
		syscall.Kill(os.Getpid(), syscall.SIGUSR2)
		return
	}
}
