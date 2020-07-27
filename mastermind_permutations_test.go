package main

import (
	"fmt"
	"testing"
)

type possibleCodes []Code

func (Codes possibleCodes) contains(c Code) bool {
	for _, co := range Codes {
		if codesEqual(co, c) {
			return true
		}
	}
	return false
}

func TestSimpleGeneration(t *testing.T) {
	colors := []CodePeg{red, orange}

	generatedCodes := possibleCodes(generateAllPossibleCodes(colors, 4))
	expectedCodes := possibleCodes{
		Code{red, red, red, red},
		Code{orange, red, red, red},
		Code{red, orange, red, red},
		Code{orange, orange, red, red},
		Code{red, red, orange, red},
		Code{orange, red, orange, red},
		Code{red, orange, orange, red},
		Code{orange, orange, orange, red},
		Code{red, red, red, orange},
		Code{orange, red, red, orange},
		Code{red, orange, red, orange},
		Code{orange, orange, red, orange},
		Code{red, red, orange, orange},
		Code{orange, red, orange, orange},
		Code{red, orange, orange, orange},
		Code{orange, orange, orange, orange},
	}

	if len(generatedCodes) != len(expectedCodes) {
		t.Fail()
		return
	}
	for _, c1 := range expectedCodes {
		if !generatedCodes.contains(c1) {
			fmt.Println(c1)
			t.Fail()
			return
		}
	}
}

func TestGeneration2(t *testing.T) {
	colors := []CodePeg{red, orange, yellow}

	generatedCodes := possibleCodes(generateAllPossibleCodes(colors, 2))
	expectedCodes := possibleCodes{
		Code{red, red},
		Code{orange, red},
		Code{yellow, red},
		Code{red, orange},
		Code{orange, orange},
		Code{yellow, orange},
		Code{red, yellow},
		Code{orange, yellow},
		Code{yellow, yellow},
	}

	if len(generatedCodes) != len(expectedCodes) {
		t.Fail()
		return
	}
	for _, c1 := range expectedCodes {
		if !generatedCodes.contains(c1) {
			fmt.Println(c1)
			t.Fail()
			return
		}
	}
}

func TestGeneration3(t *testing.T) {
	colors := []CodePeg{red, orange, yellow, green, blue, purple}

	generatedCodes := possibleCodes(generateAllPossibleCodes(colors, 1))
	expectedCodes := possibleCodes{
		Code{red},
		Code{orange},
		Code{yellow},
		Code{green},
		Code{blue},
		Code{purple},
	}

	if len(generatedCodes) != len(expectedCodes) {
		t.Fail()
		return
	}
	for _, c1 := range expectedCodes {
		if !generatedCodes.contains(c1) {
			fmt.Println(c1)
			t.Fail()
			return
		}
	}
}
