## Golang
This is a repo to showcase some projects I have done in Go. Note that all projects require the open source package [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3). This package can be installed with the go get command:

    go get github.com/mattn/go-sqlite3
Some projects also require the package golang.org/x/crypto/pbkdf2 which can also be installed with the 'go get' command.
To run a project, use
    
    go build
and run the created executable from your command line
Information on how to use each project is found in their respective folders.

# [MUD](src/mud)
Multi-User-Dungeon. A text based world where users can create accounts and interact on a server. Essentially a shell of an online text-based mmorpg.

# [MapReduce](src/mapreduce)
A distributed system to spread large jobs across multiple worker nodes. 

# [Paxos](src/paxos)
*Paxos is a family of protocols for solving consensus in a network of unreliable or fallible processors*. This program stores consistant key value pairs across multiple machines.  

# [Chord](src/chord)
A distributed hash table. 
