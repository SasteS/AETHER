using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Primitives;
using Yarp.ReverseProxy.Configuration;

var builder = WebApplication.CreateBuilder(args);

var configProvider = new InMemoryConfigProvider();
builder.Services.AddSingleton<IProxyConfigProvider>(configProvider);
builder.Services.AddReverseProxy();

var app = builder.Build();

// Endpoint to Register a new sandbox
app.MapPost("/api/routes/register", ([FromBody] ProxyRegistration request) =>
{
    configProvider.UpdateConfig(request.SandboxId, request.InternalAddress);
    return Results.Ok(new { message = $"Route registered for {request.SandboxId}" });
});

// NEW: Endpoint to Unregister/Delete a sandbox route
app.MapDelete("/api/routes/unregister/{sandboxId}", (string sandboxId) =>
{
    configProvider.RemoveConfig(sandboxId);
    return Results.Ok(new { message = $"Route removed for {sandboxId}" });
});

app.MapGet("/", () => "Aether Gateway: Active");
app.MapGet("/debug", (IProxyConfigProvider provider) => Results.Ok(provider.GetConfig()));

app.MapReverseProxy();
app.Run();

public record ProxyRegistration(string SandboxId, string InternalAddress);

public class InMemoryConfigProvider : IProxyConfigProvider
{
    private volatile CustomConfig _config;
    public InMemoryConfigProvider() => _config = new CustomConfig(new List<RouteConfig>(), new List<ClusterConfig>());
    public IProxyConfig GetConfig() => _config;

    public void UpdateConfig(string sandboxId, string internalAddress)
    {
        var routes = _config.Routes.ToList();
        var clusters = _config.Clusters.ToList();

        routes.Add(new RouteConfig
        {
            RouteId = $"route-{sandboxId}",
            ClusterId = $"cluster-{sandboxId}",
            Match = new RouteMatch { Path = $"/sandbox/{sandboxId}/{{**catch-all}}" },
            Transforms = new[] { new Dictionary<string, string> { { "PathRemovePrefix", $"/sandbox/{sandboxId}" } } }
        });

        clusters.Add(new ClusterConfig
        {
            ClusterId = $"cluster-{sandboxId}",
            Destinations = new Dictionary<string, DestinationConfig> { { "dest1", new DestinationConfig { Address = internalAddress } } }
        });

        var oldConfig = _config;
        _config = new CustomConfig(routes, clusters);
        oldConfig.SignalChange();
    }

    // NEW: Logic to remove routes from memory
    public void RemoveConfig(string sandboxId)
    {
        var routes = _config.Routes.Where(r => r.RouteId != $"route-{sandboxId}").ToList();
        var clusters = _config.Clusters.Where(c => c.ClusterId != $"cluster-{sandboxId}").ToList();

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
            Routes = routes; Clusters = clusters;
            ChangeToken = new CancellationChangeToken(_cts.Token);
        }
        public void SignalChange() => _cts.Cancel();
    }
}