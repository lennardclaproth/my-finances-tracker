import type { StorybookConfig } from '@storybook/sveltekit';

const config: StorybookConfig = {
  "stories": [
    "../src/**/*.mdx",
    "../src/**/*.stories.@(js|ts|svelte)"
  ],
  "addons": [
    "@storybook/addon-svelte-csf",
    "@chromatic-com/storybook",
    "@storybook/addon-vitest",
    "@storybook/addon-a11y",
    "@storybook/addon-docs"
  ],
  "framework": "@storybook/sveltekit",
  // Serve SvelteKit's static/ at the root so /fonts/*.woff2 (referenced by @font-face in app.css)
  // resolve in Storybook exactly as they do in the running app.
  "staticDirs": ["../static"]
};
export default config;