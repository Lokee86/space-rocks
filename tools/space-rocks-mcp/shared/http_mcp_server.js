import { createServer } from "node:http";

import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";

import {
  PROTECTED_RESOURCE_METADATA,
  PROTECTED_RESOURCE_METADATA_URL,
  validateBearerAccessToken,
} from "./oauth_auth.js";

function sendJson(res, statusCode, body, headers = {}) {
  res.writeHead(statusCode, {
    "content-type": "application/json",
    ...headers,
  });
  res.end(JSON.stringify(body));
}

function sendUnauthorized(res) {
  const challenge = `Bearer resource_metadata="${PROTECTED_RESOURCE_METADATA_URL}"`;
  sendJson(
    res,
    401,
    { error: "Unauthorized" },
    {
      "WWW-Authenticate": challenge,
      "Access-Control-Expose-Headers": "Mcp-Session-Id, WWW-Authenticate",
    }
  );
}

function getBearerToken(req) {
  const header = req.headers.authorization;
  if (typeof header !== "string") {
    return null;
  }

  const match = header.match(/^Bearer\s+(.+)$/i);
  if (!match) {
    return null;
  }

  return match[1].trim() || null;
}

export function listenMcpHttpServer({
  port,
  mcpPath = "/mcp",
  label,
  createMcpServer,
}) {
  const serverLabel = label ?? "Space Rocks";

  const httpServer = createServer(async (req, res) => {
    if (!req.url) {
      res.writeHead(400).end("Missing URL");
      return;
    }

    const url = new URL(req.url, `http://${req.headers.host ?? "localhost"}`);

    if (req.method === "GET" && url.pathname === "/") {
      res.writeHead(200, { "content-type": "text/plain" });
      res.end(`${serverLabel} MCP server is running`);
      return;
    }

    if (req.method === "GET" && url.pathname === "/.well-known/oauth-protected-resource") {
      sendJson(res, 200, {
        resource: PROTECTED_RESOURCE_METADATA.resource,
        authorization_servers: PROTECTED_RESOURCE_METADATA.authorization_servers,
      });
      return;
    }

    if (req.method === "OPTIONS" && url.pathname === mcpPath) {
      res.writeHead(204, {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "POST, GET, DELETE, OPTIONS",
        "Access-Control-Allow-Headers": "content-type, mcp-session-id, mcp-protocol-version, authorization",
        "Access-Control-Expose-Headers": "Mcp-Session-Id, WWW-Authenticate",
      });
      res.end();
      return;
    }

    const allowedMethods = new Set(["POST", "GET", "DELETE"]);

    if (url.pathname === mcpPath && req.method && allowedMethods.has(req.method)) {
      res.setHeader("Access-Control-Allow-Origin", "*");
      res.setHeader("Access-Control-Expose-Headers", "Mcp-Session-Id, WWW-Authenticate");

      const token = getBearerToken(req);
      if (!token) {
        sendUnauthorized(res);
        return;
      }

      try {
        await validateBearerAccessToken(token);
      } catch (error) {
        console.error(`Invalid bearer token for ${serverLabel} MCP request:`, error?.message ?? error);
        sendUnauthorized(res);
        return;
      }

      const mcpServer = await Promise.resolve(createMcpServer());
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
        console.error(`Error handling ${serverLabel} MCP request:`, error);

        if (!res.headersSent) {
          res.writeHead(500).end("Internal server error");
        }
      }

      return;
    }

    res.writeHead(404).end("Not Found");
  });

  httpServer.listen(port, "127.0.0.1", () => {
    console.log(
      `${serverLabel} MCP server listening on http://127.0.0.1:${port}${mcpPath}`
    );
  });

  return httpServer;
}
