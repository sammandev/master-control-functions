# Use Shadcn-Svelte Theme Vocabulary

The frontend uses shadcn-svelte token names in `frontend/src/app.css` so registry components can use the owned theme without renaming their color utilities. This replaced the custom vocabulary at the cost of a one-time migration; no compatibility aliases remain. `primary` represents the brand accent, `accent` and `muted` are backgrounds, and `muted-foreground` represents subdued text.

App-specific status, Device, modal-overlay, and mono tokens remain in the same layer. Both themes must provide readable text and keyboard focus; the dark destructive foreground uses the dark background token to meet normal-text contrast requirements.

The current styling contract is published in [the frontend README](../../frontend/README.md#styling-contract). Later approved workspace decisions supersede the original UI Refresh proposal's fixed width, section cards, and density toggle: production uses full-width unframed workspaces and no density preference. Shared layout and feedback classes remain where they still have consumers.
