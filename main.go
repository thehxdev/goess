package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const GUESS_LIMIT = 10

var (
	address    string
	errlog            = log.New(os.Stderr, "[ERROR]\t", log.Lshortfile|log.Ldate|log.Ltime)
	inflog            = log.New(os.Stderr, "[INFO]\t", log.Lshortfile|log.Ldate|log.Ltime)
	conn_map          = &sync.Map{}
	conn_count uint64 = 0
	mu                = &sync.Mutex{}
)

type Game struct {
	id   uint64
	conn net.Conn
}

func main() {
	flag.StringVar(&address, "addr", "127.0.0.1:8000", "listen address")
	flag.Parse()

	l, err := net.Listen("tcp", address)
	if err != nil {
		errlog.Fatal(err)
	}

	serverContext, serverCancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		l.Close()
		conn_map.Range(func(id, game any) bool {
			inflog.Printf("closing connection with id %d", id.(uint64))
			game.(Game).conn.Close()
			return true
		})
		serverCancel()
	}()

	go func() {
		inflog.Println("listening on", address)
		for {
			conn, err := l.Accept()
			if err != nil {
				errlog.Println(err)
				break
			}
			go func(conn net.Conn) {
				mu.Lock()
				g := Game{
					id:   conn_count,
					conn: conn,
				}
				conn_count += 1
				mu.Unlock()

				remoteAddr := conn.RemoteAddr().String()
				conn_map.Store(g.id, g)

				defer func() {
					conn_map.Delete(g.id)
					conn.Close()
					inflog.Printf("connection from %s with id %d closed", remoteAddr, g.id)
				}()

				inflog.Printf("new connection from %s", remoteAddr)

				if err := g.startGame(); err != nil && err != net.ErrClosed && err != io.EOF {
					errlog.Println(err)
				}
			}(conn)
		}
	}()

	<-serverContext.Done()
}

func (g *Game) startGame() error {
	randnum := rand.IntN(101)
	conn := g.conn
	counter := 1
	buf := make([]byte, 4)

	reader := bufio.NewReader(conn)
	fmt.Fprintf(
		conn,
		"Welcome to goess game!\nGuess a number between 1 and 100 (Both ends are included)\nYou only have %d guesses!\n\n",
		GUESS_LIMIT,
	)

	for {
		if counter > GUESS_LIMIT {
			fmt.Fprintf(conn, "You failed!\nSee you later...\n")
			break
		}
		fmt.Fprintf(conn, "Guess %d => ", counter)
		_, err := reader.Read(buf)
		if err != nil {
			return err
		}

		guessed, err := parseInt(buf)
		if err != nil {
			fmt.Fprintln(conn, err)
			continue
		}

		if guessed > randnum {
			fmt.Fprintln(conn, "Not correct. Your guess is too high.")
		} else if guessed < randnum {
			fmt.Fprintln(conn, "Not correct. Your guess is too low.")
		} else {
			fmt.Fprintln(conn, "Correct!!! See you later...")
			break
		}

		counter += 1
	}

	return nil
}

func parseInt(buf []byte) (int, error) {
	invalidErr := fmt.Errorf("Invalid number")
	num := 0
	i := 0

	if buf[0] == '\n' {
		return -1, invalidErr
	}

	for {
		b := buf[i]
		if b == '\n' {
			break
		}
		if b > '9' || b < '0' {
			return -1, invalidErr
		}
		num = (num * 10) + (int(b) - '0')
		i += 1
	}
	return num, nil
}
