package filegunner

import (
	"github.com/fsnotify/fsnotify"
)

type CreationEvent struct {
	FileName string
}

type FileWatcher struct {
	fsNotifyWatcher *fsnotify.Watcher
}

// todo: real logging
type LogFn func(v ...any)
type ErrFn func(v ...any)
type EventFn func(evt CreationEvent)

func NewWatcher(dir string, logFn LogFn, errFn ErrFn, eventFn EventFn) (*FileWatcher, error) {

	fsnWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-fsnWatcher.Events:
				if !ok {
					return
				}
				logFn("event: ", event)
				if event.Has(fsnotify.Create) {
					logFn("file created: ", event.Name)
					eventFn(CreationEvent{FileName: event.Name})
				}
			case err, ok := <-fsnWatcher.Errors:
				if !ok {
					return
				}
				errFn(err)
			}
		}
	}()

	// Add a path.
	err = fsnWatcher.Add(dir)
	if err != nil {
		return nil, err
	}

	watcherWrapper := FileWatcher{
		fsNotifyWatcher: fsnWatcher,
	}

	return &watcherWrapper, nil
}

func (w *FileWatcher) Close() error {
	return w.fsNotifyWatcher.Close()
}
