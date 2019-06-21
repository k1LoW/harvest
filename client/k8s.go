package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/k1LoW/harvest/client/k8s"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type K8sClient struct {
	contextName string
	namespace   string
	pod         string
	podFilter   *regexp.Regexp
	clientset   *kubernetes.Clientset
	lineChan    chan Line
	logger      *zap.Logger
}

// NewK8sClient ...
func NewK8sClient(l *zap.Logger, host, path string) (Client, error) {
	contextName := host
	splited := strings.Split(path, "/")
	ns := splited[1]
	p := splited[2]
	pRegexp := regexp.MustCompile(strings.Replace(strings.Replace(p, ".*", "*", -1), "*", ".*", -1))

	clientset, err := k8s.NewKubeClientSet(contextName)
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		contextName: contextName,
		namespace:   ns,
		pod:         p,
		podFilter:   pRegexp,
		clientset:   clientset,
		lineChan:    make(chan Line),
		logger:      l,
	}, nil
}

// Read ...
func (c *K8sClient) Read(ctx context.Context, st *time.Time, et *time.Time) error {
	var sinceSeconds int64
	sinceSeconds = time.Now().Unix() - st.Unix()
	return c.Stream(ctx, false, &sinceSeconds)
}

// Tailf ...
func (c *K8sClient) Tailf(ctx context.Context) error {
	var sinceSeconds int64
	sinceSeconds = 1
	return c.Stream(ctx, true, &sinceSeconds)
}

// Ls ...
func (c *K8sClient) Ls(ctx context.Context, st *time.Time, et *time.Time) error {
	return nil
}

// Copy ...
func (c *K8sClient) Copy(ctx context.Context, filePath string, dstDir string) error {
	return nil
}

// RandomOne ...
func (c *K8sClient) RandomOne(ctx context.Context) error {
	return nil
}

// Out ...
func (c *K8sClient) Out() <-chan Line {
	return c.lineChan
}

// Reference code:
// https://github.com/wercker/stern/blob/473d1b605673d8f4bfe5f86b3748d02c87d339d7/stern/main.go
// https://github.com/wercker/stern/blob/473d1b605673d8f4bfe5f86b3748d02c87d339d7/stern/tail.go

//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

// targetContainer is a target to watch
type targetContainer struct {
	namespace string
	pod       string
	container string
}

func (tc *targetContainer) getID() string {
	return fmt.Sprintf("%s-%s-%s", tc.namespace, tc.pod, tc.container)
}

// Stream ...
func (c *K8sClient) Stream(ctx context.Context, follow bool, sinceSeconds *int64) error {
	defer close(c.lineChan)
	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	added, removed, err := watchContainers(innerCtx, c.clientset.CoreV1().Pods(c.namespace), c.podFilter)
	if err != nil {
		return err
	}

	tails := make(map[string]*Tail)

	go func() {
		for tc := range added {
			id := tc.getID()
			if tails[id] != nil {
				continue
			}

			tail := NewTail(c.logger, c.lineChan, c.contextName, tc.namespace, tc.pod, tc.container)
			tails[id] = tail

			tail.Start(innerCtx, c.clientset.CoreV1().Pods(tc.namespace), follow, sinceSeconds)
		}
	}()

	go func() {
		for tc := range removed {
			id := tc.getID()
			if tails[id] == nil {
				delete(tails, id)
				continue
			}
			tails[id].Close()
			delete(tails, id)
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
	L:
		for {
			select {
			case <-ticker.C:
				for id, t := range tails {
					if t.Closed {
						delete(tails, id)
					}
				}
				if len(tails) == 0 {
					cancel()
				}
			case <-innerCtx.Done():
				break L
			}
		}
	}()

	<-innerCtx.Done()

	return nil
}

func watchContainers(ctx context.Context, i v1.PodInterface, podFilter *regexp.Regexp) (chan *targetContainer, chan *targetContainer, error) {
	watcher, err := i.Watch(metav1.ListOptions{Watch: true})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to set up watch")
	}

	added := make(chan *targetContainer)
	removed := make(chan *targetContainer)

	go func() {
		for {
			select {
			case e := <-watcher.ResultChan():
				if e.Object == nil {
					// Closed because of error
					return
				}

				pod := e.Object.(*corev1.Pod)

				if !podFilter.MatchString(pod.Name) {
					continue
				}

				switch e.Type {
				case watch.Added, watch.Modified:
					var statuses []corev1.ContainerStatus
					statuses = append(statuses, pod.Status.InitContainerStatuses...)
					statuses = append(statuses, pod.Status.ContainerStatuses...)

					for _, c := range statuses {
						added <- &targetContainer{
							namespace: pod.Namespace,
							pod:       pod.Name,
							container: c.Name,
						}
					}
				case watch.Deleted:
					var containers []corev1.Container
					containers = append(containers, pod.Spec.Containers...)
					containers = append(containers, pod.Spec.InitContainers...)

					for _, c := range containers {
						removed <- &targetContainer{
							namespace: pod.Namespace,
							pod:       pod.Name,
							container: c.Name,
						}
					}
				}
			case <-ctx.Done():
				watcher.Stop()
				close(added)
				close(removed)
				return
			}
		}
	}()

	return added, removed, nil
}

type Tail struct {
	ContextName   string
	Namespace     string
	PodName       string
	ContainerName string
	Closed        bool
	lineChan      chan Line
	req           *rest.Request
	closed        chan struct{}
	logger        *zap.Logger
}

// NewTail returns a new tail for a Kubernetes container inside a pod
func NewTail(l *zap.Logger, lineChan chan Line, contextName, namespace, podName, containerName string) *Tail {
	return &Tail{
		ContextName:   contextName,
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		lineChan:      lineChan,
		closed:        make(chan struct{}),
		logger:        l,
	}
}

// Start starts tailing
func (t *Tail) Start(ctx context.Context, i v1.PodInterface, follow bool, sinceSeconds *int64) {
	tz := "+0000"
	go func() {
		t.logger.Info(fmt.Sprintf("Open stream: /%s/%s/%s", t.Namespace, t.PodName, t.ContainerName))
		req := i.GetLogs(t.PodName, &corev1.PodLogOptions{
			Follow:       follow,
			Timestamps:   false,
			Container:    t.ContainerName,
			SinceSeconds: sinceSeconds,
		})

		stream, err := req.Stream()
		if err != nil {
			t.logger.Error(fmt.Sprintf("Error opening stream to /%s/%s/%s", t.Namespace, t.PodName, t.ContainerName))
			return
		}
		defer stream.Close()

		go func() {
			<-t.closed
			stream.Close()
		}()

		reader := bufio.NewReader(stream)
	L:
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					t.logger.Error(fmt.Sprintf("%s", err))
				}
				break L
			}

			t.lineChan <- Line{
				Host:     t.ContextName,
				Path:     strings.Join([]string{"", t.Namespace, t.PodName, t.ContainerName}, "/"),
				Content:  strings.TrimSuffix(string(line), "\n"),
				TimeZone: tz,
			}

			select {
			case <-ctx.Done():
				break L
			default:
			}
		}
		t.Close()
	}()
}

// Close stops tailing
func (t *Tail) Close() {
	t.logger.Info(fmt.Sprintf("Close stream to /%s/%s/%s", t.Namespace, t.PodName, t.ContainerName))
	t.Closed = true
	close(t.closed)
}
