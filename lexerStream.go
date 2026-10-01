package govaluate

import (
	"sync"
	"unicode/utf8"
)

type lexerStream struct {
	sourceString string
	source       []rune
	strPosition  int
	position     int
	length       int
}

var lexerStreamPool = sync.Pool{
	New: func() interface{} {
		return new(lexerStream)
	},
}

func newLexerStream(source string) *lexerStream {
	ret := lexerStreamPool.Get().(*lexerStream)
	if ret.source == nil {
		ret.source = make([]rune, 0, len(source))
	}
	for _, character := range source {
		ret.source = append(ret.source, character)
	}
	ret.sourceString = source
	ret.position = 0
	ret.strPosition = 0
	ret.length = len(ret.source)
	return ret
}

func (stream *lexerStream) readCharacter() rune {
	character := stream.source[stream.position]
	stream.position += 1
	// Advance strPosition by the number of bytes this rune actually occupies in
	// the original source string. utf8.RuneLen cannot be used here: ranging over
	// invalid UTF-8 yields utf8.RuneError (3 bytes) for a single source byte, so
	// RuneLen would overcount and push strPosition past len(sourceString).
	_, size := utf8.DecodeRuneInString(stream.sourceString[stream.strPosition:])
	stream.strPosition += size
	return character
}

func (stream *lexerStream) rewind(amount int) {
	if amount < 0 {
		// Advance forward by -amount runes (used as rewind(-1)).
		for i := 0; i < -amount; i++ {
			if stream.position >= stream.length {
				break
			}
			_, size := utf8.DecodeRuneInString(stream.sourceString[stream.strPosition:])
			stream.strPosition += size
			stream.position += 1
		}
		return
	}
	for i := 0; i < amount; i++ {
		if stream.position <= 0 {
			break
		}
		stream.position -= 1
		_, size := utf8.DecodeLastRuneInString(stream.sourceString[:stream.strPosition])
		stream.strPosition -= size
	}
}

func (stream lexerStream) canRead() bool {
	return stream.position < stream.length
}

func (stream *lexerStream) close() {
	stream.source = stream.source[:0]
	lexerStreamPool.Put(stream)
}
