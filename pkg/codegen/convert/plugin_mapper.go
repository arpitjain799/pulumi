// Copyright 2016-2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package convert

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

// Workspace is the current workspace.
// This is used to get the list of plugins installed in the workspace.
// It's analogous to the workspace package, but scoped down to just the parts we need.
//
// This should probably be used to replace a load of our currently hardcoded for real world (i.e actual file
// system, actual http calls) plugin workspace code, but for now we're keeping it scoped just to help out with
// testing the mapper code.
type Workspace interface {
	// GetPlugins returns the list of plugins installed in the workspace.
	GetPlugins() ([]workspace.PluginInfo, error)
}

type defaultWorkspace struct{}

func (defaultWorkspace) GetPlugins() ([]workspace.PluginInfo, error) {
	return workspace.GetPlugins()
}

// DefaultWorkspace returns a default workspace implementation
// that uses the workspace module directly to get plugin info.
func DefaultWorkspace() Workspace {
	return defaultWorkspace{}
}

// ProviderFactory creates a provider for a given package and version.
type ProviderFactory func(tokens.Package, *semver.Version) (plugin.Provider, error)

// hostManagedProvider is Provider built from a plugin.Host.
type hostManagedProvider struct {
	plugin.Provider

	host plugin.Host
}

var _ plugin.Provider = (*hostManagedProvider)(nil)

func (pc *hostManagedProvider) Close() error {
	return pc.host.CloseProvider(pc.Provider)
}

// ProviderFactoryFromHost builds a ProviderFactory
// that uses the given plugin host to create providers.
func ProviderFactoryFromHost(host plugin.Host) ProviderFactory {
	return func(pkg tokens.Package, version *semver.Version) (plugin.Provider, error) {
		provider, err := host.Provider(pkg, version)
		if err != nil {
			desc := pkg.String()
			if version != nil {
				desc += "@" + version.String()
			}
			return nil, fmt.Errorf("load plugin %v: %w", desc, err)
		}

		return &hostManagedProvider{
			Provider: provider,
			host:     host,
		}, nil
	}
}

type mapperPluginSpec struct {
	name    tokens.Package
	version semver.Version
}

type pluginMapper struct {
	providerFactory ProviderFactory
	conversionKey   string
	plugins         []mapperPluginSpec
	entries         map[string][]byte
}

func NewPluginMapper(ws Workspace,
	providerFactory ProviderFactory,
	key string, mappings []string,
) (Mapper, error) {
	contract.Requiref(providerFactory != nil, "providerFactory", "must not be nil")
	contract.Requiref(ws != nil, "ws", "must not be nil")

	entries := map[string][]byte{}

	// Enumerate _all_ our installed plugins to ask for any mappings they provide. This allows users to
	// convert aws terraform code for example by just having 'pulumi-aws' plugin locally, without needing to
	// specify it anywhere on the command line, and without tf2pulumi needing to know about every possible
	// plugin.
	allPlugins, err := ws.GetPlugins()
	if err != nil {
		return nil, fmt.Errorf("could not get plugins: %w", err)
	}

	// First assumption we only care about the latest version of each plugin. If we add support to get a
	// mapping for plugin version 1, it seems unlikely that we would remove support for that mapping in v2, so
	// the latest version should in most cases be fine. If a user case comes up where this is not fine we can
	// provide the manual workaround that this is based on what is locally installed, not what is published
	// and so the user can just delete the higher version plugins from their cache.
	latestVersions := make(map[string]semver.Version)
	for _, plugin := range allPlugins {
		if plugin.Kind != workspace.ResourcePlugin {
			continue
		}

		if cur, has := latestVersions[plugin.Name]; has {
			if plugin.Version.GT(cur) {
				latestVersions[plugin.Name] = *plugin.Version
			}
		} else {
			latestVersions[plugin.Name] = *plugin.Version
		}
	}

	// We now have a list of plugin specs (i.e. a name and version), save that list because we don't want to
	// iterate all the plugins now because the convert might not even ask for any mappings.
	plugins := make([]mapperPluginSpec, 0)
	for pkg, version := range latestVersions {
		plugins = append(plugins, mapperPluginSpec{
			name:    tokens.Package(pkg),
			version: version,
		})
	}

	// These take precedence over any plugin returned mappings, but we want to error early if we can't read
	// any of these.
	for _, path := range mappings {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not read mapping file '%s': %w", path, err)
		}

		// Mapping file names are assumed to be the provider key.
		provider := filepath.Base(path)
		// strip the extension
		dotIndex := strings.LastIndex(provider, ".")
		if dotIndex != -1 {
			provider = provider[0:dotIndex]
		}

		entries[provider] = data
	}
	return &pluginMapper{
		providerFactory: providerFactory,
		conversionKey:   key,
		plugins:         plugins,
		entries:         entries,
	}, nil
}

