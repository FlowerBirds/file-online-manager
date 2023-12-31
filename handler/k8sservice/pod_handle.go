package k8sservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"file-online-manager/model"
	"file-online-manager/util"
	"fmt"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func InitK8sClient() *kubernetes.Clientset {
	kubeconfig := os.Getenv("KUBECONFIG")
	var config *rest.Config
	var err error
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			util.Println("load config failed: ", err)
			config = nil
		}
	}
	if config == nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			util.Println("load in cluster config failed: ", err)
			config = nil
		}
	}

	if config != nil {
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			util.Println(err)
			return nil
		}
		return clientset
	}
	return nil
}

func RestartPodHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.FormValue("namespace")
	name := r.FormValue("name")
	if name == "" || namespace == "" {
		util.Error(w, errors.New("empty pod params"))
		return
	}
	if namespace == "kube-system" {
		util.Error(w, errors.New("forbidden to restart kube-system pod"))
		return
	}
	util.Println("restart:", namespace, name)
	hostname := os.Getenv("HOSTNAME")
	clientset := InitK8sClient()
	if clientset != nil {
		ctx := context.Background()

		for _, podName := range strings.Split(name, ",") {
			if podName == "" {
				continue
			}
			_, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				util.Error(w, err)
				return
			}
			// forbidden to restart itself
			if hostname == podName {
				util.Println("Ignore restart pod: ", podName)
				continue
			}

			util.Println("Found and delete pod: ", podName)
			err = clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
			if err != nil {
				util.Error(w, err)
				return
			}
		}
		response := model.Response{Code: 200, Message: "Restart pod successfully", Data: true}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		return
	}
	util.Error(w, errors.New("invalid config"))
}

func ListPodHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.FormValue("namespace")
	if namespace == "" {
		util.Error(w, errors.New("empty pod params"))
		return
	}
	clientset := InitK8sClient()
	if clientset != nil {
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			util.Println(err)
			util.Error(w, err)
			return
		}
		var podInfos []model.PodInfo
		for _, pod := range pods.Items {

			var RestartCount int32 = 0
			if len(pod.Status.ContainerStatuses) > 0 {
				RestartCount = pod.Status.ContainerStatuses[0].RestartCount
			}

			podInfo := model.PodInfo{
				Name:      pod.ObjectMeta.Name,
				Namespace: pod.ObjectMeta.Namespace,
				Status:    getStatus(pod),
				Ready:     fmt.Sprintf("%d/%d", aliveContainer(pod.Status.ContainerStatuses), len(pod.Spec.Containers)),
				Restarts:  RestartCount,
				Age:       calculateAge(pod.Status),
				IP:        pod.Status.PodIP,
				Node:      pod.Spec.NodeName,
			}
			podInfos = append(podInfos, podInfo)
		}
		response := model.Response{Code: 200, Message: "List pods successfully", Data: podInfos}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		return
	}
	util.Error(w, errors.New("invalid config"))
}

func calculateAge(status v1.PodStatus) string {
	var creationTime metav1.Time
	if status.StartTime != nil {
		creationTime = *status.StartTime
	} else {
		return ""
	}
	now := metav1.Now()
	duration := now.Sub(creationTime.Time)
	// util.Println(duration.String())
	duration, err := time.ParseDuration(duration.String())
	if err != nil {
		return ""
	}

	var result []string
	if days := int(duration.Hours() / 24); days > 0 {
		result = append(result, fmt.Sprintf("%dd", days))
		duration -= time.Duration(days) * 24 * time.Hour
	}
	if hours := int(duration.Hours()); hours > 0 {
		result = append(result, fmt.Sprintf("%dh", hours))
		duration -= time.Duration(hours) * time.Hour
	}
	if minutes := int(duration.Minutes()); minutes > 0 {
		result = append(result, fmt.Sprintf("%dm", minutes))
		duration -= time.Duration(minutes) * time.Minute
	}
	if seconds := int(duration.Seconds()); seconds > 0 {
		result = append(result, fmt.Sprintf("%ds", seconds))
	}
	return strings.Join(result, " ")
}

func aliveContainer(containers []v1.ContainerStatus) int {
	i := 0
	for _, c := range containers {
		// fmt.Printf("- %s: %s\n", c.Name, c.State)
		if *c.Started && c.State.Running != nil {
			i++
		}
	}
	return i
}

func getStatus(pod v1.Pod) string {
	for _, c := range pod.Status.ContainerStatuses {
		if c.State.Waiting != nil {
			return c.State.Waiting.Reason
		}
		if c.State.Terminated != nil {
			return c.State.Terminated.Reason
		}
	}
	return string(pod.Status.Phase)
}

