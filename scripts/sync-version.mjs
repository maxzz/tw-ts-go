#!/usr/bin/env node

import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const pkgPath = join(root, "package.json");
const out = join(root, "twts", "8-version.go");
const bump = process.argv.includes("--bump");

const pkg = JSON.parse(readFileSync(pkgPath, "utf8"));

if (bump) {
    const parts = pkg.version.split(".");
    if (parts.length !== 3 || parts.some((p) => Number.isNaN(Number(p)))) {
        throw new Error(`Invalid semver in package.json: ${pkg.version}`);
    }
    parts[2] = String(Number(parts[2]) + 1);
    pkg.version = parts.join(".");
    writeFileSync(pkgPath, `${JSON.stringify(pkg, null, 4)}\n`);
}

writeFileSync(
    out,
    [
        "package twts",
        "",
        "// Version is synced from package.json by scripts/sync-version.mjs.",
        `var Version = ${JSON.stringify(pkg.version)}`,
        "",
    ].join("\n"),
);
