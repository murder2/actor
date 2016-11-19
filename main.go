package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

import "fmt"


type ActionType int

const (
	ActionTypePrint = iota
)

type ActionPut struct {
	Type	ActionType `json:"type" binding:"required"`
}

type ActorsPost struct {
	Capabilities	[]string `json:"capabilities"`
	Port		int `json:"port"`
}


var actions map[string]ActionType


func config(c *gin.Context) {
        c.JSON(http.StatusBadRequest, gin.H{
            "message": "not implemented",
        })
}


func actionsPut(c *gin.Context) {
	var uuid string

	uuid = c.Param("uuid")

	if _, found := actions[uuid]; !found {
		c.Status(http.StatusBadRequest)
		return
	}

	var json ActionPut
	if err := c.BindJSON(&json); err != nil {
		return
	}

	if json.Type != ActionTypePrint {
		c.Status(http.StatusBadRequest)
		return
	}

	actions[uuid] = ActionTypePrint
	c.Status(http.StatusOK)
}

func actionsGet(c *gin.Context) {
	var uuid string

	uuid = c.Param("uuid")

	if _, found := actions[uuid]; !found {
		c.Status(http.StatusBadRequest)
		return
	}

	action_type, _ := actions[uuid]
	switch action_type {
	case ActionTypePrint:
		fmt.Println(uuid)
		c.Status(http.StatusOK)
	default:
		c.Status(http.StatusBadRequest)
	}
}

func actionsDelete(c *gin.Context) {
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

	res, err := http.Post("http://" + os.Args[1] + "/actors/", "application/json", bytes.NewBuffer(dump))

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

	actions = make(map[string]ActionType)

	router := gin.Default()

	router.POST("/config", config)

	{
		actions := router.Group("/actions")
		actions.GET("/:uuid", actionsGet)
		actions.PUT("/:uuid", actionsPut)
		actions.DELETE("/:uuid", actionsDelete)
	}

	router.Run()
}
