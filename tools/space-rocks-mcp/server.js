import { createServer } from "node:http";
import { promises as fs } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import { z } from "zod";

const port = Number(process.env.PORT ?? 8787);
const MCP_PATH = "/mcp";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// server.js lives at: space-rocks/tools/space-rocks-mcp/server.js
// Repo root is two levels up.
const REPO_ROOT = path.resolve(
  process.env.SPACE_ROCKS_REPO ?? path.join(__dirname, "../..")
);

const IGNORED_DIRS = new Set([
  ".git",
  "node_modules",
  ".godot",
  ".import",
  "dist",
  "build",
  "target",
  "tmp",
]);

const TEXT_EXTENSIONS = new Set([
  ".go",
  ".gd",
  ".ts",
  ".js",
  ".mjs",
  ".py",
  ".pyi",
  ".json",
  ".toml",
  ".yaml",
  ".yml",
  ".md",
  ".txt",
  ".tscn",
  ".tres",
  ".cfg",
  ".gitignore",
  ".mod",
  ".sum",
]);

function textResponse(text) {
  return {
    content: [
      {
        type: "text",
        text,
      },
    ],
  };
}

function repoPath(relativePath = ".") {
  const resolved = path.resolve(REPO_ROOT, relativePath);

  if (resolved !== REPO_ROOT && !resolved.startsWith(REPO_ROOT + path.sep)) {
    throw new Error("Path escapes Space Rocks repo root");
  }

  return resolved;
}

function repoRelative(absolutePath) {
  return path.relative(REPO_ROOT, absolutePath).replaceAll("\\", "/");
}

function isProbablyTextFile(filePath) {
  const base = path.basename(filePath);
  const ext = path.extname(filePath);
  return TEXT_EXTENSIONS.has(ext) || TEXT_EXTENSIONS.has(base);
}

async function walkDirectory(root, maxFiles) {
  const results = [];

  async function walk(current) {
    if (results.length >= maxFiles) return;

    const entries = await fs.readdir(current, { withFileTypes: true });

    for (const entry of entries) {
      if (results.length >= maxFiles) return;

      if (entry.name.startsWith(".") && entry.name !== ".gitignore" && entry.name !== ".github") {
        continue;
      }

      const fullPath = path.join(current, entry.name);
      const relPath = repoRelative(fullPath);

      if (entry.isDirectory()) {
        if (IGNORED_DIRS.has(entry.name)) continue;

        results.push(`${relPath}/`);
        await walk(fullPath);
        continue;
      }

      if (entry.isFile()) {
        results.push(relPath);
      }
    }
  }

  await walk(root);
  return results;
}

async function searchText(root, query, maxFiles, maxMatches) {
  const matches = [];
  let scanned = 0;
  const lowerQuery = query.toLowerCase();

  async function walk(current) {
    if (scanned >= maxFiles || matches.length >= maxMatches) return;

    const entries = await fs.readdir(current, { withFileTypes: true });

    for (const entry of entries) {
      if (scanned >= maxFiles || matches.length >= maxMatches) return;

      const fullPath = path.join(current, entry.name);

      if (entry.isDirectory()) {
        if (IGNORED_DIRS.has(entry.name)) continue;

        await walk(fullPath);
        continue;
      }

      if (!entry.isFile() || !isProbablyTextFile(fullPath)) continue;

      scanned++;

      let text;
      try {
        text = await fs.readFile(fullPath, "utf8");
      } catch {
        continue;
      }

      const lines = text.split(/\r?\n/);

      for (let i = 0; i < lines.length; i++) {
        if (lines[i].toLowerCase().includes(lowerQuery)) {
          matches.push({
            file: repoRelative(fullPath),
            line: i + 1,
            text: lines[i].slice(0, 300),
          });

          if (matches.length >= maxMatches) break;
        }
      }
    }
  }

  await walk(root);

  return {
    scanned_files: scanned,
    matches,
  };
}

