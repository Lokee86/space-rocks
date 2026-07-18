import { promises as fs } from "node:fs";
import path from "node:path";

import { defaultProcessJobManager } from "./job_manager.js";
import { WORKSPACE_ROOT } from "./paths.js";

const MAX_ARGS = 200;
const MAX_ARG_CHARS = 4096;
const MAX_ENV_ENTRIES = 50;
const MAX_ENV_VALUE_CHARS = 4096;
const REDACTED_ARG = "[redacted]";
const SENSITIVE_FLAG_PARTS = ["token", "password", "secret", "api-key", "apikey", "auth"];
const BLOCKED_ENV_NAMES = new Set([
  "PATH", "PATHEXT", "NODE_OPTIONS", "PYTHONPATH", "RUBYOPT", "BUNDLE_GEMFILE", "GEM_HOME", "GEM_PATH", "COMSPEC", "SHELL",
]);
const WINDOWS_GODOT_EXECUTABLE = "C:\\Godot_v4.6.3-stable_win64.exe";

export const RESTRICTED_COMMAND_REGISTRY = Object.freeze({
  go: "go", gofmt: "gofmt", python: "python", pytest: "pytest", ruby: "ruby", bundle: "bundle", rails: "rails",
  npm: "npm", node: "node", godot: "godot", rg: "rg", grep: "grep", find: "find", ls: "ls", cat: "cat",
  sed: "sed", head: "head", tail: "tail", wc: "wc", diff: "diff",
});
export const COMMAND_REGISTRY = RESTRICTED_COMMAND_REGISTRY;

function message(error) { return error instanceof Error ? error.message : String(error); }
function isWithin(root, target) {
  const relative = path.relative(root, target);
  return relative === "" || (relative !== ".." && !relative.startsWith(`..${path.sep}`) && !path.isAbsolute(relative));
}
function containsNul(value) { return value.includes("\0"); }
function flagName(argument) { return argument.split("=", 1)[0].toLowerCase(); }
function isSensitiveFlag(argument) {
  const name = flagName(argument);
  return name.startsWith("-") && SENSITIVE_FLAG_PARTS.some((part) => name.includes(part));
}

function redactArgs(args) {
  const publicArgs = [];
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (!isSensitiveFlag(argument)) {
      publicArgs.push(argument);
      continue;
    }
    const equals = argument.indexOf("=");
    publicArgs.push(equals >= 0 ? `${argument.slice(0, equals)}=${REDACTED_ARG}` : argument);
    if (equals < 0 && index + 1 < args.length) {
      publicArgs.push(REDACTED_ARG);
      index += 1;
    }
  }
  return publicArgs;
}

function hasFlag(args, flags) {
  return args.some((argument) => flags.includes(argument) || flags.some((flag) => argument.startsWith(`${flag}=`)));
}
function hasAttachedShortFlag(args, flags) {
  return args.some((argument) => flags.some((flag) => argument === flag || (argument.startsWith(flag) && !argument.startsWith("--") && argument.length > flag.length)));
}

function rejectPolicy(command, args) {
  if (command === "node" && (hasAttachedShortFlag(args, ["-e", "-p"]) || hasFlag(args, ["--eval", "--print"]))) throw new Error("node eval and print modes are not allowed");
  if (command === "python") {
    if (hasAttachedShortFlag(args, ["-c"])) throw new Error("python -c is not allowed");
    for (let index = 0; index < args.length; index += 1) {
      if (args[index] === "-m" || args[index] === "--module") {
        if (args[index + 1] !== "pytest") throw new Error("python -m is restricted to pytest");
      } else if (args[index].startsWith("-m") && args[index] !== "-m") {
        if (args[index].slice(2) !== "pytest") throw new Error("python -m is restricted to pytest");
      }
    }
  }
  if (command === "ruby" && hasAttachedShortFlag(args, ["-e"])) throw new Error("ruby -e is not allowed");
  if (command === "npm" && args.some((argument) => argument === "exec" || argument === "x")) throw new Error("npm exec and npm x are not allowed");
  if (command === "rails" && args.some((argument) => argument === "runner" || argument === "console")) throw new Error("rails runner and console are not allowed");
  if (command === "go") {
    const envIndex = args.indexOf("env");
    const cleanIndex = args.indexOf("clean");
    if (envIndex >= 0 && args.slice(envIndex + 1).some((argument) => argument === "-w" || argument.startsWith("-w="))) throw new Error("go env -w is not allowed");
    if (cleanIndex >= 0 && args.slice(cleanIndex + 1).some((argument) => argument === "-modcache")) throw new Error("go clean -modcache is not allowed");
  }
  if (command === "bundle") {
    if (args[0] === "exec") {
      const nested = args[1] ? path.basename(args[1]).replace(/\.bat$/i, "") : "";
      if (!["rails", "rake", "ruby", "rspec"].includes(nested)) throw new Error("bundle exec is restricted to rails, rake, ruby, or rspec");
    }
    if (args[0] === "config" && args.slice(1).some((argument) => argument === "--global" || argument === "--system" || argument.startsWith("--global=") || argument.startsWith("--system="))) {
      throw new Error("bundle config global and system settings are not allowed");
    }
  }
}

