import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";

import { listenMcpHttpServer } from "./shared/http_mcp_server.js";
import { registerChromeDevtoolsProxyTools } from "./shared/chrome_devtools_proxy_tools.js";
import { registerEngineForgeWriteTools } from "./shared/engineforge_write_tools.js";
import { registerPlasmicReadTools, registerPlasmicWriteTools } from "./shared/plasmic_tools.js";
import { registerRepoWriteTools } from "./shared/repo_write_tools.js";

const port = Number(process.env.PORT ?? 8788);
const chromeDevtoolsEnabled = process.env.ENABLE_CHROME_DEVTOOLS === "1";

function createMcpServer() {
  const server = new McpServer({
    name: "space-rocks-write-mcp",
    version: "0.1.0",
  });

  registerRepoWriteTools(server);
  registerEngineForgeWriteTools(server);

  if (chromeDevtoolsEnabled) {
    registerChromeDevtoolsProxyTools(server);
    registerPlasmicReadTools(server);
    registerPlasmicWriteTools(server);
  }

  return server;
}

if (chromeDevtoolsEnabled) {
  console.log("Chrome DevTools / Plasmic write bridge enabled");
}

listenMcpHttpServer({
  port,
  label: "Space Rocks write",
  createMcpServer,
});
