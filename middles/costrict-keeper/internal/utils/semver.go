package utils

import (
	"errors"

	"github.com/Masterminds/semver/v3"
)

// ErrVersionNotAllowed indicates that a version is not in the allowed range
var ErrVersionNotAllowed = errors.New("version not allowed")

/**
 * Check if a version is within the allowed version range
 * @param {string} version - The version string to check (e.g., "1.2.3")
 * @param {string} allowedVersions - The version constraint string (e.g., ">=1.0.0,<2.0.0")
 * @returns {error} Returns ErrVersionNotAllowed if version doesn't match constraints, nil otherwise
 * @description
 * - Parses version using semver.Parse()
 * - Parses constraint using semver.NewConstraint()
 * - Validates version against constraint using Constraints.Check()
 * @throws
 * - Version parse error (semver.Parse)
 * - Constraint parse error (semver.NewConstraint)
 * @example
 * err := IsVersionAllowed("1.5.0", ">=1.0.0,<2.0.0")
 * if err != nil { log.Fatal(err) }
 */
func IsVersionAllowed(version string, allowedVersions string) error {
	constraints, err := semver.NewConstraint(allowedVersions)
	if err != nil {
		return err
	}

	versionParsed, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	if !constraints.Check(versionParsed) {
		return ErrVersionNotAllowed
	}

	return nil
}
