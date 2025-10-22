# 🐊 Gator — RSS Feed Aggregator CLI

Gator is a command-line RSS feed aggregator written in Go.  
It allows users to subscribe to RSS feeds, fetch and store posts in a local PostgreSQL database, and view the latest posts right from the terminal.

---

## 📦 Prerequisites

To run Gator locally, you'll need:

- **Go** (version 1.22 or later) → [Install Go](https://go.dev/doc/install)
- **PostgreSQL** → [Install PostgreSQL](https://www.postgresql.org/download/)

Make sure both are available in your terminal:

```bash
go version
psql --version


CONFIGURATION
Create a .gatorconfig.json file in your home directory (or wherever your app expects it) with content like this:
{
  "db_url": "postgres://<user>:<password>@localhost:5432/<dbname>?sslmode=disable"
}
Replace <user>, <password>, and <dbname> with your own PostgreSQL credentials.  


Usage

Add a new feed
go run . addfeed "https://example.com/rss"

List all feeds
go run . feeds

Create a new user
go run . register myusername
