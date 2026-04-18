package web

import "strings"

type mediaType uint8

const (
	mediaUnknown mediaType = iota
	mediaJSON
	mediaGOB
	mediaOctetStream
	mediaAvro
	mediaXML
)

const mediaTypeSlots = int(mediaXML) + 1

func acceptMediaType(header string) mediaType {
	mt := parseMediaType(header)
	if mt == mediaUnknown {
		return mediaJSON
	}
	return mt
}

func parseMediaType(header string) mediaType {
	if header == "" {
		return mediaUnknown
	}

	switch header {
	case "application/json", "*/*":
		return mediaJSON
	case "application/x-gob":
		return mediaGOB
	case "application/octet-stream":
		return mediaOctetStream
	case "application/x-avro":
		return mediaAvro
	case "application/xml", "text/xml":
		return mediaXML
	}

	// Fast prefix path for values with parameters or media-ranges, e.g.
	// "application/json; charset=utf-8" or "application/json, */*".
	switch {
	case strings.HasPrefix(header, "application/json"):
		return mediaJSON
	case strings.HasPrefix(header, "application/x-gob"):
		return mediaGOB
	case strings.HasPrefix(header, "application/octet-stream"):
		return mediaOctetStream
	case strings.HasPrefix(header, "application/x-avro"):
		return mediaAvro
	case strings.HasPrefix(header, "application/xml"), strings.HasPrefix(header, "text/xml"):
		return mediaXML
	default:
		return mediaUnknown
	}
}

func contentTypeForMedia(mt mediaType) string {
	switch mt {
	case mediaGOB:
		return "application/x-gob"
	case mediaOctetStream:
		return "application/octet-stream"
	case mediaAvro:
		return "application/x-avro"
	case mediaXML:
		return "application/xml"
	default:
		return "application/json"
	}
}
