package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func main() {
	args := os.Args
	if len(args) > 2 {
		fmt.Println("too many arguments")
		os.Exit(1)
	}

	p := args[1] // path to Richard Burns Rally installation directory
	p = path.Join(p, "rsfdata/cars")

	_, err := os.Stat(p)
	if err != nil {
		fmt.Printf("error reading directory: %v\n", err)
		os.Exit(1)
	}

	carDirs := []string{}

	if err := filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error traversing filepath: %v", err)
			return err
		}

		// skip the root cars folder
		if path == p {
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
		f, err := os.Open(path.Join(p, car))
		if err != nil {
			fmt.Printf("error opening dir %q %v\n", car, err)
			os.Exit(1)
		}
		files, err := f.Readdir(0)
		if err != nil {
			fmt.Printf("error reading dir %q %v\n", car, err)
			os.Exit(1)
		}
		for _, file := range files {
			if file.IsDir() || len(filepath.Ext(file.Name())) > 0 {
				continue
			}
			carMap[file.Name()] = car
		}
	}

	b, err := json.MarshalIndent(carMap, "", "  ")
	if err != nil {
		fmt.Printf("error marshaling map to JSON: %v\n", err)
	}

	_, err = os.Stat("data")
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir("data", 0755); err != nil {
				panic(err)
			}
		}
	}

	f, err := os.OpenFile("data/cars.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	_, err = f.Write(b)
	if err != nil {
		panic(err)
	}
}
