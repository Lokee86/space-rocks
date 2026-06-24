import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";

import { listenMcpHttpServer } from "./shared/http_mcp_server.js";
import { registerRepoReadonlyTools } from "./shared/repo_readonly_tools.js";
import { registerEngineForgeReadonlyTools } from "./shared/engineforge_readonly_tools.js";
import { registerChromeDevtoolsProxyTools } from "./shared/chrome_devtools_proxy_tools.js";
import { registerPlasmicReadTools, registerPlasmicWriteTools } from "./shared/plasmic_tools.js";

const port = Number(process.env.PORT ?? 8789);
const chromeDevtoolsEnabled = process.env.ENABLE_CHROME_DEVTOOLS === "1";

function createMcpServer() {
  const server = new McpServer({
    name: "space-rocks-info-mcp",
    version: "0.3.0",
  });

  registerRepoReadonlyTools(server);
  registerEngineForgeReadonlyTools(server);
  if (chromeDevtoolsEnabled) {
    registerChromeDevtoolsProxyTools(server);
    registerPlasmicReadTools(server);
    registerPlasmicWriteTools(server);
  }

  return server;
}

if (chromeDevtoolsEnabled) {
  console.log("Chrome DevTools MCP proxy enabled");
}

listenMcpHttpServer({
  port,
  label: "Space Rocks info",
  createMcpServer,
});
