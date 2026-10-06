package twts

import "testing"

func TestTransformFixtures(t *testing.T) {
	cases := []struct{ in, want string }{
		{`import type { Foo, Bar as Baz } from './a.ts';`, `import { type Foo, type Bar as Baz } from "./a";`},
		{"import type {\n  Foo,\n  Bar,\n} from './a.tsx';", "import {\n  type Foo,\n  type Bar,\n} from \"./a\";"},
		{`import { type Foo, bar } from './b.ts';`, `import { type Foo, bar } from "./b";`},
		{`import './c.css';`, `import "./c.css";`},
		{`import './c.ts';`, `import "./c";`},
		{`export { X } from './d.tsx';`, `export { X } from "./d";`},
		{`export type { Y } from './e.ts';`, `export { type Y } from "./e";`},
		{`export type { Foo, Bar as Baz } from './a.ts';`, `export { type Foo, type Bar as Baz } from "./a";`},
		{`export * from './d.ts';`, `export * from "./d";`},
		{`export * as NS from './g.tsx';`, `export * as NS from "./g";`},
		{`export type * as NS from './g.ts';`, `export type * as NS from "./g";`},
		{`export { X } from '@/editor/2-file/index.ts';`, `export { X } from "@/editor/2-file";`},
		{`export type Foo = './a.ts';`, `export type Foo = './a.ts';`},
		{`import type Foo from './f.ts';`, `import type Foo from "./f";`},
		{`import type * as NS from './g.ts';`, `import type * as NS from "./g";`},
		{`import type Foo, { Bar } from './h.ts';`, `import type Foo, { type Bar } from "./h";`},
		{`import type from './i.ts';`, `import type from "./i";`},
		{`import type, { Foo } from './j.ts';`, `import type, { Foo } from "./j";`},
		{`import "already.ts";`, `import "already";`},
		{`const { x } = await import('./z.tsx');`, `const { x } = await import('./z.tsx');`},
		{`import { a, b } from "react";`, `import { a, b } from "react";`},
		{`import React, * as ReactDOM from 'react';`, `import React, * as ReactDOM from "react";`},
		{`import * as path from 'node:path';`, `import * as path from "node:path";`},
		{`import type { A } from './file.d.ts';`, `import { type A } from "./file.d.ts";`},
		{`export const x = await import('./a.ts');`, `export const x = await import('./a.ts');`},
		{"import type { A } from './a.ts';\nexport function F(){ return <div className=\"from './nope.ts'\">x</div>; }\nexport { A } from './b.tsx';\nconst x = import('./c.ts');", "import { type A } from \"./a\";\nexport function F(){ return <div className=\"from './nope.ts'\">x</div>; }\nexport { A } from \"./b\";\nconst x = import('./c.ts');"},
		{`vi.mock('@/editor/0-core/8-lib/main-api.ts', () => import('./main-api-mock.ts'));`, `vi.mock('@/editor/0-core/8-lib/main-api.ts', () => import('./main-api-mock.ts'));`},
		{`import { type MenuAction, runMenuAction } from '@/editor/0-core/menu-actions/index.ts';`, `import { type MenuAction, runMenuAction } from "@/editor/0-core/menu-actions";`},
		{`import { StreamsSelector } from '@/editor/6-streams/index.tsx';`, `import { StreamsSelector } from "@/editor/6-streams";`},
		{`import { setBatchFiles } from '@/editor/2-file/index';`, `import { setBatchFiles } from "@/editor/2-file";`},
		{`import { loadViewsSideEffects } from './views-load-side-effects/index.ts';`, `import { loadViewsSideEffects } from "./views-load-side-effects";`},
		{`import { local } from './index.ts';`, `import { local } from ".";`},
		{`import { parent } from '../index.tsx';`, `import { parent } from "..";`},
		{`import { readFile } from './foo.test.ts';`, `import { readFile } from "./foo.test.ts";`},
		{`import { readFile } from './foo.test.tsx';`, `import { readFile } from "./foo.test.tsx";`},
		{`import type { Foo as Bar } from './a.ts'`, `import { type Foo as Bar } from "./a"`},
		{"import type {\n  // keep\n  Foo,\n} from './a.ts';", "import {\n  // keep\n  type Foo,\n} from \"./a\";"},
		{"import type {\r\n  Foo,\r\n} from './a.ts';", "import {\r\n  type Foo,\r\n} from \"./a\";"},
		{`import { a } from 'it\'s.ts';`, `import { a } from "it's";`},
		{`import('./c.ts');`, `import("./c");`},
	}

	for i, tc := range cases {
		for _, jsx := range []bool{false, true} {
			got, err := Transform(tc.in, jsx)
			if err != nil {
				t.Fatalf("case %d jsx=%v: %v", i, jsx, err)
			}
			if got != tc.want {
				t.Fatalf("case %d jsx=%v\n in: %q\nexp: %q\ngot: %q", i, jsx, tc.in, tc.want, got)
			}
			again, err := Transform(got, jsx)
			if err != nil {
				t.Fatalf("case %d second pass: %v", i, err)
			}
			if again != got {
				t.Fatalf("case %d not idempotent\n once: %q\n twice: %q", i, got, again)
			}
		}
	}
}
