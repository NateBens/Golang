package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"

	"golang.org/x/crypto/pbkdf2"

	crand "crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
)

func checkForUser(tx *sql.Tx, username string) (bool, error) {
	var name string
	dbuser := tx.QueryRow("SELECT name FROM players WHERE name=?", username)
	switch err := dbuser.Scan(&name); err {
	case sql.ErrNoRows:
		fmt.Printf("User %v does not exist\n", username)
		return false, tx.Commit()
	case nil:
		fmt.Printf("User %v found!\n", username)
		return true, tx.Commit()
	default:
		fmt.Printf("Error checking for user.\n")
		tx.Rollback()
		return false, err
	}
}

func loginUser(tx *sql.Tx, password string, username string) (bool, error) {
	dbplayer, err := tx.Query("SELECT salt,hash FROM players WHERE name=?", username)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	defer dbplayer.Close()
	var salt64, hash64 string
	for dbplayer.Next() {

		if err := dbplayer.Scan(&salt64, &hash64); err != nil {
			tx.Rollback()
			return false, err
		}
	}
	salt, err := base64.StdEncoding.DecodeString(salt64)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	hash, err := base64.StdEncoding.DecodeString(hash64)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	inputhash := pbkdf2.Key(
		[]byte(password),
		salt,
		64*1024,
		32,
		sha256.New)
	if subtle.ConstantTimeCompare(hash, inputhash) != 1 {
		return false, tx.Commit()
	} else {
		return true, tx.Commit()
	}
}
func createUser(tx *sql.Tx, password string, username string) error {
	salt := make([]byte, 32)
	_, err := crand.Read(salt)
	if err != nil {
		log.Print("Error creating salt: ", err)
	}
	salt64 := base64.StdEncoding.EncodeToString(salt)
	hash := pbkdf2.Key(
		[]byte(password),
		salt,
		64*1024,
		32,
		sha256.New)
	hash64 := base64.StdEncoding.EncodeToString(hash)

	tx.Exec("INSERT INTO players (name,salt,hash) VALUES (?,?,?)", username, salt64, hash64)
	return tx.Commit()
}

func getRoom(tx *sql.Tx, rooms []*Room, ID int) (*Room, error) {
	dbroom, err := tx.Query("SELECT id, zone_id, name, description FROM rooms WHERE id=?", ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer dbroom.Close()
	var room *Room
	for dbroom.Next() {
		var id, zone_id int
		var name, description string
		if err := dbroom.Scan(&id, &zone_id, &name, &description); err != nil {
			tx.Rollback()
			return nil, err
		}
		for i := range rooms {
			if rooms[i].ID == id {
				room = rooms[i]
			}
		}
	}
	if err := dbroom.Err(); err != nil {
		tx.Rollback()
		return nil, err
	}
	return room, tx.Commit()
}
func readInZones(zones map[int]Zone, tx1 *sql.Tx) (map[int]Zone, error) {
	dbzones, err := tx1.Query("SELECT id, name FROM zones ORDER BY id")
	if err != nil {
		tx1.Rollback()
		return nil, fmt.Errorf("reading a room from the database: %v", err)
	}

	defer dbzones.Close()
	for dbzones.Next() {
		var id int
		var name string
		if err := dbzones.Scan(&id, &name); err != nil {
			tx1.Rollback()
			return nil, err
		}
		rooms := make([]*Room, 0)
		z := Zone{id, name, rooms}
		zones[id] = z
	}
	if err := dbzones.Err(); err != nil {
		tx1.Rollback()
		return nil, err
	}
	return zones, tx1.Commit()
}
func readInRooms(tx *sql.Tx, zones map[int]Zone) ([]*Room, error) {
	dbrooms, err := tx.Query("SELECT id, zone_id, name, description FROM rooms ORDER BY id")
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer dbrooms.Close()
	var rooms []*Room
	for dbrooms.Next() {
		var id, zone_id int
		var name, description string
		if err := dbrooms.Scan(&id, &zone_id, &name, &description); err != nil {
			tx.Rollback()
			return nil, err
		}
		z := zones[zone_id]
		zone := &z
		var exits [6]Exit
		room := Room{id, zone, name, description, exits}
		rooms = append(rooms, &room)
	}
	if err := dbrooms.Err(); err != nil {
		tx.Rollback()
		return nil, err
	}
	return rooms, tx.Commit()
}
func readInExits(tx *sql.Tx, rooms []*Room) ([]*Room, error) {
	dbexits, err := tx.Query("SELECT from_room_id, to_room_id, direction, description FROM exits ORDER BY from_room_id")
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer dbexits.Close()
	dirs := [6]string{"n", "e", "w", "s", "u", "d"}
	for dbexits.Next() {
		var from_room_id, to_room_id int
		var direction, description string
		if err := dbexits.Scan(&from_room_id, &to_room_id, &direction, &description); err != nil {
			tx.Rollback()
			return nil, err
		}
		exit := Exit{nil, description}
		for i := range rooms {
			if rooms[i].ID == to_room_id {
				exit.To = rooms[i]
			}
		}
		var dirIndex int
		for i := range dirs {
			if dirs[i] == direction {
				dirIndex = i
			}
		}
		for i := range rooms {
			if rooms[i].ID == from_room_id {
				rooms[i].Exits[dirIndex] = exit
			}
		}

	}
	if err := dbexits.Err(); err != nil {
		tx.Rollback()
		return nil, err
	}
	return rooms, tx.Commit()
}
func fillDirectionLookup() {
	dirLookup["n"] = 0
	dirLookup["north"] = 0
	dirLookup["e"] = 1
	dirLookup["east"] = 1
	dirLookup["w"] = 2
	dirLookup["west"] = 2
	dirLookup["s"] = 3
	dirLookup["south"] = 3
	dirLookup["u"] = 4
	dirLookup["up"] = 4
	dirLookup["d"] = 5
	dirLookup["down"] = 5
	dirLookupInt[0] = "north"
	dirLookupInt[1] = "east"
	dirLookupInt[2] = "west"
	dirLookupInt[3] = "south"
	dirLookupInt[4] = "up"
	dirLookupInt[5] = "down"
}
