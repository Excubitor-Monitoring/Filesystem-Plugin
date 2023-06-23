package main

import (
	"Filesystem-Plugin/partitions"
	"embed"
	"encoding/json"
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
	return modules.NewVersion(0, 0, 1)
}

func (p *PluginImpl) TickFunction() []shared.PluginMessage {
	blockDevices, err := partitions.ParseBlockDevices()
	if err != nil {
		p.logger.Error("Error on retrieving block devices", "ErrorMsg", err.Error())
		return nil
	}

	partitionsJSON, err := json.Marshal(blockDevices)
	if err != nil {
		p.logger.Error("Error on marshalling block devices:", "ErrorMsg", err.Error())
		return nil
	}

	return []shared.PluginMessage{
		{
			Monitor: "FS.Partitions",
			Body:    string(partitionsJSON),
		},
	}
}

func (p *PluginImpl) GetComponents() []modules.Component {
	return []modules.Component{
		{
			TabName: "Filesystems",
			JSFile:  "index.js",
			Tag:     "filesystem-usage",
		},
	}
}

//go:embed frontend
var componentFiles embed.FS

func (p *PluginImpl) GetComponentFile(path string) []byte {
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
	}).With("internal", "true")

	impl := &PluginImpl{logger: logger}

	var pluginMap = map[string]plugin.Plugin{
		"module": &shared.ModulePlugin{Impl: impl},
	}

	logger.Debug("Serving plugin 'filesystem'!")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
