package deployment

import (
	gonet "net"

	boshaction "bosh/agent/action"

	bpdepman "boshprovisioner/deployment/manifest"
)

type Deployment struct {
	Manifest bpdepman.Manifest

	Name string

	Releases []Release

	Networks []Network

	Jobs []Job

	CompilationInstance Instance
}

type Release struct {
	Name    string
	Version string

	// Not offical BOSH manifest construct
	URL string
}

const (
	NetworkTypeManual  = bpdepman.NetworkTypeManual
	NetworkTypeDynamic = bpdepman.NetworkTypeDynamic
	NetworkTypeVip     = bpdepman.NetworkTypeVip
)

var NetworkTypes = []string{NetworkTypeManual, NetworkTypeDynamic, NetworkTypeVip}

type Network struct {
	Name string
	Type string
}

type Job struct {
	Name string

	Templates []Template

	Instances []Instance
}

type Template struct {
	Name string

	Release *Release
}

type Instance struct {
	Index int

	// Denormalized values to avoid passing dep/job/instance tuple
	JobName        string
	DeploymentName string

	// Watch time will vary depending if an instance is a canary
	WatchTime bpdepman.WatchTime

	Properties Properties

	NetworkAssociations []NetworkAssociation

	// Represents current state of an associated VM
	CurrentState boshaction.GetStateV1ApplySpec
}

type Properties map[string]interface{}

type NetworkAssociation struct {
	Network *Network

	StaticIP gonet.IP

	// StaticIP might equal to nil, though that does not indicate
	// that this instance does not need a static IP
	MustHaveStaticIP bool
}

// populateFromManifest populates deployment information
// interpreted from deployment manifest.
func (d *Deployment) populateFromManifest(manifest bpdepman.Manifest) {
	d.populateNetworks(manifest)
	d.populateReleases(manifest)
	d.populateCompilationInstances(manifest)
	d.populateJobs(manifest)
	d.Manifest = manifest
}

func (d *Deployment) populateNetworks(manifest bpdepman.Manifest) {
	for _, manNet := range manifest.Deployment.Networks {
		d.Networks = append(d.Networks, Network{
			Name: manNet.Name,
			Type: manNet.Type,
		})
	}
}

func (d *Deployment) populateReleases(manifest bpdepman.Manifest) {
	for _, manRelease := range manifest.Deployment.Releases {
		d.Releases = append(d.Releases, Release{
			Name:    manRelease.Name,
			Version: manRelease.Version,
			URL:     manRelease.URL,
		})
	}
}

func (d *Deployment) populateCompilationInstances(manifest bpdepman.Manifest) {
	network := d.findNetworkOrDefault(manifest.Deployment.Compilation.NetworkName)

	d.CompilationInstance = Instance{
		Index: 0,

		JobName:        "compilation",
		DeploymentName: manifest.Deployment.Name,

		NetworkAssociations: []NetworkAssociation{
			NetworkAssociation{Network: network},
		},
	}
}

func (d *Deployment) populateJobs(manifest bpdepman.Manifest) {
	for _, manJob := range manifest.Deployment.Jobs {
		d.Jobs = append(d.Jobs, d.buildJob(manifest.Deployment, manJob))
	}
}

func (d *Deployment) buildJob(manDep bpdepman.Deployment, manJob bpdepman.Job) Job {
	job := Job{Name: manJob.Name}

	for i := 0; i < manJob.Instances; i++ {
		watchTime := manDep.InstanceWatchTime(manJob, i)
		properties := manDep.InstanceProperties(manJob, i)
		netAssocs := d.buildNetworkAssociations(manJob, i)

		job.Instances = append(job.Instances, Instance{
			Index: i,

			JobName:        manJob.Name,
			DeploymentName: manDep.Name,

			WatchTime:  watchTime,
			Properties: Properties(properties),

			NetworkAssociations: netAssocs,
		})
	}

	// todo check to make sure release is present
	for _, manTemplate := range manJob.Templates {
		release := d.findReleaseOrDefault(manTemplate.ReleaseName)

		job.Templates = append(job.Templates, Template{
			Name:    manTemplate.Name,
			Release: release,
		})
	}

	return job
}

func (d *Deployment) buildNetworkAssociations(manJob bpdepman.Job, i int) []NetworkAssociation {
	var netAssocs []NetworkAssociation

	for _, manNa := range manJob.NetworkAssociations {
		var staticIP gonet.IP

		network := d.findNetworkOrDefault(manNa.NetworkName)

		if i < len(manNa.StaticIPs) {
			staticIP = manNa.StaticIPs[i]
		}

		netAssocs = append(netAssocs, NetworkAssociation{
			Network:  network,
			StaticIP: staticIP,

			MustHaveStaticIP: len(manNa.StaticIPs) > 0,
		})
	}

	return netAssocs
}

func (d *Deployment) findReleaseOrDefault(name string) *Release {
	// todo check against empty release name
	for _, release := range d.Releases {
		if release.Name == name {
			return &release
		}
	}

	// Assume if there is only one release it's default
	if len(d.Releases) == 1 {
		return &d.Releases[0]
	}

	return nil
}

func (d *Deployment) findNetworkOrDefault(name string) *Network {
	for _, net := range d.Networks {
		if net.Name == name {
			return &net
		}
	}

	return nil
}
