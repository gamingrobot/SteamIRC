package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Philipp15b/go-steam"
	. "github.com/Philipp15b/go-steam/internal/steamlang"
	"log"
	"net"
	"os"
	"sync"
)

var loginDetails steam.LogOnDetails

type LockingIRCConnections struct {
	sync.RWMutex
	byId      map[int64]*IRCConnection
	currentId int64
}

func (c *LockingIRCConnections) deleteIRCConnection(id int64) {
	c.Lock()
	defer c.Unlock()
	delete(c.byId, id)
}

func (c *LockingIRCConnections) addIRCConnection(conn *IRCConnection) int64 {
	c.Lock()
	defer c.Unlock()
	c.currentId += 1
	retid := c.currentId
	c.byId[retid] = conn
	return retid
}

var ircConnections *LockingIRCConnections

func main() {
	hostcfg := flag.String("listen", "localhost:6667", "<host>:<port>")
	flag.Parse()
	file, _ := os.Open("steamauth.cfg")
	decoder := json.NewDecoder(file)
	loginDetails = steam.LogOnDetails{}
	decoder.Decode(&loginDetails)
	ircConnections = &LockingIRCConnections{
		byId:      make(map[int64]*IRCConnection),
		currentId: 0,
	}
	// Listen for incoming connections.
	l, err := net.Listen("tcp", *hostcfg)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + *hostcfg)
	steamClient := steam.NewClient()
	server := steamClient.Connect()
	log.Println("Connecting to steam server:", server)
	go connectToSteam(steamClient, loginDetails)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		fmt.Println("Got Connection")
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		// Handle connections in a new goroutine.
		irc := &IRCConnection{Connection: conn, ConnectionState: ConnectionNone, Steam: steamClient}
		ircConnections.addIRCConnection(irc)
		irc.Start()
	}
}

func connectToSteam(steamClient *steam.Client, login steam.LogOnDetails) {
	for event := range steamClient.Events() {
		switch e := event.(type) { //Events that should *not* be passed to web
		case *steam.ConnectedEvent:
			log.Println("Connected to steam")
			steamClient.Auth.LogOn(login)
		case *steam.LoggedOnEvent:
			log.Println("Logged on steam as", login.Username)
			steamClient.Social.SetPersonaState(EPersonaState_Online)
		case *steam.LoggedOffEvent:
			log.Println("Logged off steam")
			steamClient.Auth.LogOn(login)
		case *steam.DisconnectedEvent:
			log.Println("Disconnected to steam")
		case *steam.MachineAuthUpdateEvent:
		case *steam.LoginKeyEvent:
		case steam.FatalErrorEvent:
			steamClient.Connect() // please do some real error handling here
			log.Print("FatalError", e)
		case error:
			log.Println(e)
		default:
			handleSteamEvent(event)
		}
	}
}

func handleSteamEvent(event interface{}) {
	switch e := event.(type) { //Events that are not part of login
	}
}
