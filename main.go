package main

import (
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

type UserInfo struct {
	createdAt string
	subs      []string
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

func appendSub(u *UserInfo, sub string) {
	u.subs = append(u.subs, sub)
}

func parseUsers(users []*User) map[string]*UserInfo {
	info := make(map[string]*UserInfo)

	for _, user := range users {
		for _, sub := range user.Subs {
			if _, ok := info[sub.Email]; !ok {
				info[sub.Email] = &UserInfo{sub.CreatedAt, make([]string, 0)}
			}

			appendSub(info[sub.Email], user.Email)
		}
	}

	return info
}

func breadthFirstSearch(start string, info map[string]*UserInfo) (map[string]struct{}, map[string]string) {
	visited := make(map[string]struct{})

	var yes struct{}
	visited[start] = yes

	parent := make(map[string]string)
	parent[start] = StartParent

	queue := make([]string, 0)
	queue = append(queue, start)

	for len(queue) > 0 {
		email := queue[First]

		if user, ok := info[email]; ok {
			for _, to := range user.subs {
				if _, ok := visited[to]; !ok {
					visited[to] = yes

					queue = append(queue, to)
					parent[to] = email
				}
			}
		}

		queue = queue[1:]
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

func findShortestPath(from string, to string, info map[string]*UserInfo) []string {
	visited, parent := breadthFirstSearch(from, info)

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

func makePathInfo(id int, from string, to string, path []string, info map[string]*UserInfo) PathInfo {
	if path == nil || from == to {
		return PathInfo{id, from, to, nil}
	}

	plots := make([]Subscriber, 0)

	for i := 1; i < len(path)-1; i++ {
		email := path[i]
		date := info[email].createdAt
		plots = append(plots, Subscriber{email, date})
	}

	return PathInfo{id, from, to, plots}
}

func findPaths(betweenUsers [][]string, users []*User) []PathInfo {
	info := parseUsers(users)
	paths := make([]PathInfo, 0)
	id := StartID

	for _, between := range betweenUsers {
		path := findShortestPath(between[First], between[Second], info)
		pathInfo := makePathInfo(id, between[First], between[Second], path, info)
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
