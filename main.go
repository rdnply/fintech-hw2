package main

import (
	"container/list"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
)

const (
	StartParent string = "startParent"
	First       int    = 0
	Second      int    = 1
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

type PathElement struct {
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type PathInfo struct {
	ID   int           `json:"id"`
	From string        `json:"from"`
	To   string        `json:"to"`
	Path []PathElement `json:"path"`
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

func parseUsers(users []*User) (map[string][]string, map[string]string) {
	contacts := make(map[string][]string)
	created := make(map[string]string)
	for _, user := range users {
		created[user.Email] = user.CreatedAt
		for _, sub := range user.Subs {
			if _, ok := contacts[sub.Email]; !ok {
				contacts[sub.Email] = make([]string, 0)
			}
			contacts[sub.Email] = append(contacts[sub.Email], user.Email)
		}
	}

	return contacts, created
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

func makePathInfo(id int, from string, to string, path []string, created map[string]string) PathInfo {
	plots := make([]PathElement, 0)
	for i := 1; i < len(path)-1; i++ {
		email := path[i]
		date := created[email]
		plots = append(plots, PathElement{email, date})
	}

	return PathInfo{id, from, to, plots}
}

func findPaths(betweenUsers [][]string, users []*User) []PathInfo {
	contacts, created := parseUsers(users)
	paths := make([]PathInfo, 0)
	for id, between := range betweenUsers {
		path := findShortestPath(between[First], between[Second], contacts)
		pathInfo := makePathInfo(id+1, between[First], between[Second], path, created)
		paths = append(paths, pathInfo)
	}

	return paths
}

func main() {
	users, err := readJSONFile("users.json")
	if err != nil {
		log.Fatal(err)
	}

	betweenUsers, err := readCSVFile("input.csv")
	if err != nil {
		log.Fatal(err)
	}

	paths := findPaths(betweenUsers, users)

	jsonPaths, err := json.Marshal(paths)
	if err != nil {
		log.Fatalf("can't write as json format ")
	}

	err = ioutil.WriteFile("result.json", jsonPaths, 0644)
	if err != nil {
		log.Fatalln("can't write json data in file")
	}
}
