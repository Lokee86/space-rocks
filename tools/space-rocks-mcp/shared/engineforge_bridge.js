import { promises as fs } from "node:fs";
import path from "node:path";

import { REPO_ROOT } from "./paths.js";

const LOCAL_HOSTS = new Set(["127.0.0.1", "localhost", "::1"]);

export const GODOT_PROJECT_ROOT = path.resolve(
  process.env.SPACE_ROCKS_GODOT_PROJECT ?? path.join(REPO_ROOT, "client")
);

export const ENGINEFORGE_BRIDGE_FILE = path.resolve(
  process.env.ENGINEFORGE_BRIDGE_FILE ??
    path.join(GODOT_PROJECT_ROOT, ".godot/engineforge/bridge.json")
);

function normalizeBridgeUrl(rawUrl) {
  const url = new URL(rawUrl);

  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new Error("EngineForge bridge URL must use http or https");
  }

  if (!LOCAL_HOSTS.has(url.hostname)) {
    throw new Error("Refusing to call non-local EngineForge bridge URL");
  }

  return url.toString().replace(/\/$/, "");
}

function bridgeUrlFromData(data) {
  if (typeof data.url === "string") {
    return data.url;
  }

  if (typeof data.base_url === "string") {
    return data.base_url;
  }

  if (typeof data.baseUrl === "string") {
    return data.baseUrl;
  }

  const host = data.host ?? "127.0.0.1";
  const port = Number(data.port);

  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    throw new Error("EngineForge bridge data does not contain a valid URL or port");
  }

  return `http://${host}:${port}`;
}

export async function getEngineForgeBridgeInfo() {
  if (process.env.ENGINEFORGE_BRIDGE_URL) {
    return {
      baseUrl: normalizeBridgeUrl(process.env.ENGINEFORGE_BRIDGE_URL),
      bridgeFile: ENGINEFORGE_BRIDGE_FILE,
      source: "ENGINEFORGE_BRIDGE_URL",
      raw: null,
    };
  }

  const raw = await fs.readFile(ENGINEFORGE_BRIDGE_FILE, "utf8");
  const data = JSON.parse(raw);

  return {
    baseUrl: normalizeBridgeUrl(bridgeUrlFromData(data)),
    bridgeFile: ENGINEFORGE_BRIDGE_FILE,
    source: "bridge.json",
    raw: data,
  };
}

export async function engineForgeRequest(route, options = {}) {
  const bridge = await getEngineForgeBridgeInfo();

  const response = await fetch(`${bridge.baseUrl}${route}`, {
    method: options.method ?? "GET",
    headers: {
      "content-type": "application/json",
      ...(options.headers ?? {}),
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  const text = await response.text();

  const result = {
    ok: response.ok,
    status: response.status,
    route,
    text,
  };

  try {
    result.json = JSON.parse(text);
  } catch {
    result.json = null;
  }

  if (!response.ok && !options.allowFailure) {
    throw new Error(`EngineForge bridge request failed: ${response.status} ${text}`);
  }

  return result;
}

export async function engineForgeCommand(category, action, params = {}, options = {}) {
  return engineForgeRequest("/command", {
    method: "POST",
    body: {
      category,
      action,
      params,
    },
    allowFailure: options.allowFailure,
  });
}

function isSuccessfulBridgeCommandResult(result) {
  if (!result.ok) {
    return false;
  }

  if (
    result.json &&
    typeof result.json === "object" &&
    result.json.success === false
  ) {
    return false;
  }

  return true;
}

export async function tryEngineForgeCommands(commandCandidates, params = {}) {
  const attempts = [];

  for (const candidate of commandCandidates) {
    const commandParams = {
      ...params,
      ...(candidate.params ?? {}),
    };

    const result = await engineForgeCommand(
      candidate.category,
      candidate.action,
      commandParams,
      {
      allowFailure: true,
      }
    );

    attempts.push({
      category: candidate.category,
      action: candidate.action,
      params: commandParams,
      ok: result.ok,
      status: result.status,
      json: result.json,
      text: result.text,
    });

    if (isSuccessfulBridgeCommandResult(result)) {
      return {
        ok: true,
        category: candidate.category,
        action: candidate.action,
        params: commandParams,
        result: result.json ?? result.text,
        attempts,
      };
    }
  }

  return {
    ok: false,
    attempts,
  };
}

export async function probeEngineForgeRoutes() {
  const routes = [
    "/",
    "/status",
    "/capabilities",
    "/commands",
    "/schema",
    "/openapi.json",
    "/docs",
  ];

  const results = [];

  for (const route of routes) {
    const result = await engineForgeRequest(route, {
      method: "GET",
      allowFailure: true,
    });

    results.push({
      route,
      ok: result.ok,
      status: result.status,
      json: result.json,
      text: result.text.slice(0, 2000),
    });
  }

  return results;
}
