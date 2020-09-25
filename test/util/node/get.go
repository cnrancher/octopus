// +build test

package node

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rancher/octopus/test/util/fuzz"
)

func GetValidWorker(ctx context.Context, k8sCli client.Client) (string, error) {
	var workers, err = getWorkerSet(ctx, k8sCli)
	if err != nil {
		return "", err
	}

	var idx = rand.Intn(workers.Len())
	var name = workers.UnsortedList()[idx]
	return name, nil
}

func GetInvalidWorker(ctx context.Context, k8sCli client.Client) (string, error) {
	var workers, err = getWorkerSet(ctx, k8sCli)
	if err != nil {
		return "", err
	}

	var invalidWorker string
	for {
		invalidWorker = fmt.Sprintf("i-%s", strings.ToLower(fuzz.String(6)))
		if !workers.Has(invalidWorker) {
			break
		}
	}
	return invalidWorker, nil
}

func getWorkerSet(ctx context.Context, k8sCli client.Client) (sets.String, error) {
	var workers = sets.NewString()
	var err = wait.Poll(3*time.Second, 30*time.Second, func() (bool, error) {
		var list = corev1.NodeList{}
		if err := k8sCli.List(ctx, &list); err != nil {
			return false, err
		}
		for _, node := range list.Items {
			if IsOnlyWorker(&node) {
				workers.Insert(node.Name)
			}
		}
		if workers.Len() == 0 {
			return false, errors.New("no workers")
		}
		return true, nil
	})
	return workers, err
}
