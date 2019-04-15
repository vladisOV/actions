package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bobappleyard/readline"
	"github.com/urfave/cli"
)

type Action struct {
	Id          *string `json:"id"`
	Description string  `json:"description"`
	Result      string  `json:"result"`
	Timestamp   string  `json:"timestamp"`
}

type ActionResponse struct {
	Token string
}

type QueryParam struct {
	name  string
	value string
}

type AuthRequest struct {
	username string
	password string
}

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
			Usage:   "actions all",
			Action:  getAllActions,
		},
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "actions new -d DESCRIPTION -r RESULT",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "desc, d"},
				cli.StringFlag{Name: "result, r"},
			},
			Action: createAction,
		},
		{
			Name:    "login",
			Aliases: []string{"l"},
			Usage:   "actions login and follow instructions",
			Action:  auth,
		},
		{
			Name:  "by",
			Usage: "get actions by param",
			Subcommands: []cli.Command{
				{
					Name:    "desc",
					Aliases: []string{"d"},
					Usage:   "actions by desc/d DESCRIPTION VALUE",
					Action:  getActionsByDesc,
				},
				{
					Name:    "res",
					Aliases: []string{"r"},
					Usage:   "actions by res/r RESULT VALUE",
					Action:  getActionsByResult,
				},
			},
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "actions update -id ID OF ACTION -d DESCRIPTION -r RESULT",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "id"},
				cli.StringFlag{Name: "d"},
				cli.StringFlag{Name: "r"},
			},
			Action: updateAction,
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

func getActionsByResult(c *cli.Context) error {
	var token = getToken()
	actions := getActionsRequest(token, QueryParam{"result", c.Args().Get(0)})
	printActions(actions)
	return nil
}

func getActionsByDesc(c *cli.Context) error {
	var token = getToken()
	actions := getActionsRequest(token, QueryParam{"description", c.Args().Get(0)})
	printActions(actions)
	return nil
}

func getAllActions(c *cli.Context) error {
	var token = getToken()
	actions := getActionsRequest(token, QueryParam{})
	printActions(actions)
	return nil
}

func updateAction(c *cli.Context) error {
	var id = c.String("id")
	if len(id) == 0 {
		fmt.Println("Provide correct id.")
		return nil
	}
	var description = c.String("d")
	var result = c.String("r")
	var token = getToken()

	if len(description) > 0 && len(result) > 0 {
		var action = Action{
			Id:          &id,
			Description: description,
			Result:      result,
		}
		createActionRequest(action, token)
	} else {
		var actions = getActionsRequest(token, QueryParam{name: "id", value: id})
		if len(actions) > 0 {
			var action = actions[0]
			if len(description) > 0 {
				action.Description = description
			} else {
				action.Result = result
			}
		} else {
			fmt.Println("Action has not been found by id %s", id)
		}
	}
	return nil
}

func createAction(c *cli.Context) error {
	var token = getToken()
	var action = Action{
		Description: c.String("desc"),
		Result:      c.String("result"),
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.999"),
	}
	saved := createActionRequest(action, token)
	if saved != (Action{}) {
		printDelimiter()
		printAction(saved)
	}
	return nil
}

func auth(c *cli.Context) error {
	var authRequest = AuthRequest{}
	fmt.Println("Enter your username:")
	userline, err := readline.String("> ")
	if err != nil {
		return nil
	}
	authRequest.username = userline
	fmt.Println("Enter your password:")
	passline, err := readline.String("> ")
	if err != nil {
		return nil
	}
	authRequest.password = passline
	token := loginRequest(authRequest)
	if len(token) == 0 {
		fmt.Printf("Incorrect username or/and password.\n")
		return nil
	}
	saveToken(token)
	fmt.Printf("Successfully logged in.\n")
	return nil
}

func loginRequest(authRequest AuthRequest) string {
	jsonData := map[string]string{"username": authRequest.username, "password": authRequest.password}
	response, err := http.Post("http://localhost:8080/auth", "application/json", marshalBody(jsonData))
	if err != nil {
		fmt.Printf("Failed to authenticate %s\n", err)
		return ""
	}
	data, _ := ioutil.ReadAll(response.Body)

	var ar ActionResponse
	json.Unmarshal([]byte(data), &ar)

	return ar.Token
}

func getActionsRequest(token string, param QueryParam) []Action {
	req, _ := http.NewRequest("GET", "http://localhost:8080/api/item", nil)
	if (param != QueryParam{}) {
		q := req.URL.Query()
		q.Add(param.name, param.value)
		req.URL.RawQuery = q.Encode()
	}

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

func createActionRequest(action Action, token string) Action {
	jsonValue, _ := json.Marshal(action)

	req, err := http.NewRequest("POST", "http://localhost:8080/api/item", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		fmt.Printf("Request to actions storage failed with error %s\n", err)
		return Action{}
	}
	data, _ := ioutil.ReadAll(response.Body)
	var saved Action
	json.Unmarshal([]byte(data), &saved)
	return saved
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

func marshalBody(body map[string]string) io.Reader {
	jsonValue, _ := json.Marshal(body)
	return bytes.NewBuffer(jsonValue)
}

func printAction(action Action) {
	fmt.Printf("Id : %s\n", *action.Id)
	fmt.Printf("Description : %s\n", action.Description)
	fmt.Printf("Result : %s\n", action.Result)
	fmt.Printf("Timestamp : %s\n", action.Timestamp)
	printDelimiter()
	time.Sleep(100 * time.Millisecond)
}

func printDelimiter() {
	fmt.Printf("----------------------------------\n")
}
