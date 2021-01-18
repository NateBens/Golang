package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
)

var commands = make(map[string]func([]string))

func getAddress() string {
	localaddress := getLocalAddress()
	address := (localaddress + ":" + port)
	return address
}

func commandHandler(commandFields []string) {
	command := commandFields[0]
	value, ok := commands[command]
	if ok {
		args := commandFields[1:]
		value(args)
	} else {
		fmt.Printf("[%s] Command not recognized\n", command)
	}
}
func addCommand(command string, action func([]string)) {
	commands[command] = action
}
func fillCommands() {
	addCommand("help", doHelp)
	addCommand("port", doPort)
	addCommand("quit", doQuit)
	addCommand("create", doCreate)
	addCommand("ping", doPing)
	addCommand("get", doGet)
	addCommand("put", doPut)
	addCommand("put10", doPut10)
	addCommand("delete", doDelete)
	addCommand("dump", doDump)
	addCommand("join", doJoin)
}

func doDump(args []string) {
	if len(args) != 0 {
		log.Printf("'dump' takes no arguments.")
		return
	}
	var node Node
	var junk Nothing
	address := getAddress()
	if err := call(address, "Server.Dump", junk, &node); err != nil {
		log.Fatalf("Calling Server.Dump: %v", err)
	}

	log.Println("Neighborhood")
	predInt := hashString(node.Pred).Text(16)
	log.Printf("pred:	%v.. (%v)\n", predInt[:8], node.Pred)
	addressInt := hashString(address).Text(16)
	log.Printf("self:	%v.. (%v)\n", addressInt[:8], address)
	succ0Int := hashString(node.Succ[0]).Text(16)
	succ1Int := hashString(node.Succ[1]).Text(16)
	succ2Int := hashString(node.Succ[2]).Text(16)
	log.Printf("succ 0:	%v.. (%v)\n", succ0Int[:8], node.Succ[0])
	log.Printf("succ 1:	%v.. (%v)\n", succ1Int[:8], node.Succ[1])
	log.Printf("succ 2:	%v.. (%v)\n", succ2Int[:8], node.Succ[2])

	log.Println("Finger Table")
	for i := len(node.Finger) - 1; i >= 0; i-- {
		fingerInt := hashString(node.Finger[i]).Text(16)
		log.Printf("[%v]:	%v.. (%v)\n", i, fingerInt[:8], node.Finger[i])
		if i != 0 {
			if node.Finger[i] == node.Finger[i-1] {
				break
			}
		}
	}

	log.Println("Data items")
	for k, v := range node.Bucket {
		keyInt := hashString(k).Text(16)
		log.Printf("[%v] => [%v] ... %v\n", k, v, keyInt[:8])
	}
}

