#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const binaryName = process.platform === "win32" ? "twts.exe" : "twts";

const platformMap = {
    darwin: "darwin",
    linux: "linux",
    win32: "windows",
};

const archMap = {
    x64: "amd64",
    arm64: "arm64",
};

const os = platformMap[process.platform];
const arch = archMap[process.arch];

if (!os || !arch) {
    console.error(`twts: unsupported platform ${process.platform}/${process.arch}`);
    process.exit(1);
}

const binaryPath = join(__dirname, "..", "binaries", `${os}-${arch}`, binaryName);

if (!existsSync(binaryPath)) {
    console.error(`twts: binary not found for ${os}-${arch} at ${binaryPath}`);
    console.error("twts: reinstall the package or run npm run build");
    process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), { stdio: "inherit" });
process.exit(result.status ?? 1);
