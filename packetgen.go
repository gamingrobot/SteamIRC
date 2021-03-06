package main

import (
	"fmt"
	"os"
)

func GenerateIRCMessage(code string, username string, data string) string {
	return fmt.Sprintf(":steam %s %s %s\r\n", code, username, data)
}

func GenerateIRCMessageBin(code string, username string, data string) []byte {
	return []byte(GenerateIRCMessage(code, username, data))
}

func GetWelcomePackets(IRCUsername string) []byte {
	hostname, e := os.Hostname()
	if e != nil {
		hostname = "Unknown"
	}

	pack := ""
	pack += GenerateIRCMessage(RplWelcome, IRCUsername, ":Welcome to SteamIRC")
	pack += GenerateIRCMessage(RplYourHost, IRCUsername, fmt.Sprintf(":Host is: %s", hostname))
	pack += GenerateIRCMessage(RplCreated, IRCUsername, ":This server was first made on 07/07/2014")
	pack += GenerateIRCMessage(RplMyInfo, IRCUsername, fmt.Sprintf(":%s steamIRC DOQRSZaghilopswz CFILMPQSbcefgijklmnopqrstvz bkloveqjfI", hostname))
	pack += GenerateIRCMessage(RplMotdStart, IRCUsername, ":Filling in a MOTD here because I have to.")
	pack += GenerateIRCMessage(RplMotdEnd, IRCUsername, ":done")
	return []byte(pack)
}

func GenerateIRCPrivateMessage(content string, room string, username string) []byte {
	return []byte(fmt.Sprintf(":%s!~%s@steam PRIVMSG %s :%s\r\n", username, username, room, content))
}
