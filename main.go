package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gamingrobot/steamgo"
	"net"
	"os"
	"strings"
	"time"
)

const (
	ConnectionNone int = 0
	ConnectionAuthed int = 1
	ConnectionConnected int = 2
)

var loginDetails steamgo.LogOnDetails

func main() {
	hostcfg := flag.String("listen", "localhost:6667", "<host>:<port>")
	flag.Parse()
	file, _ := os.Open("steamauth.cfg")
	decoder := json.NewDecoder(file)
	loginDetails = steamgo.LogOnDetails{}
	decoder.Decode(&loginDetails)
	// Listen for incoming connections.
	l, err := net.Listen("tcp", *hostcfg)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + *hostcfg)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		fmt.Println("Got Connection")
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		// Handle connections in a new goroutine.
		go handleIRCConn(conn)
	}
}

// Handles incoming requests.
func handleIRCConn(conn net.Conn) {
	var ConnectionStage int = ConnectionNone
	var IRCUsername string

	hostname, e := os.Hostname()
	if e != nil {
		hostname = "Unknown"
	}

	reader := bufio.NewReader(conn)
	for {
		lineb, _, err := reader.ReadLine()
		line := string(lineb)

		if err != nil {
			return
		}

		fmt.Println(line, "Connection", ConnectionStage)

		if CheckPrefix(line, "QUIT ") {
			conn.Close()
			return
		}

		if CheckPrefix(line, "PASS ") && ConnectionStage == ConnectionNone {
			//Do password checking here
			ConnectionStage = ConnectionAuthed
		}

		if CheckPrefix(line, "NICK ") && ConnectionStage == ConnectionAuthed {
			//Check to make sure ircusername matches steam username
			IRCUsername = strings.Split(line, " ")[1]
			conn.Write(GetWelcomePackets(IRCUsername, hostname))
		} else if CheckPrefix(line, "NICK ") && ConnectionStage == ConnectionNone {
			IRCUsername = strings.Split(line, " ")[1]
			conn.Write(GetWelcomePackets(IRCUsername, hostname))
			conn.Write(GenerateIRCPrivateMessage("Please login use the PASS: steampassword", IRCUsername, "SYS"))			
		}

		if CheckPrefix(line, "USER ") && ConnectionStage == ConnectionAuthed {
			if IRCUsername != "" {
				ConnectionStage = ConnectionConnected
				go PingClient(conn)
			}
		}

		if CheckPrefix(line, "MENTION") && ConnectionStage == ConnectionConnected {
		}

		if CheckPrefix(line, "ALL") && ConnectionStage == ConnectionConnected {
		}

		if CheckPrefix(line, "JOIN ##friends") && ConnectionStage == ConnectionConnected {
			conn.Write([]byte(fmt.Sprintf(":%s!~%s@steam JOIN ##friends * :Blah\r\n", IRCUsername, IRCUsername)))
		}

		if CheckPrefix(line, "MODE ##FRIENDS") && ConnectionStage == ConnectionConnected {
			conn.Write(GenerateIRCMessageBin(RplChannelModeIs, IRCUsername, "##friends +ns"))
			conn.Write(GenerateIRCMessageBin(RplChannelCreated, IRCUsername, "##friends 1401629312"))
		}
	}

}

func PingClient(conn net.Conn) {
	for {
		_, e := conn.Write([]byte(fmt.Sprintf("PING :%d\r\n", int32(time.Now().Unix()))))
		if e != nil {
			break
		}
		time.Sleep(time.Second * 30)
	}
}

func CheckPrefix(line string, prefix string) bool {
	st := strings.Split(line, " ")
	st[0] = strings.ToUpper(st[0])
	final := strings.Join(st, " ")
	return strings.HasPrefix(final, prefix)
}