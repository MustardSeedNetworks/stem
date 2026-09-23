# Code index

## Authentication state

`internal/authstate.Publish` durably publishes a protected `auth.json` record.
`State.Encode`, `Decode` and `State.Save` validate the single-principal schema.
Complete credentials use foundation's released passkey codec. The password hash
is opaque here; the password factory/verifier owns format validation. Invalid
present state is never bootstrap permission. The fail-closed secure loader and
manager adapter follow; login and recovery do not yet call this package.
An `ErrUncertain` result requires authentication shutdown and reload, not an
in-memory rollback: the replacement may already be visible on disk.
