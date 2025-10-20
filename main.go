package main

import (
	"fmt"
	"os"

	config "github.com/satyamdash/BlogAggregator/internal"
)

type state struct {
	cfg *config.Config
}

type commands struct {
	CommandHandlerStore map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {

	val, ok := c.CommandHandlerStore[cmd.name]
	if !ok {
		return fmt.Errorf("command does not exist")
	}
	return val(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.CommandHandlerStore[name] = f
}

type command struct {
	name     string
	argslice []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.argslice) < 2 {
		return fmt.Errorf("not enough arguments were provided")
	}

	if len(cmd.argslice) < 3 {
		return fmt.Errorf("a username is required")
	}

	s.cfg.SetUser(cmd.argslice[2])
	fmt.Printf("the user %s has been set", s.cfg.Current_User_Name)
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Print(err)
	}
	st := &state{cfg: cfg}
	cmds := &commands{CommandHandlerStore: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	args := os.Args
	fmt.Printf("size of args %d", len(args))
	if len(args) < 2 {
		fmt.Println("usage: go run . <command> [args...]")
		os.Exit(1)
	}
	// for idx, arg := range args {
	// 	fmt.Printf("%d, %s", idx, arg)
	// }
	cmdName := args[1]
	cmd := command{
		name:     cmdName,
		argslice: args,
	}
	if err := cmds.run(st, cmd); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
