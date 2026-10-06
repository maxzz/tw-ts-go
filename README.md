# twts

Rewrite TypeScript import and export statements in `.ts` and `.tsx` files.

`twts` scans a folder or a file, rewrites import and re-export statements in place, and can check which files still break the rules. **Rewriting is the default.** Pass `-c` to list non-compliant files and leave them unchanged.

Built with Go. Distributed on npm as `tw-ts-go` with a small Node.js launcher. The command name is `twts`.

- **Repository:** [github.com/maxzz/tw-ts-go](https://github.com/maxzz/tw-ts-go)
- **npm package:** [npmjs.com/package/tw-ts-go](https://www.npmjs.com/package/tw-ts-go)

## Contents

- [Installation](#installation)
- [Usage](#usage)
  - [CLI](#cli)
  - [Go library](#go-library)
- [Options](#options)
- [What gets scanned](#what-gets-scanned)
- [Rewrite rules](#rewrite-rules)
- [Check output](#check-output)
- [Development](#development)
- [Publishing to npm](#publishing-to-npm)
- [License](#license)

## Installation

```bash
npm install -D tw-ts-go
```

Or run without installing:

```bash
npx twts src
```

## Usage

### CLI

```bash
twts [options] <path>
```

`<path>` is required. It is a folder or a single `.ts` / `.tsx` file.

```bash
# Rewrite imports under src (default)
twts src

# Rewrite one file
twts src/App.tsx

# List files that still break the rules. Does not write.
twts -c src

# Scan only the folder itself
twts --no-recursive src

# Print full paths instead of an indented tree
twts -c --no-tree src
```

On Windows, when `twts` is attached to a console, it prints a colored report and waits for a key press before closing the window. Piped and non-console runs print the report and exit.

### Go library

```go
import "twts/twts"

next, err := twts.Transform(source, false)
result, err := twts.RunScan("src", twts.ScanOptions{Recursive: true, Check: true})
```

`Check: true` fills `result.Hits` with non-compliant files and does not write. The default (`Check: false`) rewrites those files.

## Options

| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--check` | `-c` | off | Verify files and print the non-compliant names. Does not modify files. Exit code 1 when any file fails. |
| `--recursive` | — | **on** | Walk subfolders. |
| `--no-recursive` | — | off | Scan only the selected folder. A file argument is still that one file. |
| `--tree` | — | **on** | Print matching files as a folder tree with indentation. |
| `--no-tree` | — | off | Print each matching file as a full path, one per line. |
| `--help` | `-h` | off | Show usage, options, and examples. |

`--recursive=false` and `--tree=false` turn those switches off. `--recursive=true` and `--tree=true` turn them on. The last switch wins.

**Arguments**

| Argument | Default | Description |
|----------|---------|-------------|
| `path` | required | One folder or one file to process. |

**Examples**

```bash
twts src
twts -c src
twts --no-recursive src
twts -c --no-tree src/App.tsx
```

## What gets scanned

`twts` reads `.ts` and `.tsx` files, including `.d.ts`, `.test.ts`, and `.test.tsx`.

Directories named `node_modules`, `dist`, and `.git` are skipped. With `--no-recursive`, subfolders are skipped as well.

A file is non-compliant when rewriting it would change the text. Check mode prints those names. Fix mode writes the rewritten text and prints the names of the files it updated.

## Rewrite rules

A statement is rewritten when its line starts with `import` or `export` (leading spaces and tabs are allowed). The whole statement is rewritten, including a `from` clause on a following line.

- Module paths use double quotes.
- A trailing `.ts` or `.tsx` is removed.
- A path ending in `/index.ts`, `/index.tsx`, or `/index` uses the folder. `./index.ts` becomes `"."` and `../index.ts` becomes `".."`.
- These extensions stay: `.css`, `.js`, `.d.ts`, `.test.ts`, `.test.tsx`.
- `import type { A, B }` becomes `import { type A, type B }`. `export type { A, B }` becomes `export { type A, type B }`. Each type gets its own `type` keyword.
- `import type Name`, `import type * as Name`, and `export type * as Name` stay type-only. A default type has no brace form.
- `import type Name, { A }` keeps `import type` on the default and writes `type` on each braced name.

An `export` that is not a re-export is left as written. That includes `export function`, `export const`, `export type Name = …`, `await import()`, `const x = import()`, and `vi.mock(..., () => import(...))`. A line that itself starts with `import(` is rewritten.

### Cases

| Case | Input | Output |
| --- | --- | --- |
| Grouped type import | `import type { Foo, Bar as Baz } from './a.ts';` | `import { type Foo, type Bar as Baz } from "./a";` |
| Grouped type import, several lines | `import type {\n  Foo,\n  Bar,\n} from './a.tsx';` | `import {\n  type Foo,\n  type Bar,\n} from "./a";` |
| Type and value in one import | `import { type Foo, bar } from './b.ts';` | `import { type Foo, bar } from "./b";` |
| Alias inside a grouped type import | `import type { Foo as Bar } from './a.ts'` | `import { type Foo as Bar } from "./a"` |
| Side-effect import, non-script file | `import './c.css';` | `import "./c.css";` |
| Side-effect import, script file | `import './c.ts';` | `import "./c";` |
| Side-effect import, quotes already double | `import "already.ts";` | `import "already";` |
| Default type import | `import type Foo from './f.ts';` | `import type Foo from "./f";` |
| Namespace type import | `import type * as NS from './g.ts';` | `import type * as NS from "./g";` |
| Default type plus named types | `import type Foo, { Bar } from './h.ts';` | `import type Foo, { type Bar } from "./h";` |
| Default binding named `type` | `import type from './i.ts';` | `import type from "./i";` |
| Default binding named `type`, plus a value | `import type, { Foo } from './j.ts';` | `import type, { Foo } from "./j";` |
| Value import, package name | `import { a, b } from "react";` | `import { a, b } from "react";` |
| Default plus namespace | `import React, * as ReactDOM from 'react';` | `import React, * as ReactDOM from "react";` |
| Namespace import | `import * as path from 'node:path';` | `import * as path from "node:path";` |
| Declaration file | `import type { A } from './file.d.ts';` | `import { type A } from "./file.d.ts";` |
| Folder barrel, `.ts` | `import { type MenuAction, runMenuAction } from '@/editor/0-core/menu-actions/index.ts';` | `import { type MenuAction, runMenuAction } from "@/editor/0-core/menu-actions";` |
| Folder barrel, `.tsx` | `import { StreamsSelector } from '@/editor/6-streams/index.tsx';` | `import { StreamsSelector } from "@/editor/6-streams";` |
| Folder barrel, extension already removed | `import { setBatchFiles } from '@/editor/2-file/index';` | `import { setBatchFiles } from "@/editor/2-file";` |
| Relative folder barrel | `import { loadViewsSideEffects } from './views-load-side-effects/index.ts';` | `import { loadViewsSideEffects } from "./views-load-side-effects";` |
| Current folder barrel | `import { local } from './index.ts';` | `import { local } from ".";` |
| Parent folder barrel | `import { parent } from '../index.tsx';` | `import { parent } from "..";` |
| Vitest module, `.test.ts` | `import { readFile } from './foo.test.ts';` | `import { readFile } from "./foo.test.ts";` |
| Vitest module, `.test.tsx` | `import { readFile } from './foo.test.tsx';` | `import { readFile } from "./foo.test.tsx";` |
| Re-export | `export { X } from './d.tsx';` | `export { X } from "./d";` |
| Type-only re-export | `export type { Y } from './e.ts';` | `export { type Y } from "./e";` |
| Grouped type re-export | `export type { Foo, Bar as Baz } from './a.ts';` | `export { type Foo, type Bar as Baz } from "./a";` |
| Re-export everything | `export * from './d.ts';` | `export * from "./d";` |
| Re-export namespace | `export * as NS from './g.tsx';` | `export * as NS from "./g";` |
| Type-only namespace re-export | `export type * as NS from './g.ts';` | `export type * as NS from "./g";` |
| Re-export of a folder barrel | `export { X } from '@/editor/2-file/index.ts';` | `export { X } from "@/editor/2-file";` |
| Type alias, not a re-export | `export type Foo = './a.ts';` | `export type Foo = './a.ts';` |
| Dynamic import | `const { x } = await import('./z.tsx');` | `const { x } = await import('./z.tsx');` |
| Dynamic import inside an export | `export const x = await import('./a.ts');` | `export const x = await import('./a.ts');` |
| Vitest mock factory | `vi.mock('@/editor/0-core/8-lib/main-api.ts', () => import('./main-api-mock.ts'));` | `vi.mock('@/editor/0-core/8-lib/main-api.ts', () => import('./main-api-mock.ts'));` |

`\n` in the table is a real line break.

## Check output

Colors:

| Color | Meaning |
|-------|---------|
| Cyan | Operation title and folder names |
| Yellow | Non-compliant file names and the failure summary |
| Green | Files updated by a fix, and a clean check |
| Gray | Version and counts |
| Red | Errors |

Tree mode (default):

```text
Check imports

Scanned 4 files

src
  app.ts
  lib
    util.ts

2 files do not match import rules.
```

Flat mode (`--no-tree`):

```text
C:\proj\src\app.ts
C:\proj\src\lib\util.ts
```

Fix mode uses the same shapes for the files it updated.

## Development

```bash
# Check the sample tree (no writes)
npm run dev

# Rewrite the sample tree
npm run fix

# Go tests
npm test

# Build the Windows binary (default)
npm run build

# Build all platform binaries (for npm publish)
npm run build:all
```

`testdata` holds a small tree with both clean and non-compliant files. `npm run dev` checks it. `npm run fix` rewrites it.

## Publishing to npm

1. Log in: `npm login`
2. Build all platform binaries: `npm run build:all`
3. Publish: `npm publish` (or `npm run to-npm`)

`prepublishOnly` runs `build:all` automatically before publish. The published package name is `tw-ts-go`. The installed command is `twts`.

## License

MIT — see the `license` field in [package.json](./package.json).
