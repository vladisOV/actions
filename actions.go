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
	Username string `json:"username"`
	Password string `json:"password"`
}

type ErrorResponse struct {
	Timestamp string
	Path      string
	Status    string
	Error     string
	Message   string
}

var (
	version = "1.0.0"
	url     = "http://localhost:8080/"
)

func main() {
	app := cli.NewApp()
	app.Name = "actions"
	app.Usage = "add/get/edit your actions"
	app.Version = version
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "Get all actions for current user",
			Action:  getAllActions,
		},
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "Create new action",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "desc, d"},
				cli.StringFlag{Name: "result, r"},
			},
			Action: createAction,
		},
		{
			Name:    "login",
			Aliases: []string{"l"},
			Usage:   "Log in into app.",
			Action:  auth,
		},
		{
			Name:  "by",
			Usage: "Get actions by param. Just put your description/result as argument",
			Subcommands: []cli.Command{
				{
					Name:    "desc",
					Aliases: []string{"d"},
					Usage:   "Get actions by description",
					Action:  getActionsByDesc,
				},
				{
					Name:    "res",
					Aliases: []string{"r"},
					Usage:   "Get actions by result",
					Action:  getActionsByResult,
				},
				{
					Name:   "date",
					Usage:  "Get actions by date yyyy-MM-dd",
					Action: getActionsByDate,
				},
			},
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "Update action by id.",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "id"},
				cli.StringFlag{Name: "d"},
				cli.StringFlag{Name: "r"},
			},
			Action: updateAction,
		},
		{
			Name:    "register",
			Aliases: []string{"r"},
			Usage:   "Register new user",
			Action:  register,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

//TODO refactor, generic func
func getActionsByDate(c *cli.Context) error {
	var token = getToken()
	actions := getActionsByParam(token, QueryParam{"date", c.Args().Get(0)})
	printActions(actions)
	return nil
}

func getActionsByResult(c *cli.Context) error {
	var token = getToken()
	actions := getActionsByParam(token, QueryParam{"result", c.Args().Get(0)})
	printActions(actions)
	return nil
}

func getActionsByDesc(c *cli.Context) error {
	var token = getToken()
	actions := getActionsByParam(token, QueryParam{"description", c.Args().Get(0)})
	printActions(actions)
	return nil
}

func getAllActions(c *cli.Context) error {
	var token = getToken()
	actions := getActionsByParam(token, QueryParam{})
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
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.999"),
		}
		var saved = createActionRequest(action, token)
		printAction(saved)
	} else {
		var action = getSingleActionByParam(token, QueryParam{name: "id", value: id})
		if (action != Action{}) {
			if len(description) > 0 {
				action.Description = description
			} else {
				action.Result = result
			}
			var saved = createActionRequest(action, token)
			printAction(saved)
		} else {
			fmt.Printf("Action has not been found by id %s\n", id)
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
	authRequest.Username = userline
	fmt.Println("Enter your password:")
	passline, err := readline.String("> ")
	if err != nil {
		return nil
	}
	authRequest.Password = passline
	token := loginRequest(authRequest)
	if len(token) == 0 {
		fmt.Printf("Incorrect username or/and password.\n")
		return nil
	}
	saveToken(token)
	fmt.Printf("Successfully logged in.\n")
	return nil
}

func register(c *cli.Context) error {
	var authRequest = AuthRequest{}
	fmt.Println("Enter username:")
	userline, err := readline.String("> ")
	if err != nil {
		return nil
	}
	authRequest.Username = userline
	fmt.Println("Enter your password:")
	passline, err := readline.String("> ")
	if err != nil {
		return nil
	}
	var password = passline
	fmt.Println("Repeat your password:")
	passline, err = readline.String("> ")
	if err != nil {
		return nil
	}
	if password != passline || password == "" {
		fmt.Printf("Passwords does not match.\n")
		return nil
	}
	authRequest.Password = password
	response := registerRequest(authRequest)
	fmt.Printf(response + "\n")
	return nil
}

func registerRequest(authRequest AuthRequest) string {
	response, err := http.Post(url+"register", "application/json", marshalBody(authRequest))
	if err != nil {
		return "Smth went wrong. User hasn't been saved."
	}
	statusCode := response.StatusCode
	if statusCode == 200 {
		return "User has been saved successfully. Use login command to log in."
	}
	return "Smth went wrong. User hasn't been saved."
}

func loginRequest(authRequest AuthRequest) string {
	response, err := http.Post(url+"auth", "application/json", marshalBody(authRequest))
	if err != nil {
		fmt.Printf("Failed to authenticate %s\n", err)
		return ""
	}
	data, _ := ioutil.ReadAll(response.Body)

	var ar ActionResponse
	json.Unmarshal([]byte(data), &ar)

	return ar.Token
}

func getActionsByParam(token string, param QueryParam) []Action {
	data, statusCode := getActionsRequest(token, param)
	if checkBadRequest(data, statusCode) {
		return nil
	}
	var actions []Action
	json.Unmarshal([]byte(data), &actions)
	return actions
}

func getSingleActionByParam(token string, param QueryParam) Action {
	data, statusCode := getActionsRequest(token, param)
	if checkBadRequest(data, statusCode) {
		return Action{}
	}
	var action Action
	json.Unmarshal([]byte(data), &action)
	return action
}

func getActionsRequest(token string, param QueryParam) ([]byte, int) {
	req, _ := http.NewRequest("GET", url+"api/item", nil)
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
		return nil, 500
	}

	data, _ := ioutil.ReadAll(response.Body)
	return data, response.StatusCode
}

func createActionRequest(action Action, token string) Action {
	jsonValue, _ := json.Marshal(action)

	req, err := http.NewRequest("POST", url+"api/item", bytes.NewBuffer(jsonValue))
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

func checkBadRequest(data []byte, statusCode int) bool {
	if statusCode == 400 {
		var error ErrorResponse
		json.Unmarshal([]byte(data), &error)
		fmt.Printf("Request to actions storage failed with error '%s'\n", error.Message)
		return true
	}
	return false
}

func marshalBody(v interface{}) io.Reader {
	jsonValue, _ := json.Marshal(v)
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
