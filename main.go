package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	ctx context.Context
}

type commands struct {
	CommandHandlerStore map[string]func(*state, command) error
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.cfg.Current_User_Name == "" {
			return fmt.Errorf("no user logged in; please log in first")
		}
		user, err := s.db.GetUser(s.ctx, s.cfg.Current_User_Name)
		if err != nil {
			return fmt.Errorf("failed to fetch logged-in user: %v", err)
		}
		return handler(s, cmd, user)
	}
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
	_, err := s.db.GetUser(s.ctx, cmd.argslice[2])
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
	user, err := s.db.CreateUser(s.ctx, userparams)
	if err != nil {
		fmt.Println(err)
		return err
	}
	s.cfg.SetUser(userparams.Name)
	fmt.Printf("User %s created successfully\n", userparams.Name)
	fmt.Printf("User %s details \n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	fmt.Println("Inside Reset Handler")
	return s.db.DeleteAllUser(s.ctx)
}

func handlerLogAllUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(s.ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		if s.cfg.Current_User_Name == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	read, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rssfeed RSSFeed
	if err := xml.Unmarshal(read, &rssfeed); err != nil {
		return nil, err
	}
	return &rssfeed, nil

}

func handlerAggWebsite(s *state, cmd command) error {
	fmt.Println("Inside AggWebsite Handler")
	time_between_reqs := cmd.argslice[2]
	timeBetweenRequests, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		fmt.Printf("Collecting feeds every %s", time_between_reqs)
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.argslice) < 4 {
		return fmt.Errorf("not enough arguments")
	}
	feedparams := database.CreateFeedParams{
		Name:   cmd.argslice[2],
		Url:    cmd.argslice[3],
		UserID: user.ID,
	}
	dbfeed, err := s.db.CreateFeed(s.ctx, feedparams)
	if err != nil {
		return err
	}
	feedfollowparams := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: dbfeed.ID,
	}
	_, err = s.db.CreateFeedFollow(s.ctx, feedfollowparams)
	if err != nil {
		return err
	}
	fmt.Printf("%v", dbfeed)
	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	ctx := s.ctx
	dbfeed, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}
	for _, feed := range dbfeed {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		dbuser, err := s.db.GetUserById(ctx, feed.UserID)
		if err != nil {
			return err
		}
		fmt.Println(dbuser.Name)

	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.argslice) < 3 {
		return fmt.Errorf("not enough arguments")
	}
	feed, err := s.db.GetFeedNameByUrl(s.ctx, cmd.argslice[2])
	if err != nil {
		return err
	}
	feedfollowparams := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	dbfollow, err := s.db.CreateFeedFollow(s.ctx, feedfollowparams)
	if err != nil {
		return err
	}
	fmt.Println(dbfollow.FeedName)
	fmt.Println(user.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	dbfeed, err := s.db.GetFeedFollowsForUser(s.ctx, user.ID)
	if err != nil {
		return err
	}
	for _, dbf := range dbfeed {
		fmt.Println(dbf.FeedName)
		fmt.Println(dbf.UserName)
	}
	return nil

}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.argslice) < 3 {
		return fmt.Errorf("not enough arguments")
	}
	feed, err := s.db.GetFeedNameByUrl(s.ctx, cmd.argslice[2])
	if err != nil {
		return err
	}
	feedunfollow := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	if err := s.db.DeleteFeedFollow(s.ctx, feedunfollow); err != nil {
		return err
	}

	return nil
}

func scrapeFeeds(s *state) error {
	dbfeed, err := s.db.GetNextFeedToFetch(s.ctx)
	if err != nil {
		return err
	}
	feedfetchparam := database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:     time.Now(),
		ID:            dbfeed.ID,
	}
	dbfeed, err = s.db.MarkFeedFetched(s.ctx, feedfetchparam)
	if err != nil {
		return err
	}
	rssfeed, err := fetchFeed(s.ctx, dbfeed.Url)
	if err != nil {
		return err
	}

	for _, item := range rssfeed.Channel.Item {
		// Try to parse the published date safely
		var publishedAt time.Time
		if item.PubDate != "" {
			// Common RSS date format: "Mon, 02 Jan 2006 15:04:05 MST"
			t, err := time.Parse(time.RFC1123Z, item.PubDate)
			if err != nil {
				// Try fallback without timezone offset
				t, err = time.Parse(time.RFC1123, item.PubDate)
				if err != nil {
					// If parsing fails, use current time
					publishedAt = time.Now().UTC()
				} else {
					publishedAt = t.UTC()
				}
			} else {
				publishedAt = t.UTC()
			}
		} else {
			publishedAt = time.Now().UTC()
		}
		createpostparam := database.CreatePostParams{
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			FeedID:      dbfeed.ID,
			PublishedAt: publishedAt,
		}

		_, err := s.db.CreatePost(s.ctx, createpostparam)
		if err != nil {
			// Check if the error is a duplicate URL (unique violation)
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				// Ignore duplicate post
				continue
			}
			// Log and continue for other errors
			log.Printf("Error inserting post %q: %v", item.Title, err)
			continue
		}
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	page := 3
	if len(cmd.argslice) > 1 {
		intVal, err := strconv.Atoi(cmd.argslice[1])
		if err != nil {
			return err
		}
		limit = intVal
	}
	if len(cmd.argslice) > 2 {
		intVal, err := strconv.Atoi(cmd.argslice[2])
		if err != nil {
			return err
		}
		page = intVal
	}
	offset := (page - 1) * limit
	postuserparam := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
		Offset: int32(offset),
	}
	_, err := s.db.GetPostsForUser(s.ctx, postuserparam)
	if err != nil {
		return err
	}
	return nil

}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get DB URL from env
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set in environment")
	}
	fmt.Println("Connecting to DB:", dbURL)

	// Open DB
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	// Initialize queries
	dbQueries := database.New(db)
	cfg, err := config.Read()
	if err != nil {
		fmt.Print(err)
	}
	st := &state{cfg: cfg,
		db:  dbQueries,
		ctx: context.Background()}
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
	case "reset":
		cmds.register(cmdName, handlerReset)
	case "users":
		cmds.register(cmdName, handlerLogAllUsers)
	case "agg":
		cmds.register(cmdName, handlerAggWebsite)
	case "addfeed":
		cmds.register(cmdName, middlewareLoggedIn(handlerAddFeed))
	case "feeds":
		cmds.register(cmdName, handlerGetFeeds)
	case "follow":
		cmds.register(cmdName, middlewareLoggedIn(handlerFollow))
	case "following":
		cmds.register(cmdName, middlewareLoggedIn(handlerFollowing))
	case "unfollow":
		cmds.register(cmdName, middlewareLoggedIn(handlerUnfollow))
	case "browse":
		cmds.register(cmdName, middlewareLoggedIn(handlerBrowse))
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
