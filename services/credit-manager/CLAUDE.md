# CLAUDE.md

This file provides guidance to CoStrict when working with code in this repository.

## Project Overview

This is the **CoStrict Credit Manager** (`zgsm-admin-system`), a Vue 3 + TypeScript admin dashboard for managing user quotas, credits, subscriptions, and usage statistics. It serves as the backend management system for CoStrict.

## Tech Stack

- **Framework:** Vue 3 (Composition API, `<script setup lang="ts">`)
- **Build Tool:** Vite 6
- **State Management:** Pinia (single store: `src/store/user.ts`)
- **Router:** Vue Router 4 with custom auth guards
- **UI Library:** Naive UI (custom dark theme overrides in `src/theme-overrides.ts`)
- **Styling:** Tailwind CSS v4 + Less (scoped styles in `.vue` files)
- **I18n:** Vue I18n (`src/locales/zh.json`, `src/locales/en.json`)
- **HTTP Client:** Axios with custom interceptors (`src/utils/request.ts`)
- **Testing:** Vitest with happy-dom
- **Auth:** OIDC-based, token stored via js-cookie

## Common Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Start dev server on port 9527 (mock enabled) |
| `npm run dev:mock` | Start dev server with mock mode explicitly |
| `npm run build` | Type-check + production build |
| `npm run build-only` | Production build without type-check |
| `npm run type-check` | Run `vue-tsc --build` |
| `npm run preview` | Preview production build |
| `npm run lint` | Run ESLint with auto-fix |
| `npm run format` | Format `src/` with Prettier |
| `npm run test` | Run all Vitest tests once |
| `npm run test:watch` | Run Vitest in watch mode |
| `npm run test:ui` | Run Vitest with UI |

To run a single test file: `npx vitest run src/__tests__/store.user.spec.ts`

## Project Structure

```
src/
  api/
    mods/          # API endpoint modules (quota.mod.ts, activity.mod.ts)
    bos/           # Business object types (quota.bo.ts, activity.bo.ts)
  services/        # Service layer (auth.ts, user.ts) — wraps API calls with business logic
  store/           # Pinia stores (user.ts)
  router/
    index.ts       # Route definitions + PUBLIC_ROUTES / NO_FOOTER_ROUTES / ZH_ONLY_ROUTES
    guards/
      auth.ts      # Auth guard: background auth that doesn't block rendering
  views/           # Page-level components
    Home/          # Main dashboard with sidebar menu (profile / subscription / usage / activity)
      components/  # Section components (profile-section, usage-section, etc.)
      hook/        # Composables: useMenu, useProfile, useSubscription, useUsageStatistics, useResponsive
    Credits/       # Credit records page
    Reward/        # Credit reward plan page
    Subscribe/     # Subscription page
    Login/         # Login page
    Preview/       # Markdown preview page
  components/      # Shared layout components (common-header, common-content, common-footer, common-card)
  composables/     # Shared composables (useMetaTags.ts)
  config/          # Config files (meta.ts for social sharing)
  directives/      # Vue custom directives
  locales/         # i18n messages (zh.json, en.json)
  utils/
    request.ts     # Axios instance with interceptors (auto-attaches Bearer token, handles 401)
    token.ts       # Token manager (js-cookie + URL state cleanup)
    i18n.ts        # i18n helper for non-component contexts
    date.ts        # Date formatting utilities
    copy.ts        # Clipboard utilities
  assets/          # Static assets (images, fonts, base.less)
  theme-overrides.ts  # Naive UI global theme overrides (dark theme)
```

## Key Architecture Patterns

### API Layer

APIs are organized in two layers:
- **`src/api/mods/`** — thin HTTP wrappers that call `get()`/`post()` from `src/utils/request.ts`
- **`src/api/bos/`** — TypeScript interfaces for request/response payloads
- **`src/services/`** — business logic services that wrap API calls (e.g., `AuthService`, `UserService`)

All API responses follow a consistent envelope: `{ code: number, message: string, data: T }`. The axios interceptor in `request.ts` unwraps the response and auto-redirects to `/login` on 401.

### Authentication Flow

1. Token is stored in a cookie via `js-cookie` (`src/utils/token.ts`)
2. Hash token (`?state=xxx`) is exchanged for an access token via `/oidc-auth/api/v1/manager/token`
3. Auth guard (`src/router/guards/auth.ts`) uses **background authentication** — it calls `next()` immediately and authenticates in the background. This avoids blocking the UI, but means views must handle their own loading/unauthenticated states.
4. `AuthService` lazily imports the Pinia store (`await import('@/store/user')`) to avoid initialization order issues.
5. Public routes (`PUBLIC_ROUTES` in `src/router/index.ts`) skip auth checks.

### HTTP Request Pipeline

- `src/utils/request.ts` — single Axios instance with interceptors:
  - **Request**: injects `Authorization: Bearer <token>` from cookie.
  - **Response**: on non-zero `code`, shows Naive UI error message. On 401, redirects to `/login`.
  - The Naive UI `message` instance is set lazily via `setMessageInstance()` called from `App.vue`.
- Exported helpers: `request()`, `get()`, `post()`, `put()`, `del()` — all generic, returning `Promise<T>`.

### Page State Management (Home View)

The Home page (`src/views/Home/home-page.vue`) is organized into menu sections driven by composables in `src/views/Home/hook/`:
- `useMenu.ts` — menu state and navigation
- `useProfile.ts` — user info, quota data, invite code
- `useSubscription.ts` — order/subscription data with pagination
- `useUsageStatistics.ts` — usage consumption records with time range filtering
- `useResponsive.ts` — responsive layout detection (mobile vs desktop)

