package main

import (
	"container/list"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	StartParent string = "startParent"
)

type void struct{}

var member void

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

func readJSONFile(filename string) ([]*User, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("unable to read input json file " + filename)
	}
	defer f.Close()

	byteData, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.New("unable to read input json file as a byte array " + filename)
	}

	var users []*User
	json.Unmarshal(byteData, &users)

	return users, nil
}

func readCSVFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("unable to read input file " + filePath)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()

	if err != nil {
		return nil, errors.New("unable to parse file as CSV for " + filePath)
	}

	return records, nil
}

func parseContacts(users []*User) map[string][]string {
	contacts := make(map[string][]string)
	for _, user := range users {
		for _, sub := range user.Subs {
			if _, ok := contacts[sub.Email]; !ok {
				contacts[sub.Email] = make([]string, 0)
			}
			contacts[sub.Email] = append(contacts[sub.Email], user.Email)
		}
	}

	return contacts
}


func breadthFirstSearch(start string, contacts map[string][]string) (map[string]void, map[string]string) {
	visited := make(map[string]void)
	visited[start] = member

	parent := make(map[string]string)
	parent[start] = StartParent

	queue := list.New()
	queue.PushBack(start)

	for queue.Len() > 0 {
		node := queue.Front()
		email := node.Value.(string)
		for _, to := range contacts[email] {
			if _, ok := visited[to]; !ok {
				visited[to] = member
				queue.PushBack(to)
				parent[to] = email
			}
		}
		queue.Remove(node)
	}

	return visited, parent
}

func reversePath(path []string) {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
}

func findShortestPath(from string, to string, contacts map[string][]string) []string {
	visited, parent := breadthFirstSearch(from, contacts)

	if _, ok := visited[to]; !ok {
		return nil
	}

	path := make([]string, 0)
	for v := to; v != StartParent; v = parent[v] {
		path = append(path, v)
	}

	reversePath(path)

	return path
}

func main() {
	users, err := readJSONFile("users.json")
	if err != nil {
		log.Fatal(err)
	}

	contacts := parseContacts(users)

	path := findShortestPath("mako1332@rambler.ru", "mosquito371@mail.ru", contacts)

	for _, el := range path {
		fmt.Println(el)
	}
}
