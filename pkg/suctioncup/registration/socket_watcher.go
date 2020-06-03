package registration

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/rancher/octopus/pkg/suctioncup/adaptor"
	"github.com/rancher/octopus/pkg/suctioncup/event"
)

func newSocketWatcher(log logr.Logger, dir string, adaptors adaptor.Adaptors, adaptorNotifier event.AdaptorNotifier) (*socketWatcher, error) {
	var fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &socketWatcher{
		log:       log,
		dir:       dir,
		set:       adaptors,
		notifier:  adaptorNotifier,
		fsWatcher: fsWatcher,
	}, nil
}

type socketWatcher struct {
	log      logr.Logger
	dir      string
	set      adaptor.Adaptors
	notifier event.AdaptorNotifier

	fsWatcher *fsnotify.Watcher
}

func (w *socketWatcher) Watch(adp adaptor.Adaptor) error {
	var path = filepath.Join(w.dir, adp.GetEndpoint())
	if err := w.fsWatcher.Add(path); err != nil {
		return errors.Wrapf(err, "failed to add watching on %s", path)
	}

	w.set.Put(adp)
	// use another loop to reduce the blocking of rpc,
	// at the same time, that loop ensures that all links will be updated.
	w.notifier.NoticeAdaptorRegistered(adp.GetName())

	w.log.V(2).Info("Watching path", "path", path)
	return nil
}

func (w *socketWatcher) Start(stop <-chan struct{}) {
loop:
	for {
		select {
		case ent, ok := <-w.fsWatcher.Events:
			if !ok {
				break loop
			}
			if ent.Op&fsnotify.Remove == fsnotify.Remove {
				var path = ent.Name
				var _, endpoint = filepath.Split(path)
				if adp := w.set.Get(endpoint); adp != nil {
					w.notifier.NoticeAdaptorUnregistered(adp.GetName())
					w.set.Delete(adp.GetEndpoint())
					_ = w.fsWatcher.Remove(path)
					w.log.V(2).Info("Unwatching path", "path", path)
				}
			}
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				break loop
			}
			if err != nil {
				w.log.Error(err, "Receive error")
			}
		case <-stop:
			_ = w.fsWatcher.Close()
			break loop
		}
	}
}
