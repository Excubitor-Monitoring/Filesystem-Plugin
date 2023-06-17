package partitions

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type Partition struct {
	Name           string       `json:"name"`
	Size           uint64       `json:"size"`
	Used           uint64       `json:"used"`
	FilesystemType string       `json:"type"`
	Filesystem     string       `json:"filesystem"`
	MountPoints    []MountPoint `json:"mount_points"`
}

func getPartitions(blockDeviceName string) ([]Partition, error) {
	var partitions []Partition

	blockDeviceFolder, err := os.ReadDir("/sys/block/" + blockDeviceName)
	if err != nil {
		return nil, fmt.Errorf("error on reading block device stats of %s: %w", blockDeviceName, err)
	}

	for _, content := range blockDeviceFolder {
		if strings.HasPrefix(content.Name(), blockDeviceName) {
			var partition Partition
			partition.Name = content.Name()

			partitionFolderFS := os.DirFS("/sys/block/" + blockDeviceName + "/" + partition.Name)

			partitionSizeFile, err := fs.ReadFile(partitionFolderFS, "size")
			if err != nil {
				return nil, fmt.Errorf("error on reading size of partition %s: %w", blockDeviceName, err)
			}

			partition.Size, err = strconv.ParseUint(string(partitionSizeFile)[:len(partitionSizeFile)-1], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("error on parsing size of partition %s: %w", blockDeviceName, err)
			}

			var mountPoints []MountPoint

			holdersFolder, err := os.ReadDir("/sys/block/" + blockDeviceName + "/" + partition.Name + "/holders")
			if len(holdersFolder) != 0 {
				for _, holder := range holdersFolder {
					if strings.HasPrefix(holder.Name(), "dm") {
						dmName, err := os.ReadFile("/sys/block/" + blockDeviceName + "/" + partition.Name + "/holders/" + holder.Name() + "/dm/name")
						if err != nil {
							return nil, fmt.Errorf("error when reading name of device mapper partition name of %s: %w", partition.Name, err)
						}

						mountPoints, err = getMountPoints("mapper/" + string(dmName)[:len(dmName)-1])
						if err != nil {
							return nil, fmt.Errorf("error on getting mount points for device mapper volume %s of partition %s: %w", dmName, partition.Name, err)
						}
					}
				}
			} else {
				mountPoints, err = getMountPoints(partition.Name)
				if err != nil {
					return nil, fmt.Errorf("error on getting mount points for partition %s: %w", partition.Name, err)
				}
			}

			if len(mountPoints) > 0 {
				stat := syscall.Statfs_t{}

				if err := syscall.Statfs(mountPoints[0].Path, &stat); err != nil {
					return nil, fmt.Errorf("error on statfs syscall using path %s on partition %s: %w", mountPoints[0].Path, partition.Name, err)
				}

				if stat.Bsize < 0 {
					return nil, errors.New("negative block size")
				}

				partition.Used = partition.Size - (stat.Bfree * uint64(stat.Bsize) / 1024)
				partition.Filesystem = mountPoints[0].Filesystem
				partition.FilesystemType = FSType(stat.Type).Name()

				for _, mp := range mountPoints {
					partition.MountPoints = append(partition.MountPoints, mp)
				}

			}

			partitions = append(partitions, partition)
		}
	}

	return partitions, nil
}
