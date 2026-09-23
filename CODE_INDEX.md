# Code index

## Authentication state

`internal/authstate.Publish` durably publishes a protected `auth.json` record.
The next AUTH-STEM increment supplies its single-principal schema, fail-closed
loader and manager adapter. It is not yet called by login or recovery.
An `ErrUncertain` result requires authentication shutdown and reload, not an
in-memory rollback: the replacement may already be visible on disk.
