import { z } from "zod";

import { engineForgeCommand } from "./engineforge_bridge.js";
import { textResponse } from "./responses.js";

function jsonText(value) {
  return textResponse(JSON.stringify(value, null, 2));
}

async function runBridgeCommand(category, action, params = {}) {
  const result = await engineForgeCommand(category, action, params, {
    allowFailure: true,
  });

  if (
    result.json &&
    typeof result.json === "object" &&
    result.json.success === false
  ) {
    throw new Error(result.json.error ?? `EngineForge command failed: ${category}.${action}`);
  }

  return result.json ?? result.text;
}

const vector2Schema = z.union([
  z.array(z.number()).min(2).max(2),
  z.object({
    _type: z.literal("Vector2").optional(),
    x: z.number(),
    y: z.number(),
  }),
]);

const vector3Schema = z.union([
  z.array(z.number()).min(3).max(3),
  z.object({
    _type: z.literal("Vector3").optional(),
    x: z.number(),
    y: z.number(),
    z: z.number(),
  }),
]);

export function registerEngineForgeWriteTools(server) {
  server.registerTool(
    "godot_scene_create",
    {
      title: "Create Godot scene",
      description: "Create and open a new Godot scene through the EngineForge bridge.",
      inputSchema: {
        name: z.string(),
        path: z.string().optional(),
        rootType: z.string().optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("scene", "create", params));
    }
  );

  server.registerTool(
    "godot_scene_open",
    {
      title: "Open Godot scene",
      description: "Open an existing Godot scene through the EngineForge bridge.",
      inputSchema: {
        path: z.string(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("scene", "open", params));
    }
  );

  server.registerTool(
    "godot_scene_save",
    {
      title: "Save Godot scene",
      description: "Save the current Godot scene through the EngineForge bridge.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runBridgeCommand("scene", "save"));
    }
  );

  server.registerTool(
    "godot_node_create",
    {
      title: "Create Godot node",
      description: "Create a node in the current Godot scene.",
      inputSchema: {
        type: z.string(),
        name: z.string().optional(),
        parentPath: z.string().optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("node", "create", params));
    }
  );

  server.registerTool(
    "godot_node_delete",
    {
      title: "Delete Godot node",
      description: "Delete a node from the current Godot scene.",
      inputSchema: {
        nodePath: z.string(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("node", "delete", params));
    }
  );

  server.registerTool(
    "godot_node_duplicate",
    {
      title: "Duplicate Godot node",
      description: "Duplicate a node in the current Godot scene.",
      inputSchema: {
        nodePath: z.string(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("node", "duplicate", params));
    }
  );

  server.registerTool(
    "godot_node_set_property",
    {
      title: "Set Godot node property",
      description: "Set a property on a Godot node.",
      inputSchema: {
        nodePath: z.string(),
        propertyName: z.string(),
        value: z.unknown(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("node", "setProperty", params));
    }
  );

  server.registerTool(
    "godot_node_reparent",
    {
      title: "Reparent Godot node",
      description: "Move a node under a different parent node.",
      inputSchema: {
        nodePath: z.string(),
        newParentPath: z.string(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("node", "reparent", params));
    }
  );

  server.registerTool(
    "godot_node_set_transform",
    {
      title: "Set Godot node transform",
      description: "Set transform fields on a Godot 2D, Control, or 3D node.",
      inputSchema: {
        nodePath: z.string(),
        position: z.union([vector2Schema, vector3Schema]).optional(),
        rotation: z.union([z.number(), vector3Schema]).optional(),
        scale: z.union([vector2Schema, vector3Schema]).optional(),
        size: vector2Schema.optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("node", "setTransform", params));
    }
  );

  server.registerTool(
    "godot_script_attach",
    {
      title: "Attach Godot script",
      description: "Attach an existing GDScript to a Godot node.",
      inputSchema: {
        nodePath: z.string(),
        scriptPath: z.string(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("script", "attach", params));
    }
  );

  server.registerTool(
    "godot_script_detach",
    {
      title: "Detach Godot script",
      description: "Detach the script from a Godot node.",
      inputSchema: {
        nodePath: z.string(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("script", "detach", params));
    }
  );

  server.registerTool(
    "godot_script_create",
    {
      title: "Create Godot script",
      description: "Create a new GDScript through the EngineForge bridge.",
      inputSchema: {
        name: z.string(),
        path: z.string().optional(),
        extends: z.string().optional(),
        className: z.string().optional(),
        includeReady: z.boolean().optional(),
        includeProcess: z.boolean().optional(),
        includePhysicsProcess: z.boolean().optional(),
        includeInput: z.boolean().optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("script", "create", params));
    }
  );

  server.registerTool(
    "godot_script_edit",
    {
      title: "Edit Godot script",
      description: "Edit an existing GDScript by full replacement or exact text replacement.",
      inputSchema: {
        path: z.string(),
        contents: z.string().optional(),
        oldText: z.string().optional(),
        newText: z.string().optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("script", "edit", params));
    }
  );

  server.registerTool(
    "godot_script_delete",
    {
      title: "Delete Godot script",
      description: "Delete a GDScript through the EngineForge bridge.",
      inputSchema: {
        path: z.string(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("script", "delete", params));
    }
  );

  server.registerTool(
    "godot_resource_create",
    {
      title: "Create Godot resource",
      description: "Create a Godot resource through the EngineForge bridge.",
      inputSchema: {
        type: z.string(),
        path: z.string(),
        properties: z.record(z.unknown()).optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("resource", "create", params));
    }
  );

  server.registerTool(
    "godot_resource_create_material",
    {
      title: "Create Godot material",
      description: "Create a Godot material resource through the EngineForge bridge.",
      inputSchema: {
        type: z.string().optional(),
        path: z.string(),
        properties: z.record(z.unknown()).optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("resource", "createMaterial", params));
    }
  );

  server.registerTool(
    "godot_resource_set_material_property",
    {
      title: "Set Godot material property",
      description: "Set a property on a Godot material resource.",
      inputSchema: {
        path: z.string(),
        propertyName: z.string(),
        value: z.unknown(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("resource", "setMaterialProperty", params));
    }
  );

  server.registerTool(
    "godot_editor_play",
    {
      title: "Play Godot project",
      description: "Start play mode in the Godot editor.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runBridgeCommand("editor", "play"));
    }
  );

  server.registerTool(
    "godot_editor_stop",
    {
      title: "Stop Godot project",
      description: "Stop play mode in the Godot editor.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runBridgeCommand("editor", "stop"));
    }
  );

  server.registerTool(
    "godot_editor_pause",
    {
      title: "Pause Godot project",
      description: "Attempt to pause play mode through the Godot bridge.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runBridgeCommand("editor", "pause"));
    }
  );

  server.registerTool(
    "godot_console_clear",
    {
      title: "Clear Godot console logs",
      description: "Clear console logs exposed by the EngineForge bridge.",
      inputSchema: {},
    },
    async () => {
      return jsonText(await runBridgeCommand("console", "clear"));
    }
  );

  server.registerTool(
    "godot_animation_play",
    {
      title: "Play Godot animation",
      description: "Play an animation through the EngineForge bridge.",
      inputSchema: {
        nodePath: z.string().optional(),
        animation: z.string().optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("animation", "play", params));
    }
  );

  server.registerTool(
    "godot_animation_stop",
    {
      title: "Stop Godot animation",
      description: "Stop an animation through the EngineForge bridge.",
      inputSchema: {
        nodePath: z.string().optional(),
      },
    },
    async (params) => {
      return jsonText(await runBridgeCommand("animation", "stop", params));
    }
  );
}
