using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Primitives;
using Yarp.ReverseProxy.Configuration;

var builder = WebApplication.CreateBuilder(args);

// 1. Setup the Dynamic Configuration Provider
var configProvider = new InMemoryConfigProvider();
builder.Services.AddSingleton<IProxyConfigProvider>(configProvider);
builder.Services.AddReverseProxy();

var app = builder.Build();

// 2. The API Endpoint for the Go Orchestrator
// Go will call this: POST http://localhost:5005/api/routes/register
app.MapPost("/api/routes/register", ([FromBody] ProxyRegistration request) =>
{
    configProvider.UpdateConfig(request.SandboxId, request.InternalAddress);
    return Results.Ok(new { message = $"Route registered for {request.SandboxId}" });
});

// 3. Landing page and Debug view
app.MapGet("/", () => "Aether Gateway: Active & Listening for registrations...");
app.MapGet("/debug", (IProxyConfigProvider provider) => Results.Ok(provider.GetConfig()));

app.MapReverseProxy();
app.Run();

// --- Data Structures ---
public record ProxyRegistration(string SandboxId, string InternalAddress);

// --- The Enterprise "In-Memory" Proxy Manager ---
public class InMemoryConfigProvider : IProxyConfigProvider
{
    private volatile CustomConfig _config;

    public InMemoryConfigProvider()
    {
        _config = new CustomConfig(new List<RouteConfig>(), new List<ClusterConfig>());
    }

    public IProxyConfig GetConfig() => _config;

    public void UpdateConfig(string sandboxId, string internalAddress)
    {
        var routes = _config.Routes.ToList();
        var clusters = _config.Clusters.ToList();

        // Create a route: /sandbox/{id}/{**catch-all}
        var route = new RouteConfig
        {
            RouteId = $"route-{sandboxId}",
            ClusterId = $"cluster-{sandboxId}",
            Match = new RouteMatch { Path = $"/sandbox/{sandboxId}/{{**catch-all}}" },
            Transforms = new[] { new Dictionary<string, string> { { "PathRemovePrefix", $"/sandbox/{sandboxId}" } } }
        };

        // Create a cluster: Point to the internal Docker DNS name
        var cluster = new ClusterConfig
        {
            ClusterId = $"cluster-{sandboxId}",
            Destinations = new Dictionary<string, DestinationConfig>
            {
                { "dest1", new DestinationConfig { Address = internalAddress } }
            }
        };

        routes.Add(route);
        clusters.Add(cluster);

        // Swap the old config for the new one and signal YARP to refresh
        var oldConfig = _config;
        _config = new CustomConfig(routes, clusters);
        oldConfig.SignalChange();
    }

    private class CustomConfig : IProxyConfig
    {
        private readonly CancellationTokenSource _cts = new();
        public IReadOnlyList<RouteConfig> Routes { get; }
        public IReadOnlyList<ClusterConfig> Clusters { get; }
        public IChangeToken ChangeToken { get; }

        public CustomConfig(IReadOnlyList<RouteConfig> routes, IReadOnlyList<ClusterConfig> clusters)
        {
            Routes = routes;
            Clusters = clusters;
            ChangeToken = new CancellationChangeToken(_cts.Token);
        }

        public void SignalChange() => _cts.Cancel();
    }
}