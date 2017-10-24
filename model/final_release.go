package model

import (
	"fmt"
	//"io/ioutil"
	"path/filepath"

	//"github.com/SUSE/fissile/util"

	//"github.com/cppforlife/go-semi-semantic/version"
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

//func (r *Release) validateDevPathStructure() error {
//	if err := util.ValidatePath(r.Path, true, "release directory"); err != nil {
//		return err
//	}
//
//	if err := util.ValidatePath(r.getDevReleasesDir(), true, "release 'dev_releases' directory"); err != nil {
//		return err
//	}
//
//	if err := util.ValidatePath(r.getDevReleaseConfigDir(), true, "release config directory"); err != nil {
//		return err
//	}
//
//	return util.ValidatePath(r.getDevReleaseFinalConfigFile(), false, "release final config file")
//}
//
//func (r *Release) validateSpecificDevReleasePathStructure() error {
//	if err := util.ValidatePath(r.getDevReleaseManifestsDir(), true, "release dev manifests directory"); err != nil {
//		return err
//	}
//
//	return util.ValidatePath(r.getDevReleaseIndexPath(), false, "release index file")
//}
//
//func (r *Release) getDevReleaseManifestFilename() string {
//	return fmt.Sprintf("%s-%s.yml", r.Name, r.Version)
//}
//
//func (r *Release) getDevReleaseManifestsDir() string {
//	return filepath.Join(r.getDevReleasesDir(), r.Name)
//}
//
//func (r *Release) getDevReleaseIndexPath() string {
//	return filepath.Join(r.getDevReleaseManifestsDir(), "index.yml")
//}
//
//func (r *Release) getDevReleasesDir() string {
//	return filepath.Join(r.Path, "dev_releases")
//}
//
//func (r *Release) getDevReleaseConfigDir() string {
//	return filepath.Join(r.Path, "config")
//}

//func (r *Release) getDevReleaseFinalConfigFile() string {
//	return filepath.Join(r.getDevReleaseConfigDir(), "final.yml")
//}
//
func (r *Release) getReleaseConfigFile() string {
	return filepath.Join(r.Path, "release.MF")
}
