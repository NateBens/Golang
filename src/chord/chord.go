package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	defaultHost          = "localhost"
	defaultPort          = "3410"
	defaultSuccessorSize = 3
	maxSteps             = 32
)

var port = defaultPort

func main() {
	fmt.Println("Welcome to the chord server!")
	localaddress := getLocalAddress()
	log.Printf("Found Local Address: %v\n", localaddress)
	address := (localaddress + ":" + port)
	log.Printf("Address: %v\n", address)
	fillCommands()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		commandFields := strings.Fields(input)
		if len(commandFields) != 0 {
			commandHandler(commandFields)
		} else {
			continue
		}
	}
}

func getLocalAddress() string {
	var localaddress string

	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}

	// find the first non-loopback interface with an IP address
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localaddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localaddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}

	return localaddress
}

/*func maina() {
	fmt.Println("Welcome to the chord server!")
	port := ":3410"
	_ = port //probably need to delete this later
	fillCommands()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		commandFields := strings.Fields(input)
		command := commandFields[0]
		args := commandFields[1:]
		commandHandler(command, args)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}*/

/*func mainb() {
	var isServer bool
	var isClient bool
	var address string
	flag.BoolVar(&isServer, "server", false, "start as a server")
	flag.BoolVar(&isClient, "client", false, "start as a client")
	flag.Parse()

	if isServer && isClient {
		log.Fatalf("Cannot be server and client")
	}
	if !isServer && !isClient {
		printUsage()
	}

	switch flag.NArg() {
	case 0:
		if isClient {
			address = defaultHost + ":" + defaultPort
		} else {
			address = ":" + defaultPort
		}
	case 1:
		//user specified the address
		address = flag.Arg(0)
	default:
		printUsage()
	}

	if isClient {
		shell(address)
	} else {
		server(address)
	}

}*/
