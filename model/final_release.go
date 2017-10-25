package model

import (
	"fmt"
	"path/filepath"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// NewFinalRelease will create an instance of a BOSH final release
func NewFinalRelease(path, releaseName, version, boshCacheDir string) (*Release, error) {
	release := &Release{
		Path:            path,
		Name:            releaseName,
		Version:         version,
		DevBOSHCacheDir: boshCacheDir,
	}

	if releaseName == "" {
		releaseName, err := release.getFinalReleaseName()
		if err != nil {
			return nil, err
		}

		release.Name = releaseName
	}

	if version == "" {
		version, err := release.getFinalReleaseVersion()
		if err != nil {
			return nil, err
		}

		release.Version = version
	}

	if err := release.loadFinalReleaseMetadata(); err != nil {
		return nil, err
	}

	if err := release.loadPackages(); err != nil {
		return nil, err
	}

	if err := release.loadDependenciesForPackages(); err != nil {
		return nil, err
	}

	if err := release.loadJobs(); err != nil {
		return nil, err
	}

	if err := release.loadLicense(); err != nil {
		return nil, err
	}

	return release, nil
}

func (r *Release) getFinalReleaseName() (ver string, err error) {
	var releaseConfig map[interface{}]interface{}
	var name string

	releaseConfigContent, err := ioutil.ReadFile(r.getReleaseConfigFile())
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal([]byte(releaseConfigContent), &releaseConfig); err != nil {
		return "", err
	}

	if value, ok := releaseConfig["name"]; !ok {
			return "", fmt.Errorf("name not exists in the release.MF file for release: %s", r.Path)
	} else if name, ok = value.(string); !ok {
		return "", fmt.Errorf("name was not a string in release: %s, type: %T, value: %v", r.Path, value, value)
	}

	return name, nil
}

func (r *Release) getFinalReleaseVersion() (ver string, err error) {
	var releaseConfig map[interface{}]interface{}
	var version string

	releaseConfigContent, err := ioutil.ReadFile(r.getReleaseConfigFile())
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal([]byte(releaseConfigContent), &releaseConfig); err != nil {
		return "", err
	}

	if value, ok := releaseConfig["version"]; !ok {
		return "", fmt.Errorf("version not exists in the release.MF file for release: %s", r.Path)
	} else if version, ok = value.(string); !ok {
		return "", fmt.Errorf("version was not a string in release: %s, type: %T, value: %v", r.Path, value, value)
	}

	return version, nil
}

func (r *Release) getReleaseConfigFile() string {
	return filepath.Join(r.Path, "release.MF")
}
