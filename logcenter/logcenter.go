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

var guests []*net.TCPConn
var count int

func main() {
	c := make(chan os.Signal, 1)
	laddr, err := net.ResolveTCPAddr("tcp4", ":46714")
	l, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closeall()
	go counting()
	go broadcast()
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Fatal("Error accepting: ", err)
		}
		log.Println("[WELCOME] ", conn.RemoteAddr())

		go func() {
			guests = append(guests, conn)
			scanner := bufio.NewScanner(conn)
			i := 0
			for scanner.Scan() {
				log.Printf("[%d][%s]%s\n", i, conn.RemoteAddr(), scanner.Text())
				i++
			}
			if err := scanner.Err(); err != nil {
				log.Fatal("Reading input:", err)
			}
			log.Println("[GOODBYE] ", conn.RemoteAddr())
		}()
	}
	signal.Notify(c, os.Interrupt, syscall.SIGHUP)
	for {
		s := <-c
		closeall()
		log.Fatal("Got signal: ", s)
	}
}

func counting() {
	count = len(guests)
	for {
		time.Sleep(time.Second * 1)
		cur := len(guests)
		if count == cur {
			continue
		}
		count = cur
		//fmt.Printf("MSG: %d guests online.\n", count)
	}
}

func broadcast() {
	for {
		time.Sleep(time.Second * 5)
		for _, c := range guests {
			now := fmt.Sprintf("%s", time.Now().Local())
			c.Write([]byte("Now Time: " + now + "\n"))
		}
	}
}

func closeall() {
	for _, c := range guests {
		c.Close()
	}
}
