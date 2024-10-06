package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BlindGarret/filegunner"
	"gopkg.in/yaml.v3"
)

func main() {
	// Read in configuration
	configLocation := os.Getenv("CONFIG_PATH")
	if configLocation == "" {
		configLocation = "config.yaml"
	}
	buf, err := os.ReadFile(configLocation)
	if err != nil {
		log.Fatal(err)
	}
	c := filegunner.Configuration{}
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Configuration Details:")
	fmt.Printf("Run Without Filewatcher: %t\n", c.RunNoWatcher)
	fmt.Printf("Run With Verbosity Flag for FileWatcher: %t\n", c.VerboseFileWatcher)
	fmt.Printf("Input Directory: %s\n", c.InputDir)
	if c.HistoryDir != nil {
		fmt.Printf("History Directory: %s\n", *c.HistoryDir)
	}

	httpClient := filegunner.NewHttpClientWrapper()
	var mailer filegunner.Mailer
	if c.RunMode == filegunner.DryRun {
		mailer, err = filegunner.NewDryRunMailer(osCreate, c.LogDir)
		if err != nil {
			panic(err)
		}
	} else {
		mailer = filegunner.NewMailgunMailer(httpClient, readFileContents, c.Services)
	}

	logFn := log.Default().Println
	if !c.VerboseFileWatcher {
		logFn = noop
	}

	if !dirExists(c.InputDir) {
		log.Fatal("input directory does not exist", c.InputDir)
	}

	if c.HistoryDir != nil && !dirExists(*c.HistoryDir) {
		log.Fatal("history directory does not exist", *c.HistoryDir)
	}

	if !c.RunNoWatcher {
		// Start filewatcher on input directory
		watcher, err := filegunner.NewWatcher(c.InputDir, logFn, log.Fatal, event(c, logFn, mailer))
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()
	}

	// read existing files
	fmt.Println("Reading Directory")
	fs, err := os.ReadDir(c.InputDir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d files...", len(fs))
	for _, f := range fs {
		fmt.Printf("Working on File: %s \n", f.Name())
		event(c, logFn, mailer)(filegunner.CreationEvent{FileName: filepath.Join(c.InputDir, f.Name())})
	}

	if !c.RunNoWatcher {
		// Block main goroutine forever.
		fmt.Println("Waiting for File Events...")
		<-make(chan struct{})
	}
}

func noop(_ ...any) {
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("for path: %s\n", path)
		return false
	}
	return true
}

func event(cfg filegunner.Configuration, logfn filegunner.LogFn, mailer filegunner.Mailer) filegunner.EventFn {
	return func(evt filegunner.CreationEvent) {
		now := time.Now()
		if !strings.HasSuffix(evt.FileName, ".maildata.json") {
			logfn("File isn't maildata. Skipping. ", evt.FileName)
			return
		}

		buf, err := os.ReadFile(evt.FileName)
		if err != nil {
			logfn("error reading file: ", evt.FileName, err)
			return
		}

		req := filegunner.MailRequest{}
		err = json.Unmarshal(buf, &req)
		if err != nil {
			logfn("error unmarshalling JSON file: ", evt.FileName, err)
			return
		}

		if req.Attachments != nil {
			// root the attachments
			for i, s := range req.Attachments {
				req.Attachments[i].FilePath = filepath.Join(cfg.InputDir, s.FilePath)
			}
		}

		err = mailer.Send(req, evt.FileName, now)
		if err != nil {
			logfn("error sending: ", evt.FileName, err)
			return
		}

		if cfg.HistoryDir != nil {
			fileName := strconv.FormatInt(now.Unix(), 10) + "." + filepath.Base(evt.FileName)
			err = os.WriteFile(filepath.Join(*cfg.HistoryDir, fileName), buf, 0644)
			if err != nil {
				logfn("error creating historical save for file: ", evt.FileName, err)
				// we won't return here, as we don't want to redo this functionality move on to deletion
			}

			if req.Attachments != nil {
				for _, attachment := range req.Attachments {
					bs, err := readFileContents(attachment.FilePath)
					if err != nil {
						logfn("error reading attachment for history: ", attachment.FilePath, err)
						continue
					}
					fileName := strconv.FormatInt(now.Unix(), 10) + "." + filepath.Base(attachment.FilePath)
					err = os.WriteFile(filepath.Join(*cfg.HistoryDir, fileName), bs, 0644)
					if err != nil {
						logfn("error creating historical save for file: ", attachment.FilePath, err)
						// we won't return here, as we don't want to redo this functionality move on to deletion
					}
				}
			}
		}

		err = os.Remove(evt.FileName)
		if err != nil {
			logfn("error removing file: ", evt.FileName, err)
		}
		if req.Attachments != nil {
			for _, attachment := range req.Attachments {
				err = os.Remove(attachment.FilePath)
				if err != nil {
					logfn("error removing file: ", attachment.FilePath, err)
				}
			}
		}
	}
}

func osCreate(s string) (io.WriteCloser, error) {
	return os.Create(s)
}

func readFileContents(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
