package p2p

import (
	"math/rand"
)

var adjectives = []string{
	"adamant", "adept", "adventurous", "arcadian", "auspicious",
	"awesome", "blossoming", "bold", "brave", "bubbly",
	"charming", "chatty", "circular", "colorful", "considerate",
	"cosmic", "cuddly", "curious", "dazzling", "delighted",
	"dynamic", "eager", "ecstatic", "effervescent", "elated",
	"elegant", "elvish", "fanciful", "fearless", "festive",
	"fluffy", "forceful", "glorious", "goofy", "graceful",
	"gutsy", "happy", "hobbity", "jedi", "mystical",
	"potteresque", "rebel", "slytherin", "valiant", "witty",
}

var nouns = []string{
	"aardvark", "accordion", "apple", "apricot", "asteroid",
	"bantha", "bee", "bison", "brachiosaur", "bubble",
	"cactus", "capsicum", "clarinet", "cloud", "cowbell",
	"crab", "cuckoo", "cupcake", "cymbal", "diplodocus",
	"donkey", "dragon", "drum", "dobby", "elf",
	"eel", "emu", "ent", "falcon", "fern",
	"flamingo", "giraffe", "glacier", "goose", "grapefruit",
	"gryffindor", "hamster", "harmonica", "hedgehog", "honeybee",
	"hufflepuff", "hummingbird", "lightsaber", "mandalorian", "mushroom",
	"orchestra", "orcrist", "phoenix", "ringwraith", "snitch",
	"stormtrooper", "tatooine", "wand", "wookiee", "yoda",
}

func NewSession() string {
	noun := nouns[rand.Intn(len(nouns))]
	adjective := adjectives[rand.Intn(len(adjectives))]
	return noun + "-" + adjective
}
