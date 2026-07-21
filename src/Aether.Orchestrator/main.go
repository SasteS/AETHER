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

// Global Docker client to be reused by API calls
var dockerClient *client.Client

// --- NEW: CORS MIDDLEWARE ---
// This function wraps our handlers to allow the React Frontend to communicate
// with this API from a different port (Cross-Origin).
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set headers to allow requests from your React dev server
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Important: Browsers send an "OPTIONS" request before the real POST
		// to check if they are allowed to talk to the server.
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Proceed to the actual logic
		next(w, r)
	}
}

func main() {
	fmt.Println("🌌 AETHER ORCHESTRATOR v3.1 (CORS Enabled)")

	// 1. Initialize Docker Client
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("❌ Docker Connection Failed: %s", err)
	}
	defer dockerClient.Close()

	// 2. Ensure Infrastructure is ready
	ctx := context.Background()
	_, _ = dockerClient.NetworkCreate(ctx, "aether-network", types.NetworkCreate{
		CheckDuplicate: true,
	})
	fmt.Println("🌐 Infrastructure: aether-network is ready.")

	// 3. Define API Routes wrapped with CORS middleware
	http.HandleFunc("/provision", enableCORS(handleProvision))

	http.HandleFunc("/health", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Healthy")
	}))

	// 4. Start the permanent Service
	port := ":8081"
	fmt.Printf("📡 API Engine listening on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleProvision(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests for provisioning
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	sandboxId := fmt.Sprintf("sb-%d", time.Now().Unix()%10000)
	imageName := "docker.io/library/nginx:alpine"

	fmt.Printf("🚀 API Trigger: Provisioning [%s]...\n", sandboxId)

	// 1. Pull Image
	out, err := dockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err == nil {
		io.Copy(io.Discard, out)
		out.Close()
	}

	// 2. Create with Resource Governance (512MB / 0.5 CPU)
	resp, err := dockerClient.ContainerCreate(ctx,
		&container.Config{
			Image:    imageName,
			Hostname: sandboxId,
		},
		&container.HostConfig{
			Resources: container.Resources{
				Memory:   512 * 1024 * 1024,
				NanoCPUs: 500000000,
			},
			AutoRemove: true,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"aether-network": {},
			},
		}, nil, sandboxId)

	if err != nil {
		fmt.Printf("❌ Creation Failed: %s\n", err)
		http.Error(w, "Orchestration Error", http.StatusInternalServerError)
		return
	}

	// 3. Start
	dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	fmt.Printf("✅ Container %s is online.\n", sandboxId)

	// --- Service Discovery: Notify .NET Gateway ---
	fmt.Println("📢 Notifying .NET Gateway...")
	payload, _ := json.Marshal(map[string]string{
		"SandboxId":       sandboxId,
		"InternalAddress": fmt.Sprintf("http://%s:80", sandboxId),
	})

	_, err = http.Post("http://localhost:5005/api/routes/register", "application/json", bytes.NewBuffer(payload))

	if err != nil {
		fmt.Printf("⚠️  Gateway notify failed: %s\n", err)
	}

	// 4. Respond to the Caller (React)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"id":     sandboxId,
		"url":    fmt.Sprintf("http://localhost:5005/sandbox/%s/", sandboxId),
	})
}
