package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var seed *bool
var vlc *bool
var progress int64

const clearScreen = "\033[H\033[2J"

// Exit statuses.
const (
	_ = iota
	exitNoTorrentProvided
	exitErrorCreatingClient
	exitErrorAddingTorrent
)

func main() {
	// Set up flags.
	seed = flag.Bool("seed", true, "Seed after finished downloading")
	vlc = flag.Bool("vlc", false, "Open vlc to play the file")
	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(exitNoTorrentProvided)
	}

	// Start up the torrent client.
	client, err := NewClient(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error adding magnet \"%s\": %s\n", flag.Arg(0), err)
		os.Exit(exitErrorAddingTorrent)
	}

	// Http handler.
	go func() {
		http.HandleFunc("/", client.GetFile)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Open vlc to play.
	if *vlc {
		go func() {
			for !client.ReadyForPlayback() {
				time.Sleep(time.Second)
			}
			log.Printf("Playing in vlc")

			// @todo decide command to run based on os.
			if err := exec.Command("open", "-a", "vlc", "http://localhost:8080").Start(); err != nil {
				log.Printf("Error opening vlc: %s\n", err)
			}
		}()
	}

	// Cli render loop.
	for true {
		client.Render()
		time.Sleep(time.Second)
	}
}
