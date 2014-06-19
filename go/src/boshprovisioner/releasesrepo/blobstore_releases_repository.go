package releasesrepo

import (
	boshblob "bosh/blobstore"
	bosherr "bosh/errors"
	boshlog "bosh/logger"

	bpdep "boshprovisioner/deployment"
	bpdload "boshprovisioner/downloader"
	bpindex "boshprovisioner/index"
)

type BlobstoreReleasesRepository struct {
	downloader bpdload.Downloader
	blobstore  boshblob.Blobstore
	index      bpindex.Index
	logger     boshlog.Logger
}

type blobstoreRecord struct {
	BlobID      string
	Fingerprint string
}

func NewBlobstoreReleasesRepository(
	downloader bpdload.Downloader,
	blobstore boshblob.Blobstore,
	index bpindex.Index,
	logger boshlog.Logger,
) BlobstoreReleasesRepository {
	return BlobstoreReleasesRepository{
		downloader: downloader,
		blobstore:  blobstore,
		index:      index,
		logger:     logger,
	}
}

func (rr BlobstoreReleasesRepository) Pull(release bpdep.Release) error {
	var record blobstoreRecord

	err := rr.index.Find(release, &record)
	if err == nil {
		return nil // nothing to do if already exists
	} else if err != bpindex.ErrNotFound {
		return bosherr.WrapError(err, "Finding release in index")
	}

	path, err := rr.downloader.Download(release.URL)
	if err != nil {
		return bosherr.WrapError(err, "Downloading release")
	}

	defer rr.downloader.CleanUp(path)

	blobID, fingerprint, err := rr.blobstore.Create(path)
	if err != nil {
		return bosherr.WrapError(err, "Creating release blob")
	}

	record = blobstoreRecord{
		BlobID:      blobID,
		Fingerprint: fingerprint,
	}

	err = rr.index.Save(release, record)
	if err != nil {
		// todo delete from blobstore

		return bosherr.WrapError(err, "Saving release to index")
	}

	return nil
}

func (rr BlobstoreReleasesRepository) KeepOnly(releasesToKeep []bpdep.Release) error {
	var allReleases []bpdep.Release

	err := rr.index.ListKeys(&allReleases)
	if err != nil {
		return bosherr.WrapError(err, "Listing releases in index")
	}

	for _, foundRelease := range allReleases {
		var keep bool

		for _, releaseToKeep := range releasesToKeep {
			if foundRelease == releaseToKeep {
				keep = true
				break
			}
		}

		if keep {
			continue
		}

		var record blobstoreRecord

		err := rr.index.Find(foundRelease, &record)
		if err != nil {
			return bosherr.WrapError(err, "Finding release to delete")
		}

		// todo delete from blobstore

		err = rr.index.Remove(foundRelease)
		if err != nil {
			return bosherr.WrapError(err, "Deleting release from index")
		}
	}

	return nil
}
