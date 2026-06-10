import { z } from "zod";

import { textResponse } from "./responses.js";
import {
  getEngineForgeBridgeInfo,
  engineForgeCommand,
  engineForgeRequest,
  probeEngineForgeRoutes,
  tryEngineForgeCommands,
} from "./engineforge_bridge.js";

function jsonText(value) {
  return textResponse(JSON.stringify(value, null, 2));
}

const READONLY_COMMANDS = {
  project_info: [
    { category: "project", action: "getInfo" },
    { category: "project", action: "scan" },
  ],

  current_scene: [
    { category: "scene", action: "getActive" },
    { category: "editor", action: "getState" },
  ],

  scene_tree: [
    { category: "scene", action: "getTree" },
  ],

  node_properties: [
    { category: "node", action: "getProperties" },
  ],

  editor_errors: [
    { category: "console", action: "getLogs" },
    { category: "editor", action: "getState" },
  ],
};

async function runReadonlyCommand(kind, params = {}) {
  const candidates = READONLY_COMMANDS[kind];

  if (!candidates) {
    throw new Error(`Unknown read-only EngineForge command kind: ${kind}`);
  }

  return tryEngineForgeCommands(candidates, params);
}

export function registerEngineForgeReadonlyTools(server) {
  server.registerTool(
    "engineforge_bridge_info",
    {
      title: "EngineForge bridge info",
      description: "Show local EngineForge/Godot bridge discovery info.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await getEngineForgeBridgeInfo());
    }
  );

  server.registerTool(
    "engineforge_status",
    {
      title: "EngineForge status",
      description: "Read /status from the local EngineForge/Godot bridge.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await engineForgeRequest("/status"));
    }
  );

  server.registerTool(
    "engineforge_probe_routes",
    {
      title: "Probe EngineForge bridge routes",
      description: "Probe common read-only discovery routes on the local bridge.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await probeEngineForgeRoutes());
    }
  );

  server.registerTool(
    "engineforge_probe_command",
    {
      title: "Probe EngineForge command",
      description: "Probe a single EngineForge category/action command shape for diagnostics.",
      inputSchema: {
        category: z.string(),
        action: z.string(),
        params: z.record(z.unknown()).optional(),
      },
    },
    async ({ category, action, params = {} }) => {
      return jsonText(await engineForgeCommand(category, action, params, {
        allowFailure: true,
      }));
    }
  );

  server.registerTool(
    "godot_project_info",
    {
      title: "Godot project info",
      description: "Try read-only bridge commands for project info or project scan.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runReadonlyCommand("project_info"));
    }
  );

  server.registerTool(
    "godot_current_scene",
    {
      title: "Godot current scene",
      description: "Try read-only bridge commands for the current open scene.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runReadonlyCommand("current_scene"));
    }
  );

  server.registerTool(
    "godot_scene_tree",
    {
      title: "Godot scene tree",
      description: "Try read-only bridge commands for the active scene tree.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runReadonlyCommand("scene_tree"));
    }
  );

  server.registerTool(
    "godot_selected_node",
    {
      title: "Godot selected node",
      description: "Try read-only bridge commands for the selected Godot node.",
      inputSchema: {},
    },
    async () => {
      return jsonText({
        ok: false,
        supported: false,
        reason: "The installed EngineForge Godot bridge does not expose a selected-node read command. Use godot_scene_tree and godot_node_properties instead.",
      });
    }
  );

  server.registerTool(
    "godot_node_properties",
    {
      title: "Godot node properties",
      description: "Try read-only bridge commands for properties of a Godot node.",
      inputSchema: {
        node_path: z.string(),
      },
    },
    async ({ node_path }) => {
      return jsonText(await runReadonlyCommand("node_properties", {
        nodePath: node_path,
      }));
    }
  );

  server.registerTool(
    "godot_editor_errors",
    {
      title: "Godot editor errors",
      description: "Try read-only bridge commands for Godot editor errors or logs.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runReadonlyCommand("editor_errors"));
    }
  );
}
