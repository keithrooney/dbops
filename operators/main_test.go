package main

import (
	"context"
	"math/rand"
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fk8s "k8s.io/client-go/kubernetes/fake"
)

func TestCreateStrategy(t *testing.T) {
	clientset := fk8s.NewSimpleClientset()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: strconv.Itoa(rand.Int()),
		},
	}
	if _, err := clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{}); err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	databases := []*Database{
		NewDatabase(strconv.Itoa(rand.Int())),
		NewDatabase(namespace.ObjectMeta.Name),
	}

	strategy := CreateStrategy{
		Client: clientset,
	}

	for _, database := range databases {

		if err := strategy.Execute(&StrategyContext{Database: database}); err != nil {
			t.Fail()
			t.Logf(err.Error())
		}
		if _, err := clientset.CoreV1().Namespaces().Get(context.TODO(), database.Namespace, metav1.GetOptions{}); err != nil {
			t.Fail()
			t.Logf(err.Error())
		}
		if _, err := clientset.AppsV1().Deployments(database.Namespace).Get(context.TODO(), database.Name, metav1.GetOptions{}); err != nil {
			t.Fail()
			t.Logf(err.Error())
		}

	}

}
