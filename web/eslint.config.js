import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    rules: {
      // `const { [key]: _, ...rest } = obj` is the idiomatic way to omit a key,
      // and `_`-prefixed params document a signature we must match but do not use.
      '@typescript-eslint/no-unused-vars': [
        'error',
        { varsIgnorePattern: '^_', argsIgnorePattern: '^_' },
      ],
      // This codebase deliberately co-locates a field predicate with the component
      // it selects (isPublishField/PublishButton, isActionField/ActionButton, ...).
      // The rule only guards Vite fast-refresh ergonomics in dev, so keep it visible
      // as a warning rather than splitting four modules apart to satisfy it.
      'react-refresh/only-export-components': 'warn',
    },
  },
])
