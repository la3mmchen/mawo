package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/la3mmchen/mawo/internal/types"
)

var (
	AppVersion string
)

/*
	index returns a 302 redirect if / is called
*/
func index(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMovedPermanently)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "see /container-resources"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Something went wrong.", 500)
	}
	w.Write(jsonResp)
	return
}

/*
	health just returns "alive" if called
*/
func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "alive"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Something went wrong.", 500)
	}
	w.Write(jsonResp)
	return
}

/*
	version returns the current deployed version
*/
func version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(AppVersion))
	return
}

/*
	containerResources query the kubernetes api for pods with the given podLabels.
		will return a formated json that contains the resource configuration for the filtered pods.
*/
func containerResources(w http.ResponseWriter, r *http.Request) {
	label, ok := r.URL.Query()["pod-label"]
	if !ok || len(label[0]) < 1 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "URL param 'pod-label' is missing.", 422)
		return
	}

	// creates the in-cluster config for every request to make sure we get a connection
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Print(err)
		http.Error(w, "Something went wrong.", 500)
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Print(err)
		http.Error(w, "Something went wrong.", 500)
		return
	}
	// read by query label
	podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{LabelSelector: strings.Join(label, ",")})

	// something went wrong
	if err != nil {
		http.Error(w, "Something went wrong.", 500)
		return
	}

	// we haven't found anything
	if errors.IsNotFound(err) {
		http.Error(w, "None found", 404)
		return
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
			m := map[string]string{
				"container_name": container.Name,
				"pod_name":       pod.Metadata.Name,
				"namespace":      pod.Metadata.Namespace,
				"mem_req":        container.Resources.Requests.Memory,
				"mem_limit":      container.Resources.Limits.Memory,
				"cpu_req":        container.Resources.Requests.Cpu,
				"cpu_limit":      container.Resources.Limits.Cpu,
			}
			queryResults = append(queryResults, m)
		}
	}

	asJson, err := json.Marshal(queryResults)
	if err != nil {
		http.Error(w, "Something went wrong.", 500)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(asJson)

	return
}

func main() {

	// check that we run on kubernetes before we start the server
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	_, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// read config from env
	httpPort := getEnv("MAWO_PORT", "80")

	http.HandleFunc("/", index)
	http.HandleFunc("/version", version)
	http.HandleFunc("/health", health)
	http.HandleFunc("/container-resources", containerResources)

	fmt.Printf("Starting server at port %v\n", httpPort)

	if err := http.ListenAndServe(fmt.Sprintf(":%v", httpPort), logRequest(http.DefaultServeMux)); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
