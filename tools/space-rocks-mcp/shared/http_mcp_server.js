import { createServer } from "node:http";

import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";

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

    if (req.method === "OPTIONS" && url.pathname === mcpPath) {
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

    if (url.pathname === mcpPath && req.method && allowedMethods.has(req.method)) {
      res.setHeader("Access-Control-Allow-Origin", "*");
      res.setHeader("Access-Control-Expose-Headers", "Mcp-Session-Id");

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
