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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/blang/semver"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

type testWorkspace struct {
	infos []workspace.PluginInfo
}

func (ws *testWorkspace) GetPlugins() ([]workspace.PluginInfo, error) {
	return ws.infos, nil
}

type testProvider struct {
	pkg     tokens.Package
	mapping func(key string) ([]byte, string, error)
}

func (prov *testProvider) SignalCancellation() error {
	return nil
}

func (prov *testProvider) Close() error {
	return nil
}

func (prov *testProvider) Pkg() tokens.Package {
	return prov.pkg
}

func (prov *testProvider) GetSchema(version int) ([]byte, error) {
	return nil, errors.New("unsupported")
}

func (prov *testProvider) CheckConfig(urn resource.URN, olds,
	news resource.PropertyMap, allowUnknowns bool,
) (resource.PropertyMap, []plugin.CheckFailure, error) {
	return nil, nil, errors.New("unsupported")
}

func (prov *testProvider) DiffConfig(urn resource.URN, olds, news resource.PropertyMap,
	allowUnknowns bool, ignoreChanges []string,
) (plugin.DiffResult, error) {
	return plugin.DiffResult{}, errors.New("unsupported")
}

func (prov *testProvider) Configure(inputs resource.PropertyMap) error {
	return nil
}

func (prov *testProvider) Check(urn resource.URN,
	olds, news resource.PropertyMap, _ bool, _ []byte,
) (resource.PropertyMap, []plugin.CheckFailure, error) {
	return nil, nil, errors.New("unsupported")
}

func (prov *testProvider) Create(urn resource.URN, props resource.PropertyMap, timeout float64,
	preview bool,
) (resource.ID, resource.PropertyMap, resource.Status, error) {
	return "", nil, resource.StatusOK, errors.New("unsupported")
}

func (prov *testProvider) Read(urn resource.URN, id resource.ID,
	inputs, state resource.PropertyMap,
) (plugin.ReadResult, resource.Status, error) {
	return plugin.ReadResult{}, resource.StatusUnknown, errors.New("unsupported")
}

func (prov *testProvider) Diff(urn resource.URN, id resource.ID,
	olds resource.PropertyMap, news resource.PropertyMap, _ bool, _ []string,
) (plugin.DiffResult, error) {
	return plugin.DiffResult{}, errors.New("unsupported")
}

func (prov *testProvider) Update(urn resource.URN, id resource.ID,
	olds resource.PropertyMap, news resource.PropertyMap, timeout float64,
	ignoreChanges []string, preview bool,
) (resource.PropertyMap, resource.Status, error) {
	return nil, resource.StatusOK, errors.New("unsupported")
}

func (prov *testProvider) Delete(urn resource.URN,
	id resource.ID, props resource.PropertyMap, timeout float64,
) (resource.Status, error) {
	return resource.StatusOK, errors.New("unsupported")
}

func (prov *testProvider) Construct(info plugin.ConstructInfo, typ tokens.Type, name tokens.QName, parent resource.URN,
	inputs resource.PropertyMap, options plugin.ConstructOptions,
) (plugin.ConstructResult, error) {
	return plugin.ConstructResult{}, errors.New("unsupported")
}

func (prov *testProvider) Invoke(tok tokens.ModuleMember,
	args resource.PropertyMap,
) (resource.PropertyMap, []plugin.CheckFailure, error) {
	return nil, nil, errors.New("unsupported")
}

func (prov *testProvider) StreamInvoke(
	tok tokens.ModuleMember, args resource.PropertyMap,
	onNext func(resource.PropertyMap) error,
) ([]plugin.CheckFailure, error) {
	return nil, fmt.Errorf("not implemented")
}

func (prov *testProvider) Call(tok tokens.ModuleMember, args resource.PropertyMap, info plugin.CallInfo,
	options plugin.CallOptions,
) (plugin.CallResult, error) {
	return plugin.CallResult{}, errors.New("unsupported")
}

func (prov *testProvider) GetPluginInfo() (workspace.PluginInfo, error) {
	return workspace.PluginInfo{}, errors.New("unsupported")
}

func (prov *testProvider) GetMapping(key string) ([]byte, string, error) {
	return prov.mapping(key)
}

func semverMustParse(s string) *semver.Version {
	v := semver.MustParse(s)
	return &v
}

func TestPluginMapper_InstalledPluginMatches(t *testing.T) {
	t.Parallel()

	ws := &testWorkspace{
		infos: []workspace.PluginInfo{
			{
				Name:    "provider",
				Kind:    workspace.ResourcePlugin,
				Version: semverMustParse("1.0.0"),
			},
		},
	}
	testProvider := &testProvider{
		pkg: tokens.Package("provider"),
		mapping: func(key string) ([]byte, string, error) {
			assert.Equal(t, "key", key)
			return []byte("data"), "provider", nil
		},
	}

	provider := func(pkg tokens.Package, version *semver.Version) (plugin.Provider, error) {
		assert.Equal(t, pkg, testProvider.pkg, "unexpected package %s", pkg)
		return testProvider, nil
	}

	mapper, err := NewPluginMapper(ws, provider, "key", nil)
	assert.NoError(t, err)
	assert.NotNil(t, mapper)

	data, err := mapper.GetMapping("provider")
	assert.NoError(t, err)
	assert.Equal(t, []byte("data"), data)
}

