package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"pokedexCLI/internal/pokecache"
	"strings"
	"time"
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
	callback    func(*Config, *pokecache.Cache) error
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
			description: "Displays Location areas going forward",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays Location areas going backward",
			callback:    commandMapB,
		},
	}
}

func main() {
	config := &Config{
		Next: "https://pokeapi.co/api/v2/location-area/",
	}

	cache := pokecache.NewCache(1*time.Minute, 10*time.Second)

	initializeCommands()
	startRepl(config, cache)

}

func cleanInput(text string) []string {
	trimmedText := strings.ToLower(strings.TrimSpace(text))
	words := strings.Fields(trimmedText)

	return words
}

func startRepl(config *Config, cache *pokecache.Cache) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")

		scanner.Scan()

		input := scanner.Text()

		if command, exists := commandRegistery[cleanInput(input)[0]]; exists {
			err := command.callback(config, cache)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}

	}
}

func commandExit(config *Config, cache *pokecache.Cache) error {

	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config, cache *pokecache.Cache) error {

	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, command := range commandRegistery {
		fmt.Printf("%v: %v\n", command.name, command.description)
	}
	return nil
}

func commandMap(config *Config, cache *pokecache.Cache) error {
	locations, err := fetchLocations(config.Next, cache)
	if err != nil {
		fmt.Println("no more areas forward")
		return nil
	}

	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}

	config.Next = locations.Next
	config.Previous = locations.Previous

	return nil
}

func commandMapB(config *Config, cache *pokecache.Cache) error {
	locations, err := fetchLocations(config.Previous, cache)
	if err != nil {
		fmt.Println("no more areas backward")
		return nil
	}

	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}

	config.Next = locations.Next
	config.Previous = locations.Previous

	return nil
}

func fetchLocations(url string, cache *pokecache.Cache) (LocationAreas, error) {
	if url == "" {
		return LocationAreas{}, errors.New("no more locations in this direction")
	}

	if data, found := cache.Get(url); found {
		fmt.Println("Cache found")
		var body LocationAreas
		err := json.Unmarshal(data, &body)
		if err != nil {
			return LocationAreas{}, err
		}
		return body, nil
	}

	res, err := http.Get(url)
	if err != nil {
		return LocationAreas{}, err
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		return LocationAreas{}, errors.New("bad statusCode")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return LocationAreas{}, err
	}

	cache.Add(url, data)

	var body LocationAreas

	err = json.Unmarshal(data, &body)
	if err != nil {
		return LocationAreas{}, err
	}

	return body, nil
}
