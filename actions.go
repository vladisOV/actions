package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli"
)

type Action struct {
	Id          string
	Description string
	Result      string
	Timestamp   string
}

type AuthResponse struct {
	token string
}

var authToken = ""

func main() {
	app := cli.NewApp()
	app.Name = "actions"
	app.Usage = "add/get/edit your actions"
	app.Version = "1.0.0"
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "get list of actions",
			Action: func(c *cli.Context) error {
				actions := getAllActions()
				printActions(actions)
				return nil
			},
		},
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "add new action",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "desc, d"},
				cli.StringFlag{Name: "result, r"},
			},
			Action: func(c *cli.Context) error {
				saved := createAction(c.String("desc"), c.String("result"))
				if saved != (Action{}) {
					printDelimiter()
					printAction(saved)
				}
				return nil
			},
		},
		{
			Name:    "login",
			Aliases: []string{"a"},
			Usage:   "Log in actions-api",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "username, u"},
				cli.StringFlag{Name: "password, pw"},
			},
			Action: func(c *cli.Context) error {
				authToken = authenticate(c.String("username"), c.String("password"))
				fmt.Printf("token : %s\n", authToken)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getAllActions() []Action {
	response, err := http.Get("http://localhost:8080/api/item")
	if err != nil {
		fmt.Printf("Request to actions storage failed with error %s\n", err)
		return nil
	}
	data, _ := ioutil.ReadAll(response.Body)

	var actions []Action
	json.Unmarshal([]byte(data), &actions)
	return actions
}

func createAction(desc string, result string) Action {
	jsonData := map[string]string{"description": desc, "result": result,
		"timestamp": time.Now().Format("2006-01-02T15:04:05.999")}
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post("http://localhost:8080/api/item", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("Request to actions storage failed with error %s\n", err)
		return Action{}
	}
	data, _ := ioutil.ReadAll(response.Body)
	var action Action
	json.Unmarshal([]byte(data), &action)
	return action
}

func authenticate(username string, password string) string {
	jsonData := map[string]string{"username": username, "password": password}
	fmt.Printf("Request  %s\n", jsonData)

	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post("http://localhost:8080/auth", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("Authentication failed with error %s\n", err)
		return ""
	}
	data, _ := ioutil.ReadAll(response.Body)
	var authResponse AuthResponse
	json.Unmarshal([]byte(data), &authResponse)
	fmt.Printf("token : %s\n", authToken)

	return authResponse.token
}

func printActions(actions []Action) {
	len := len(actions)
	if len > 0 {
		printDelimiter()
	}
	for _, action := range actions {
		printAction(action)
	}
}

func printAction(action Action) {
	fmt.Printf("Id : %s\n", action.Id)
	fmt.Printf("Description : %s\n", action.Description)
	fmt.Printf("Result : %s\n", action.Result)
	fmt.Printf("Timestamp : %s\n", action.Timestamp)
	printDelimiter()
	time.Sleep(100 * time.Millisecond)
}

func printDelimiter() {
	fmt.Printf("----------------------------------\n")
}
