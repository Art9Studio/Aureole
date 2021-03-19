package storage

import "fmt"

func CheckFeaturesAvailable(requiredFeatures []string, implementedFeatures map[string]bool) error {
	for _, feature := range requiredFeatures {
		if isImplemented, ok := implementedFeatures[feature]; !ok || !isImplemented {
			return fmt.Errorf("feature %s hasn't implemented", feature)
		}
	}

	return nil
}
