# PDH package

`internal/pdh` wraps the Performance Data Helper API for the typed collectors
(`NewCollector`, `NewQueryCollector`) and the configurable `performancecounter`
collector (`NewDynamicCollector`).

## Ownership

- Constructors return either a value and a nil error, or nil and an error.
  `NewCollector` and `NewDynamicCollector` close the query they opened, and
  `NewQueryCollector` removes the counters it added to the shared query,
  before they return an error. Callers never close a value they did not
  receive; `Collector.Close` and `Query.Close` accept nil receivers.
- Collectors usually assign the result to a field before checking the error,
  so after an error the field is nil. A `Build` that tolerates the error must
  check the field for nil in `Collect`; the parent `Close` can close the field
  unconditionally.
- `newCollector` holds the lock of a shared query while it adds counters, so
  its error paths use `Collector.release`, which expects the caller to hold
  the locks or be the only user. `Close` takes the locks and calls it.
- Close collectors of a `Query` before the query itself.

## Struct tags

- `perfdata:"Counter"` marks a `float64` field; `perfdata:"Counter,secondvalue"`
  receives the second raw value (the base of a ratio or average).
- `perfdata_min_build:"<build>"` skips a counter that is missing on an older
  Windows build. The value must be a decimal build number that fits a
  `uint16`; anything else is a constructor error, so a typo cannot silently
  disable the check.

## Native calls

- The wrappers return PDH status codes. A path that cannot be converted to
  UTF-16 (it contains a NUL character) returns `CstatusBadCountername`
  instead of passing a NULL path to PDH.
- Build a `pdh.Error` with `NewPdhError` only for a status that is reported as
  an error: it formats the message text with `FormatMessage`.
