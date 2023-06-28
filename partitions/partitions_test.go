package partitions

import (
	"fmt"
	"testing"
)

func TestParseBlockDevices(t *testing.T) {
	blockDevices, err := ParseBlockDevices()
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(blockDevices)
}
