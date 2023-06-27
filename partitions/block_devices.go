package partitions

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// BlockDevice serves as a model of a block device
type BlockDevice struct {
	Name       string      `json:"name"`
	Size       uint64      `json:"size"`
	Partitions []Partition `json:"partitions"`
}

// ParseBlockDevices gathers information about the block devices present in the current system.
func ParseBlockDevices() ([]BlockDevice, error) {
	logger.Trace("Parsing block devices...")

	blockDevices := make([]BlockDevice, 0)

	blockDevicesFolder, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("error on reading block devices: %w", err)
	}

	for _, blockDeviceFolder := range blockDevicesFolder {
		if strings.HasPrefix(blockDeviceFolder.Name(), "dm") {
			// Skip device mapper
			continue
		}

		sizeFile, err := os.ReadFile("/sys/block/" + blockDeviceFolder.Name() + "/size")
		if err != nil {
			return nil, fmt.Errorf("error on reading size of block device %s: %w", blockDeviceFolder.Name(), err)
		}

		size, err := strconv.ParseUint(string(sizeFile)[:len(sizeFile)-1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("error on parsing size of block device %s: %w", blockDeviceFolder.Name(), err)
		}

		partitions, err := getPartitions(blockDeviceFolder.Name())
		if err != nil {
			return nil, fmt.Errorf("error on getting partitions of block device %s: %w", blockDeviceFolder.Name(), err)
		}

		blockDevice := BlockDevice{
			Name:       blockDeviceFolder.Name(),
			Size:       size / 2,
			Partitions: partitions,
		}

		blockDevices = append(blockDevices, blockDevice)
	}

	return blockDevices, nil
}
