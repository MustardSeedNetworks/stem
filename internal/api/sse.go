// SPDX-License-Identifier: BUSL-1.1

package api

import "time"

// sseHeartbeatSec is the raw seconds value for the SSE heartbeat interval.
// Matches [sse.HeartbeatInterval] from the leaf package so the two stay in
// sync without creating an import dependency.
const sseHeartbeatSec = 15

// sseHeartbeatInterval is how often [handleSSEEvents] sends an SSE comment
// line (": heartbeat\n\n") to keep idle proxies from closing the connection.
// 15 s is a common threshold.
const sseHeartbeatInterval = sseHeartbeatSec * time.Second
