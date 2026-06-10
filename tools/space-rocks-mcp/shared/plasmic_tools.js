import { z } from "zod";

import { callChromeDevtoolsTool } from "./chrome_devtools_client.js";
import { textResponse } from "./responses.js";

const DEFAULT_PLASMIC_BASE_URL = "https://studio.plasmic.app";
const PLASMIC_AI_TOOLS_READY_SCRIPT = "() => typeof window.PLASMIC_AI_TOOLS";

function jsonTextResponse(value) {
  return textResponse(JSON.stringify(value, null, 2));
}

function safeJsonForScript(value) {
  return JSON.stringify(value).replaceAll("</script", "<\\/script");
}

async function evaluateInChrome(functionSource) {
  return callChromeDevtoolsTool("evaluate_script", {
    function: functionSource,
  });
}

function requireProjectId(projectId) {
  if (typeof projectId !== "string" || projectId.trim() === "") {
    throw new Error("project_id must be a non-empty string");
  }

  return projectId.trim();
}

function normalizeBaseUrl(baseUrl) {
  const resolved = baseUrl?.trim() || DEFAULT_PLASMIC_BASE_URL;

  if (!resolved.startsWith("https://") && !resolved.startsWith("http://")) {
    throw new Error("base_url must start with http:// or https://");
  }

  return resolved.replace(/\/+$/, "");
}

export function registerPlasmicReadTools(server) {
  server.registerTool(
    "plasmic_open_project",
    {
      title: "Open Plasmic project",
      description: "Opens a Plasmic Studio project in the Chrome DevTools MCP browser.",
      inputSchema: {
        project_id: z.string(),
        base_url: z.string().optional(),
      },
    },
    async ({ project_id, base_url }) => {
      const projectId = requireProjectId(project_id);
      const resolvedBaseUrl = normalizeBaseUrl(base_url);
      const url = `${resolvedBaseUrl}/projects/${encodeURIComponent(projectId)}/`;
      const result = await callChromeDevtoolsTool("navigate_page", { url });
      return jsonTextResponse({ url, result });
    }
  );

  server.registerTool(
    "plasmic_check_ai_tools",
    {
      title: "Check Plasmic AI tools",
      description: "Checks whether window.PLASMIC_AI_TOOLS is available in the active Chrome page.",
      inputSchema: {},
    },
    async () => {
      const result = await evaluateInChrome(PLASMIC_AI_TOOLS_READY_SCRIPT);
      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "plasmic_identify",
    {
      title: "Identify Plasmic bridge",
      description: "Identifies this MCP bridge to Plasmic Studio.",
      inputSchema: {
        model: z.string().optional(),
        client: z.string().optional(),
        skill: z.string().optional(),
      },
    },
    async ({
      model = "gpt-5.5-thinking",
      client = "chatgpt-space-rocks-mcp",
      skill = "space-rocks-plasmic-bridge@0.1.0",
    }) => {
      const payload = { model, client, skill };
      const payloadJson = safeJsonForScript(payload);

      const result = await evaluateInChrome(`() => {
        if (!window.PLASMIC_AI_TOOLS) {
          throw new Error("window.PLASMIC_AI_TOOLS is not available");
        }

        return window.PLASMIC_AI_TOOLS.identify(${payloadJson});
      }`);

      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "plasmic_read",
    {
      title: "Read Plasmic canvas",
      description: "Reads Plasmic Studio state through window.PLASMIC_AI_TOOLS.read().",
      inputSchema: {
        query: z.string().optional(),
        include_screenshot: z.boolean().optional(),
      },
    },
    async ({ query, include_screenshot }) => {
      const options = {};

      if (query !== undefined) {
        options.query = query;
      }

      if (include_screenshot !== undefined) {
        options.includeScreenshot = include_screenshot;
      }

      const optionsJson = safeJsonForScript(options);

      const result = await evaluateInChrome(`() => {
        if (!window.PLASMIC_AI_TOOLS) {
          throw new Error("window.PLASMIC_AI_TOOLS is not available");
        }

        return window.PLASMIC_AI_TOOLS.read(${optionsJson});
      }`);

      return jsonTextResponse(result);
    }
  );
}
