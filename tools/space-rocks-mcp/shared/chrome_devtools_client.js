import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

const CLIENT_INFO = {
  name: "space-rocks-chrome-devtools-proxy",
  version: "0.1.0",
};

let clientPromise = null;
let activeClient = null;

const packageRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  ".."
);

function chromeDevtoolsBinPath() {
  const packageJsonPath = path.join(
    packageRoot,
    "node_modules",
    "chrome-devtools-mcp",
    "package.json"
  );

  const packageJson = JSON.parse(readFileSync(packageJsonPath, "utf8"));
  const bin =
    typeof packageJson.bin === "string"
      ? packageJson.bin
      : packageJson.bin?.["chrome-devtools-mcp"];

  if (!bin) {
    throw new Error("chrome-devtools-mcp package does not expose a bin entry");
  }

  return path.resolve(path.dirname(packageJsonPath), bin);
}

function chromeDevtoolsCommand() {
  return process.execPath;
}

function chromeDevtoolsArgs() {
  return [chromeDevtoolsBinPath(), "--no-usage-statistics"];
}

function chromeDevtoolsEnv() {
  const env = { ...process.env };

  if (process.platform === "win32") {
    env.SystemRoot ??= "C:\\Windows";
    env.PROGRAMFILES ??= "C:\\Program Files";
    env["PROGRAMFILES(X86)"] ??= "C:\\Program Files (x86)";
  }

  return env;
}

async function createChromeDevtoolsClient() {
  const client = new Client(CLIENT_INFO);

  const transport = new StdioClientTransport({
    command: chromeDevtoolsCommand(),
    args: chromeDevtoolsArgs(),
    env: chromeDevtoolsEnv(),
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
