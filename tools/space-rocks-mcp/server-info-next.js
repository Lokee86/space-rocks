import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";

import { listenMcpHttpServer } from "./shared/http_mcp_server.js";
import { registerRepoReadonlyTools } from "./shared/repo_readonly_tools.js";
import { registerEngineForgeReadonlyTools } from "./shared/engineforge_readonly_tools.js";

const port = Number(process.env.PORT ?? 8789);

function createMcpServer() {
  const server = new McpServer({
    name: "space-rocks-info-mcp",
    version: "0.3.0",
  });

  registerRepoReadonlyTools(server);
  registerEngineForgeReadonlyTools(server);

  return server;
}

listenMcpHttpServer({
  port,
  label: "Space Rocks info",
  createMcpServer,
});
