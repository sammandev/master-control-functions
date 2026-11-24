-- Match the table's numeric, base-sensitivity text sort before paging.
CREATE COLLATION example_sort_natural (
    provider = icu,
    locale = 'und-u-kn-ks-level1',
    deterministic = false
);
