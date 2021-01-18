package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Zone struct {
	ID    int
	Name  string
	Rooms []*Room
}
type Room struct {
	ID          int
	Zone        *Zone
	Name        string
	Description string
	Exits       [6]Exit
}
type Exit struct {
	To          *Room
	Description string
}
type Player struct {
	Name     string
	Location *Room
	Homebase *Room
	Outputs  chan OutputEvent
}
type InputEvent struct {
	Player  *Player
	Command []string
	Close   bool
	Login   bool
}
type OutputEvent struct {
	Text  string
	Close bool
}

var commands = make(map[string]func(*Player, []string))
var db *sql.DB
var dirLookup = make(map[string]int)
var dirLookupInt = make(map[int]string)
var players = make(map[string]*Player)

func main() {
	log.Printf("MUD Started.")
	fillDirectionLookup()
	fillCommands()
	db, err := sql.Open("sqlite3", "./world.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//Begin a Transaction to read in Zones
	tx1, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	//Read in the zones
	var zones = make(map[int]Zone)
	zones, err = readInZones(zones, tx1)
	if err != nil {
		tx1.Rollback()
		log.Fatal(err)
	}
	/*for k := range zones {
		fmt.Printf("key[%s] value[%s]\n", k, zones[k])
	}*/
	//Begin another transaction to read in all the rooms
	tx2, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	//Read in the rooms
	rooms := make([]*Room, 0)
	rooms, err = readInRooms(tx2, zones)
	if err != nil {
		tx2.Rollback()
		log.Fatal(err)
	}
	/*for _, room := range rooms {
		fmt.Printf("room:[%s], Zone:[%s]\n", room.Name, room.Zone)
	}*/
	//Begin another transaction to read in all the exits
	tx3, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	//Read in all the exits
	rooms, err = readInExits(tx3, rooms)
	if err != nil {
		tx3.Rollback()
		fmt.Printf("Error reading in exits")
		log.Fatal(err)
	}
	/*fmt.Printf("room:[%s]\n", rooms[1].Name)
	for _, exit := range rooms[1].Exits {
		fmt.Printf("exit:[%s]\n", exit.Description)
	}*/
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	startRoom, err := getRoom(tx, rooms, 3001)
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	//creating a server with the listen function
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}

	/*closingChannel := make(chan string)
	go func() {
		for {
			s := <-closingChannel
			log.Print(s)
			log.Print("Closing Channel")
		}
	}()*/

	inputs := make(chan InputEvent)

	//Connection Goroutine listener
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatalf("Accept error: %v", err)
				continue
			}
			log.Printf("Incoming connection from %s", conn.RemoteAddr())
			fmt.Fprintf(conn, "Welcome to the MUD!\n")
			//Logging in a player
			scanner1 := bufio.NewScanner(conn)
			fmt.Fprint(conn, "Username: ")
			scanner1.Scan()
			text := scanner1.Text()
			//text = strings.TrimSuffix(text, "\n")
			username := text
			//check if user exists
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			userexists, err := checkForUser(tx, username)
			if err != nil {
				tx.Rollback()
				log.Print("Error checking if user exists")
				log.Fatal(err)
			}
			if userexists {
				for {
					tx2, err := db.Begin()
					if err != nil {
						log.Fatal(err)
					}
					fmt.Fprint(conn, "Password: ")
					scanner1.Scan()
					text = scanner1.Text()
					password := text
					correctPW, err := loginUser(tx2, password, username)
					if err != nil {
						log.Print("error logging in user")
						log.Fatal(err)
					}
					if correctPW == false {
						fmt.Fprint(conn, "Password Incorrect, try again.\n")
					} else {
						break
					}
				}
			} else {
				tx3, err := db.Begin()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Fprintf(conn, "Creating new user: %s\n", username)
				fmt.Fprintf(conn, "Password: ")
				scanner1.Scan()
				text = scanner1.Text()
				password := text
				err2 := createUser(tx3, password, username)
				if err2 != nil {
					log.Print("error creating user")
					log.Fatal(err)
				}
			}

			outputs := make(chan OutputEvent)
			player := &Player{username, startRoom, startRoom, outputs}
			inputevent := InputEvent{player, nil, false, true}
			inputs <- inputevent
			go handlePlayerConnection(db, conn, inputs, player)
		}
	}()

	//Start of the main goRoutine loop
	for {
		inputevent := <-inputs
		if inputevent.Login == true {
			if player, ok := players[inputevent.Player.Name]; ok {
				if player.Outputs != nil {
					close(player.Outputs)
					player.Outputs = nil
				}
			}
			players[inputevent.Player.Name] = inputevent.Player
		} else {
			if inputevent.Close == true {
				close(inputevent.Player.Outputs)
				inputevent.Player.Outputs = nil
				//players[inputevent.Player.Name] = nil
				delete(players, inputevent.Player.Name)
				//log.Print("Input Channel closed")
			} else {
				if inputevent.Player != nil {
					commandHandler(inputevent.Player, inputevent.Command)
				}
			}
		}
	}
}

func handlePlayerConnection(db *sql.DB, conn net.Conn, inputs chan<- InputEvent, player *Player) {
	//time.Sleep(30 * time.Second)
	fmt.Fprintf(conn, "Welcome %s! ", player.Name)
	scanner := bufio.NewScanner(conn)
	fmt.Fprintln(conn, "Enter a command or type 'quit' to quit.")
	go func() {
		for {
			outputevent, more := <-player.Outputs
			if more {
				fmt.Fprintf(conn, outputevent.Text)
			} else {
				log.Print("Output channel closed for user: ", player.Name)
				conn.Close()
				return
			}
		}
	}()
	for scanner.Scan() {
		input := scanner.Text()
		//fmt.Println(command) // Println will add back the final '\n'
		//fmt.Printf("Fields are: %q", strings.Fields(command))
		commandFields := strings.Fields(input)
		//fmt.Printf("%T\n", commandFields)
		if len(commandFields) == 0 {
			fmt.Fprintln(conn, "Please Enter a command\n")
		} else if commandFields[0] == "quit" || commandFields[0] == "Quit" {
			inputevent := InputEvent{player, commandFields, true, false}
			inputs <- inputevent
		} else {
			// pass commands to Main goroutine from here
			inputevent := InputEvent{player, commandFields, false, false}
			inputs <- inputevent
			//commandHandler(player, commandFields)
		}

		//commands["look"](input)

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
	}
}

func (p *Player) Printf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	p.Outputs <- OutputEvent{Text: msg}
}

/*for scanner.Scan() {
	input := scanner.Text()
	//fmt.Println(command) // Println will add back the final '\n'
	//fmt.Printf("Fields are: %q", strings.Fields(command))
	commandFields := strings.Fields(input)
	//fmt.Printf("%T\n", commandFields)
	if len(commandFields) == 0 {
		fmt.Fprintln(conn, "Please Enter a command")
	} else if commandFields[0] == "quit" || commandFields[0] == "Quit" {
		fmt.Fprintf(conn, "Goodbye!\n")
		conn.Close()
	} else {
		commandHandler(player, commandFields)
	}

	//commands["look"](input)

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}*/