func TestPluginMapper_MappedNameDiffersFromPulumiName(t *testing.T) {
	t.Parallel()

	ws := &testWorkspace{
		infos: []workspace.PluginInfo{
			{
				Name:    "pulumiProvider",
				Kind:    workspace.ResourcePlugin,
				Version: semverMustParse("1.0.0"),
			},
		},
	}
	testProvider := &testProvider{
		pkg: tokens.Package("pulumiProvider"),
		mapping: func(key string) ([]byte, string, error) {
			assert.Equal(t, "key", key)
			return []byte("data"), "otherProvider", nil
		},
	}

	provider := func(pkg tokens.Package, version *semver.Version) (plugin.Provider, error) {
		assert.Equal(t, pkg, testProvider.pkg, "unexpected package %s", pkg)
		return testProvider, nil
	}

	mapper, err := NewPluginMapper(ws, provider, "key", nil)
	assert.NoError(t, err)
	assert.NotNil(t, mapper)

	data, err := mapper.GetMapping("otherProvider")
	assert.NoError(t, err)
	assert.Equal(t, []byte("data"), data)
}

func TestPluginMapper_NoPluginMatches(t *testing.T) {
	t.Parallel()

	ws := &testWorkspace{
		infos: []workspace.PluginInfo{
			{
				Name:    "pulumiProvider",
				Kind:    workspace.ResourcePlugin,
				Version: semverMustParse("1.0.0"),
			},
		},
	}
	testProvider := &testProvider{
		pkg: tokens.Package("pulumiProvider"),
		mapping: func(key string) ([]byte, string, error) {
			assert.Equal(t, "key", key)
			return []byte("data"), "otherProvider", nil
		},
	}

	provider := func(pkg tokens.Package, version *semver.Version) (plugin.Provider, error) {
		assert.Equal(t, pkg, testProvider.pkg, "unexpected package %s", pkg)
		return testProvider, nil
	}

	mapper, err := NewPluginMapper(ws, provider, "key", nil)
	assert.NoError(t, err)
	assert.NotNil(t, mapper)

	data, err := mapper.GetMapping("yetAnotherProvider")
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, data)
}

func TestPluginMapper_UseMatchingNameFirst(t *testing.T) {
	t.Parallel()

	ws := &testWorkspace{
		infos: []workspace.PluginInfo{
			{
				Name:    "otherProvider",
				Kind:    workspace.ResourcePlugin,
				Version: semverMustParse("1.0.0"),
			},
			{
				Name:    "provider",
				Kind:    workspace.ResourcePlugin,
				Version: semverMustParse("1.0.0"),
			},
		},
	}
	testProvider := &testProvider{
		pkg: tokens.Package("provider"),
		mapping: func(key string) ([]byte, string, error) {
			assert.Equal(t, "key", key)
			return []byte("data"), "provider", nil
		},
	}

	provider := func(pkg tokens.Package, version *semver.Version) (plugin.Provider, error) {
		assert.Equal(t, pkg, testProvider.pkg, "unexpected package %s", pkg)
		return testProvider, nil
	}

	mapper, err := NewPluginMapper(ws, provider, "key", nil)
	assert.NoError(t, err)
	assert.NotNil(t, mapper)

	data, err := mapper.GetMapping("provider")
	assert.NoError(t, err)
	assert.Equal(t, []byte("data"), data)
}

func TestPluginMapper_MappedNamesDifferFromPulumiName(t *testing.T) {
	t.Parallel()

	ws := &testWorkspace{
		infos: []workspace.PluginInfo{
			{
				Name:    "pulumiProviderAws",
				Kind:    workspace.ResourcePlugin,
				Version: semverMustParse("1.0.0"),
			},
			{
				Name:    "pulumiProviderGcp",
				Kind:    workspace.ResourcePlugin,
				Version: semverMustParse("1.0.0"),
			},
		},
	}
	testProviderAws := &testProvider{
		pkg: tokens.Package("pulumiProviderAws"),
		mapping: func(key string) ([]byte, string, error) {
			assert.Equal(t, "key", key)
			return []byte("dataaws"), "aws", nil
		},
	}
	testProviderGcp := &testProvider{
		pkg: tokens.Package("pulumiProviderGcp"),
		mapping: func(key string) ([]byte, string, error) {
			assert.Equal(t, "key", key)
			return []byte("datagcp"), "gcp", nil
		},
	}

	provider := func(pkg tokens.Package, version *semver.Version) (plugin.Provider, error) {
		if pkg == testProviderAws.pkg {
			return testProviderAws, nil
		} else if pkg == testProviderGcp.pkg {
			return testProviderGcp, nil
		}
		assert.Fail(t, "unexpected package %s", pkg)
		return nil, fmt.Errorf("unexpected package %s", pkg)
	}

	mapper, err := NewPluginMapper(ws, provider, "key", nil)
	assert.NoError(t, err)
	assert.NotNil(t, mapper)

	// Get the mapping for the GCP provider.
	data, err := mapper.GetMapping("gcp")
	assert.NoError(t, err)
	assert.Equal(t, []byte("datagcp"), data)

	// Now get the mapping for the AWS provider, it should be cached.
	data, err = mapper.GetMapping("aws")
	assert.NoError(t, err)
	assert.Equal(t, []byte("dataaws"), data)
}
