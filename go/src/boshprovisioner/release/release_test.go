package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "boshprovisioner/release"
)

var _ = Describe("Release", func() {
	var (
		release Release
	)

	BeforeEach(func() {
		release = Release{}
	})

	Describe("ResolvedPackageDependencies", func() {
		It("returns packages in correct order with transitive deps", func() {
			pkg1 := Package{Name: "fake-package-name-1"}

			pkg2 := Package{
				Name:         "fake-package-name-2",
				Dependencies: []*Package{&pkg1},
			}

			pkg3 := Package{
				Name:         "fake-package-name-3",
				Dependencies: []*Package{&pkg2},
			}

			release.Packages = []*Package{&pkg3, &pkg2, &pkg1}

			pkgs := release.ResolvedPackageDependencies()
			Expect(pkgs).To(Equal([]*Package{&pkg1, &pkg2, &pkg3}))
		})

		It("returns packages in resolved order with multiple deps", func() {
			pkg1 := Package{Name: "fake-package-name-1"}
			pkg2 := Package{Name: "fake-package-name-2"}

			pkg3 := Package{
				Name:         "fake-package-name-3",
				Dependencies: []*Package{&pkg1, &pkg2},
			}

			release.Packages = []*Package{&pkg3, &pkg2, &pkg1}

			pkgs := release.ResolvedPackageDependencies()
			Expect(pkgs).To(Equal([]*Package{&pkg2, &pkg1, &pkg3})) // or pkg1, pkg2, pkg3
		})

		It("compiles BOSH release packages (example)", func() {
			nginx := Package{Name: "nginx"}
			genisoimage := Package{Name: "genisoimage"}
			powerdns := Package{Name: "powerdns"}
			ruby := Package{Name: "ruby"}

			blobstore := Package{
				Name:         "blobstore",
				Dependencies: []*Package{&ruby},
			}

			mysql := Package{Name: "mysql"}

			nats := Package{
				Name:         "nats",
				Dependencies: []*Package{&ruby},
			}

			common := Package{Name: "common"}
			redis := Package{Name: "redis"}
			libpq := Package{Name: "libpq"}
			postgres := Package{Name: "postgres"}

			registry := Package{
				Name:         "registry",
				Dependencies: []*Package{&libpq, &mysql, &ruby},
			}

			director := Package{
				Name:         "director",
				Dependencies: []*Package{&libpq, &mysql, &ruby},
			}

			healthMonitor := Package{
				Name:         "health_monitor",
				Dependencies: []*Package{&ruby},
			}

			release.Packages = []*Package{
				&nginx,
				&genisoimage,
				&powerdns,
				&blobstore, // before ruby
				&ruby,
				&mysql,
				&nats,
				&common,
				&director, // before libpq, postgres; after ruby
				&redis,
				&registry, // before libpq, postgres; after ruby
				&libpq,
				&postgres,
				&healthMonitor, // after ruby, libpq, postgres
			}

			pkgs := release.ResolvedPackageDependencies()

			Expect(pkgs).To(Equal([]*Package{
				&nginx,
				&genisoimage,
				&powerdns,
				&ruby,
				&mysql,
				&nats,
				&common,
				&redis,
				&libpq,
				&postgres,
				&healthMonitor, // currently always appended to the end
				&blobstore,
				&director,
				&registry,
			}))
		})
	})
})
