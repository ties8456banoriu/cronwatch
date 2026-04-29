package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cronwatch/internal/config"
)

const defaultConfigPath = "/etc/cronwatch/config.json"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to cronwatch config file")
	validate := flag.Bool("validate", false, "validate config and exit")
	flag.Parse()

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "error: config path must not be empty")
		os.Exit(1)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if *validate {
		fmt.Printf("config OK: %d job(s) loaded from %s\n", len(cfg.Jobs), *configPath)
		os.Exit(0)
	}

	log.Printf("cronwatch starting: monitoring %d job(s), check interval %v",
		len(cfg.Jobs), cfg.CheckInterval)

	// Placeholder: monitor loop will be wired in subsequent phases.
	select {}
}
