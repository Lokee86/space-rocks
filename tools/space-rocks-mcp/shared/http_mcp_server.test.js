import assert from "node:assert/strict";
import { once } from "node:events";
import { after, before, test } from "node:test";

const originalEnv = {
  AUTH0_ISSUER: process.env.AUTH0_ISSUER,
  AUTH0_AUDIENCE: process.env.AUTH0_AUDIENCE,
  RESOURCE_SERVER_URL: process.env.RESOURCE_SERVER_URL,
};

process.env.AUTH0_ISSUER = "https://tenant.example.com/";
process.env.AUTH0_AUDIENCE = "api://space-rocks";
process.env.RESOURCE_SERVER_URL = "https://api.example.com/";

const { listenMcpHttpServer } = await import("./http_mcp_server.js");

let server;
let baseUrl;
let createMcpServerCalls;

before(async () => {
  createMcpServerCalls = 0;
  server = listenMcpHttpServer({
    port: 0,
    createMcpServer: () => {
      createMcpServerCalls += 1;
      return {
        connect: async () => {},
        close: () => {},
      };
    },
  });

  await once(server, "listening");
  const address = server.address();
  assert.equal(typeof address, "object");
  assert.ok(address);
  baseUrl = `http://127.0.0.1:${address.port}`;
});

after(async () => {
  if (server) {
    await new Promise((resolve) => server.close(resolve));
  }

  for (const [key, value] of Object.entries(originalEnv)) {
    if (value === undefined) {
      delete process.env[key];
    } else {
      process.env[key] = value;
    }
  }
});

test("GET /.well-known/oauth-protected-resource returns canonical metadata", async () => {
  const response = await fetch(`${baseUrl}/.well-known/oauth-protected-resource`);
  assert.equal(response.status, 200);
  assert.equal(response.headers.get("content-type"), "application/json");

  const body = await response.json();
  assert.deepEqual(body, {
    resource: "https://api.example.com/",
    authorization_servers: ["https://tenant.example.com/"],
  });
  assert.equal("scopes" in body, false);
  assert.equal(createMcpServerCalls, 0);
});

test("POST /mcp without authorization is rejected before MCP server creation", async () => {
  const response = await fetch(`${baseUrl}/mcp`, {
    method: "POST",
    headers: {
      "content-type": "application/json",
    },
    body: JSON.stringify({}),
  });

  assert.equal(response.status, 401);
  assert.equal(
    response.headers.get("www-authenticate"),
    'Bearer resource_metadata="https://api.example.com/.well-known/oauth-protected-resource"'
  );
  assert.equal(createMcpServerCalls, 0);
});

test("POST /mcp with malformed bearer authorization is rejected before MCP server creation", async () => {
  const response = await fetch(`${baseUrl}/mcp`, {
    method: "POST",
    headers: {
      authorization: "Bearer   ",
      "content-type": "application/json",
    },
    body: JSON.stringify({}),
  });

  assert.equal(response.status, 401);
  assert.equal(
    response.headers.get("www-authenticate"),
    'Bearer resource_metadata="https://api.example.com/.well-known/oauth-protected-resource"'
  );
  assert.equal(createMcpServerCalls, 0);
});

test("OPTIONS /mcp remains public and allows authorization header", async () => {
  const response = await fetch(`${baseUrl}/mcp`, {
    method: "OPTIONS",
    headers: {
      origin: "https://example.com",
      "access-control-request-method": "POST",
      "access-control-request-headers": "authorization, content-type",
    },
  });

  assert.equal(response.status, 204);
  assert.match(response.headers.get("access-control-allow-headers") ?? "", /authorization/i);
  assert.match(response.headers.get("access-control-expose-headers") ?? "", /WWW-Authenticate/i);
  assert.equal(createMcpServerCalls, 0);
});

