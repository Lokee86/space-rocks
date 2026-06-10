import { z } from "zod";

import {
  callChromeDevtoolsTool,
  listChromeDevtoolsTools,
  restartChromeDevtoolsClient,
} from "./chrome_devtools_client.js";
import { textResponse } from "./responses.js";

function jsonTextResponse(value) {
  return textResponse(JSON.stringify(value, null, 2));
}

function parseArgumentsJson(argumentsJson) {
  if (!argumentsJson || argumentsJson.trim() === "") {
    return {};
  }

  try {
    const parsed = JSON.parse(argumentsJson);

    if (parsed === null || Array.isArray(parsed) || typeof parsed !== "object") {
      throw new Error("arguments_json must parse to a JSON object");
    }

    return parsed;
  } catch (error) {
    throw new Error(`Invalid arguments_json: ${error?.message ?? String(error)}`);
  }
}

export function registerChromeDevtoolsProxyTools(server) {
  server.registerTool(
    "chrome_devtools_ping",
    {
      title: "Chrome DevTools proxy ping",
      description: "Confirms Chrome DevTools proxy tools are registered.",
      inputSchema: {},
    },
    async () => {
      return textResponse("Chrome DevTools proxy tools are registered.");
    }
  );

  server.registerTool(
    "chrome_devtools_list_tools",
    {
      title: "List Chrome DevTools MCP tools",
      description: "Lists tools exposed by the upstream Chrome DevTools MCP server.",
      inputSchema: {},
    },
    async () => {
      const result = await listChromeDevtoolsTools();
      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "chrome_devtools_call_tool",
    {
      title: "Call Chrome DevTools MCP tool",
      description:
        "Generic proxy for calling an upstream Chrome DevTools MCP tool by name.",
      inputSchema: {
        tool_name: z.string(),
        arguments_json: z.string().optional(),
      },
    },
    async ({ tool_name, arguments_json = "{}" }) => {
      const args = parseArgumentsJson(arguments_json);
      const result = await callChromeDevtoolsTool(tool_name, args);
      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "chrome_devtools_restart",
    {
      title: "Restart Chrome DevTools MCP client",
      description: "Closes and resets the singleton Chrome DevTools MCP client.",
      inputSchema: {},
    },
    async () => {
      const result = await restartChromeDevtoolsClient();
      return textResponse(result);
    }
  );

  server.registerTool(
    "chrome_list_pages",
    {
      title: "List Chrome pages",
      description: "Lists the pages currently available in the active Chrome instance.",
      inputSchema: {},
    },
    async () => {
      const result = await callChromeDevtoolsTool("list_pages", {});
      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "chrome_navigate_page",
    {
      title: "Navigate Chrome page",
      description: "Navigates the active Chrome page to the provided URL.",
      inputSchema: {
        url: z.string(),
      },
    },
    async ({ url }) => {
      if (typeof url !== "string" || !(url.startsWith("http://") || url.startsWith("https://"))) {
        throw new Error("url must start with http:// or https://");
      }

      const result = await callChromeDevtoolsTool("navigate_page", { url });
      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "chrome_take_screenshot",
    {
      title: "Take Chrome screenshot",
      description: "Captures a screenshot from the active Chrome page.",
      inputSchema: {},
    },
    async () => {
      const result = await callChromeDevtoolsTool("take_screenshot", {});
      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "chrome_evaluate_script",
    {
      title: "Evaluate Chrome script",
      description: "Executes JavaScript in the active Chrome page.",
      inputSchema: {
        function: z.string(),
      },
    },
    async ({ function: functionCode }) => {
      const result = await callChromeDevtoolsTool("evaluate_script", {
        function: functionCode,
      });
      return jsonTextResponse(result);
    }
  );
}
