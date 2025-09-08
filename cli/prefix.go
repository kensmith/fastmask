package main

import "crypto/rand"

const (
	_prefixLen = 5
)

var (
	_lexicon    = []byte("bcdfghjkmnpqrstvwxyz")
	_numLexemes = len(_lexicon)
)

func GenPrefix() string {
	randomBytes := make([]byte, _prefixLen)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// the docs say this should never happen so it's acceptable to panic in
		// this case rather than return an error to the caller
		// https://pkg.go.dev/crypto/rand#Read
		panic("failed to generate prefix")
	}
	for i := range _prefixLen {
		randByte := randomBytes[i]
		lexeme := _lexicon[int(randByte)%_numLexemes]
		randomBytes[i] = lexeme
	}
	return string(randomBytes[:])
}