func ListNamespaceHandler(w http.ResponseWriter, r *http.Request) {
	clientset := InitK8sClient()
	if clientset != nil {
		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			util.Println(err)
			util.Error(w, err)
			return
		}
		var nsInfos []model.Namespace
		for _, ns := range namespaces.Items {
			nsInfo := model.Namespace{
				Name: ns.ObjectMeta.Name,
			}
			nsInfos = append(nsInfos, nsInfo)
		}
		response := model.Response{Code: 200, Message: "List namespace successfully", Data: nsInfos}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		return
	}
	util.Error(w, errors.New("invalid config"))
}

func PodStreamLogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	hostname := os.Getenv("HOSTNAME")

	if name == "" || namespace == "" {
		util.Error(w, errors.New("invalid query param"))
		return
	}
	if name == hostname {
		util.Error(w, errors.New("can't view itself logs due to cause to recursive access and leads to an infinite loop"))
		return
	}
	util.Println("read logs: ", namespace, name)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// ticker := time.NewTicker(time.Second * 1) // 每隔 1 秒发送一次数据
	// defer ticker.Stop()

	tailLines := int64(100)
	// containerName := ""
	clientset := InitK8sClient()
	if clientset != nil {
		stream := clientset.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{
			TailLines: &tailLines,
			Follow:    true,
		})
		if stream == nil {
			util.Error(w, errors.New("Failed to get Pod stream"))
			return
		}
		logs, err := stream.Stream(context.TODO())
		if err != nil {
			log.Printf("Failed to read Pod logs: %v\n", err)
			return
		}
		defer logs.Close()

		buf := make([]byte, 1024*1024)
		full := true
		reverseStr := ""
		for {
			// util.Println("read data")
			size, err := logs.Read(buf)
			if size > 0 {
				if buf[size-1] != 10 {
					full = false
				} else {
					full = true
				}
				readlogs := strings.Split(string(buf[:size]), "\n")
				for i, text := range readlogs {
					if len(text) == 0 {
						continue
					}
					if !full && i == len(readlogs)-1 {
						reverseStr += text
						// util.Println("reverse " + reverseStr)
						continue
					}
					data := []byte("data: " + text + "\n\n")
					if i == 0 && len(reverseStr) > 0 {
						// util.Println("add reverse " + reverseStr + text)
						data = []byte("data: " + reverseStr + text + "\n\n")
						reverseStr = ""
					}
					_, err = w.Write(data)
					if err != nil {
						fmt.Printf("Failed to write response: %v\n", err)
						return
					}
					flusher.Flush()
				}

			}
			if err != nil || size == 0 {
				break
			}
		}
		flusher.Flush() // 刷新响应，将数据发送到客户端
	} else {
		util.Error(w, errors.New("invalid k8s client"))
		return
	}
}

func ViewPodYamlHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	namespace := r.URL.Query().Get("namespace")
	if name == "" || namespace == "" {
		util.Error(w, errors.New("invalid query param"))
		return
	}
	util.Println("view yaml: ", namespace, name)
	clientset := InitK8sClient()
	if clientset != nil {
		pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			log.Printf("Failed to get pod: %v", err)
			util.Error(w, err)
			return
		}
		ownerReferences := pod.ObjectMeta.OwnerReferences
		util.Println(ownerReferences)
		if len(ownerReferences) == 0 {
			util.Error(w, errors.New("Pod does not have an owner"))
			return
		}
		replicaSetName := ownerReferences[0].Name
		replicaSet, err := clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), replicaSetName, metav1.GetOptions{})
		if err != nil {
			util.Println(err)
			util.Error(w, err)
			return
		}

		// 获取 ReplicaSet 的 ownerReferences
		ownerReferences = replicaSet.ObjectMeta.OwnerReferences
		if len(ownerReferences) == 0 {
			err = errors.New("ReplicaSet does not have an owner")
			util.Println(err)
			util.Error(w, err)
			return
		}
		deploymentName := ownerReferences[0].Name
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Failed to get deployment: %v", err)
			util.Error(w, err)
			return
		}

		deploymentYAML, err := deploymentToYAML(deployment)
		// util.Println(deploymentYAML)
		if err != nil {
			fmt.Printf("Failed to convert deployment to YAML: %v", err)
			util.Error(w, err)
			return
		}
		depInfo := model.DeploymentInfo{
			Name: deploymentName,
			Yaml: deploymentYAML,
		}
		response := model.Response{Code: 200, Message: "view deployment yaml successfully", Data: depInfo}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	} else {
		util.Error(w, errors.New("invalid k8s client"))
		return
	}
}

func deploymentToYAML(deployment *appsv1.Deployment) (string, error) {
	yamlContent, err := runtime.Encode(unstructured.UnstructuredJSONScheme, deployment)
	if err != nil {
		return "", err
	}

	// 打印YAML内容
	var buf bytes.Buffer
	_, err = buf.WriteString(string(yamlContent))
	if err != nil {
		return "", err
	}

	var obj interface{}
	err = json.Unmarshal([]byte(buf.String()), &obj)
	if err != nil {
		return "", err
	}

	// 将数据转换为YAML格式
	yamlData, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}
