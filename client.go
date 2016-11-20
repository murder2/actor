package main

import (
	"net/http"
	"os"
	"bytes"
	"strconv"
	"log"
	"strings"
	"io/ioutil"
	"encoding/json"
	"sort"
)

type LinkEventPost struct {
	Major	int `type:"major"`
	Minor	int `type:"minor"`
}

type LinkActionPost struct {
	Actor	string `type:"actor"`
	Id	int `type:"id"`
}

type LinkPost struct {
	Event LinkEventPost `type:"event"`
	Action LinkActionPost `json:"action"`
}

func linkAdd(args []string) {
	major_minor := strings.Split(args[0], "/")
	if len(major_minor) != 2 {
                panic("major/minor")
	}
	major, _ := strconv.Atoi(major_minor[0])
	minor, _ := strconv.Atoi(major_minor[1])

	var link = LinkPost{
		LinkEventPost{major, minor},
		LinkActionPost{Actor: "", Id: 0},
	}
	var dump []byte

	dump, _ = json.Marshal(link)

	res, err := http.Post("http://" + os.Args[1] + "/link", "application/json", bytes.NewBuffer(dump))
        if err != nil || res.StatusCode != http.StatusOK {
                log.Fatal(err, res)
        }
}

func link(args []string) {
	switch args[0] {
	case "new": linkAdd(args[1:])
	default:
		panic("unknown cmd: " + args[2])
	}
}

type Action struct {
	Type	string `json:"type" binding:"required"`
	Name	string `json:"name" binding:"required"`
	SoundFile	string `json:"sound_file"`
	Id	int `json:"id"`
}

type Actor struct {
	Actions	[]Action `json:"actions" binding:"required"`;
	Id	string `json:"id" binding:"required"`;
	Ip	string `json:"ip" binding:"required"`;
	Port	int `json:"port" binding:"required"`;
	Capabilities	[]string `json:"capabilities" binding:"required"`
}

type ActorList struct {
	Actors	[]Actor `json:"actors" binding:"required"`
}

type ById []Actor

func (a ById) Len() int { return len(a) }
func (a ById) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ById) Less(i, j int) bool { return a[i].Id < a[j].Id }

func (a Action) String() string {
	return string(a.Id) + "\t" + a.Name + "\t" + a.Type
}

func actionsToStrings(actions []Action) []string {
	ret := make([]string, len(actions))

	for i, a := range(actions) {
		ret[i] = a.String()
	}

	return ret
}

func actorList() {
	res, err := http.Get("http://" + os.Args[1] + "/actors")

        if err != nil || res.StatusCode != http.StatusOK {
                log.Fatal(err, res)
        }
	defer res.Body.Close()

	raw, err := ioutil.ReadAll(res.Body)
        if err != nil {
                log.Fatal(err, res)
        }

	var list ActorList
	err = json.Unmarshal(raw, &list)
        if err != nil {
                log.Fatal(err, string(raw))
        }

	sort.Sort(ById(list.Actors))

	for _, a := range(list.Actors) {
		print(a.Id, "\t",
			a.Ip, ":", a.Port, "\t",
			strings.Join(a.Capabilities, ", "), "\t",
			strings.Join(actionsToStrings(a.Actions), ", "), "\n",
		)
	}
}

func actor(args []string) {
	switch args[0] {
	case "list": actorList()
	default:
		panic("unknown cmd: " + args[0])
	}
}

func main() {
	remains := os.Args[3:]

	switch os.Args[2] {
	case "link": link(remains)
	case "actor": actor(remains)
	default:
		panic("unknown cmd: " + os.Args[1])
	}
}
