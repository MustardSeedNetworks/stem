# The Stem API Reference

**Status:** Current API (implemented) with a Target API vNext section for planned endpoints.
**Base URL:** `http://localhost:8080`

---

## Current API (Implemented)

### Health & Status

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/health` | GET | Server health check |
| `/api/stats` | GET | Current statistics |
| `/api/version` | GET | Version information |

### Mode & Settings

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/mode` | GET/POST | Get or set current operating mode |
| `/api/settings` | GET/POST | Get or update settings |

### Interfaces

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/interfaces` | GET | List available network interfaces |

### Tests

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/test/start` | POST | Start a test |
| `/api/test/stop` | POST | Stop running test |
| `/api/test/result` | GET | Get last test result |

### Modules

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/modules` | GET | List all test modules |
| `/api/modules/{name}` | GET | Get module details |

### Reflector

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/reflector/config` | GET/POST | Get or update reflector configuration |
| `/api/reflector/stats` | GET | Get reflector statistics |

### Licensing

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/license` | GET | License status |
| `/api/license/activate` | POST | Activate a license key |
| `/api/license/trial` | POST | Start a trial |

---

## Target API vNext (Planned, Not Implemented)

This section captures future endpoints that are not yet implemented. Do not rely on these for current integrations.

### Authentication

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/auth/login` | POST | Authenticate and return a token |
| `/api/auth/logout` | POST | Invalidate session |
| `/api/auth/refresh` | POST | Refresh token |

### Setup (First Run)

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/setup/status` | GET | Setup status |
| `/api/setup/complete` | POST | Complete setup |

### Test Management

| Endpoint | Method | Description |
| --- | --- | --- |
| `/api/tests` | GET | List tests with filters |
| `/api/tests/{id}` | GET | Get test details |
| `/api/tests/{id}` | DELETE | Cancel test |

### Notes

- JWT/auth flows are intentionally listed here as planned work.
- Move these endpoints into the current section only after implementation.
