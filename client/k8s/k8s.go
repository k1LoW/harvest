package k8s

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func NewKubeClientSet(contextName string) (*kubernetes.Clientset, error) {
	kubeconfig, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	clientConfig := newClientConfig(kubeconfig, contextName)
	c, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return clientset, err
}

func GetContainers(contextName string, namespace string, podFilter *regexp.Regexp) ([]string, error) {
	clientset, err := NewKubeClientSet(contextName)
	if err != nil {
		return nil, err
	}
	list, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	containers := []string{}
	for _, i := range list.Items {
		if !podFilter.MatchString(i.GetName()) {
			continue
		}
		for _, c := range i.Spec.Containers {
			containers = append(containers, strings.Join([]string{"", i.GetNamespace(), i.GetName(), c.Name}, "/"))
		}
	}
	return containers, nil
}

func newClientConfig(configPath string, contextName string) clientcmd.ClientConfig {
	configPathList := filepath.SplitList(configPath)
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{}
	if len(configPathList) <= 1 {
		configLoadingRules.ExplicitPath = configPath
	} else {
		configLoadingRules.Precedence = configPathList
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: contextName,
		},
	)
}

func getKubeConfig() (string, error) {
	var kubeconfig string

	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig, nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home directory")
	}

	kubeconfig = filepath.Join(home, ".kube/config")

	return kubeconfig, nil
}
