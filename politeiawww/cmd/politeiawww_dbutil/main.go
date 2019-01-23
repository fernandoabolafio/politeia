package main

import (
	"fmt"
	"os"

	"github.com/btcsuite/go-flags"
	"github.com/decred/politeia/politeiawww/cmd/politeiawww_dbutil/commands"
	"github.com/decred/politeia/politeiawww/cmd/politeiawww_dbutil/config"
	"github.com/decred/politeia/politeiawww/database"
	"github.com/decred/politeia/politeiawww/database/cockroachdb"
	"github.com/decred/politeia/politeiawww/database/leveldb"
)

type politeiawww_dbutil struct {
	db        database.Database
	cfg       *config.Config
	Commmands commands.Cmds
}

func setupDatabase(p *politeiawww_dbutil) error {
	cfg := p.cfg

	// fmt.Printf("")
	// Setup cockroach db for users database
	switch cfg.Database {
	case config.LevelDBOption:
		err := leveldb.CreateLevelDB(cfg.DataDir)
		if err != nil {
			return fmt.Errorf("CreateLevelDB: %v", err)
		}

		db, err := leveldb.NewLevelDB(cfg.DataDir, cfg.DBKey)
		if err != nil {
			return fmt.Errorf("NewLevelDB: %v", err)
		}
		p.db = db
		return nil
	case config.CockroachDBOption:
		err := cockroachdb.CreateCDB(cfg.DBHost, cfg.Net,
			cfg.DBRootCert, cfg.DBCertDir)
		if err != nil {
			return fmt.Errorf("CreateCDB: %v", err)
		}

		db, err := cockroachdb.NewCDB(cockroachdb.UserPoliteiawww, cfg.DBHost,
			cfg.Net, cfg.DBRootCert, cfg.DBCertDir, cfg.DBKey)
		if err != nil {
			return fmt.Errorf("NewCDB: %v", err)
		}
		p.db = db
		return nil
	}
	return fmt.Errorf("Invalid database option: %v", cfg.Database)
}

func _main() error {
	// Load config.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %v", err)
	}

	dbutil := politeiawww_dbutil{
		cfg: cfg,
	}

	// Setup database.
	err = setupDatabase(&dbutil)
	if err != nil {
		return fmt.Errorf("setup database: %v", err)
	}

	// Set commands db.
	commands.SetDatabase(dbutil.db)

	// Parse subcommand and execute
	var parser = flags.NewParser(&dbutil, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	return nil
}

func main() {
	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
