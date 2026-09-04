# Stem → Seed Feature Migration Guide

This document outlines the architectural patterns and components from Stem that can be adapted for Seed. Focus is on
reusable UI patterns, not test-specific configurations.

## 1. Component Patterns to Port

### ModuleSelector Pattern

**Stem**: `src/components/ModuleSelector.tsx`

- Color-coded module cards with enable/disable toggles
- Collapsible sections for test selection
- Auto-start configuration per module

**Seed Adaptation**:

- Replace test modules with service modules (Connectivity, Identity, Gateway, etc.)
- Keep the color-coding and enable/disable pattern
- Adapt auto-start to service health checking

### ResultHistory Pattern

**Stem**: `src/components/ResultHistory.tsx`

- Historical results browser with filtering
- Result comparison capabilities
- Export functionality

**Seed Adaptation**:

- Adapt for service deployment history
- Configuration change history
- Audit log visualization

### TestProgressBar → ServiceStatusBar

**Stem**: `src/components/TestProgressBar.tsx`

- Real-time progress with elapsed/ETA
- Step-by-step progress indication
- Status colors (running, completed, error)

**Seed Adaptation**:

- Deployment progress tracking
- Service provisioning status
- Health check progress

## 2. State Management Patterns

### Profile Store Pattern

**Stem**: `src/stores/profileStore.ts`

```typescript
// Key patterns to port:
- Backend defaults fallback (Profile → Backend → Hardcoded)
- Optimistic updates with rollback on error
- Settings by category with typed updates
- Import/Export functionality
```

**Seed Adaptation**:

- Configuration profiles (dev, staging, prod)
- Environment-specific settings
- Tenant configuration management

### API Client Pattern

**Stem**: `src/api/profiles.ts`

```typescript
// Reusable patterns:
- ApiError class with status codes
- fetchJson wrapper with credentials
- Consistent endpoint structure
```

## 3. Settings Organization

### Collapsible Section Groups

**Stem**: `src/components/settings/tests/*.tsx`

- Category-based settings organization
- Translation key structure for i18n
- Consistent checkbox/toggle patterns

**i18n Structure** (port to Seed):

```json
{
  "settings": {
    "sections": { "services": "Services", "security": "Security" },
    "services": {
      "connectivity": {
        "name": "Connectivity",
        "desc": "Network connectivity services"
      }
    }
  }
}
```

## 4. UI Components Already Shared

These components were harmonized and exist in both projects:

- ✅ Card / CardRow / CardValue
- ✅ StatusBadge
- ✅ CollapsibleSection
- ✅ ErrorBoundary
- ✅ HeaderBar
- ✅ Theme tokens (colors, spacing, typography)

## 5. Architecture Patterns

### Module-Based Organization

```text
Stem:
modules/
  ├── benchmark/     → RFC 2544
  ├── servicetest/   → Y.1564
  ├── trafficgen/    → Traffic Gen
  └── measure/       → Y.1731

Seed Adaptation:
services/
  ├── connectivity/  → Network connectivity
  ├── identity/      → User/device identity
  ├── gateway/       → API gateway
  └── monitoring/    → Observability
```

### Form Pattern with Config Types

```typescript
// Stem pattern
interface RFC2544Config {
  framesSizes: number[];
  duration: number;
  // ...
}

// Seed adaptation
interface ServiceConfig {
  endpoints: EndpointConfig[];
  healthCheck: HealthCheckConfig;
  // ...
}
```

## 6. Migration Priority

### High Priority (Port First)

1. **Profile/Configuration Store** - Multi-environment config management
2. **Result History** → Deployment/Change History
3. **i18n Structure** - Translation organization

### Medium Priority

4. **ModuleCard** → ServiceCard pattern
5. **TestProgressBar** → DeploymentProgress
6. **Settings Drawer modularization**

### Lower Priority (Nice to Have)

7. **Export/Import functionality**
8. **Auto-start patterns** → Auto-healing/auto-scaling hooks

## 7. Key Differences

| Aspect | Stem | Seed |
| -------- | ------ | ------ |
| Primary Focus | Network testing | Network services |
| Time Horizon | Short tests (seconds-minutes) | Long-running services |
| Results | Pass/fail metrics | Health/availability |
| Configuration | Test parameters | Service parameters |
| State | Transient (test runs) | Persistent (services) |

## 8. Implementation Notes

- Keep the same file structure conventions
- Use identical theme tokens and CSS patterns
- Maintain TypeScript strict mode
- Follow the same Biome lint rules
- Use identical Storybook story patterns
