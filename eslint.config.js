import js from '@eslint/js'
import ts from 'typescript-eslint'
import svelte from 'eslint-plugin-svelte'
import prettier from 'eslint-config-prettier'
import globals from 'globals'
import svelteConfig from './svelte.config.js'

/**
 * ESLint owns the conventions that need to be checked inside `.svelte` files.
 *
 * ast-grep (see `sgconfig.yml`) has no Svelte grammar, so the division of
 * labour is: ESLint enforces import boundaries and logging discipline across
 * both `.ts` and `.svelte`; ast-grep enforces the AST-shaped rules that only
 * apply to `.ts` stores.
 */
export default ts.config(
  {
    ignores: [
      'dist/',
      'node_modules/',
      'bindings/',
      '.wails/',
      '.playwright-mcp/',
      'jscpd-report/',
      // Vendored shadcn-svelte output — kept in the registry's own style.
      'src/lib/components/ui/',
    ],
  },

  js.configs.recommended,
  ...ts.configs.recommended,
  ...svelte.configs.recommended,

  {
    languageOptions: {
      globals: { ...globals.browser },
    },
  },

  {
    files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
    languageOptions: {
      parserOptions: {
        parser: ts.parser,
        extraFileExtensions: ['.svelte'],
        // Passing the real config is what lets svelte-eslint-parser resolve
        // `compilerOptions.runes`. Without it, a present-but-empty
        // `compilerOptions` key reads as `runes: false` and every `$state`
        // in the stores becomes a `no-undef` error.
        svelteConfig,
      },
    },
  },

  {
    rules: {
      // A leading underscore is this codebase's marker for "deliberately
      // unused" — snippet parameters that only exist to reach a later
      // positional one, and caught errors that are swallowed on purpose.
      // Module-level `_`-prefixed state is *not* exempted: dead store state
      // should still be reported.
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', caughtErrorsIgnorePattern: '^_' },
      ],

      // Frontend logging goes through `$lib/logger`, which adds timestamps in
      // dev and forwards warn/error to the Go logger in production.
      'no-console': 'error',

      // Go is reached through the generated bindings, never a raw
      // runtime call. `$lib/wails-bindings` is the public face of that layer —
      // it re-exports `commands` alongside `unwrapResult()`.
      'no-restricted-imports': [
        'error',
        {
          paths: [
            {
              name: '@wailsio/runtime',
              importNames: ['Call'],
              message:
                'Call Go through the generated bindings — see src/lib/wails-bindings.ts.',
            },
          ],
          patterns: [
            {
              group: ['**/bindings/**'],
              message:
                "Import from '$lib/wails-bindings' — it re-exports the commands together with unwrapResult().",
            },
          ],
        },
      ],
    },
  },

  {
    // The logger is the one place allowed to touch the console — that is its
    // whole job.
    files: ['src/lib/logger.ts'],
    rules: { 'no-console': 'off' },
  },

  {
    // CLI scripts are Node programs, not webview code.
    files: ['scripts/**/*.js'],
    languageOptions: { globals: { ...globals.node } },
    rules: { 'no-console': 'off' },
  },

  {
    // The re-export layer necessarily imports the generated file.
    files: ['src/lib/wails-bindings.ts', 'src/test/ipc.ts'],
    rules: { 'no-restricted-imports': 'off' },
  },

  // Must stay last: turns off every rule that would fight Prettier.
  prettier,
)
