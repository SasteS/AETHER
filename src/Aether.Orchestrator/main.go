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

func main() {
	fmt.Println("🌌 AETHER ORCHESTRATOR v2.0 (Dynamic Mode)")
	ctx := context.Background()

	// 1. Connect to Docker
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("❌ Docker Connection Failed: %s", err)
	}
	defer cli.Close() // Keep the connection cleanup from v1.0

	// 2. Ensure Network exists (Safety from v1.0)
	_, _ = cli.NetworkCreate(ctx, "aether-network", types.NetworkCreate{
		CheckDuplicate: true,
	})
	fmt.Println("🌐 Infrastructure: aether-network is ready.")

	// 3. Generate Unique Identity for Sandbox
	sandboxId := fmt.Sprintf("sb-%d", time.Now().Unix()%10000)
	imageName := "docker.io/library/nginx:alpine"
	fmt.Printf("🚀 Provisioning Sandbox [%s]...\n", sandboxId)

	// 4. Pull Image (Silent pull from v1.0)
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err == nil {
		io.Copy(io.Discard, out)
		out.Close()
	}

	// 5. Create Container with Enterprise Limits (Governance from v1.0)
	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:    imageName,
			Hostname: sandboxId,
		},
		&container.HostConfig{
			Resources: container.Resources{
				Memory:   512 * 1024 * 1024, // 512MB RAM
				NanoCPUs: 500000000,         // 0.5 CPU cores
			},
			AutoRemove: true,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"aether-network": {},
			},
		}, nil, sandboxId)

	if err != nil {
		log.Fatalf("❌ Creation Failed: %s", err)
	}

	// 6. Start the Sandbox
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatalf("❌ Start Failed: %s", err)
	}
	fmt.Printf("✅ Container %s is running.\n", sandboxId)

	// 7. Dynamic Registration (The v2.0 "Brain")
	fmt.Println("📢 Registering with .NET Gateway...")

	payload, _ := json.Marshal(map[string]string{
		"SandboxId":       sandboxId,
		"InternalAddress": fmt.Sprintf("http://%s:80", sandboxId),
	})

	// Call the .NET API we are about to build
	regResp, err := http.Post("http://localhost:5005/api/routes/register", "application/json", bytes.NewBuffer(payload))

	if err != nil {
		fmt.Printf("⚠️  Gateway Registration Failed (Is the Gateway running?): %s\n", err)
	} else {
		fmt.Printf("🎉 Successfully registered! Access at: http://localhost:5005/sandbox/%s/\n", sandboxId)
		regResp.Body.Close()
	}
}
