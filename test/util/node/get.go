// +build test

package node

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rancher/octopus/test/util/fuzz"
)

func GetValidWorker(ctx context.Context, k8sCli client.Client) (string, error) {
	var list = corev1.NodeList{}
	if err := k8sCli.List(ctx, &list); err != nil {
		return "", err
	}

	var workers []string
	for _, node := range list.Items {
		if IsOnlyWorker(&node) {
			workers = append(workers, node.Name)
		}
	}

	if len(workers) == 0 {
		return "", errors.New("no workers")
	}

	var idx = rand.Intn(len(workers))
	var name = workers[idx]
	return name, nil
}

func GetInvalidWorker(ctx context.Context, k8sCli client.Client) (string, error) {
	var list = corev1.NodeList{}
	if err := k8sCli.List(ctx, &list); err != nil {
		return "", err
	}

	var workers = sets.NewString()
	for _, node := range list.Items {
		if IsOnlyWorker(&node) {
			workers.Insert(node.Name)
		}
	}

	if workers.Len() == 0 {
		return "", errors.New("no workers")
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
