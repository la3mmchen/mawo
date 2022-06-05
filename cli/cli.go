package main

import (
	"os"
	"fmt"
	"strings"
	"encoding/json"
	"context"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/la3mmchen/mawo/internal/types"
)

func main() {
		// handle input params
		if (len(os.Args[1:]) == 0 ) {
			fmt.Println("***")
			fmt.Println("")
			fmt.Println("  Non-sophisticated kubernetes query by label tooling")
			fmt.Println("  Just do ./bin/mawo-cli <label-name>=<label-value>")
			fmt.Println("  eg    ./bin/mawo-cli tier=control-plane")
			fmt.Println("")
			fmt.Println("***")
			fmt.Println("")
			os.Exit(0)
		}
		queryLabel := os.Args[1:]

		// create a kubernetes config from current loaded config
    kubeconfig := filepath.Join(
        os.Getenv("HOME"), ".kube", "config",
    )

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
    }

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}

		// read by query label
		podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{LabelSelector: strings.Join(queryLabel, ","),})

		// something went wrong
		if  err != nil {
			fmt.Printf("Something went wrong. http %i", 500)
			os.Exit(500)
		}

		// we haven't found anything
		if errors.IsNotFound(err) {
			fmt.Printf("None found. http %i", 404)
			os.Exit(404)
		}

		// marshal received list of pods to json
		// 	to be able to unmarshal into a structure that only contains
		//	the values we want to see
		var podlistfiltered types.PodListFiltered
		podListJson, err := json.Marshal(podList)
		json.Unmarshal(podListJson, &podlistfiltered)

		// range and create target structure
		queryResults := make([]map[string]string, 0, len(podlistfiltered.Items))
		for _, pod := range podlistfiltered.Items {
			for _, container := range pod.Spec.Containers {
				m := map[string]string {
					"container_name": container.Name,
					"pod_name": pod.Metadata.Name,
					"namespace": pod.Metadata.Namespace,
					"mem_req": container.Resources.Requests.Memory,
					"mem_limit": container.Resources.Limits.Memory,
					"cpu_req": container.Resources.Requests.Cpu,
					"cpu_limit": container.Resources.Limits.Cpu,
				}
				queryResults = append(queryResults, m)
			}
		}

		asJson, err := json.Marshal(queryResults)
    if err != nil {
        fmt.Printf("Error: %s", err.Error())
    }

		fmt.Printf("%v \n", string(asJson))

		os.Exit(0)
}