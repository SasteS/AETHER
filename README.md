# AETHER
### Enterprise-grade Ephemeral Environments-as-a-Service (EaaS)

**Aether** is a high-performance orchestration platform designed to provision isolated, secure, and disposable full-stack "sandboxes" on demand. Built for the modern enterprise, it eliminates staging bottlenecks by giving every developer, feature branch, or CI/CD pipeline its own ephemeral environment.

---

## The Vision
In large organizations, waiting for a "free" staging environment slows down velocity. Aether allows users to spin up a complete stack (Frontend, Backend, Database) in seconds, isolated within Docker networks, and automatically destroyed after a set TTL (Time-to-Live).

## Tech Stack & Architecture
Aether uses a polyglot microservices architecture to leverage the best tool for every task:

*   **The Gateway (.NET 8 + YARP):** A high-performance dynamic reverse proxy that handles OAuth2 authentication (Keycloak) and routes traffic to ephemeral containers based on subdomains.
*   **The Orchestrator (Go):** A systems-level service communicating directly with the **Docker Engine API** for low-latency container provisioning and resource governance.
*   **The Security Vault (HashiCorp Vault):** Centralized secret management. No environment variables; secrets are injected into containers via the Vault API at runtime.
*   **The Dashboard (React + TypeScript):** A sleek, professional UI built with Vite and Tailwind CSS for real-time environment management and log streaming.
*   **Communication (gRPC):** Internal service-to-service communication using protocol buffers for high-speed, type-safe data transfer.

## Enterprise Features
- **Dynamic Routing:** Real-time proxy reconfiguration using Microsoft's YARP.
- **Resource Governance:** Hard CPU and Memory limits enforced via Docker Cgroups.
- **Network Isolation:** Automatic creation of dedicated Docker bridge networks per sandbox.
- **Secret Zero-Trust:** Integration with HashiCorp Vault for dynamic credential generation.
- **Observability:** Real-time container log streaming via WebSockets.

---

## Project Structure
/src
├── Aether.Gateway # .NET 8 Reverse Proxy & Auth
├── Aether.Orchestrator # Go Docker Controller
└── Aether.Web # React/Vite Dashboard
/infra
└── docker-compose.yaml # Platform bootstrap (Vault, Keycloak, DB)