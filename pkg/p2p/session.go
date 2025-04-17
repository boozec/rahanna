package p2p

import (
	"math/rand"
)

var adjectives = []string{
	"adamant", "adept", "adventurous", "arcadian", "auspicious",
	"awesome", "blossoming", "brave", "charming", "chatty",
	"circular", "considerate", "cubic", "curious", "delighted",
}

var nouns = []string{
	"aardvark", "accordion", "apple", "apricot", "bee",
	"brachiosaur", "cactus", "capsicum", "clarinet", "cowbell",
	"crab", "cuckoo", "cymbal", "diplodocus", "donkey",
}

func NewSession() string {
	noun := nouns[rand.Intn(len(nouns))]
	adjective := adjectives[rand.Intn(len(adjectives))]
	return noun + "-" + adjective
}
