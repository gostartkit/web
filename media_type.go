package web

import (
	"strconv"
	"strings"
)

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
	if header == "" {
		return mediaJSON
	}

	best := mediaUnknown
	bestQuality := -1.0
	for {
		part := header
		if comma := strings.IndexByte(header, ','); comma >= 0 {
			part = header[:comma]
			header = header[comma+1:]
		} else {
			header = ""
		}

		mt := parseMediaType(part)
		if mt != mediaUnknown {
			quality := mediaRangeQuality(part)
			if quality > bestQuality {
				best = mt
				bestQuality = quality
			}
		}
		if header == "" {
			break
		}
	}
	if best == mediaUnknown || bestQuality <= 0 {
		return mediaJSON
	}
	return best
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

	// Keep the common lowercase parameterized forms on a short, allocation-free
	// path while requiring a real media-type delimiter (so JSONP is not JSON).
	switch {
	case hasMediaTypePrefix(header, "application/json"), hasMediaTypePrefix(header, "*/*"):
		return mediaJSON
	case hasMediaTypePrefix(header, "application/x-gob"):
		return mediaGOB
	case hasMediaTypePrefix(header, "application/octet-stream"):
		return mediaOctetStream
	case hasMediaTypePrefix(header, "application/x-avro"):
		return mediaAvro
	case hasMediaTypePrefix(header, "application/xml"), hasMediaTypePrefix(header, "text/xml"):
		return mediaXML
	}

	end := len(header)
	if i := strings.IndexAny(header, ";,"); i >= 0 {
		end = i
	}
	token := strings.TrimSpace(header[:end])
	switch token {
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

	switch {
	case strings.EqualFold(token, "application/json"), token == "*/*":
		return mediaJSON
	case strings.EqualFold(token, "application/x-gob"):
		return mediaGOB
	case strings.EqualFold(token, "application/octet-stream"):
		return mediaOctetStream
	case strings.EqualFold(token, "application/x-avro"):
		return mediaAvro
	case strings.EqualFold(token, "application/xml"), strings.EqualFold(token, "text/xml"):
		return mediaXML
	default:
		return mediaUnknown
	}
}

func hasMediaTypePrefix(header, media string) bool {
	if len(header) <= len(media) || !strings.HasPrefix(header, media) {
		return false
	}
	switch header[len(media)] {
	case ';', ',', ' ', '\t':
		return true
	default:
		return false
	}
}

func mediaRangeQuality(value string) float64 {
	semicolon := strings.IndexByte(value, ';')
	if semicolon < 0 {
		return 1
	}
	params := value[semicolon+1:]
	for params != "" {
		param := params
		if next := strings.IndexByte(params, ';'); next >= 0 {
			param = params[:next]
			params = params[next+1:]
		} else {
			params = ""
		}
		param = strings.TrimSpace(param)
		if len(param) < 2 || (param[0] != 'q' && param[0] != 'Q') || param[1] != '=' {
			continue
		}
		q, err := strconv.ParseFloat(strings.TrimSpace(param[2:]), 64)
		if err != nil || q < 0 || q > 1 {
			return 0
		}
		return q
	}
	return 1
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
