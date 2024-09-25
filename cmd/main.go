package main

import (
	"encoding/json"
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
	buf, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	c := filegunner.Configuration{}
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		log.Fatal(err)
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

	// Start filewatcher on input directory
	watcher, err := filegunner.NewWatcher(c.InputDir, logFn, log.Fatal, event(c, logFn))
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// read existing files
	fs, err := os.ReadDir(c.InputDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fs {
		event(c, logFn)(filegunner.CreationEvent{FileName: filepath.Join(c.InputDir, f.Name())})
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func noop(_ ...any) {
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func event(cfg filegunner.Configuration, logfn filegunner.LogFn) filegunner.EventFn {
	return func(evt filegunner.CreationEvent) {
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
				req.Attachments[i] = filepath.Join(cfg.InputDir, s)
			}
		}

		service, isFound := cfg.Services[req.ServiceID]
		if !isFound {
			logfn("mailgun service not found: ", req.ServiceID)
			return
		}

		if cfg.RunMode == filegunner.DryRun {
			err = filegunner.DryRunMail(req, evt.FileName)
			if err != nil {
				logfn("error making dry run log: ", evt.FileName, err)
				return
			}
		} else {
			err = filegunner.SendMailRequest(req, service)
			if err != nil {
				logfn("error sending mail for file: ", evt.FileName, err)
				return
			}
		}

		if cfg.HistoryDir != nil {
			now := time.Now().Unix()
			fileName := strconv.FormatInt(now, 10) + "." + filepath.Base(evt.FileName)
			err = os.WriteFile(filepath.Join(*cfg.HistoryDir, fileName), buf, 0644)
			if err != nil {
				logfn("error creating historical save for file: ", evt.FileName, err)
				// we won't return here, as we don't want to redo this functionality move on to deletion
			}

			if req.Attachments != nil {
				for _, path := range req.Attachments {
					bs, err := readFileContents(path)
					if err != nil {
						logfn("error reading attachment for history: ", path, err)
						continue
					}
					fileName := strconv.FormatInt(now, 10) + "." + filepath.Base(path)
					err = os.WriteFile(filepath.Join(*cfg.HistoryDir, fileName), bs, 0644)
					if err != nil {
						logfn("error creating historical save for file: ", path, err)
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
			for _, path := range req.Attachments {
				err = os.Remove(path)
				if err != nil {
					logfn("error removing file: ", path, err)
				}
			}
		}
	}
}

func readFileContents(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
