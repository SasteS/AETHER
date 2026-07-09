# Aether: Ephemeral Environments-as-a-Service

Aether is a cloud-native orchestration platform designed to solve the "Staging Bottleneck" common in large-scale enterprise development. In traditional environments, developers often share a limited number of static staging servers, leading to configuration drift and deployment delays. Aether eliminates this friction by providing a self-service platform that provisions fully isolated, secure, and disposable full-stack sandboxes on demand.

### System Architecture and Narrative
The platform operates as a polyglot microservices ecosystem. At the edge, the **Aether Gateway**—built on .NET 8 and utilizing Microsoft’s **YARP (Yet Another Reverse Proxy)**—manages incoming traffic. It handles complex routing logic, ensuring that requests to specific subdomains are dynamically mapped to the correct ephemeral containers while enforcing strict OAuth2 authentication. 

The heavy lifting of container lifecycle management is handled by the **Aether Orchestrator**. Written in Go for its high-performance concurrency model and native support for systems programming, this service communicates directly with the Docker Engine API. It is responsible for the precise allocation of system resources, enforcing hard CPU and memory limits, and managing the creation of isolated virtual bridge networks to ensure tenant separation.

Security is treated as a core pillar through a deep integration with **HashiCorp Vault**. Rather than relying on insecure environment variables, Aether implements a zero-trust model where secrets and database credentials are dynamically generated and injected into sandboxes at runtime. The entire system is managed via the **Aether Web Dashboard**, a React-based command center that provides developers with real-time observability through WebSocket-driven log streaming and automated TTL (Time-to-Live) management.

### Technical Rationale
The choice of a polyglot stack reflects a "right tool for the job" enterprise philosophy. .NET was selected for the gateway due to its robust ecosystem for middleware and the high-performance capabilities of YARP. Go was utilized for the orchestrator to leverage its low-level efficiency when interacting with container runtimes. For internal communication, the platform employs **gRPC**, providing high-performance, type-safe contracts between services that are significantly more efficient than standard REST interfaces for internal data transfer.

---

### Project Structure
```text
aether/
├── src/
│   ├── Aether.Gateway/       # .NET 8 Reverse Proxy & OAuth2 Logic
│   ├── Aether.Orchestrator/  # Go Service for Docker API Interaction
│   └── Aether.Web/           # React TypeScript Dashboard
├── infra/
│   ├── docker-compose.yaml   # Local Platform Infrastructure (Vault, DB, Keycloak)
│   └── vault/                # Secret Management Policies & Config
├── docs/                     # Architectural Diagrams & API Contracts
└── README.md
```

### Instructions for initialization:
1. Open your terminal in your root `aether` directory.
2. Run `nano README.md` or open it in VS Code.
3. Paste the text above and save.
4. Run the following to prepare your folders:

```bash
mkdir -p src/Aether.Gateway src/Aether.Orchestrator src/Aether.Web infra docs