Data is loaded lazily per menu section via `loadMenuData()` in `home-page.vue`.

### Styling Conventions

- **Global styles:** `src/assets/base.less` (dark theme base, scrollbar styling)
- **Component styles:** Scoped Less inside `.vue` files
- **Utility classes:** Tailwind CSS v4 (check `tailwind.config.js` for custom fonts like `font-zcool`)
- **Naive UI theme:** Dark theme overrides in `src/theme-overrides.ts` — nearly all components are styled for a black background with white text

### Routing Quirks

- `scrollBehavior` in `src/router/index.ts` resets both the `.content` element scroll and `window.scrollTo(0, 0)` with a 50ms delay. This matters when navigating between views.
- `NO_FOOTER_ROUTES` hides the footer for certain paths.
- `ZH_ONLY_ROUTES` (`/subscribe`, `/annual-summary`, `/annual-summary-cover`) are blocked when `localStorage.getItem('app-language') === 'en'`.

### i18n

- Default locale is `zh` (Chinese). English is supported but some features are Chinese-only.
- Language is persisted to `localStorage` key `app-language`.
- `src/utils/i18n.ts` exports `getT()` for accessing translations outside Vue components (e.g., in Axios interceptors).

### Mock Environment

Mock data is provided via `vite-plugin-mock` during development (`npm run dev`). All backend API calls are intercepted and return synthetic data from the `mock/` directory.

- **`mock/auth.ts`** — OIDC auth mocks (`/oidc-auth/*`): token, userinfo, bind, invite-code, login URL
- **`mock/quota.ts`** — Quota mocks (`/quota-manager/*`): quota info, audit records, transfer in/out, usage statistics
- **`mock/order.ts`** — Order mocks (`/quota-order-manager/*`): orders list, create order, payment QR, invoices, quota types
- **`mock/activity.ts`** — Activity mocks (`/operational_activities/*`): user me, share record
- **`mock/_utils.ts`** — Shared `successResponse` / `errorResponse` helpers

Mock is **only enabled in development** (`command === 'serve'`) and does not affect production builds.

### Vite Config Notes

- `base: '/credit/manager'` — the app is served under this subpath
- Dev server proxies API paths to `https://zgsm.sangfor.com` (proxy is bypassed when mock is enabled for matching URLs):
  - `/quota-manager`
  - `/oidc-auth`
  - `/quota-order-manager`
  - `/operational_activities`

## Code Conventions

- **Imports**: group as Vue → third-party → `@/` aliases → relative. Use `type` imports for type-only imports.
- **File naming**: Vue files kebab-case, TS utilities camelCase, API modules `.mod.ts`, business object types `.bo.ts`.
- **Prettier**: 4-space tabs, single quotes, 100 char width, semicolons required.

## Testing

Tests are in `src/__tests__/` and use Vitest with happy-dom. Coverage is reported in text and HTML formats. Tests for stores, services, utilities, and router guards exist.

## Important Files to Know

| File | Purpose |
|------|---------|
| `src/main.ts` | App bootstrap: Pinia, router, i18n, head, directives, auth guard |
| `src/App.vue` | Root layout: header + Naive UI config provider + content + auth loading |
| `src/router/index.ts` | All routes + public route constants + language-based route blocking |
| `src/utils/request.ts` | Axios config with auth header injection and error handling |
| `src/theme-overrides.ts` | Naive UI dark theme customization |
| `vite.config.ts` | Build config, base path, dev proxy |
| `vitest.config.ts` | Test config with `@/` alias and happy-dom |

## Invoice Modal Validation Rules (`src/views/Home/components/invoice-modal.vue`)

### Common Fields (all header types)

| Field | Required | Trigger | Rule |
|-------|----------|---------|------|
| `headerType` | Yes | blur, change | must be selected |
| `invoiceType` | Yes | blur, change | must be selected |
| `email` | Yes | blur, input | required; must match `^[^\s@]+@[^\s@]+\.[^\s@]+$` |

### Company Header (`headerType === COMPANY`)

| Field | Required | Trigger | Rule |
|-------|----------|---------|------|
| `title` | Yes | blur, input | required |
| `taxNumber` | Yes | blur, input | required; exactly 18 digits or uppercase letters (`^[0-9A-Z]{18}$`) |
| `address` | **Dynamic** | blur, input | required only when `invoiceType === SPECIAL` (special VAT invoice) |
| `companyPhone` | **Dynamic** | blur, input | required only when `invoiceType === SPECIAL` |
| `bank` | **Dynamic** | blur, input | required only when `invoiceType === SPECIAL` |
| `bankAccount` | **Dynamic** | blur, input | required only when `invoiceType === SPECIAL` |

### Personal Header (`headerType === PERSONAL`)

| Field | Required | Trigger | Rule |
|-------|----------|---------|------|
| `title` | Yes | blur, input | required; must not start with a digit and length >= 2 (`^[^0-9].{1,}$`) |
| `mobile` | Yes | blur, input | required; mainland China mobile number (`^1[3-9]\d{9}$`) |

### Dynamic Behavior Notes

1. **Header type switch**: resets validation (`restoreValidation()`); switching to personal forces `invoiceType` to `VAT`.
2. **Invoice type switch**: resets validation; only company headers can select `SPECIAL`.
3. **Special invoice**: when `invoiceType === SPECIAL`, the four company-only fields (`address`, `phone`, `bank`, `bankAccount`) become required and show a red asterisk via dynamic `required` binding.
