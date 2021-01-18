package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

type Feed struct {
	Messages []string
}
type Node struct {
	Pred   string
	Succ   [defaultSuccessorSize]string
	Finger [161]string
	Bucket map[string]string
}
type Handler func(*Node)
type Server chan<- Handler

type Nothing struct{}

type Pair struct {
	a int
	b string
}

func (s Server) Post(pair []string, reply *Nothing) error {
	log.Printf("Post request received...\n")
	finished := make(chan struct{})
	s <- func(n *Node) {
		key := pair[0]
		value := pair[1]
		n.Bucket[key] = value
		log.Printf("[%v] => [%v]\n", key, value)
		finished <- struct{}{}
	}
	<-finished
	return nil
}
func (s Server) Get(key string, reply *string) error {
	log.Printf("Get request received...\n")
	finished := make(chan struct{})
	s <- func(n *Node) {
		if val, ok := n.Bucket[key]; ok {
			*reply = val
			log.Printf("[%v] => [%v]\n", key, val)
		} else {
			*reply = ""
			log.Printf("[%v] => nil\n", key)
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}
func (s Server) Delete(key string, reply *string) error {
	log.Printf("Delete request received...\n")
	finished := make(chan struct{})
	s <- func(n *Node) {
		if val, ok := n.Bucket[key]; ok {
			*reply = val
			delete(n.Bucket, key)
			log.Printf("[%v] => [%v] ==>  [%v] => nil\n", key, val, key)
		} else {
			*reply = ""
			log.Printf("[%v] => nil ==> [%v] => nil\n", key, key)
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Dump(junk Nothing, reply *Node) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		//log.Printf("Recieving dump request")
		*reply = *n
		finished <- struct{}{}
	}
	<-finished
	return nil
}
func (s Server) PutAll(bucket map[string]string, reply *Nothing) error {
	finished := make(chan struct{})
	log.Println("Receiving PutAll request from predecessor... ")
	s <- func(n *Node) {
		for k, v := range bucket {
			n.Bucket[k] = v
		}
		log.Println("Data items updated")

		finished <- struct{}{}
	}
	<-finished
	return nil
}
func (s Server) GetAll(junk Nothing, newbucket *map[string]string) error {
	finished := make(chan struct{})
	*newbucket = make(map[string]string)
	log.Println("Receiving GetAll request... ")
	s <- func(n *Node) {
		//*newbucket = n.Bucket
		for k, v := range n.Bucket {
			(*newbucket)[k] = v
			delete(n.Bucket, k)
		}
		//newbucket = bucket
		log.Println("Data items")
		for k, v := range n.Bucket {
			log.Printf("[%v] => [%v]\n", k, v)
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) RunBackgroundTasks(n *Node, reply *Nothing) error {
	//log.Printf("RUNNING BACKGROUND TASKS")
	next := 0
	for {
		//log.Printf("PRED1: %v", n.Pred)
		n.Stablize()
		//log.Printf("PRED2: %v", n.Pred)
		time.Sleep(333 * time.Millisecond)
		//log.Printf("next: %v", next)
		next = n.FixFingers(next)
		time.Sleep(333 * time.Millisecond)
		//log.Printf(n.Pred)
		n.CheckPred()
		//log.Printf(n.Pred)
		time.Sleep(333 * time.Millisecond)
	}
}

func (n *Node) Stablize() {
	var succ Node
	succAddress := n.Succ[0]
	var junk Nothing
	localAddress := getAddress()
	//log.Printf("sending Dump request (stablize) to node: %v\n", succAddress)
	if err := call(succAddress, "Server.Dump", junk, &succ); err != nil {
		//log.Printf(n.Succ[0])
		log.Printf("Succesor Node failed")
		n.Succ[0] = n.Succ[1]
		//log.Printf(n.Succ[0])
		succlist := []string{n.Succ[0], n.Succ[1], n.Succ[2]}
		if err := call(localAddress, "Server.SetSuccessors", succlist, &junk); err != nil {
			log.Fatalf("Calling Server.SetSuccessors: %v", err)
		}
		if err := call(n.Succ[0], "Server.Dump", junk, &succ); err != nil {
			log.Fatalf("Calling Server.Dump (Stabilize): %v", err)
		}
		return
		/// this probably works, the problem is that my predecessor remains unchanged and fails
	}
	//log.Printf(n.Succ[0])
	succPred := succ.Pred
	succPredInt := hashString(succPred)
	nInt := hashString(localAddress)
	succInt := hashString(succAddress)
	if between(nInt, succPredInt, succInt, false) && succPred != "" {
		//log.Printf("succPred: %v\n", succPred)
		n.Succ[0] = succPred
		n.Succ[1] = succAddress
		n.Succ[2] = succ.Succ[0]
		succlist := []string{n.Succ[0], n.Succ[1], n.Succ[2]}
		if err := call(localAddress, "Server.SetSuccessors", succlist, &junk); err != nil {
			log.Fatalf("Calling Server.SetSuccessors: %v", err)
		}
		log.Printf("Successor changed to: %v\n", succPred)
	} else {
		n.Succ[0] = succAddress
		n.Succ[1] = succ.Succ[0]
		n.Succ[2] = succ.Succ[1]
		succlist := []string{n.Succ[0], n.Succ[1], n.Succ[2]}
		if err := call(localAddress, "Server.SetSuccessors", succlist, &junk); err != nil {
			log.Fatalf("Calling Server.SetSuccessors: %v", err)
		}
	}
	//fmt.Println(n.Succ[0])
	//log.Printf("Notifying node: %v\n", n.Succ[0])
	if err := call(n.Succ[0], "Server.Notify", localAddress, &junk); err != nil {
		if err := call(localAddress, "Server.Dump", junk, &succ); err != nil {
			log.Fatalf("Calling Server.Dump (notify): %v", err)
		}
		//log.Fatalf("Calling Server.Notify: %v", err)
	}
}
func (s Server) SetSuccessors(list []string, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		n.Succ[0] = list[0]
		n.Succ[1] = list[1]
		n.Succ[2] = list[2]
		finished <- struct{}{}
	}
	<-finished
	return nil
}
func (s Server) Notify(address string, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		//log.Printf("Incoming notify from: %v\n", address)
		localAddress := getAddress()
		predInt := hashString(n.Pred)
		localInt := hashString(localAddress)
		incomingInt := hashString(address)
		//log.Printf("PREDa: %v", n.Pred)
		if (n.Pred == "") || (between(predInt, incomingInt, localInt, false)) {
			n.Pred = address
			log.Printf("Predecessor changed to: %v\n", address)
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (n *Node) FixFingers(next int) int {
	next += 1
	if next >= 161 {
		next = 1
	}
	localAddress := getAddress()
	var junk Nothing
	fingerAddressInt := jump(localAddress, next)
	fingerAddress := find(fingerAddressInt, localAddress)

	if next == 1 {
		bigaddressInt := jump(localAddress, 155)
		bigaddress := find(bigaddressInt, localAddress)
		if bigaddress == fingerAddress {
			//flood fill first 155 entries
			if err := call(localAddress, "Server.FloodFinger", fingerAddress, &junk); err != nil {
				log.Fatalf("Calling Server.FloodFinger: %v", err)
			}
			return 155
		}
	}

	list := []string{strconv.Itoa(next), fingerAddress}
	if err := call(localAddress, "Server.SetFinger", list, &junk); err != nil {
		log.Fatalf("Calling Server.SetFinger: %v", err)
	}
	return next

}
func (s Server) FloodFinger(address string, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		for i := 1; i < 156; i++ {
			n.Finger[i] = address
		}
		//log.Printf("Finger[1]...Finger[155] = %v\n", address)
		finished <- struct{}{}
	}
	<-finished
	return nil
}
func (s Server) SetFinger(list []string, reply *Nothing) error {
	finished := make(chan struct{})
	next, err := strconv.Atoi(list[0])
	if err != nil {
		log.Fatalf("Calling Server.SetFingerAtoi: %v", err)
	}
	s <- func(n *Node) {
		n.Finger[next] = list[1]
		//log.Printf("Finger[%v] = %v\n", next, list[1])
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (n *Node) CheckPred() {
	var junk Nothing
	var reply string
	//var junk1 []string
	//doDump(junk1)
	var node Node
	localAddress := getAddress()
	if err := call(localAddress, "Server.Dump", junk, &node); err != nil {
		log.Fatalf("Calling Server.Dump (Stabilize): %v", err)
	}
	//log.Printf("PRED: %v", node.Pred)
	if err := call(node.Pred, "Server.Ping", junk, &reply); err != nil {
		if err := call(localAddress, "Server.SetPred", "", &junk); err != nil {
			log.Fatalf("Calling Server.SetPred: %v", err)
		}
	}
}
func (s Server) SetPred(pred string, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		n.Pred = pred
		finished <- struct{}{}
	}
	<-finished
	return nil
}
func (s Server) Ping(junk *Nothing, reply *string) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		address := getAddress()
		*reply = "Pinged Node at " + address
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func startActor(node *Node) Server {
	ch := make(chan Handler)
	//state := new(Feed)
	go func() {
		for f := range ch {
			f(node)
		}
	}()
	return ch
}
func server(address string, node *Node) {
	actor := startActor(node)
	rpc.Register(actor)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", address)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	if err := http.Serve(l, nil); err != nil {
		log.Fatalf("http.Serve: %v", err)
	}
}

func call(address string, method string, request interface{}, response interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Printf("rpc.DialHTTP: %v", err)
		return err
	}
	defer client.Close()

	if err := client.Call(method, request, response); err != nil {
		log.Printf("client.Call: %s: %v", method, err)
		return err
	}
	return nil
}

func client(address string) {

	var junk Nothing
	if err := call(address, "Server.Post", "Hello Again!", &junk); err != nil {
		log.Fatalf("client.Call: Post: %v", err)
	}
	if err := call(address, "Server.Post", "Heidee HO!", &junk); err != nil {
		log.Fatalf("client.Call: Post: %v", err)
	}

	var lst []string
	if err := call(address, "Server.Get", 5, &lst); err != nil {
		log.Fatalf("client.Call: Get: %v", err)
	}
	for _, elt := range lst {
		log.Println(elt)
	}

}

func shell(address string) {
	log.Printf("Starting shell")
	log.Printf("Commands are get, post")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		input = strings.TrimSpace(input)

		commandFields := strings.SplitN(input, " ", 2)
		if len(commandFields) > 1 {
			commandFields[1] = strings.TrimSpace(commandFields[1])
		}

		if len(commandFields) == 0 {
			continue
		}
		switch commandFields[0] {
		case "get":
			n := 10
			if len(commandFields) == 2 {
				var err error
				if n, err = strconv.Atoi(commandFields[1]); err != nil {
					log.Fatalf("string conver failure: %v", err)
				}
			}
			var messages []string
			if err := call(address, "Server.Get", n, &messages); err != nil {
				log.Fatalf("Calling Server.Get: %v", err)
			}
			for _, elt := range messages {
				log.Println(elt)
			}
		case "post":
			if len(commandFields) != 2 {
				log.Printf("You must specify a message to post")
				continue
			}
			var junk Nothing
			if err := call(address, "Server.Post", commandFields[1], &junk); err != nil {
				log.Fatalf("Calling Server.Post: %v", err)
			}
		default:
			log.Printf("Commands are 'get' or 'post'")
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}

func printUsage() {
	log.Printf("Usage: %s [-server or -client] [address]", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

/*func (n Node) Stablize(localAddress string) {
	var succPred string
	succ := n.Succ[0]
	log.Printf("sending GetPred request to node: %v\n", succ)
	if err := call(succ, "Server.GetPred", &junk, &succPred); err != nil {
		log.Fatalf("Calling Server.GetPred: %v", err)
	}
	succPredInt := hashString(succPred)
	nInt := hashString(localAddress)
	succInt := hashString(succ)
	if between(nInt,succPredInt,succInt,false){
		n.Succ[0] = succPred
	}
}*/
