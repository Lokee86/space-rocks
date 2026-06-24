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

async function callPlasmicAiTool(toolName, args = {}) {
  const toolNameJson = safeJsonForScript(toolName);
  const argsJson = safeJsonForScript(args);

  return evaluateInChrome(`async () => {
    if (!window.PLASMIC_AI_TOOLS) {
      throw new Error("window.PLASMIC_AI_TOOLS is not available");
    }

    if (typeof window.PLASMIC_AI_TOOLS[${toolNameJson}] !== "function") {
      throw new Error("window.PLASMIC_AI_TOOLS does not expose " + ${toolNameJson});
    }

    return await window.PLASMIC_AI_TOOLS[${toolNameJson}](${argsJson});
  }`);
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
      const result = await callPlasmicAiTool("identify", payload);

      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "plasmic_read",
    {
      title: "Read Plasmic canvas",
      description: "Reads Plasmic Studio state through window.PLASMIC_AI_TOOLS.read().",
      inputSchema: {
        project: z
          .object({
            components: z.boolean().optional(),
            screenBreakpoints: z.boolean().optional(),
            globalVariants: z.boolean().optional(),
            tokens: z.boolean().optional(),
            animations: z.boolean().optional(),
          })
          .optional(),
        components: z.array(z.string()).optional(),
        elements: z
          .array(
            z.object({
              componentUuid: z.string(),
              elementUuid: z.string(),
            })
          )
          .optional(),
        tokens: z.array(z.string()).optional(),
        animations: z.array(z.string()).optional(),
      },
    },
    async ({ project, components, elements, tokens, animations }) => {
      const options = {};

      if (project !== undefined) {
        options.project = project;
      }

      if (components !== undefined) {
        options.components = components;
      }

      if (elements !== undefined) {
        options.elements = elements;
      }

      if (tokens !== undefined) {
        options.tokens = tokens;
      }

      if (animations !== undefined) {
        options.animations = animations;
      }

      const result = await callPlasmicAiTool("read", options);
      return jsonTextResponse(result);
    }
  );
}

export function registerPlasmicWriteTools(server) {
  server.registerTool(
    "plasmic_insert_html",
    {
      title: "Insert Plasmic HTML",
      description: "Inserts HTML into a Plasmic element through window.PLASMIC_AI_TOOLS.insertHtml().",
      inputSchema: {
        html: z.string(),
        componentUuid: z.string(),
        tplUuid: z.string(),
        variantUuids: z.array(z.string()).optional(),
        insertRelLoc: z
          .enum(["before", "prepend", "append", "after", "wrap", "replace"])
          .optional(),
      },
    },
    async ({ html, componentUuid, tplUuid, variantUuids, insertRelLoc }) => {
      const args = {
        html,
        componentUuid,
        tplUuid,
      };

      if (variantUuids !== undefined) {
        args.variantUuids = variantUuids;
      }

      if (insertRelLoc !== undefined) {
        args.insertRelLoc = insertRelLoc;
      }

      const result = await callPlasmicAiTool("insertHtml", args);
      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "plasmic_change_element",
    {
      title: "Change Plasmic element",
      description: "Updates a Plasmic element through window.PLASMIC_AI_TOOLS.changeElement().",
      inputSchema: {
        componentUuid: z.string(),
        variantUuids: z.array(z.string()).optional(),
        changes: z.array(
          z.object({
            tplUuid: z.string(),
            name: z.string().nullable().optional(),
            styles: z.record(z.string().nullable()).optional(),
            props: z.record(z.any()).optional(),
            unsetProps: z.array(z.string()).optional(),
          })
        ),
      },
    },
    async ({ componentUuid, variantUuids, changes }) => {
      const args = {
        componentUuid,
        changes,
      };

      if (variantUuids !== undefined) {
        args.variantUuids = variantUuids;
      }

      const result = await callPlasmicAiTool("changeElement", args);
      return jsonTextResponse(result);
    }
  );

  server.registerTool(
    "plasmic_delete_element",
    {
      title: "Delete Plasmic element",
      description: "Deletes a Plasmic element through window.PLASMIC_AI_TOOLS.deleteElement().",
      inputSchema: {
        componentUuid: z.string(),
        tplUuid: z.string(),
      },
    },
    async ({ componentUuid, tplUuid }) => {
      const result = await callPlasmicAiTool("deleteElement", {
        componentUuid,
        tplUuid,
      });

      return jsonTextResponse(result);
    }
  );
}
