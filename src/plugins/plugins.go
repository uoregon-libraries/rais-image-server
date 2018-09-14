// Package plugins holds some example (optional) plugins to demonstrate how
// RAIS could be extended.  These plugins are not necessarily fleshed out
// fully, and are intended more to show what's possible than what may be
// useful.
package plugins

import "errors"

// Skipped is an error plugins can return to state that they didn't actually
// handle a given task, and other plugins should be used instead.  It shouldn't
// generally be reported, as it's not a situation that's concerning (much like
// io.EOF when reading a file).
var Skipped = errors.New("plugin doesn't handle this feature")
