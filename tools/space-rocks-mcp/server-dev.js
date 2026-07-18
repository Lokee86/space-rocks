import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";

import { listenMcpHttpServer } from "./shared/http_mcp_server.js";
import { registerEngineForgeReadonlyTools } from "./shared/engineforge_readonly_tools.js";
import { registerEngineForgeWriteTools } from "./shared/engineforge_write_tools.js";
import { registerChromeDevtoolsProxyTools } from "./shared/chrome_devtools_proxy_tools.js";
import { registerHermesTools } from "./shared/hermes_tools.js";
import { defaultProcessJobManager } from "./shared/job_manager.js";
import { registerJobTools } from "./shared/job_tools.js";
import { registerPlasmicReadTools, registerPlasmicWriteTools } from "./shared/plasmic_tools.js";
import { registerRepoReadonlyTools } from "./shared/repo_readonly_tools.js";
import { registerRepoWriteTools } from "./shared/repo_write_tools.js";
import { registerRestrictedCommandTools } from "./shared/restricted_command_tools.js";

const port = Number(process.env.PORT ?? 8889);
const chromeDevtoolsEnabled = process.env.ENABLE_CHROME_DEVTOOLS === "1";

function createMcpServer() {
  const server = new McpServer({
    name: "space-rocks-dev-mcp",
    version: "0.1.0",
  });

  registerRepoReadonlyTools(server);
  registerRepoWriteTools({
    registerTool(name, ...args) {
      if (name === "ping") {
        return;
      }

      return server.registerTool(name, ...args);
    },
  });
  registerRestrictedCommandTools(server);
  registerJobTools(server, { manager: defaultProcessJobManager });
  registerHermesTools(server, { processJobManager: defaultProcessJobManager });
  registerEngineForgeReadonlyTools(server);
  registerEngineForgeWriteTools(server);

  if (chromeDevtoolsEnabled) {
    registerChromeDevtoolsProxyTools(server);
    registerPlasmicReadTools(server);
    registerPlasmicWriteTools(server);
  }

  return server;
}

if (chromeDevtoolsEnabled) {
  console.log("Chrome DevTools / Plasmic development bridge enabled");
}

listenMcpHttpServer({
  port,
  label: "Space Rocks development",
  createMcpServer,
});
