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
		Listen string
		Logto  string
	}
	Stats map[string]*struct {
		Col  string
		Stat string
	}
}

var guests []*net.TCPConn
var count int
var logch chan string
var c = make(chan os.Signal, 1)

func main() {
	cfg := Config{}
	err := gcfg.ReadFileInto(&cfg, "center.conf")
	if err != nil {
		log.Fatalf("Failed to parse gcfg data: %s", err)
	}

	laddr, err := net.ResolveTCPAddr("tcp4", cfg.Section.Listen)
	l, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer closeall()
	go counting()
	go broadcast()
	go writelog(cfg.Section.Logto)
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
				logch <- scanner.Text()
				log.Printf("[%d][%s]%s\n", i, conn.RemoteAddr(), scanner.Text())
				i++
			}
			if err := scanner.Err(); err != nil {
				log.Fatal("Reading input:", err)
			}
			log.Println("[GOODBYE] ", conn.RemoteAddr())
		}()
	}
	signal.Notify(c, os.Interrupt)
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

func writelog(filename string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case s := <-logch:
			f.WriteString(s)
		}
	}
}
