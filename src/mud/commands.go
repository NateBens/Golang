package main

import "strings"

func commandHandler(player *Player, commandFields []string) {
	command := commandFields[0]
	value, ok := commands[command]
	if ok {
		value(player, commandFields[1:])
	} else {
		player.Printf("[%s] Command not recognized\n", command)
	}
}
func addCommand(command string, action func(*Player, []string)) {
	commands[command] = action
}

func doLook(player *Player, args []string) {
	//fmt.Printf(player.Name)
	//fmt.Printf(" looked ... ")
	//fmt.Println(args)
	currentRoom := player.Location
	if len(args) == 0 {
		player.Printf("Current Room: %s\n", currentRoom.Name)
		player.Printf("%s", currentRoom.Description)
		player.Printf("Exits:\n")
		for i := 0; i <= 5; i++ {
			if currentRoom.Exits[i].Description != "" {
				player.Printf(" %s", dirLookupInt[i])
				player.Printf("\n")
			}
		}
		player.Printf("Players:\n")
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location {
				player.Printf(" %v\n", otherplayer.Name)
			}
		}

		player.Printf("\n")
	} else {
		dir := args[0]
		dirInt := dirLookup[dir]
		if currentRoom.Exits[dirInt].Description != "" {
			player.Printf("%s\n", currentRoom.Exits[dirInt].Description)
		} else {
			player.Printf("There is nothing interesting that way.")
		}
	}

}
func goNorth(player *Player, args []string) {
	currentRoom := player.Location
	if currentRoom.Exits[0].Description != "" {
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location && otherplayer != player {
				otherplayer.Printf("%v went North\n", player.Name)
			}
		}
		player.Location = currentRoom.Exits[0].To
		player.Printf("%s\n", player.Location.Description)
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location && otherplayer != player {
				otherplayer.Printf("%v came from the South\n", player.Name)
			}
		}
	} else {
		player.Printf("You cannot go that way.")
	}
}
func goEast(player *Player, args []string) {
	currentRoom := player.Location
	if currentRoom.Exits[1].Description != "" {
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location && otherplayer != player {
				otherplayer.Printf("%v went East\n", player.Name)
			}
		}
		player.Location = currentRoom.Exits[1].To
		player.Printf("%s\n", player.Location.Description)
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location && otherplayer != player {
				otherplayer.Printf("%v came from the West\n", player.Name)
			}
		}
	} else {
		player.Printf("You cannot go that way.")
	}
}
func goWest(player *Player, args []string) {
	currentRoom := player.Location
	if currentRoom.Exits[2].Description != "" {
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location && otherplayer != player {
				otherplayer.Printf("%v went West\n", player.Name)
			}
		}
		player.Location = currentRoom.Exits[2].To
		player.Printf("%s\n", player.Location.Description)
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location && otherplayer != player {
				otherplayer.Printf("%v came from the East\n", player.Name)
			}
		}
	} else {
		player.Printf("You cannot go that way.")
	}
}
func goSouth(player *Player, args []string) {
	currentRoom := player.Location
	if currentRoom.Exits[3].Description != "" {
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location && otherplayer != player {
				otherplayer.Printf("%v went South\n", player.Name)
			}
		}
		player.Location = currentRoom.Exits[3].To
		player.Printf("%s\n", player.Location.Description)
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location && otherplayer != player {
				otherplayer.Printf("%v came from the North\n", player.Name)
			}
		}
	} else {
		player.Printf("You cannot go that way.")
	}
}
func goUp(player *Player, args []string) {
	currentRoom := player.Location
	if currentRoom.Exits[4].Description != "" {
		player.Location = currentRoom.Exits[4].To
		player.Printf("%s\n", player.Location.Description)
	} else {
		player.Printf("You cannot go that way.")
	}
}
func goDown(player *Player, args []string) {
	currentRoom := player.Location
	if currentRoom.Exits[5].Description != "" {
		player.Location = currentRoom.Exits[5].To
		player.Printf("%s\n", player.Location.Description)
	} else {
		player.Printf("You cannot go that way.")
	}
}
func playerRecall(player *Player, args []string) {
	for _, otherplayer := range players {
		if otherplayer.Location == player.Location && otherplayer != player {
			otherplayer.Printf("%v recalled\n", player.Name)
		}
	}
	player.Location = player.Homebase
	player.Printf("%s\n", player.Location.Description)
}
func doGossip(player *Player, args []string) {
	if len(args) == 0 {
		player.Printf("You must type a message to gossip!\n")
	} else {
		message := strings.Join(args, " ")
		//name := player.Name
		//name = strings.TrimSuffix(name, "\n")
		//player.Printf(name)
		//player.Printf(player.Name)
		for _, otherplayer := range players {
			otherplayer.Printf("%v:[Gossip] %v \n", player.Name, message)
		}
	}
}
func doSay(player *Player, args []string) {
	if len(args) == 0 {
		player.Printf("You must type a message to say!\n")
	} else {
		message := strings.Join(args, " ")
		for _, otherplayer := range players {
			if otherplayer.Location == player.Location {
				otherplayer.Printf("%v:[Say] %v \n", player.Name, message)
			}
		}
	}
}
func doTell(player *Player, args []string) {
	if len(args) < 2 {
		player.Printf("You must type a players name and a message to tell!\n")
	} else {
		if otherplayer, ok := players[args[0]]; ok {
			message := strings.Join(args[1:], " ")
			otherplayer.Printf("%v:[Tell] %v \n", player.Name, message)
		} else {
			player.Printf("Player '%s' not found\n", args[0])
		}
	}
}
func doShout(player *Player, args []string) {
	if len(args) == 0 {
		player.Printf("You must type a message to shout!\n")
	} else {
		message := strings.Join(args, " ")
		for _, otherplayer := range players {
			if otherplayer.Location.Zone.ID == player.Location.Zone.ID {
				otherplayer.Printf("%v: [Shout] %v \n", player.Name, message)
			}
		}
	}
}
func doWhere(player *Player, args []string) {
	for _, otherplayer := range players {
		if otherplayer.Location.Zone.ID == player.Location.Zone.ID {
			player.Printf("Player: %v Room: %v \n", otherplayer.Name, otherplayer.Location.Name)
		}
	}

}
func doLeeroy(player *Player, args []string) {
	for _, otherplayer := range players {
		if otherplayer.Location == player.Location {
			otherplayer.Printf("%v Let's out a powerful shout of \nLEEEEEERRROOOOOOOOYYYYY MMJJJEENNNKKKIINNNSSS!\nYou know whatever he is about to do will be bold, mighty and reckless.\n", player.Name)
		}
	}
}
func doNee(player *Player, args []string) {
	if len(args) == 0 {
		player.Printf("You must designate a player to say NEE to!\n")
	} else {
		if otherplayer, ok := players[args[0]]; ok {
			if otherplayer.Location == player.Location {
				otherplayer.Printf("%v: whispers 'NEE!' into your ear, causing you to wince in pain.\n", player.Name)
				player.Printf("You whisper 'NEE!' into %v's ear, causing them to wince in pain.\n", args[0])
			} else {
				player.Printf("You can only say NEE to someone in the same room as you\n")
			}
		} else {
			player.Printf("Player '%s' not found\n", args[0])
		}
	}
}
func doCoconuts(player *Player, args []string) {
	for _, otherplayer := range players {
		if otherplayer.Location == player.Location {
			otherplayer.Printf("%v claps two halves of a coconut together and trots as if he's riding a horse,\n you get the feeling he is on a very important quest. \n", player.Name)
		}
	}
}

func fillCommands() {
	addCommand("look", doLook)
	addCommand("l", doLook)
	addCommand("north", goNorth)
	addCommand("east", goEast)
	addCommand("west", goWest)
	addCommand("south", goSouth)
	addCommand("up", goUp)
	addCommand("down", goDown)
	addCommand("n", goNorth)
	addCommand("e", goEast)
	addCommand("w", goWest)
	addCommand("s", goSouth)
	addCommand("u", goUp)
	addCommand("d", goDown)
	addCommand("recall", playerRecall)
	addCommand("gossip", doGossip)
	addCommand("say", doSay)
	addCommand("tell", doTell)
	addCommand("shout", doShout)
	addCommand("where", doWhere)
	addCommand("leeroy", doLeeroy)
	addCommand("nee", doNee)
	addCommand("coconuts", doCoconuts)
}