func doGet(args []string) {
	if len(args) != 1 {
		log.Printf("'get' requires one argument, a key.")
		return
	}

	var node Node
	var junk Nothing
	localaddress := getAddress()
	if err := call(localaddress, "Server.Dump", junk, &node); err != nil {
		log.Fatalf("Calling Server.Dump: %v", err)
	}
	idInt := hashString(args[0])
	address := find(idInt, node.Succ[0])

	key := args[0]
	var value string
	log.Printf("sending get request to node: %v\n", address)
	if err := call(address, "Server.Get", key, &value); err != nil {
		log.Fatalf("Calling Server.Get: %v", err)
	}
	if value == "" {
		log.Printf("[%v] => nil\n", key)
	} else {
		log.Printf("[%v] => [%v]\n", key, value)
	}

}
func doDelete(args []string) {
	if len(args) != 1 {
		log.Printf("'delete' requires one argument, a key.")
		return
	}

	var node Node
	var junk Nothing
	localaddress := getAddress()
	if err := call(localaddress, "Server.Dump", junk, &node); err != nil {
		log.Fatalf("Calling Server.Dump: %v", err)
	}
	idInt := hashString(args[0])
	address := find(idInt, node.Succ[0])

	key := args[0]
	var value string
	log.Printf("sending delete request to node: %v\n", address)
	if err := call(address, "Server.Delete", key, &value); err != nil {
		log.Fatalf("Calling Server.Delete: %v", err)
	}
	if value == "" {
		log.Printf("[%v] => nil ==> [%v] => nil\n", key, key)
	} else {
		log.Printf("[%v] => [%v] ==>  [%v] => nil\n", key, value, key)
	}
}
func doPut(args []string) {
	if len(args) != 2 {
		log.Printf("'put' requires two arguments, a key, and a value.")
		return
	}
	var node Node
	var junk Nothing
	localaddress := getAddress()
	if err := call(localaddress, "Server.Dump", junk, &node); err != nil {
		log.Fatalf("Calling Server.Dump: %v", err)
	}
	idInt := hashString(args[0])
	address := find(idInt, node.Succ[0])
	key := args[0]
	value := args[1]

	log.Printf("sending put request to node: %v\n", address)
	if err := call(address, "Server.Post", args[0:], &junk); err != nil {
		log.Fatalf("Calling Server.Post: %v", err)
	}
	log.Printf("[%v] => [%v]\n", key, value)
}
func doPut10(args []string) {
	putargs := make([]string, 2)
	putargs[0] = "a"
	putargs[1] = "b"
	doPut(putargs)
	putargs[0] = "x"
	putargs[1] = "y"
	doPut(putargs)
	putargs[0] = "nate"
	putargs[1] = "benson"
	doPut(putargs)
	putargs[0] = "apple"
	putargs[1] = "sauce"
	doPut(putargs)
	putargs[0] = "cheese"
	putargs[1] = "burger"
	doPut(putargs)
	putargs[0] = "123"
	putargs[1] = "456"
	doPut(putargs)
	putargs[0] = "cookie"
	putargs[1] = "monster"
	doPut(putargs)
	putargs[0] = "cs"
	putargs[1] = "3410"
	doPut(putargs)
	putargs[0] = "chicken"
	putargs[1] = "sandwich"
	doPut(putargs)
	putargs[0] = "jimmy"
	putargs[1] = "johns"
	doPut(putargs)
}
func doPing(args []string) {
	if len(args) != 1 {
		fmt.Printf("'ping' requires one argument, an address to ping")
		return
	}
	address := args[0]
	var response string
	var junk Nothing
	log.Printf("sending ping request to node: %v\n", address)
	if err := call(address, "Server.Ping", &junk, &response); err != nil {
		log.Fatalf("Calling Server.Ping: %v", err)
	}
	log.Printf(response + "\n")
}

func doCreate(args []string) {
	log.Printf("Creating new ring")
	node := &Node{}
	address := getAddress()
	//Call server with address
	go server(address, node)
	//set the predecessor
	node.Pred = ""
	//set the successor
	node.Succ[0] = address

	// Create the bucket
	node.Bucket = make(map[string]string)
	go func() {
		var junk Nothing
		if err := call(address, "Server.RunBackgroundTasks", &node, &junk); err != nil {
			log.Fatalf("Calling Server.RunBackgroundTasks: %v", err)
		}
	}()
}

func doJoin(args []string) {
	if len(args) != 1 {
		log.Printf("'join' requires one argument, an address of an existing ring.")
		return
	}
	node := &Node{}
	address := getAddress()
	joinAddress := args[0]
	//Call server with address
	go server(address, node)
	//set the predecessor
	node.Pred = ""
	//set the successor
	addressInt := hashString(address)
	joinAddress = find(addressInt, joinAddress)
	node.Succ[0] = joinAddress
	//Create the bucket
	var junk Nothing
	node.Bucket = make(map[string]string)
	//m :=  make(map[string]string)
	log.Printf("Sending GetAll request to new successor: %v", node.Succ[0])
	if err := call(node.Succ[0], "Server.GetAll", junk, &node.Bucket); err != nil {
		log.Fatalf("Calling Server.Dump: %v", err)
	}
	//node.Bucket = m

	go func() {
		var junk Nothing
		if err := call(address, "Server.RunBackgroundTasks", &node, &junk); err != nil {
			log.Fatalf("Calling Server.RunBackgroundTasks: %v", err)
		}
	}()
}

