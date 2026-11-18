import { Tabs as TabsPrimitive } from 'bits-ui';
import TabsList from './tabs-list.svelte';
import TabsTrigger from './tabs-trigger.svelte';

// Tabs namespace wrapper (shadcn-svelte shape): pages import
// `#lib/components/ui/tabs/index.ts` and never touch raw bits-ui. Root and Content
// need no extra styling, so they re-export the primitives directly.
export const Root = TabsPrimitive.Root;
export const Content = TabsPrimitive.Content;
export const List = TabsList;
export const Trigger = TabsTrigger;
