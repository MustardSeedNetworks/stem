# Distribution & Licensing Roadmap

**Product:** Stem
**Model:** Commercial (paid license)
**Status:** Development - NOT FOR PUBLIC DISTRIBUTION

---

## Current State (Development)

All distribution channels are **locked down**:

| Channel | Status | Notes |
|---------|--------|-------|
| Container registry | DISABLED | No `container-push` target |
| Public downloads | DISABLED | No public artifacts |
| Package repos | DISABLED | .deb/.rpm stay local |

### Local Development Only

```bash
make container   # Builds locally only
make deb         # Creates dist/stem_*.deb (local)
make rpm         # Creates dist/stem_*.rpm (local)
```

---

## Distribution Strategy

### License Validation Required

Before any public/commercial distribution:

1. **License server integration** or offline license validation
2. **Hardware fingerprinting** (optional, for appliance model)
3. **Expiration/renewal** handling
4. **Feature gating** by license tier (modules)

### Existing License Code

Stem has license infrastructure skeleton at `internal/license/`. Needs:
- [ ] License key generation
- [ ] Validation logic
- [ ] Module-based feature gating
- [ ] Grace period handling

### Proposed License Tiers

| Tier | Modules | Target |
|------|---------|--------|
| Trial | All modules, 30-day limit | Evaluation |
| Benchmark | RFC 2544 only | Basic testing |
| Professional | All modules | Enterprise |
| OEM | White-label, custom modules | Partners |

---

## Module Licensing

Stem's modular architecture enables per-module licensing:

| Module | Color | Could be separate license |
|--------|-------|--------------------------|
| Benchmark | Red | Base tier |
| ServiceTest | Orange | Pro tier |
| TrafficGen | Yellow | Pro tier |
| Measure | Blue | Pro tier |
| Certify | Green | Pro tier |

---

## Deployment Channels (Future)

When ready for commercial release:

### 1. Private Container Registry
```bash
# Future - requires auth
CONTAINER_REGISTRY=registry.mustardseednetworks.com
make container-push
```

### 2. Customer Portal
- Authenticated download of .deb/.rpm
- License key provisioning
- Module activation

### 3. Appliance Image
- Pre-installed on test equipment
- Hardware-locked license
- Calibration certificates

---

## Pre-Release Checklist

- [ ] License validation implemented (`internal/license/`)
- [ ] Module gating by license tier
- [ ] License server deployed (or offline validation)
- [ ] Private registry configured
- [ ] Customer portal ready
- [ ] EULA/Terms of Service finalized
- [ ] Pricing per module/tier determined
- [ ] Support infrastructure ready

---

## Version Strategy

**Single source of truth:** Git tags

```bash
git tag v0.24.0          # Creates version
make build              # Embeds version via ldflags
./bin/stem --version    # Shows v0.24.0
```

- `package.json` version is `0.0.0` (ignored, real version from API)
- Container tags match git tags
- All artifacts include version in filename

---

## Security Considerations

1. **No secrets in containers** - Config injected at runtime
2. **License validation on startup** - App won't run without valid license
3. **Module activation** - Unlicensed modules disabled
4. **Tamper detection** - Binary signing (future)
5. **Calibration data** - Signed, tamper-evident

---

*Last updated: 2025-01-19*
*Status: Development lockdown*
