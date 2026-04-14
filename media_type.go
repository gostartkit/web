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

	// Accept may include multiple values and parameters; keep only the first media range.
	if i := strings.IndexByte(header, ','); i >= 0 {
		header = header[:i]
	}
	if i := strings.IndexByte(header, ';'); i >= 0 {
		header = header[:i]
	}
	header = strings.TrimSpace(header)

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
