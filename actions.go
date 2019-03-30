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
				data := getAllItems()
				fmt.Println(data)
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
				saved := createItem(c.String("desc"), c.String("result"))
				fmt.Println(saved)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getAllItems() string {
	response, err := http.Get("http://localhost:8080/api/item")
	if err != nil {
		return fmt.Sprintf("Request to notes storage failed with error %s\n", err)
	}
	data, _ := ioutil.ReadAll(response.Body)
	return fmt.Sprintf(string(data))
}

func createItem(desc string, result string) string {
	jsonData := map[string]string{"description": desc, "result": result,
		"timestamp": time.Now().Format("2006-01-02T15:04:05.999")}
	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post("http://localhost:8080/api/item", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Sprintf("Request to notes storage failed with error %s\n", err)
	}
	data, _ := ioutil.ReadAll(response.Body)
	return fmt.Sprintf(string(data))
}
