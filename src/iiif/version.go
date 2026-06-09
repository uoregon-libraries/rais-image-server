package iiif

// Version represents which IIIF Image API specification a request or response
// adheres to.  RAIS can serve more than one version simultaneously (on separate
// endpoints), so most parsing and info.json generation has to be told which
// version it's operating under.
type Version int

const (
	// VersionUnknown is the zero value, used when a version hasn't been set
	VersionUnknown Version = 0
	// V2 is the IIIF Image API 2.1 specification
	V2 Version = 2
	// V3 is the IIIF Image API 3.0 specification
	V3 Version = 3
)

// String returns a human-readable label for the version
func (v Version) String() string {
	switch v {
	case V2:
		return "2"
	case V3:
		return "3"
	default:
		return "unknown"
	}
}

// Valid returns whether the version is one RAIS knows how to serve
func (v Version) Valid() bool {
	return v == V2 || v == V3
}

// ContextURI returns the JSON-LD @context value for the version's info.json
func (v Version) ContextURI() string {
	switch v {
	case V3:
		return "http://iiif.io/api/image/3/context.json"
	default:
		return "http://iiif.io/api/image/2/context.json"
	}
}

// conformanceLabel returns the value used for a compliance level in the info
// response.  In v2 this is a full URI (e.g.,
// "http://iiif.io/api/image/2/level2.json"); in v3 it's a short string (e.g.,
// "level2").  The level argument is 0, 1, or 2.
func (v Version) conformanceLabel(level int) string {
	switch v {
	case V3:
		switch level {
		case 2:
			return "level2"
		case 1:
			return "level1"
		default:
			return "level0"
		}
	default:
		switch level {
		case 2:
			return "http://iiif.io/api/image/2/level2.json"
		case 1:
			return "http://iiif.io/api/image/2/level1.json"
		default:
			return "http://iiif.io/api/image/2/level0.json"
		}
	}
}
