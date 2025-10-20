package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	config "github.com/satyamdash/BlogAggregator/internal"
	"github.com/satyamdash/BlogAggregator/internal/database"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
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
	fmt.Println("Registering Command")
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
	_, err := s.db.GetUser(context.Background(), cmd.argslice[2])
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
		os.Exit(1)
	}

	s.cfg.SetUser(cmd.argslice[2])
	fmt.Printf("the user %s has been set", s.cfg.Current_User_Name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	fmt.Println("Inside Register Handler")
	if len(cmd.argslice) < 3 {
		return fmt.Errorf("a username is required")
	}
	fmt.Println(cmd.argslice[2])
	userparams := database.CreateUserParams{
		ID:        uuid.New(),
		Name:      cmd.argslice[2],
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	user, err := s.db.CreateUser(context.Background(), userparams)
	if err != nil {
		return err
	}
	s.cfg.SetUser(userparams.Name)
	fmt.Printf("User %s created successfully\n", userparams.Name)
	fmt.Printf("User %s details \n", user)
	return nil
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error reading dburl from env")
	}
	dbQueries := database.New(db)
	cfg, err := config.Read()
	if err != nil {
		fmt.Print(err)
	}
	st := &state{cfg: cfg,
		db: dbQueries}
	cmds := &commands{CommandHandlerStore: make(map[string]func(*state, command) error)}

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
	switch cmdName {
	case "login":
		fmt.Println("Login handler initiate")
		cmds.register(cmdName, handlerLogin)
	case "register":
		fmt.Println("Register handler initiate")
		cmds.register(cmdName, handlerRegister)
	}

	cmd := command{
		name:     cmdName,
		argslice: args,
	}
	if err := cmds.run(st, cmd); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
