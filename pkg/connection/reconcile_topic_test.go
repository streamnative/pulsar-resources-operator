// Copyright 2025 StreamNative
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package connection

import (
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

type fakeSchemaAdmin struct {
	admin.DummyPulsarAdmin
	schema    *resourcev1alpha1.SchemaInfo
	schemaErr error
	uploads   []*admin.SchemaParams
}

func (f *fakeSchemaAdmin) GetSchema(string) (*resourcev1alpha1.SchemaInfo, error) {
	return f.schema, f.schemaErr
}

func (f *fakeSchemaAdmin) UploadSchema(_ string, params *admin.SchemaParams) error {
	f.uploads = append(f.uploads, params)
	return nil
}

func TestApplySchemaSkipsEquivalentSchema(t *testing.T) {
	t.Parallel()

	fake := &fakeSchemaAdmin{
		schema: &resourcev1alpha1.SchemaInfo{
			Type: "AVRO",
			Schema: `{
	  "name": "Example",
	  "type": "record",
	  "fields": [
	    {
	      "name": "foo",
	      "type": "string"
	    },
	    {
	      "name": "bar",
	      "type": "int"
	    }
	  ]
	}`,
			Properties: map[string]string{},
		},
	}

	topic := &resourcev1alpha1.PulsarTopic{}
	topic.Spec.Name = "persistent://public/default/example"
	topic.Spec.SchemaInfo = &resourcev1alpha1.SchemaInfo{
		Type:       "AVRO",
		Schema:     `{"name":"Example","type":"record","fields":[{"name":"foo","type":"string"},{"name":"bar","type":"int"}]}`,
		Properties: nil,
	}

	if err := applySchema(fake, topic, logr.Discard()); err != nil {
		t.Fatalf("applySchema() error = %v", err)
	}

	if len(fake.uploads) != 0 {
		t.Fatalf("expected no schema upload when schemas are equivalent, got %d", len(fake.uploads))
	}
}

func TestSchemasEqual(t *testing.T) {
	t.Parallel()

	avroSpec := &resourcev1alpha1.SchemaInfo{
		Type:   "AVRO",
		Schema: `{"name":"Example","type":"record","fields":[{"name":"foo","type":"string"},{"name":"bar","type":"int"}]}`,
	}

	avroActualFormatted := &resourcev1alpha1.SchemaInfo{
		Type: "AVRO",
		Schema: `{
  "type": "record",
  "name": "Example",
  "fields": [
    {
      "type": "string",
      "name": "foo"
    },
    {
      "type": "int",
      "name": "bar"
    }
  ]
}`,
	}

	cases := []struct {
		name     string
		desired  *resourcev1alpha1.SchemaInfo
		current  *resourcev1alpha1.SchemaInfo
		expected bool
	}{
		{
			name:     "both nil",
			desired:  nil,
			current:  nil,
			expected: true,
		},
		{
			name:     "current missing",
			desired:  avroSpec,
			current:  nil,
			expected: false,
		},
		{
			name: "type mismatch",
			desired: &resourcev1alpha1.SchemaInfo{
				Type:   "JSON",
				Schema: `{"type":"object"}`,
			},
			current: &resourcev1alpha1.SchemaInfo{
				Type:   "AVRO",
				Schema: `{"type":"record"}`,
			},
			expected: false,
		},
		{
			name:     "equivalent json formatting",
			desired:  avroSpec,
			current:  avroActualFormatted,
			expected: true,
		},
		{
			name: "properties nil vs empty",
			desired: &resourcev1alpha1.SchemaInfo{
				Type:   "JSON",
				Schema: `{"type":"object"}`,
			},
			current: &resourcev1alpha1.SchemaInfo{
				Type:       "JSON",
				Schema:     `{"type":"object"}`,
				Properties: map[string]string{},
			},
			expected: true,
		},
		{
			name: "properties differ",
			desired: &resourcev1alpha1.SchemaInfo{
				Type:       "JSON",
				Schema:     `{"type":"object"}`,
				Properties: map[string]string{"env": "prod"},
			},
			current: &resourcev1alpha1.SchemaInfo{
				Type:       "JSON",
				Schema:     `{"type":"object"}`,
				Properties: map[string]string{"env": "dev"},
			},
			expected: false,
		},
		{
			name: "different schema content",
			desired: &resourcev1alpha1.SchemaInfo{
				Type:   "AVRO",
				Schema: `{"name":"Example","type":"record","fields":[{"name":"foo","type":"string"}]}`,
			},
			current: &resourcev1alpha1.SchemaInfo{
				Type:   "AVRO",
				Schema: `{"name":"Example","type":"record","fields":[{"name":"foo","type":"int"}]}`,
			},
			expected: false,
		},
	}

	for i := range cases {
		tc := cases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := reflect.DeepEqual(tc.desired, tc.current); got != tc.expected {
				t.Fatalf("reflect.DeepEqual() = %v, want %v", got, tc.expected)
			}
		})
	}
}
