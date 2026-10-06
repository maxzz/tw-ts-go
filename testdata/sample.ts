import type { Foo, Bar as Baz } from './a.ts';
import { ready } from './nested/index';
import './styles.css';

export type { Y } from './e.ts';

export const kept = "import type { No } from './no.ts'";
