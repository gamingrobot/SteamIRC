package main

import (
	"bufio"
	"fmt"
	"github.com/Philipp15b/go-steam"
	"log"
	"net"
	"sync"
	"time"
	"strings"
)

const (
	ConnectionNone      int = 0
	ConnectionAuthed    int = 1
	ConnectionConnected int = 2
	ConnectionInRoom    int = 3
)

type IRCConnection struct {
	sync.RWMutex
	Username        string
	Connection      net.Conn
	ConnectionState int
	Steam           *steam.Client
}

func (c *IRCConnection) Start() {
	reader := bufio.NewReader(c.Connection)
	go PingClient(c.Connection)
	for {
		lineb, _, err := reader.ReadLine()
		line := string(lineb)

		if err != nil {
			return
		}

		fmt.Println(line, "Connection", c.ConnectionState)

		if CheckPrefix(line, "QUIT") {
			c.Connection.Close()
			return
		}

		if CheckPrefix(line, "PASS") && c.ConnectionState == ConnectionNone {
			//Do password checking here
			c.ConnectionState = ConnectionAuthed
		}

		if CheckPrefix(line, "NICK") && c.ConnectionState == ConnectionAuthed {
			//Check to make sure ircusername matches steam username
			c.Username = strings.Split(line, " ")[1]
			c.Connection.Write(GetWelcomePackets(c.Username))
			c.JoinRoom()
		} else if CheckPrefix(line, "NICK ") && c.ConnectionState == ConnectionNone {
			c.Username = strings.Split(line, " ")[1]
			c.Connection.Write(GetWelcomePackets(c.Username))
			c.Connection.Write(GenerateIRCPrivateMessage("Please login use the PASS: steampassword", c.Username, "SYS"))
		}

		if CheckPrefix(line, "USER") && c.ConnectionState == ConnectionAuthed {
			if c.Username != "" {
				c.ConnectionState = ConnectionConnected
			}
		}

		if CheckPrefix(line, "MENTION") && c.ConnectionState == ConnectionConnected {
		}

		if CheckPrefix(line, "ALL") && c.ConnectionState == ConnectionConnected {
		}

		if CheckPrefix(line, "JOIN ##friends") && c.ConnectionState == ConnectionConnected {
			c.JoinRoom()
		}

		if CheckPrefix(line, "PART ##friends") && c.ConnectionState == ConnectionInRoom {
			c.JoinRoom()
		}

		if CheckPrefix(line, "MODE ##friends") && c.ConnectionState == ConnectionInRoom {
			c.Connection.Write(GenerateIRCMessageBin(RplChannelModeIs, c.Username, "##friends +ns"))
			c.Connection.Write(GenerateIRCMessageBin(RplChannelCreated, c.Username, "##friends 1401629312"))
		}
	}
}

func (c *IRCConnection) JoinRoom() {
	c.Connection.Write([]byte(fmt.Sprintf(":%s!~%s@steam JOIN ##friends\r\n", c.Username, c.Username)))
	for id, friend := range c.Steam.Social.Friends.GetCopy() {
		log.Println(id, friend.Name)
		c.Connection.Write(GenerateIRCMessageBin(RplNamReply, c.Username, fmt.Sprintf("@ ##friends :@%s %s", c.Username, FixName(friend.Name))))

	}
	c.Connection.Write(GenerateIRCMessageBin(RplEndOfNames, c.Username, "##friends :End of /NAMES list."))
	c.ConnectionState = ConnectionInRoom
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


func FixName(name string) string {
	return strings.Replace(name, " ", "_", -1)	
}