package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/cconger/tw-fc/twitch"
)

func main() {
	clientID := flag.String("client-id", "", "specify a client-id to identify to twitch with")
	frequency := flag.Duration("p", 1*time.Minute, "specify a golang parsable duration to wait between downloads")
	number := flag.Int("n", 100, "specify the number of streams to download image for")
	outDir := flag.String("out", "captures", "path to output images to")

	flag.Parse()

	if *clientID == "" {
		log.Fatal("Client ID Required")
	}

	client := twitch.NewClient(twitch.WithClientID(*clientID))
	ctx := context.Background()

	err := os.MkdirAll(*outDir, 0775)
	if err != nil {
		log.Fatalf("Unable to create output directory %s", outDir)
	}
	log.Printf("Downloading top %d streams every %s to %s", *number, frequency.String(), *outDir)
	err = triggerDownload(ctx, client, *number, *outDir)
	log.Printf("Finished")
	timer := time.NewTicker(*frequency)

	for {
		select {
		case <-timer.C:
			log.Printf("Downloading top %d streams", *number)
			err := triggerDownload(ctx, client, *number, *outDir)
			log.Printf("Finished")
			if err != nil {
				log.Fatal(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

type download struct {
	url      string
	filename string
}

func triggerDownload(ctx context.Context, client *twitch.Client, number int, outDir string) error {
	streams, err := client.GetTopStreams(context.Background(), number)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got %d streams", len(streams))
	var wg sync.WaitGroup
	c := make(chan download)
	workerCount := 32
	log.Printf("Starting %d workers...", workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dl := range c {
				err = DownloadURLToFile(ctx, dl.url, dl.filename)
				if err != nil {
					log.Printf("Err downloading file: %s", err.Error())
					continue
				}
			}
		}()
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	for _, s := range streams {
		url, err := twitch.GetURLForStreamShot(&s)
		if err != nil {
			log.Printf("Err getting url: %s", err.Error())
			continue
		}
		filename := path.Join(outDir, fmt.Sprintf("%d_%s.jpg", s.ID, timestamp))
		c <- download{
			url:      url,
			filename: filename,
		}
	}
	close(c)
	wg.Wait()

	return nil
}

//TODO: better defaults
var client = http.Client{}

func DownloadURLToFile(ctx context.Context, url string, filename string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
