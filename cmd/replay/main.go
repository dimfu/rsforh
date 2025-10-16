package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/dimfu/rsforh/internal/config"
)

func main() {
	cfg, err := config.Setup()
	if err != nil {
		fmt.Printf("error initializing config: %v\n", err)
		os.Exit(1)
	}

	args := os.Args

	var installPath string

	if len(cfg.RBRInstallationPath) > 0 {
		installPath = cfg.RBRInstallationPath
		fmt.Printf("using installation path from config: %s\n", installPath)
	} else {
		if len(args) < 2 {
			fmt.Println("please provide the path to Richard Burns Rally installation directory")
			os.Exit(1)
		}
		installPath = args[1]
	}

	if len(cfg.RBRInstallationPath) == 0 {
		cfg.Set(config.RBRInstallationPath, installPath)
	}

	p := path.Join(installPath, "RichardBurnsRally_SSE.exe")

	if _, err := os.Stat(p); err != nil {
		fmt.Printf("error reading directory: %v\n", err)
		os.Exit(1)
	}

	// TODO: use set replay data from online rally later
	cmd := exec.Command(p, "-autologon", "replay", "-autologonparam1", "test.rpl")
	cmd.Dir = installPath  
	cmd.Stdout = os.Stdout 
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to run exe: %v", err)
	}
}
