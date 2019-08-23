package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	xmpp "gosrc.io/xmpp"
	stanza "gosrc.io/xmpp/stanza"
)

type iSig int

const (
	iExit iSig = iota
	iFail
)

var osSignals = make(chan os.Signal, 1)
var signals = make(chan iSig, 1)

func main() {
	//set up os signal listener
	signal.Notify(osSignals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for s := range osSignals {
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				fmt.Println("Got told by os signal to exit")
				signals <- iExit
			}
		}
	}()

	//read credentials and stuff from environment
	xmppUser, ok := os.LookupEnv("XMPP_USER")
	if !ok {
		fmt.Println("No user specified, failing")
		os.Exit(2)
	}

	xmppPw, ok := os.LookupEnv("XMPP_PW")
	if !ok {
		fmt.Println("No liveness password specified, failing")
		os.Exit(2)
	}

	xmppAuthDomain, ok := os.LookupEnv("XMPP_AUTH_DOMAIN")
	if !ok {
		fmt.Println("no xmpp auth domain specified")
		os.Exit(2)
	}

	xmppPort, ok := os.LookupEnv("XMPP_SERVER")
	if !ok || xmppPort == "" {
		xmppPort = "5222"
	}

	xmppServer, ok := os.LookupEnv("XMPP_PORT")
	if !ok {
		fmt.Println("no xmpp server specified")
		os.Exit(2)
	}

	jid := xmppUser + "@" + xmppAuthDomain
	address := xmppServer + ":" + xmppPort
	config := xmpp.Config{
		Address:      address,
		Jid:          jid,
		Password:     xmppPw,
		StreamLogger: os.Stdout,
		Insecure:     true,
		TLSConfig:    &tls.Config{InsecureSkipVerify: true},
	}

	router := xmpp.NewRouter()
	router.HandleFunc("message", handleMessage)
	router.HandleFunc("iq", handleIq)
	router.HandleFunc("presence", handlePresence)

	go connectClient(config, router)

	//keep process running
	for s := range signals {
		switch s {
		case iExit:
			shutdown()
			os.Exit(0)
		case iFail:
			shutdown()
			os.Exit(1)
		}
	}
}

func shutdown() {

}

func handleMessage(s xmpp.Sender, p stanza.Packet) {

}

func handleIq(s xmpp.Sender, p stanza.Packet) {

}

func handlePresence(s xmpp.Sender, p stanza.Packet) {

}

func connectClient(c xmpp.Config, r *xmpp.Router) {
	client, err := xmpp.NewClient(c, r)
	if err != nil {
		fmt.Printf("unable to create client: %s\n", err.Error())
		signals <- iFail
	}

	cm := xmpp.NewStreamManager(client, nil)
	err = cm.Run()
	if err != nil {
		fmt.Printf("xmpp connection manager returned with error: %s\n", err.Error())
		signals <- iFail
	}
}