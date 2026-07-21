package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var dockerClient *client.Client

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	fmt.Println("🌌 AETHER ORCHESTRATOR v5.0 (State Sync Enabled)")
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("❌ Docker Error: %s", err)
	}
	defer dockerClient.Close()

	ctx := context.Background()
	_, _ = dockerClient.NetworkCreate(ctx, "aether-network", types.NetworkCreate{CheckDuplicate: true})

	http.HandleFunc("/provision", enableCORS(handleProvision))
	http.HandleFunc("/terminate", enableCORS(handleTerminate))
	http.HandleFunc("/sandboxes", enableCORS(handleList)) // NEW: List active sandboxes

	port := ":8081"
	fmt.Printf("📡 API Engine listening on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// NEW: Handlers that queries Docker to find currently running Aether containers
func handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list := make([]map[string]string, 0)
	for _, c := range containers {
		if len(c.Names) > 0 {
			name := c.Names[0] // e.g., "/sb-1234"
			// Only include containers started by Aether (prefix sb-)
			if len(name) > 4 && name[1:4] == "sb-" {
				id := name[1:] // remove leading slash
				list = append(list, map[string]string{
					"id":  id,
					"url": fmt.Sprintf("http://localhost:5005/sandbox/%s/", id),
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func handleProvision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	sandboxId := fmt.Sprintf("sb-%d", time.Now().Unix()%10000)
	imageName := "docker.io/library/nginx:alpine"

	out, _ := dockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
	io.Copy(io.Discard, out)
	out.Close()

	resp, err := dockerClient.ContainerCreate(ctx,
		&container.Config{Image: imageName, Hostname: sandboxId},
		&container.HostConfig{
			Resources:  container.Resources{Memory: 512 * 1024 * 1024, NanoCPUs: 500000000},
			AutoRemove: true,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{"aether-network": {}},
		}, nil, sandboxId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})

	payload, _ := json.Marshal(map[string]string{
		"SandboxId":       sandboxId,
		"InternalAddress": fmt.Sprintf("http://%s:80", sandboxId),
	})
	_, _ = http.Post("http://localhost:5005/api/routes/register", "application/json", bytes.NewBuffer(payload))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": sandboxId, "url": fmt.Sprintf("http://localhost:5005/sandbox/%s/", sandboxId)})
}

func handleTerminate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	ctx := context.Background()
	timeout := 5
	_ = dockerClient.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:5005/api/routes/unregister/%s", id), nil)
	_, _ = http.DefaultClient.Do(req)
	w.WriteHeader(http.StatusOK)
}
