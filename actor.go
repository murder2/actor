package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/exec"
)

import "fmt"


type ActionType int
const (
	ActionTypePrint ActionType = iota
	ActionTypeSound ActionType = iota
)
func ActionTypeStringIsValid(t string) bool {
	switch t {
		case "print":
		case "sound":
		default:
			return false
	}

	return true
}
func ActionTypeStringToActionType(t string) ActionType {
	switch t {
		case "print":
			return ActionTypePrint
		case "sound":
			return ActionTypeSound
	}

	panic("Check for it before")
}

type ActionPut struct {
	Type		string `json:"type" binding:"required"`
	Name		string `json:"name"`

	SoundFile	string `json:"sound_file"`
}

type ActorsPost struct {
	Capabilities	[]string `json:"capabilities"`
	Port		int `json:"port"`
}

type Action struct {}

var actions map[string]func()


func config(c *gin.Context) {
        c.JSON(http.StatusBadRequest, gin.H{
            "message": "not implemented",
        })
}


var still_running bool
func playSound(path string) {
	if (still_running) {
		return
	}

	still_running = true

	cmd := exec.Command("mpv", path)

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		still_running = false
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
		still_running = false
	}

	still_running = false
}

func (a *Action) Put(c *gin.Context) {
	var uuid string

	uuid = c.Param("uuid")

	if _, found := actions[uuid]; found {
		log.Print(uuid, ": not found in actions")
		c.Status(http.StatusBadRequest)
		return
	}

	var json ActionPut
	if err := c.BindJSON(&json); err != nil {
		log.Print("fail to parse json: ", err)
		c.Status(http.StatusBadRequest)
		return
	}

	if !ActionTypeStringIsValid(json.Type) {
		log.Print("invalid type: ", json.Type)
		c.Status(http.StatusBadRequest)
		return
	}

	var action func()
	switch ActionTypeStringToActionType(json.Type) {
	case ActionTypePrint:
		action = func() {
			fmt.Println(uuid)
		}
	case ActionTypeSound:
		action = func() {
			playSound(json.SoundFile)
		}
	default:
		log.Print("unknown type", json.Type)
		c.Status(http.StatusBadRequest)
		return
	}

	actions[uuid] = action
	c.Status(http.StatusOK)
}

func (a *Action) Get(c *gin.Context) {
	var uuid string

	uuid = c.Param("uuid")

	if _, found := actions[uuid]; !found {
		c.Status(http.StatusBadRequest)
		return
	}

	actions[uuid]()

	c.Status(http.StatusOK)
}

func (a *Action) Delete(c *gin.Context) {
	var uuid string

	uuid = c.Param("uuid")

	if _, found := actions[uuid]; !found {
		c.Status(http.StatusBadRequest)
		return
	}

	delete(actions, uuid)
	c.Status(http.StatusOK)
}

func advertise(self ActorsPost) {
	var dump []byte

	dump, _ = json.Marshal(self)

	res, err := http.Post("http://" + os.Args[1] + "/actors", "application/json", bytes.NewBuffer(dump))

	if err != nil || res.StatusCode != http.StatusOK {
		log.Fatal(err, res)
		return
	}
}


func main() {
	var self = ActorsPost{
		Capabilities: []string{"print"},
		Port: 8080, // TODO do not hardcode
	}
	advertise(self)

	actions = make(map[string]func())

	router := gin.Default()

	//router.POST("/config", config)

	{
		var action *Action = new(Action)
		actions := router.Group("/actions")
		actions.GET("/:uuid", action.Get)
		actions.PUT("/:uuid", action.Put)
		actions.DELETE("/:uuid", action.Delete)
	}

	router.Run()
}
