package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	mrand "math/rand"
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
	callback    func(*Config, *pokecache.Cache, []string) error
}

var pokedex = make(map[string]Pokemon)

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
		"explore": {
			name:        "explore <location>",
			description: "Displays possible pokemons at specific Location ",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch <pokemon>",
			description: "Displays possible pokemons at specific Location ",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect <pokemon>",
			description: "Displays if pokemon is in pokedex ",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex ",
			description: "Displays caught pokemons in pokedex ",
			callback:    commandPokedex,
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

		words := cleanInput(input)
		if len(words) == 0 {
			fmt.Printf("Unknown commmand")
		}
		cmdName := words[0]
		args := words[1:]

		if command, exists := commandRegistery[cmdName]; exists {
			err := command.callback(config, cache, args)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}

	}
}

func commandExit(config *Config, cache *pokecache.Cache, args []string) error {

	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config, cache *pokecache.Cache, args []string) error {

	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, command := range commandRegistery {
		fmt.Printf("%v: %v\n", command.name, command.description)
	}
	return nil
}

func commandMap(config *Config, cache *pokecache.Cache, args []string) error {
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

func commandMapB(config *Config, cache *pokecache.Cache, args []string) error {
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

func commandCatch(config *Config, cache *pokecache.Cache, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide a pokemon to catch")
	}
	pokemon_str := args[0]
	url := "https://pokeapi.co/api/v2/pokemon/" + pokemon_str

	res, err := http.Get(url)
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

	cache.Add(url, data)

	var pokemon Pokemon
	err = json.Unmarshal(data, &pokemon)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)

	exp_chance := mrand.Intn(250) - pokemon.BaseExperience
	if exp_chance < 0 {
		exp_chance = 0
	}
	chance := 0.3 + float64(exp_chance)*0.0026 //max exp_chance = 0,65
	player_chance := mrand.Float64()
	if player_chance < chance {
		fmt.Printf("%s escaped!\n", pokemon.Name)
		return nil
	}

	fmt.Printf("%s was caught!\n", pokemon.Name)
	pokedex[pokemon.Name] = pokemon
	return nil
}

func commandExplore(config *Config, cache *pokecache.Cache, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide a location area to explore")
	}
	location := args[0]
	url := "https://pokeapi.co/api/v2/location-area/" + location

	if data, found := cache.Get(url); found {
		//fmt.Println("Explored location found in cache")
		return printPokemonFromData(data)
	}

	res, err := http.Get(url)
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

	cache.Add(url, data)
	return printPokemonFromData(data)
}

func commandInspect(config *Config, cache *pokecache.Cache, args []string) error {
	if len(args) == 0 {
		return errors.New("please provide a pokemon area to inspect")
	}
	pokemon_str := args[0]
	pokemon, exists := pokedex[pokemon_str]
	if exists {
		//fmt.Printf("found %s in pokedex\n", pokemon_str)
		printDetailedPokemonFromData(pokemon)
		return nil
	} else {
		fmt.Println("Pokemon hasn't been caught")
		return nil
	}

	return nil
}
func commandPokedex(config *Config, cache *pokecache.Cache, args []string) error {
	if len(pokedex) < 1 {
		fmt.Println("Pokedex empty")
		return nil
	}
	for pokemon := range pokedex {
		fmt.Printf("- %v\n", pokemon)
	}

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

func printPokemonFromData(data []byte) error {
	var exploreData ExploreLocation
	err := json.Unmarshal(data, &exploreData)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemons:")
	for _, p := range exploreData.PokemonEncounters {
		fmt.Printf("%s\n", p.Pokemon.Name)
	}
	return nil
}

func printDetailedPokemonFromData(pokemon Pokemon) error {
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %v\n", pokemon.Height)
	fmt.Printf("Height: %v\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("-%v : %v\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("- %v\n", t.Type.Name)
	}

	return nil
}
