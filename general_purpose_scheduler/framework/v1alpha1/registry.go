/*
Copyright 2019 The Kubernetes Authors.
Modifications copyright (C) 2020 Alex Neo

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
  "fmt"
  "k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"
	"k8s.io/apimachinery/pkg/runtime"
)

// PluginFactory is a function that builds a plugin.
type PluginFactory = func(obj runtime.Object, f FrameworkHandle) (Plugin, error)

// Registry is a collection of all available plugins. The framework uses a
// registry to enable and initialize configured plugins.
// All plugins must be in the registry before initializing the framework.
type Registry map[string]PluginFactory

// DecodeInto decodes configuration whose type is *runtime.Unknown to the interface into.
func DecodeInto(obj runtime.Object, into interface{}) error {
	if obj == nil {
		return nil
	}
	configuration, ok := obj.(*runtime.Unknown)
	if !ok {
		return fmt.Errorf("want args of type runtime.Unknown, got %T", obj)
	}
	if configuration.Raw == nil {
		return nil
	}

	switch configuration.ContentType {
	// If ContentType is empty, it means ContentTypeJSON by default.
	case runtime.ContentTypeJSON, "":
		return json.Unmarshal(configuration.Raw, into)
	case runtime.ContentTypeYAML:
		return yaml.Unmarshal(configuration.Raw, into)
	default:
		return fmt.Errorf("not supported content type %s", configuration.ContentType)
	}
}
