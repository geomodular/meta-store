package utils

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

// ToProtoTimestamp converts given time.Time to timestamppb.Timestamp
func ToProtoTimestamp(ts time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: ts.Unix(), // This also does UTC conversion.
		Nanos:   int32(ts.Nanosecond()),
	}
}

// FromProtoTimestamp converts given timestamppb.Timestamp to time.Time
func FromProtoTimestamp(timestamp *timestamppb.Timestamp) time.Time {
	if timestamp == nil {
		return time.Time{}
	}
	return time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
}
