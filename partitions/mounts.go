package partitions

import (
	"fmt"
	"os"
	"strings"
)

// MountOption is a slice of exactly two strings. The first string is the key of the mount option, the second string
// is the value. If there is no value for a given MountOption, the second string is just an empty string.
type MountOption [2]string

// MountPoint contains parameters typically present in unix mount points.
// Those are Path, the path in which the device is mounted, Filesystem, a string representation of the filesystem
// present in the device and Options, a slice of MountOption
type MountPoint struct {
	Path       string
	Filesystem string
	Options    []MountOption
}

// getMountPoints gathers information about mount points of a certain partition defined the argument partitionName.
func getMountPoints(partitionName string) ([]MountPoint, error) {
	var mountPoints []MountPoint

	mounts, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("error on reading mounts: %w", err)
	}

	mountsLines := strings.Split(string(mounts), "\n")
	for _, line := range mountsLines {
		columns := strings.Split(line, " ")
		if columns[0] == "/dev/"+partitionName {
			mountPoint := MountPoint{
				Path:       columns[1],
				Filesystem: columns[2],
				Options:    []MountOption{},
			}

			options := strings.Split(columns[3], ",")

			for _, option := range options {
				if strings.Contains(option, "=") {
					optionKeyValue := strings.Split(option, "=")
					mountPoint.Options = append(mountPoint.Options, MountOption{optionKeyValue[0], optionKeyValue[1]})
				} else {
					mountPoint.Options = append(mountPoint.Options, MountOption{option, ""})
				}
			}

			mountPoints = append(mountPoints, mountPoint)
		}
	}

	return mountPoints, nil
}
