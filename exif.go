package main

import (
	"bytes"
	"fmt"

	"github.com/rwcarlsen/goexif/exif"
)

func extractEXIF(data []byte) map[string]string {
	result := make(map[string]string)
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return result
	}

	fields := []exif.FieldName{
		exif.Make, exif.Model, exif.DateTime,
		exif.GPSLatitude, exif.GPSLongitude,
		exif.GPSAltitude, exif.FocalLength,
		exif.ExposureTime, exif.FNumber,
		exif.ISOSpeedRatings,
	}

	for _, f := range fields {
		tag, err := x.Get(f)
		if err != nil {
			continue
		}
		result[string(f)] = tag.String()
	}

	// Decode GPS coords to decimal degrees
	if lat, lon, err := x.LatLong(); err == nil {
		result["GPSDecimal"] = fmt.Sprintf("%.6f, %.6f", lat, lon)
	}

	return result
}

func exifSummary(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteString("\n\n[Image metadata]")
	for k, v := range m {
		fmt.Fprintf(&buf, "\n%s: %s", k, v)
	}
	return buf.String()
}
