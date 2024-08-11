package config

import "slices"

// Returns copy of environment
//
// Free to modify or delete, actual env configuration will be untoched
func Enivronment() []string {
	return slices.Clone(environment[:])
}

func Flags() []flagSetter {
	return slices.Clone(flags[:])
}

func Else() []elseSetter {
	return slices.Clone(elses[:])
}

func Override() []overrideContainer {
	return slices.Clone(override[:])
}
