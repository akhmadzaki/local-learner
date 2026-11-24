package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/akhmadzaki/go-backend/config"
	"github.com/akhmadzaki/go-backend/db"
)

const usage = `Database migrations (golang-migrate)

Usage:
  migrate <command> [version]

Commands:
  up         Apply all pending migrations
  down       Roll back the most recently applied migration
  to <n>     Migrate to a specific version (up or down)
  drop       Drop every managed table (runs all down migrations)
  version    Print the current schema version

Examples:
  migrate up
  migrate to 2
`

func main() {
	cfg := config.Load()

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	cmd := os.Args[1]
	var err error

	switch cmd {
	case "up":
		err = db.MigrateUp(&cfg.Database)
	case "down":
		err = db.MigrateDown(&cfg.Database)
	case "to":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "migrate: 'to' requires a version argument, e.g. 'migrate to 2'")
			os.Exit(2)
		}
		var n int
		n, err = strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "migrate: invalid version %q\n", os.Args[2])
			os.Exit(2)
		}
		err = db.MigrateTo(&cfg.Database, uint(n))
	case "drop":
		err = db.MigrateDrop(&cfg.Database)
	case "version":
		version, dirty, verr := db.MigrateVersion(&cfg.Database)
		if verr != nil {
			fmt.Fprintf(os.Stderr, "migrate: %v\n", verr)
			os.Exit(1)
		}
		fmt.Printf("version: %d (dirty: %v)\n", version, dirty)
		return
	case "help", "-h", "--help":
		fmt.Print(usage)
		return
	default:
		fmt.Fprintf(os.Stderr, "migrate: unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("migrate: %s OK\n", cmd)
}