function resolveDefaultExecutable(command, platform, environment) {
  if (command === "node") return process.execPath;
  if (platform === "win32" && command === "npm") return "npm.cmd";
  if (platform === "win32" && (command === "bundle" || command === "rails")) return `${command}.bat`;
  if (command === "godot") return environment.GODOT_EXECUTABLE || (platform === "win32" ? WINDOWS_GODOT_EXECUTABLE : "godot");
  return command;
}

export class RestrictedCommandService {
  constructor({ processJobManager = defaultProcessJobManager, root = WORKSPACE_ROOT, fileSystem = fs, platform = process.platform, processEnvironment = process.env, resolveExecutable } = {}) {
    this.processJobManager = processJobManager;
    this.root = path.resolve(root);
    this.fileSystem = fileSystem;
    this.platform = platform;
    this.processEnvironment = processEnvironment;
    this.resolveExecutable = resolveExecutable ?? ((command) => resolveDefaultExecutable(command, this.platform, this.processEnvironment));
  }

  listCommands() { return Object.keys(RESTRICTED_COMMAND_REGISTRY); }

  async start({ command, args = [], cwd = ".", timeoutMs, env } = {}) {
    this._validateCommand(command);
    this._validateArgs(args);
    rejectPolicy(command, args);
    const resolvedCwd = await this._resolveCwd(cwd);
    const environment = this._resolveEnvironment(env);
    const executable = this.resolveExecutable(command);
    return this.processJobManager.start({
      command: executable,
      args: [...args],
      publicArgs: redactArgs(args),
      cwd: resolvedCwd,
      env: environment,
      timeoutMs,
      metadata: { kind: "workspace_command", command },
    });
  }

  _validateCommand(command) {
    if (typeof command !== "string" || !Object.hasOwn(RESTRICTED_COMMAND_REGISTRY, command)) throw new Error(`Unknown restricted command: ${command}`);
  }

  _validateArgs(args) {
    if (!Array.isArray(args) || args.length > MAX_ARGS || args.some((argument) => typeof argument !== "string")) throw new Error(`args must contain at most ${MAX_ARGS} strings`);
    for (const argument of args) {
      if (argument.length > MAX_ARG_CHARS) throw new Error(`each argument must be at most ${MAX_ARG_CHARS} characters`);
      if (containsNul(argument)) throw new Error("arguments cannot contain NUL characters");
    }
  }

  _resolveEnvironment(overrides) {
    if (overrides === undefined) return { ...this.processEnvironment };
    if (!overrides || typeof overrides !== "object" || Array.isArray(overrides)) throw new Error("env must be an object");
    const entries = Object.entries(overrides);
    if (entries.length > MAX_ENV_ENTRIES) throw new Error(`env cannot contain more than ${MAX_ENV_ENTRIES} entries`);
    const safe = {};
    for (const [name, value] of entries) {
      if (!/^[A-Z_][A-Z0-9_]*$/.test(name) || BLOCKED_ENV_NAMES.has(name) || name.startsWith("GIT_") || name.startsWith("SSH_")) throw new Error(`environment override is not allowed: ${name}`);
      if (typeof value !== "string" || value.length > MAX_ENV_VALUE_CHARS || containsNul(value)) throw new Error(`environment value ${name} must be a string of at most ${MAX_ENV_VALUE_CHARS} characters without NUL`);
      safe[name] = value;
    }
    return { ...this.processEnvironment, ...safe };
  }

  async _resolveCwd(requestedCwd) {
    if (typeof requestedCwd !== "string" || containsNul(requestedCwd)) throw new Error("cwd must be a valid workspace path");
    const lexicalRoot = path.resolve(this.root);
    const lexicalCwd = path.resolve(lexicalRoot, requestedCwd);
    if (!isWithin(lexicalRoot, lexicalCwd)) throw new Error("cwd escapes workspace root");
    let realRoot;
    let realCwd;
    try {
      realRoot = await this.fileSystem.realpath(lexicalRoot);
      realCwd = await this.fileSystem.realpath(lexicalCwd);
    } catch (error) {
      throw new Error(`cwd must exist inside workspace root: ${message(error)}`);
    }
    if (!isWithin(realRoot, realCwd)) throw new Error("cwd resolves outside workspace root");
    const stat = await this.fileSystem.stat(realCwd);
    if (!stat.isDirectory()) throw new Error("cwd must be an existing directory");
    return realCwd;
  }
}

export const defaultRestrictedCommandService = new RestrictedCommandService();
