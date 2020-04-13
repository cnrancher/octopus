package node

import (
	"context"
	"math/rand"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetValidWorker(ctx context.Context, k8sCli client.Client) (string, error) {
	var list = corev1.NodeList{}
	if err := k8sCli.List(ctx, &list); err != nil {
		return "", err
	}

	var workers []string
	for _, node := range list.Items {
		if labels.Set(node.GetLabels()).Has("node-role.kubernetes.io/master") {
			continue
		}
		workers = append(workers, node.Name)
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

	var workers []string
	for _, node := range list.Items {
		if labels.Set(node.GetLabels()).Has("node-role.kubernetes.io/master") {
			continue
		}
		workers = append(workers, node.Name)
	}

	if len(workers) == 0 {
		return "", errors.New("no workers")
	}

	var idx = rand.Intn(len(workers))
	var name = workers[idx]

	// reserve
	var nameRunes = []rune(name)
	var nameRunesSize = len(nameRunes)
	for i := nameRunesSize/2 - 1; i >= 0; i-- {
		var opp = nameRunesSize - 1 - i
		nameRunes[i], nameRunes[opp] = nameRunes[opp], nameRunes[i]
	}

	return "i-" + string(nameRunes), nil
}
