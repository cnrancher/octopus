package registration

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"

	api "github.com/rancher/octopus/pkg/adaptor/api/v1alpha1"
)

func newSocketWatcher() (*socketWatcher, error) {
	var fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = fsWatcher.Add(api.LimbSocket)
	if err != nil {
		return nil, err
	}

	return &socketWatcher{
		fsWatcher: fsWatcher,
	}, nil
}

type socketWatcher struct {
	fsWatcher *fsnotify.Watcher
}

func (w *socketWatcher) Watch(stop <-chan struct{}) error {
loop:
	for {
		select {
		case ent, ok := <-w.fsWatcher.Events:
			if !ok {
				break loop
			}
			if ent.Op&fsnotify.Remove == fsnotify.Remove {
				return errors.Errorf("%s has removed", ent.Name)
			}
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				break loop
			}
			return errors.Wrapf(err, "receive error from fs watcher")
		case <-stop:
			_ = w.fsWatcher.Close()
			break loop
		}
	}

	return nil
}