// getMappingForPlugin calls GetMapping on the given plugin and returns it's result. Currently if looking up
// the "terraform" mapping and getting an empty result this will fallback to also asking for the "tf" mapping.
// This is because tfbridge providers originally only replied to "tf", while new ones reply (with the same
// answer) to both "tf" and "terraform".
func (l *pluginMapper) getMappingForPlugin(pluginSpec mapperPluginSpec) ([]byte, string, error) {
	providerPlugin, err := l.providerFactory(pluginSpec.name, &pluginSpec.version)
	if err != nil {
		// We should maybe be lenient here and ignore errors but for now assume it's better to fail out on
		// things like providers failing to start.
		return nil, "", fmt.Errorf("could not create provider '%s': %w", pluginSpec.name, err)
	}
	contract.IgnoreClose(providerPlugin)

	data, mappedProvider, err := providerPlugin.GetMapping(l.conversionKey)
	if err != nil {
		// This was an error calling GetMapping, not just that GetMapping returned a nil result. It's fine for
		// GetMapping to return (nil, "", nil) as that simply indicates that the plugin doesn't have a mapping
		// for the requested key.
		return nil, "", fmt.Errorf("could not get mapping for provider '%s': %w", pluginSpec.name, err)
	}
	// A provider should return non-empty results if it has a mapping.
	if mappedProvider != "" && len(data) != 0 {
		return data, mappedProvider, nil
	}
	// If a provider returns (empty, "provider") we also treat that as no mapping, because only the slice part
	// gets returned to the converter plugin and it needs to assume that empty means no mapping, but we warn
	// that this is unexpected.
	if mappedProvider != "" && len(data) == 0 {
		logging.Warningf(
			"provider '%s' returned empty data but a filled provider name '%s' for '%s', "+
				"this is unexpected behaviour assuming no mapping", pluginSpec.name, mappedProvider, l.conversionKey)
	}
	// TODO: Temporary hack to work around the fact that most of the plugins return a mapping for "tf" but
	// not "terraform" but they're the same thing.
	if l.conversionKey == "terraform" {
		// Copy-pasta of the above _but_ we'll delete this whole if block once the plugins have had a
		// chance to update.
		data, mappedProvider, err := providerPlugin.GetMapping("tf")
		if err != nil {
			// This was an error calling GetMapping, not just that GetMapping returned a nil result. It's fine
			// for GetMapping to return (nil, nil) as that simply indicates that the plugin doesn't have a
			// mapping for the requested key.
			return nil, "", fmt.Errorf("could not get mapping for provider '%s': %w", pluginSpec.name, err)
		}
		// A provider should return non-empty results if it has a mapping.
		if mappedProvider != "" && len(data) != 0 {
			return data, mappedProvider, nil
		}
		// If a provider returns (empty, "provider") we also treat that as no mapping, because only the slice part
		// gets returned to the converter plugin and it needs to assume that empty means no mapping, but we warn
		// that this is unexpected.
		if mappedProvider != "" && len(data) == 0 {
			logging.Warningf(
				"provider '%s' returned empty data but a filled provider name '%s' for '%s', "+
					"this is unexpected behaviour assuming no mapping", pluginSpec.name, mappedProvider, l.conversionKey)
		}
	}

	return nil, "", err
}

func (l *pluginMapper) GetMapping(provider string) ([]byte, error) {
	// If we already have an entry for this provider, use it
	if entry, has := l.entries[provider]; has {
		return entry, nil
	}

	// No entry yet, we're going to _try_ to get the plugin that matches the provider name as generally these
	// will match up and it saves us doing a linear search through all plugins.
	matchIdx := -1
	for i, pluginSpec := range l.plugins {
		if pluginSpec.name == tokens.Package(provider) {
			matchIdx = i
			break
		}
	}

	if matchIdx != -1 {
		// Pop this out the list and try and get its mapping. We pop by doing a swap and remove of the last
		// element as order doesn't matter here.
		pluginSpec := l.plugins[matchIdx]
		l.plugins[matchIdx] = l.plugins[len(l.plugins)-1]
		l.plugins = l.plugins[:len(l.plugins)-1]

		data, mappedProvider, err := l.getMappingForPlugin(pluginSpec)
		if err != nil {
			return nil, err
		}
		if mappedProvider != "" {
			contract.Assertf(len(data) != 0,
				"getMappingForPlugin returned empty data but non-empty provider name, %s", mappedProvider)

			// Don't overwrite entries, the first wins
			if _, has := l.entries[mappedProvider]; !has {
				l.entries[mappedProvider] = data
			}
			// If we got a mapping for this provider we can return it
			if mappedProvider == provider {
				return data, nil
			}
		}

		// We didn't find the mapping for the provider were looking for in the pulumi plugin of the same name,
		// so fallback to the linear search.
	}

	// No entry yet, start popping providers off the plugin list and return the first one that returns
	// conversion data for this provider for the given key we're looking for. Second assumption is that only
	// one pulumi provider will provide a mapping for each source mapping. This _might_ change in the future
	// if we for example add support to convert terraform to azure/aws-native, or multiple community members
	// bridge the same terraform provider. But as above the decisions here are based on what's locally
	// installed so the user can manually edit their plugin cache to be the set of plugins they want to use.
	for {
		if len(l.plugins) == 0 {
			// No plugins left to look in, return that we don't have a mapping
			return []byte{}, nil
		}

		// Pop the first spec off the plugins list
		pluginSpec := l.plugins[0]
		l.plugins = l.plugins[1:]

		data, mappedProvider, err := l.getMappingForPlugin(pluginSpec)
		if err != nil {
			return nil, err
		}
		if mappedProvider != "" {
			contract.Assertf(len(data) != 0,
				"getMappingForPlugin returned empty data but non-empty provider name, %s", mappedProvider)

			// Don't overwrite entries, the first wins
			if _, has := l.entries[mappedProvider]; !has {
				l.entries[mappedProvider] = data
			}
			// If this was the provider we we're looking for we can now return it
			if mappedProvider == provider {
				return data, nil
			}
		}
	}
}
