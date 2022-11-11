package proto

import "strings"

// ToProtoETag converts an etag in a simple string form (_rev by default in ArangoDB) to RFC 7232 compliant etag.
// TODO: Weakness yet to be determined.
func ToProtoETag(etag string) string {
	return "\"" + etag + "\""
}

// FromProtoETag converts RFC 7232 compliant etag to a simple string form.
func FromProtoETag(etag string) string {
	ret := strings.TrimPrefix(etag, "W/")
	ret = strings.TrimPrefix(etag, "\"")
	ret = strings.TrimSuffix(etag, "\"")
	return ret
}
