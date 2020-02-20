package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Subscriber struct {
	Email     string `json:"Email"`
	CreatedAt string `json:"Created_at"`
}

type User struct {
	Nick      string       `json:"Nick"`
	Email     string       `json:"Email"`
	CreatedAt string       `json:"Created_at"`
	Subs      []Subscriber `json:"Subscribers"`
}

func readJSONFile(filename string) ([]User, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("unable to read input json file " + filename)
	}
	defer f.Close()

	byteData, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.New("unable to read input json file as a byte array " + filename)
	}

	users := make([]User, 0)
	json.Unmarshal(byteData, &users)

	return users, nil
}

func main() {
	users, err := readJSONFile("users.json")
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(users); i++ {
		fmt.Println(users[i].Email)
	}
}
