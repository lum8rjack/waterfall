package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	version string = "v2023.6.1"
)

func main() {
	config := flag.String("config", "", "YAML config file to use")
	lout := flag.String("out", "", "File to save log output to")
	ver := flag.Bool("version", false, "Display version of waterfall")
	flag.Parse()

	// Print version
	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	// Check if config file was provided
	if *config == "" {
		log.Fatalln("No config file provided")
	}

	// Check log output file
	if *lout != "" {
		f, err := os.OpenFile(*lout, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		// Log to stdout and file
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
	}

	// Setup config
	conf, err := NewConfig(*config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Directory: %s\n", conf.Directory)

	// Start the workflow
	err = conf.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Completed
	log.Println("Completed!")

}
