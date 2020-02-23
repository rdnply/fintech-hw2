package main

import (
	"container/list"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	First = iota
	Second
	StartParent = "startParent"
	StartID     = 1
)

type void struct{}

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

type PathInfo struct {
	ID   int          `json:"id"`
	From string       `json:"from"`
	To   string       `json:"to"`
	Path []Subscriber `json:"path,omitempty"`
}

func (s *Subscriber) MarshalJSON() ([]byte, error) {
	type alias struct {
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}

	var a alias = alias(*s)

	return json.Marshal(&a)
}

func readJSONFile(fileName string) ([]*User, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read input json file %s, %v", fileName, err)
	}
	defer f.Close()

	byteData, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read input json file as a byte array %s, %v", fileName, err)
	}

	var users []*User

	err = json.Unmarshal(byteData, &users)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal byte data in json %v", err)
	}

	return users, nil
}

func readCSVFile(fileName string) ([][]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file %s, %v", fileName, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to parse file as CSV for %s, %v", fileName, err)
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

	var yes void
	visited[start] = yes

	parent := make(map[string]string)
	parent[start] = StartParent

	queue := list.New()
	queue.PushBack(start)

	for queue.Len() > 0 {
		node := queue.Front()
		email := node.Value.(string)

		for _, to := range contacts[email] {
			if _, ok := visited[to]; !ok {
				visited[to] = yes

				queue.PushBack(to)

				parent[to] = email
			}
		}

		queue.Remove(node)
	}

	return visited, parent
}

// nolint: gomnd
func reversePath(path []string) []string {
	rev := make([]string, len(path))
	j := len(path) - 1

	for _, v := range path {
		rev[j] = v
		j--
	}

	return rev
}

func findShortestPath(from string, to string, contacts map[string][]string) []string {
	visited, parent := breadthFirstSearch(from, contacts)

	if _, ok := visited[to]; !ok {
		return nil
	}

	revPath := make([]string, 0)
	for v := to; v != StartParent; v = parent[v] {
		revPath = append(revPath, v)
	}

	normalPath := reversePath(revPath)

	return normalPath
}

func makePathInfo(id int, from string, to string, path []string, created map[string]string) PathInfo {
	if path == nil || from == to {
		return PathInfo{id, from, to, nil}
	}

	plots := make([]Subscriber, 0)

	for i := 1; i < len(path)-1; i++ {
		email := path[i]
		date := created[email]
		plots = append(plots, Subscriber{email, date})
	}

	return PathInfo{id, from, to, plots}
}

func findPaths(betweenUsers [][]string, users []*User) []PathInfo {
	contacts, created := parseUsers(users)
	paths := make([]PathInfo, 0)
	id := StartID

	for _, between := range betweenUsers {
		path := findShortestPath(between[First], between[Second], contacts)
		pathInfo := makePathInfo(id, between[First], between[Second], path, created)
		paths = append(paths, pathInfo)
		id++
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
