package releasesrepo

import (
	bpdep "boshprovisioner/deployment"
)

// ReleasesRepository manages collection of releases
// available to use for provisioning
type ReleasesRepository interface {
	// Pull downloads/copies/retrieves a release
	Pull(bpdep.Release) error

	// KeepOnly deletes all releases but the provided ones
	KeepOnly([]bpdep.Release) error
}
