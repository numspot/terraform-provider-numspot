package core

import "errors"

// errMissingCreateID is returned when a create response does not carry a
// parsable JSON201 id (unexpected 2xx shape, mangled body, or a stop-retry
// status that is not a real success). Create flows must fail with this error
// instead of dereferencing a nil JSON201 payload and crashing the plugin.
var errMissingCreateID = errors.New("unexpected create response: missing resource id")
