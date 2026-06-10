import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";

import { listenMcpHttpServer } from "./shared/http_mcp_server.js";
import { registerEngineForgeWriteTools } from "./shared/engineforge_write_tools.js";
import { registerRepoWriteTools } from "./shared/repo_write_tools.js";

const port = Number(process.env.PORT ?? 8788);

function createMcpServer() {
  const server = new McpServer({
    name: "space-rocks-write-mcp",
    version: "0.1.0",
  });

  registerRepoWriteTools(server);
  registerEngineForgeWriteTools(server);

  return server;
}

listenMcpHttpServer({
  port,
  label: "Space Rocks write",
  createMcpServer,
});
