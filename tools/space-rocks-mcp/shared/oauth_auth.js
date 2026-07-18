import dotenv from "dotenv";
import { createRemoteJWKSet, jwtVerify } from "jose";
import { fileURLToPath } from "node:url";

dotenv.config({ path: fileURLToPath(new URL("../.env", import.meta.url)) });

function requireEnv(name) {
  const value = process.env[name];
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

function parseHttpsUrl(value, name) {
  let url;
  try {
    url = new URL(value);
  } catch {
    throw new Error(`Invalid ${name}: must be an absolute URL`);
  }

  if (url.protocol !== "https:") {
    throw new Error(`Invalid ${name}: must use https:`);
  }

  return url;
}

function normalizeIssuer(value) {
  return value.endsWith("/") ? value : `${value}/`;
}

const rawIssuer = requireEnv("AUTH0_ISSUER");
const rawAudience = requireEnv("AUTH0_AUDIENCE");
const rawResourceServerUrl = requireEnv("RESOURCE_SERVER_URL");

const issuerUrl = parseHttpsUrl(rawIssuer, "AUTH0_ISSUER");
const resourceServerUrl = parseHttpsUrl(rawResourceServerUrl, "RESOURCE_SERVER_URL");

export const AUTH0_ISSUER = normalizeIssuer(issuerUrl.toString());
export const AUTH0_AUDIENCE = rawAudience;
export const RESOURCE_SERVER_URL = rawResourceServerUrl;

export const PROTECTED_RESOURCE_METADATA_URL = new URL(
  ".well-known/oauth-protected-resource",
  RESOURCE_SERVER_URL
).toString();

export const PROTECTED_RESOURCE_METADATA = {
  resource: RESOURCE_SERVER_URL,
  authorization_servers: [AUTH0_ISSUER],
  metadata_url: PROTECTED_RESOURCE_METADATA_URL,
};

const jwksUrl = new URL(".well-known/jwks.json", AUTH0_ISSUER);
const remoteJwks = createRemoteJWKSet(jwksUrl);

export async function validateBearerAccessToken(token) {
  const { payload } = await jwtVerify(token, remoteJwks, {
    algorithms: ["RS256"],
    issuer: AUTH0_ISSUER,
    audience: AUTH0_AUDIENCE,
  });

  return payload;
}
