package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type LocationAreas struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []Location `json:"results"`
}

type Config struct {
	Next     string
	Previous string
}

type Location struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

var commandRegistery map[string]cliCommand

func initializeCommands() {
	commandRegistery = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays Location areas",
			callback:    commandMap,
		},
	}
}

func main() {
	config := &Config{
		Next: "https://pokeapi.co/api/v2/location-area/",
	}

	initializeCommands()
	startRepl(config)

}

func cleanInput(text string) []string {
	trimmedText := strings.ToLower(strings.TrimSpace(text))
	words := strings.Fields(trimmedText)

	return words
}

func startRepl(config *Config) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")

		scanner.Scan()

		input := scanner.Text()

		if command, exists := commandRegistery[cleanInput(input)[0]]; exists {
			err := command.callback(config)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}

	}
}

func commandExit(config *Config) error {

	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config) error {

	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, command := range commandRegistery {
		fmt.Printf("%v: %v\n", command.name, command.description)
	}
	return nil
}

func commandMap(config *Config) error {

	var body LocationAreas

	if config.Next == "" {
		fmt.Println("No more locations to show.")
		return nil
	}

	res, err := http.Get(config.Next)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		return errors.New("bad statusCode")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &body)
	if err != nil {
		return err
	}

	for _, location := range body.Results {
		fmt.Println(location.Name)
	}

	config.Next = body.Next
	config.Previous = body.Previous

	return nil
}
