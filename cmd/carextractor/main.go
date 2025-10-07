package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

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

	p := path.Join(installPath, "rsfdata/cars")

	if _, err := os.Stat(p); err != nil {
		fmt.Printf("error reading directory: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.RBRInstallationPath) == 0 {
		cfg.Set(config.RBRInstallationPath, installPath)
	}

	carDirs := []string{}

	if err := filepath.Walk(p, func(currPath string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error traversing filepath: %v\n", err)
			return err
		}

		// skip the root cars folder
		if currPath == p {
			return nil
		}

		if info.IsDir() && info.Name() != "setups" {
			carDirs = append(carDirs, info.Name())
		}
		return nil
	}); err != nil {
		fmt.Printf("error reading dirs: %v\n", err)
		os.Exit(1)
	}

	if len(carDirs) == 0 {
		fmt.Println("no cars found in directory")
		os.Exit(1)
	}

	carMap := make(map[string]string)
	for _, car := range carDirs {
		carPath := path.Join(p, car)
		f, err := os.Open(carPath)
		if err != nil {
			fmt.Printf("error opening dir %q: %v\n", car, err)
			os.Exit(1)
		}
		files, err := f.Readdir(0)
		if err != nil {
			fmt.Printf("error reading dir %q: %v\n", car, err)
			os.Exit(1)
		}
		for _, file := range files {
			if file.IsDir() || len(filepath.Ext(file.Name())) > 0 {
				continue
			}
			carMap[file.Name()] = car
		}
		f.Close()
	}

	b, err := json.MarshalIndent(carMap, "", "  ")
	if err != nil {
		fmt.Printf("error marshaling map to JSON: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err := os.Mkdir("data", 0755); err != nil {
			fmt.Printf("error creating data directory: %v\n", err)
			os.Exit(1)
		}
	}

	f, err := os.OpenFile("data/cars.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("error creating cars.json: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.Write(b); err != nil {
		fmt.Printf("error writing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("car data successfully saved to data/cars.json")
}
