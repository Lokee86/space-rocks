import { spawn } from "node:child_process";
import path from "node:path";

import { REPO_ROOT } from "./paths.js";

const COMMANDS = {
  go_server_tests: {
    cwd: path.join(REPO_ROOT, "services/game-server"),
    command: "go",
    args: ["test", "-buildvcs=false", "./..."],
  },

  godot_unit_tests: {
    cwd: REPO_ROOT,
    command: "godot",
    args: [
      "--headless",
      "--path",
      "client",
      "-s",
      "res://addons/gut/gut_cmdln.gd",
      "-gdir=res://tests/unit",
      "-ginclude_subdirs",
      "-gexit",
    ],
  },

  tools_boundary_tests: {
    cwd: REPO_ROOT,
    command: "python",
    args: ["-m", "pytest", "tools/tests"],
  },

  data_sync_tests: {
    cwd: REPO_ROOT,
    command: "python",
    args: ["-m", "pytest", "tools/data_sync/tests"],
  },
};

export function listAllowedCommands() {
  return Object.keys(COMMANDS);
}

export function runAllowedCommand(name) {
  const config = COMMANDS[name];

  if (!config) {
    throw new Error(`Unknown allowed command: ${name}`);
  }

  return new Promise((resolve) => {
    const child = spawn(config.command, config.args, {
      cwd: config.cwd,
      shell: false,
      windowsHide: true,
    });

    let stdout = "";
    let stderr = "";

    child.stdout.on("data", (chunk) => {
      stdout += chunk.toString();
    });

    child.stderr.on("data", (chunk) => {
      stderr += chunk.toString();
    });

    child.on("error", (error) => {
      resolve({
        exit_code: -1,
        stdout,
        stderr: `${stderr}${error.message}`,
      });
    });

    child.on("close", (code) => {
      resolve({
        exit_code: code ?? -1,
        stdout,
        stderr,
      });
    });
  });
}
