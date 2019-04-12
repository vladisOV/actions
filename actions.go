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

type ActionResponse struct {
	Token string
}

func main() {
	app := cli.NewApp()
	app.Name = "actions"
	app.Usage = "add/get/edit your actions"
	app.Version = "1.0.0"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "token",
			Value: "da",
			Usage: "auth token",
			// Hidden: true,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "get list of actions",
			Action:  getAllActions,
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
			Aliases: []string{"n"},
			Usage:   "log in into app",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "user, u"},
				cli.StringFlag{Name: "password, pw"},
			},
			Action: func(c *cli.Context) error {
				token := authenticate(c.String("user"), c.String("password"))
				saveToken(token)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func saveToken(token string) {
	ioutil.WriteFile("token", []byte(token), 0644)
}

func getToken() string {
	token, err := ioutil.ReadFile("token")
	if err != nil || !isAuthorized(string(token)) {
		fmt.Printf("Unauthorized, you have to log in first!\n")
	}
	return string(token)
}

func getAllActions(c *cli.Context) error {
	var token = getToken()
	actions := getAllActionsRequest(token)
	printActions(actions)
	return nil
}

func authenticate(username string, password string) string {
	jsonData := map[string]string{"username": username, "password": password}
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post("http://localhost:8080/auth", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("Failed to authenticate %s\n", err)
		return ""
	}
	data, _ := ioutil.ReadAll(response.Body)

	var ar ActionResponse
	json.Unmarshal([]byte(data), &ar)

	return ar.Token
}

func getAllActionsRequest(token string) []Action {
	req, err := http.NewRequest("GET", "http://localhost:8080/api/item", nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}
	fmt.Printf(token)

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	response, err := client.Do(req)

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
		// return nil
	}
	data, _ := ioutil.ReadAll(response.Body)
	var action Action
	json.Unmarshal([]byte(data), &action)
	return action
}

func isAuthorized(token string) bool {
	return len(token) > 0
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
