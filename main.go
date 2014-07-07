package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gamingrobot/steamgo"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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
	var ConnectionStage int = 0
	var IRCUsername string

	/*hostname, e := os.Hostname()
	if e != nil {
		hostname = "Unknown"
	}*/

	reader := bufio.NewReader(conn)
	for {
		lineb, _, err := reader.ReadLine()
		line := string(lineb)

		if err != nil {
			return
		}

		fmt.Println(line)

		if strings.HasPrefix(line, "QUIT ") {
			conn.Close()
			return
		}

		if strings.HasPrefix(line, "PASS ") && ConnectionStage == 0 {
			ConnectionStage++
		}

		// try and parse the string as a number to see what would happen
		linen := strings.TrimSpace(string(lineb))
		_, err = strconv.ParseInt(linen, 10, 64)
		if err == nil && ConnectionStage == -1 {
		}

		if strings.HasPrefix(line, "USER ") && ConnectionStage == 1 {
			if IRCUsername != "" {
				ConnectionStage++
			}
		}

		if strings.HasPrefix(strings.ToUpper(line), "MENTION") && ConnectionStage == 2 {
		}

		if strings.HasPrefix(strings.ToUpper(line), "ALL") && ConnectionStage == 2 {
		}

		if strings.HasPrefix(line, "JOIN ##friends") && ConnectionStage == 2 {
			conn.Write([]byte(fmt.Sprintf(":%s!~%s@twitter.com JOIN ##friends * :Blah\r\n", IRCUsername, IRCUsername)))
		}

		if strings.HasPrefix(line, "MODE ##friends") && ConnectionStage == 2 {
			conn.Write(GenerateIRCMessageBin(RplChannelModeIs, IRCUsername, "##friends +ns"))
			conn.Write(GenerateIRCMessageBin(RplChannelCreated, IRCUsername, "##friends 1401629312"))
			go PingClient(conn)
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
