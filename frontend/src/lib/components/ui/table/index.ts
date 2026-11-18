import Table from './table.svelte';
import TableHeader from './table-header.svelte';
import TableBody from './table-body.svelte';
import TableRow from './table-row.svelte';
import TableHead from './table-head.svelte';
import TableCell from './table-cell.svelte';

// Table namespace wrapper (shadcn-svelte shape): list pages import
// `#lib/components/ui/table/index.ts` and share one table chrome (ticket 06).
export { Table, TableHeader, TableBody, TableRow, TableHead, TableCell };
