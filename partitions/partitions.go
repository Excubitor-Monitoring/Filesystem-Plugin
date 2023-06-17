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

type BlockDevice struct {
	Name       string      `json:"name"`
	Size       uint64      `json:"size"`
	Partitions []Partition `json:"partitions"`
}

type Partition struct {
	Name           string   `json:"name"`
	Size           uint64   `json:"size"`
	Used           uint64   `json:"used"`
	FilesystemType string   `json:"type"`
	Filesystem     string   `json:"filesystem"`
	MountPoints    []string `json:"mount_points"`
}

func ParseBlockDevices() ([]BlockDevice, error) {
	blockDevices := make([]BlockDevice, 0)

	blockDevicesFolder, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("error on reading block devices: %w", err)
	}

	for _, block := range blockDevicesFolder {
		blockDeviceFolderFS := os.DirFS("/sys/block/" + block.Name())

		sizeFile, err := fs.ReadFile(blockDeviceFolderFS, "size")
		if err != nil {
			return nil, fmt.Errorf("error on reading size of block device %s: %w", block.Name(), err)
		}

		size, err := strconv.ParseUint(string(sizeFile)[:len(sizeFile)-1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("error on parsing size of block device %s: %w", block.Name(), err)
		}

		blockDeviceFolder, err := os.ReadDir("/sys/block/" + block.Name())
		if err != nil {
			return nil, fmt.Errorf("error on reading block device stats of %s: %w", block.Name(), err)
		}

		partitions := make([]Partition, 0)

		for _, content := range blockDeviceFolder {
			if strings.HasPrefix(content.Name(), block.Name()) {
				var partition Partition
				partition.Name = content.Name()

				partitionFolderFS := os.DirFS("/sys/block/" + block.Name() + "/" + content.Name())

				partitionSizeFile, err := fs.ReadFile(partitionFolderFS, "size")
				if err != nil {
					return nil, fmt.Errorf("error on reading size of partition %s: %w", block.Name(), err)
				}

				partition.Size, err = strconv.ParseUint(string(partitionSizeFile)[:len(partitionSizeFile)-1], 10, 32)
				if err != nil {
					return nil, fmt.Errorf("error on parsing size of partition %s: %w", block.Name(), err)
				}

				var mountPoints []MountPoint

				holdersFolder, err := os.ReadDir("/sys/block/" + block.Name() + "/" + content.Name() + "/holders")
				if len(holdersFolder) != 0 {
					for _, holder := range holdersFolder {
						if strings.HasPrefix(holder.Name(), "dm") {
							dmName, err := os.ReadFile("/sys/block/" + block.Name() + "/" + content.Name() + "/holders/" + holder.Name() + "/dm/name")
							if err != nil {
								return nil, fmt.Errorf("error when reading name of device mapper partition name of %s: %w", content.Name(), err)
							}

							mountPoints, err = getMountPoints("mapper/" + string(dmName)[:len(dmName)-1])
							if err != nil {
								return nil, fmt.Errorf("error on getting mount points for device mapper volume %s of partition %s: %w", dmName, content.Name(), err)
							}
						}
					}
				} else {
					mountPoints, err = getMountPoints(content.Name())
					if err != nil {
						return nil, fmt.Errorf("error on getting mount points for partition %s: %w", content.Name(), err)
					}
				}

				if len(mountPoints) > 0 {
					stat := syscall.Statfs_t{}

					if err := syscall.Statfs(mountPoints[0].Path, &stat); err != nil {
						return nil, fmt.Errorf("error on statfs syscall using path %s on partition %s: %w", mountPoints[0].Device, content.Name(), err)
					}

					if stat.Bsize < 0 {
						return nil, errors.New("negative block size")
					}

					partition.Used = partition.Size - (stat.Bfree * uint64(stat.Bsize) / 1024)
					partition.Filesystem = mountPoints[0].Filesystem
					partition.FilesystemType = FSType(stat.Type).Name()

					for _, mp := range mountPoints {
						partition.MountPoints = append(partition.MountPoints, mp.Path)
					}

				}

				partitions = append(partitions, partition)
			}
		}

		device := BlockDevice{
			Name:       block.Name(),
			Size:       size / 2,
			Partitions: partitions,
		}

		blockDevices = append(blockDevices, device)
	}

	return blockDevices, nil
}

type MountOption [2]string

type MountPoint struct {
	Device     string
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
				Device:     columns[0],
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