function createMcpServer() {
  const server = new McpServer({
    name: "space-rocks-mcp",
    version: "0.2.0",
  });

  server.registerTool(
    "ping",
    {
      title: "Ping",
      description: "Simple connection test for the Space Rocks MCP server.",
      inputSchema: {
        message: z.string().optional(),
      },
    },
    async ({ message }) => {
      return textResponse(`MCP server is reachable. Message: ${message ?? "none"}`);
    }
  );

  server.registerTool(
    "repo_root",
    {
      title: "Show repo root",
      description: "Returns the configured Space Rocks repo root.",
      inputSchema: {},
    },
    async () => {
      return textResponse(REPO_ROOT);
    }
  );

  server.registerTool(
    "list_repo_tree",
    {
      title: "List repo tree",
      description: "List files and directories under a repo-relative path.",
      inputSchema: {
        path: z.string().optional(),
        max_files: z.number().int().min(1).max(2000).optional(),
      },
    },
    async ({ path: requestedPath = ".", max_files = 500 }) => {
      const root = repoPath(requestedPath);
      const entries = await walkDirectory(root, max_files);
      return textResponse(entries.join("\n"));
    }
  );

  server.registerTool(
    "read_repo_file",
    {
      title: "Read repo file",
      description: "Read a text file from the Space Rocks repo by repo-relative path.",
      inputSchema: {
        path: z.string(),
        max_chars: z.number().int().min(1).max(50000).optional(),
      },
    },
    async ({ path: requestedPath, max_chars = 20000 }) => {
      const filePath = repoPath(requestedPath);

      if (!isProbablyTextFile(filePath)) {
        throw new Error("Refusing to read non-text or unsupported file type");
      }

      const text = await fs.readFile(filePath, "utf8");

      if (text.length > max_chars) {
        return textResponse(`${text.slice(0, max_chars)}\n\n[TRUNCATED at ${max_chars} chars]`);
      }

      return textResponse(text);
    }
  );

  server.registerTool(
    "search_repo_text",
    {
      title: "Search repo text",
      description: "Search text files in the Space Rocks repo for a string.",
      inputSchema: {
        query: z.string(),
        path: z.string().optional(),
        max_files: z.number().int().min(1).max(1000).optional(),
        max_matches: z.number().int().min(1).max(200).optional(),
      },
    },
    async ({
      query,
      path: requestedPath = ".",
      max_files = 300,
      max_matches = 50,
    }) => {
      const root = repoPath(requestedPath);
      const result = await searchText(root, query, max_files, max_matches);
      return textResponse(JSON.stringify(result, null, 2));
    }
  );

  return server;
}

const httpServer = createServer(async (req, res) => {
  if (!req.url) {
    res.writeHead(400).end("Missing URL");
    return;
  }

  const url = new URL(req.url, `http://${req.headers.host ?? "localhost"}`);

  if (req.method === "GET" && url.pathname === "/") {
    res.writeHead(200, { "content-type": "text/plain" });
    res.end("Space Rocks MCP server is running");
    return;
  }

  if (req.method === "OPTIONS" && url.pathname === MCP_PATH) {
    res.writeHead(204, {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "POST, GET, DELETE, OPTIONS",
      "Access-Control-Allow-Headers": "content-type, mcp-session-id, mcp-protocol-version",
      "Access-Control-Expose-Headers": "Mcp-Session-Id",
    });
    res.end();
    return;
  }

  const allowedMethods = new Set(["POST", "GET", "DELETE"]);

  if (url.pathname === MCP_PATH && req.method && allowedMethods.has(req.method)) {
    res.setHeader("Access-Control-Allow-Origin", "*");
    res.setHeader("Access-Control-Expose-Headers", "Mcp-Session-Id");

    const mcpServer = createMcpServer();
    const transport = new StreamableHTTPServerTransport({
      sessionIdGenerator: undefined,
      enableJsonResponse: true,
    });

    res.on("close", () => {
      transport.close();
      mcpServer.close();
    });

    try {
      await mcpServer.connect(transport);
      await transport.handleRequest(req, res);
    } catch (error) {
      console.error("Error handling MCP request:", error);

      if (!res.headersSent) {
        res.writeHead(500).end("Internal server error");
      }
    }

    return;
  }

  res.writeHead(404).end("Not Found");
});

httpServer.listen(port, "127.0.0.1", () => {
  console.log(`Space Rocks MCP server listening on http://127.0.0.1:${port}${MCP_PATH}`);
  console.log(`Repo root: ${REPO_ROOT}`);
});
