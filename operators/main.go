package main

import (
	"context"
	"log"
	"math/rand"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Database struct {
	Name      string
	Namespace string
	Replicas  int32
}

func NewDatabase(namespace string) *Database {
	return &Database{
		Name:      strconv.Itoa(rand.Int()),
		Namespace: namespace,
		Replicas:  int32(1),
	}
}

func NewConfig() *rest.Config {
	return &rest.Config{
		Host: "192.168.49.2:8443",
		TLSClientConfig: rest.TLSClientConfig{
			CertFile: "/home/krooney/.minikube/profiles/minikube/client.crt",
			KeyFile:  "/home/krooney/.minikube/profiles/minikube/client.key",
			CAFile:   "/home/krooney/.minikube/ca.crt",
		},
	}
}

func NewClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(NewConfig())
}

func NewNamespace(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func NewDeployment(database *Database) *appsv1.Deployment {
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      database.Name,
			Namespace: database.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &database.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "redis",
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "redis",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           "krooney/redis:7.0.2",
							Name:            "db",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"redis-server",
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 6379,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "redis",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: filepath.Join("/", "var", "lib", "redis", database.Namespace, database.Name),
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
					},
				},
			},
		},
	}
}

type CreateStrategy struct {
	Client kubernetes.Interface
}

type StrategyContext struct {
	Database *Database
}

func (strategy *CreateStrategy) Execute(sc *StrategyContext) error {
	namespaces := strategy.Client.CoreV1().Namespaces()
	namespace, err := namespaces.Get(context.TODO(), sc.Database.Namespace, metav1.GetOptions{})
	if err != nil {
		namespace = NewNamespace(sc.Database.Namespace)
		_, err = namespaces.Create(context.TODO(), namespace, metav1.CreateOptions{})
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
	}
	deployment := NewDeployment(sc.Database)
	deployments := strategy.Client.AppsV1().Deployments(namespace.Name)
	if _, err = deployments.Create(context.TODO(), deployment, metav1.CreateOptions{}); err != nil {
		log.Fatal(err.Error())
		return err
	}
	return nil
}

type Account struct {
	ID    string
	Name  string
	Email string
}

func NewAccount() *Account {
	return &Account{
		ID:    uuid.New().String(),
		Name:  "DBOps",
		Email: "development@dbops.com",
	}
}

func main() {

	account := NewAccount()

	databases := []*Database{
		NewDatabase(account.ID),
	}

	client, err := NewClient()
	if err != nil {
		panic(err.Error())
	}

	strategy := CreateStrategy{Client: client}
	for _, database := range databases {
		strategy.Execute(&StrategyContext{Database: database})
	}

}
