using Yarp.ReverseProxy.Configuration;

var builder = WebApplication.CreateBuilder(args);

// 1. Add YARP Services
// We use 'LoadFromMemory' so we can update routes 
// without restarting the Gateway.
//builder.Services.AddReverseProxy()
//    .LoadFromMemory(new List<RouteConfig>(), new List<ClusterConfig>());
// 1. Add YARP Services with a test route to the Go-created sandbox
builder.Services.AddReverseProxy()
    .LoadFromMemory(new[]
    {
        new RouteConfig
        {
            RouteId = "sandbox-route",
            ClusterId = "sandbox-cluster",
            Match = new RouteMatch { Path = "/test/{**catch-all}" },
            // ADD THIS BLOCK:
            Transforms = new List<Dictionary<string, string>>
            {
                new Dictionary<string, string> { { "PathRemovePrefix", "/test" } }
            }
        }
    },
    new[]
    {
        new ClusterConfig
        {
            ClusterId = "sandbox-cluster",
            Destinations = new Dictionary<string, DestinationConfig>
            {
                // Use the EXACT name from your Docker Desktop: eloquent_aryabhata
                { "dest1", new DestinationConfig { Address = "http://sandbox1:80" } }
            }
        }
    });

var app = builder.Build();

// Landing page for verification
app.MapGet("/", () => Results.Text("Aether Gateway is Online\n" +
                                   "Mode: Dynamic Reverse Proxy\n" +
                                   "Runtime: .NET 8 in Docker"));

// Debug endpoint to see active sandboxes
app.MapGet("/debug/config", (IProxyConfigProvider configProvider) =>
{
    var config = configProvider.GetConfig();
    return Results.Ok(new { config.Routes, config.Clusters });
});

app.MapReverseProxy();

app.Run();