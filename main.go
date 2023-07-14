package main

import (
	"Filesystem-Plugin/partitions"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/Excubitor-Monitoring/Excubitor-Backend/pkg/shared"
	"github.com/Excubitor-Monitoring/Excubitor-Backend/pkg/shared/modules"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
)

type PluginImpl struct {
	logger hclog.Logger
}

func (p *PluginImpl) GetName() string {
	p.logger.Trace("Name is retrieved from plugin.")
	return "Filesystems"
}

func (p *PluginImpl) GetVersion() modules.Version {
	p.logger.Trace("Version is retrieved from plugin.")
	return modules.NewVersion(0, 0, 1)
}

func (p *PluginImpl) TickFunction() []shared.PluginMessage {
	p.logger.Trace("Plugin is ticked.")
	blockDevices, err := partitions.ParseBlockDevices()
	if err != nil {
		p.logger.Error("Error on retrieving block devices", "ErrorMsg", err.Error())
		return nil
	}

	p.logger.Debug("Marshalling block devices into JSON.")
	partitionsJSON, err := json.Marshal(blockDevices)
	if err != nil {
		p.logger.Error("Error on marshalling block devices:", "ErrorMsg", err.Error())
		return nil
	}

	p.logger.Debug("Returning plugin messages.")
	return []shared.PluginMessage{
		{
			Monitor: "FS.Partitions",
			Body:    string(partitionsJSON),
		},
	}
}

func (p *PluginImpl) GetComponents() []modules.Component {
	p.logger.Trace("Components are retrieved from plugin.")
	return []modules.Component{
		{
			TabName: "Statistics",
			JSFile:  "index.js",
			Tag:     "filesystem-usage",
		},
	}
}

//go:embed frontend
var componentFiles embed.FS

func (p *PluginImpl) GetComponentFile(path string) []byte {
	p.logger.Trace(fmt.Sprintf("Attempt to retrieve component file with path '%s'.", path))
	content, err := componentFiles.ReadFile("frontend/" + path)
	if err != nil {
		return make([]byte, 0)
	}

	return content
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MODULE_PLUGIN",
	MagicCookieValue: "EXCUBITOR",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:       hclog.Trace,
		Output:      os.Stderr,
		DisableTime: true,
		JSONFormat:  true,
	})

	logger.Debug("Initializing plugin implementation...")

	impl := &PluginImpl{logger: logger}

	var pluginMap = map[string]plugin.Plugin{
		"module": &shared.ModulePlugin{Impl: impl},
	}

	logger.Debug("Serving plugin 'Filesystems'!")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