func find(id *big.Int, startID string) string {
	found := false
	var nextNode Node
	nextNodeAddress := startID
	var junk Nothing
	i := 0
	for !found && i < maxSteps {
		if err := call(nextNodeAddress, "Server.Dump", junk, &nextNode); err != nil {
			log.Fatalf("Calling Server.Dump (find): %v", err)
		}
		found, nextNodeAddress = nextNode.FindSuccessor(id, nextNodeAddress)
		i++
	}
	if found {
		return nextNodeAddress
	} else {
		nothing := []string{}
		doDump(nothing)
		log.Fatalf("Node Unable to find Successor")
		return ""
	}
}

func (n Node) FindSuccessor(idInt *big.Int, nAddress string) (bool, string) {
	nInt := hashString(nAddress)
	//idInt := hashString(id)
	succInt := hashString(n.Succ[0])
	var CPNode string
	if between(nInt, idInt, succInt, true) {
		return true, n.Succ[0]
	} else {
		CPNode = n.ClosestPrecedingNode(idInt)
		return false, CPNode
	}
}

func (n Node) ClosestPrecedingNode(id *big.Int) string {
	localaddress := getAddress()
	localaddressInt := hashString(localaddress)
	for i := 160; i < 1; i-- {
		iInt := hashString(n.Finger[i])
		if between(localaddressInt, iInt, id, false) {
			return n.Finger[i]
		}
	}
	return n.Succ[0]
}

func doHelp(args []string) {
	fmt.Printf("List of Commands:\n")
	for k := range commands {
		fmt.Printf("  [%s] \n", k)
	}
}
func doQuit(args []string) {
	var node Node
	var junk Nothing
	localaddress := getAddress()
	if err := call(localaddress, "Server.Dump", junk, &node); err != nil {
		log.Fatalf("Calling Server.Dump: %v", err)
	}
	log.Printf("Sending data entries to successor: %v", node.Succ[0])
	if err := call(node.Succ[0], "Server.PutAll", node.Bucket, &junk); err != nil {
		log.Fatalf("Calling Server.Dump: %v", err)
	}
	fmt.Printf("Goodbye!\n")
	os.Exit(0)
}
func doPort(args []string) {
	if len(args) != 1 {
		fmt.Println("'port' command takes one argument, a port number")
	} else {
		port = args[0]
		log.Printf("changing port to: %v\n", port)
		address := getAddress()
		log.Printf("Address: %v\n", address)

	}
}

/*func doGetAlpha(args []string) {
	if len(args) != 2 {
		log.Printf("'get' requires two arguments, an address and a key.")
		return
	}
	address := args[0]
	key := args[1]
	var value string
	log.Printf("sending get request to node: %v\n", address)
	if err := call(address, "Server.Get", key, &value); err != nil {
		log.Fatalf("Calling Server.Get: %v", err)
	}
	if value == "" {
		log.Printf("Key: [%v] did not map to any value\n", key)
	} else {
		log.Printf("Key: [%v] came back with value: [%v]\n", key, value)
	}

}
func doDeleteAlp(args []string) {
	if len(args) != 2 {
		log.Printf("'delete' requires two arguments, an address and a key.")
		return
	}
	address := args[0]
	key := args[1]
	var value string
	log.Printf("sending delete request to node: %v\n", address)
	if err := call(address, "Server.Delete", key, &value); err != nil {
		log.Fatalf("Calling Server.Delete: %v", err)
	}
	if value == "" {
		log.Printf("Key: [%v] did not map to any value\n", key)
	} else {
		log.Printf("Entry with key: [%v] value: [%v] deleted\n", key, value)
	}
}
func doPutAlpha(args []string) {
	if len(args) != 3 {
		log.Printf("'put' requires three arguments, an address, a key, and a value.")
		return
	}
	address := args[0]
	key := args[1]
	value := args[2]
	log.Printf("sending put request to node: %v\n", address)
	var junk Nothing
	if err := call(address, "Server.Post", args[1:], &junk); err != nil {
		log.Fatalf("Calling Server.Post: %v", err)
	}
	log.Printf("Key: [%v] now has value: [%v]\n", key, value)
}*/
