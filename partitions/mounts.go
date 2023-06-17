package partitions

import (
	"fmt"
	"os"
	"strings"
)

type MountOption [2]string

type MountPoint struct {
	Path       string
	Filesystem string
	Options    []MountOption
}

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
