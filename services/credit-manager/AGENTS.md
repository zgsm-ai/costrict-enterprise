# AGENTS.md - Coding Guidelines for ZGSM Admin System

## Build & Development Commands

```bash
# Development server (runs on port 9527)
npm run dev

# Build for production
npm run build

# Type checking only
npm run type-check

# Lint and auto-fix
npm run lint

# Format with Prettier
npm run format

# Preview production build
npm run preview
```

> Note: No test framework is configured in this project.

## Tech Stack

- **Framework**: Vue 3 with Composition API (`<script setup>`)
- **Language**: TypeScript (strict mode)
- **Build Tool**: Vite 6
- **State Management**: Pinia
- **UI Library**: Naive UI
- **Styling**: Tailwind CSS + Less
- **Routing**: Vue Router 4
- **i18n**: Vue I18n
- **HTTP**: Axios

## Code Style Guidelines

### Formatting (Prettier)

- Semi-colons: **required**
- Quotes: **single quotes**
- Print width: **100 characters**
- Tab width: **4 spaces** (no tabs)
- Vue: no indentation for `<script>` and `<style>`
- Single attribute per line in templates

### Imports

- Use `@/` alias for src directory imports
- Group imports: Vue → third-party → `@/` aliases → relative
- Use `type` imports for type-only imports
- Example:

```typescript
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { NPopover } from 'naive-ui';
import { useUserStore } from '@/store/user';
import type { QuotaList } from '@/api/bos/quota.bo';
```

### Naming Conventions

- **Components**: PascalCase (e.g., `CommonHeader.vue`)
- **Files**: kebab-case for Vue files, camelCase for TS files
- **Variables/functions**: camelCase
- **Constants**: UPPER_SNAKE_CASE or camelCase
- **Types/Interfaces**: PascalCase
- **Store**: use descriptive names like `useUserStore`
- **Hooks**: prefix with `use` (e.g., `useProfile.ts`)
- **API modules**: suffix with `.mod.ts`
- **Business objects**: suffix with `.bo.ts`

### Vue Components

- Always use `<script setup lang="ts">`
- Use `scoped` styles with `lang="less"`
- Template attributes: one per line when multiple attributes
- Component names in PascalCase in templates
- Use Naive UI components (prefix with `n-`)

### TypeScript

- Strict mode enabled
- Always define return types for exported functions
- Use interfaces over types for object shapes
- Use `unknown` over `any` when possible
- Props and emits must be typed

### Error Handling

- Use try-catch for async operations
- Log errors with `console.error()`
- Use Naive UI's message API for user feedback
- Handle API errors in request interceptors

### API Patterns

- Place API calls in `src/api/mods/`
- Define types in `src/api/bos/`
- Use request utilities from `src/utils/request.ts`
- Return types: `Promise<ApiResponse<T>>`

### State Management

- Use Pinia stores in `src/store/`
- Use `$patch` for multiple state updates
- Use `storeToRefs` for reactive destructuring

### File Organization

```
src/
  api/
    mods/      # API modules
    bos/       # Business objects (types)
  components/  # Shared components
  composables/ # Shared composables
  router/      # Vue Router
  services/    # Business logic
  store/       # Pinia stores
  utils/       # Utilities
  views/       # Page components
    ViewName/
      components/  # View-specific components
      hooks/       # View-specific hooks
      const.ts     # View constants
      interface.ts # View types
```

## ESLint Rules

- Vue essential rules enabled
- TypeScript recommended rules
- Prettier integration (formatting handled by Prettier)
- Ignores: `dist/`, `dist-ssr/`, `coverage/`

## Key Conventions

- Base path: `/credit/manager`
- Default dev port: `9527`
- i18n keys: nested with dot notation (e.g., `common.header.logout`)
- API base URLs proxied in vite.config.ts
- Use composition API patterns (ref, reactive, computed)
- Prefer async/await over promise chains
