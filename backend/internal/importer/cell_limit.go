package importer

import "unicode/utf16"

func cellTextLength(text string) int {
	length := 0
	for _, r := range text {
		length += utf16.RuneLen(r)
	}
	return length
}
