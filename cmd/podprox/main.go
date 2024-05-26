package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const BufferSize = 4096
const RemoteManifestPath = "/etc/config/remote.yaml"

func deletePod(pod_name string) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Print(err)
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Print(err)
		return
	}

	err = clientset.CoreV1().Pods("default").Delete(context.TODO(), pod_name, metav1.DeleteOptions{})
	if err != nil {
		log.Print(err)
		return
	}
}

func createPod(pod_name string) string {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Print(err)
		return ""
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Print(err)
		return ""
	}

	// Read the YAML file
	data, err := os.ReadFile(RemoteManifestPath)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		log.Print(err.Error())
	}
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Print("YAML file is not a Pod")
	}

	pod.Name = pod_name

	// Create the Pod in the default namespace
	pod, err = clientset.CoreV1().Pods("default").Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create pod: %v", err)
	}

	// Wait for the Pod to be running
	log.Println("Waiting for Pod to be running...")
	for {
		p, err := clientset.CoreV1().Pods("default").Get(context.TODO(), pod.Name, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Failed to get pod: %v", err)
		}
		if p.Status.Phase == v1.PodRunning {
			pod = p
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Return first port of first container
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			return fmt.Sprintf("%s:%d", pod.Status.PodIP, port.ContainerPort)
		}
	}

	return ""
}

func proxy_to_remote(reader *bufio.Reader, writer *bufio.Writer) {
	buffer := make([]byte, BufferSize)
	for {
		log.Print("Reading from origin")
		n_read, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Print("Client closed connection")
			} else {
				log.Printf("Error reading from client: %v", err)
			}
			return
		}

		n_wrt, err := writer.Write(buffer[:n_read])
		if err != nil {
			log.Print("Remote closed connection")
			return
		}
		writer.Flush()

		log.Printf("Wrote %d bytes", n_wrt)
	}
}

func handleRequest(origin_conn net.Conn) {
	remote_addr := origin_conn.RemoteAddr().String()
	remote_addr = strings.Replace(remote_addr, ":", ".", 1)
	host := createPod(remote_addr)

	proxy_conn, err := net.Dial("tcp", host)
	if err != nil {
		log.Print(err)
		origin_conn.Close()
		return
	}

	proxy_writer := bufio.NewWriter(proxy_conn)
	proxy_reader := bufio.NewReader(proxy_conn)

	origin_reader := bufio.NewReader(origin_conn)
	origin_writer := bufio.NewWriter(origin_conn)

	var wg sync.WaitGroup

	wg.Add(2)
	// There and Back Again
	go func() {
		defer wg.Done()
		proxy_to_remote(origin_reader, proxy_writer)
	}()
	go func() {
		defer wg.Done()
		proxy_to_remote(proxy_reader, origin_writer)
	}()
	wg.Wait()
	proxy_conn.Close()
	origin_conn.Close()
	deletePod(remote_addr)
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleRequest(conn)
	}
}
