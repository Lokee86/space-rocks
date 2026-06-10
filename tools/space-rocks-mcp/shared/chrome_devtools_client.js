import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

const CLIENT_INFO = {
  name: "space-rocks-chrome-devtools-proxy",
  version: "0.1.0",
};

let clientPromise = null;
let activeClient = null;

function npxCommand() {
  return process.platform === "win32" ? "npx.cmd" : "npx";
}

async function createChromeDevtoolsClient() {
  const client = new Client(CLIENT_INFO);

  const transport = new StdioClientTransport({
    command: npxCommand(),
    args: ["chrome-devtools-mcp@latest", "--no-usage-statistics"],
  });

  await client.connect(transport);
  activeClient = client;

  return client;
}

export async function getChromeDevtoolsClient() {
  if (!clientPromise) {
    clientPromise = createChromeDevtoolsClient().catch((error) => {
      clientPromise = null;
      activeClient = null;
      throw error;
    });
  }

  return clientPromise;
}

export async function restartChromeDevtoolsClient() {
  let closeMessage = "";

  if (activeClient) {
    try {
      await activeClient.close();
    } catch (error) {
      closeMessage = ` Previous close failed: ${error?.message ?? String(error)}`;
    }
  }

  clientPromise = null;
  activeClient = null;

  return `Chrome DevTools MCP client restarted.${closeMessage}`;
}

export async function listChromeDevtoolsTools() {
  const client = await getChromeDevtoolsClient();
  return client.listTools();
}

export async function callChromeDevtoolsTool(toolName, args = {}) {
  if (typeof toolName !== "string" || toolName.trim() === "") {
    throw new Error("toolName must be a non-empty string");
  }

  const client = await getChromeDevtoolsClient();

  return client.callTool({
    name: toolName,
    arguments: args ?? {},
  });
}
