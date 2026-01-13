// Copyright 2026 StreamNative
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
	"testing"

	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/rest"
	"github.com/go-logr/logr"

	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	"github.com/streamnative/pulsar-resources-operator/pkg/admin"
)

type fakePulsarAdmin struct {
	admin.DummyPulsarAdmin

	getSchemaInfoWithVersionFn        func(topic string) (*resourcev1alpha1.SchemaInfo, int64, error)
	getSchemaVersionBySchemaInfoFn    func(topic string, info *resourcev1alpha1.SchemaInfo) (int64, error)
	uploadSchemaFn                    func(topic string, params *admin.SchemaParams) error
	getSchemaInfoWithVersionCalls     int
	getSchemaVersionBySchemaInfoCalls int
	uploadSchemaCalls                 int
}

func (f *fakePulsarAdmin) GetSchemaInfoWithVersion(topic string) (*resourcev1alpha1.SchemaInfo, int64, error) {
	f.getSchemaInfoWithVersionCalls++
	return f.getSchemaInfoWithVersionFn(topic)
}

func (f *fakePulsarAdmin) GetSchemaVersionBySchemaInfo(topic string, info *resourcev1alpha1.SchemaInfo) (int64, error) {
	f.getSchemaVersionBySchemaInfoCalls++
	return f.getSchemaVersionBySchemaInfoFn(topic, info)
}

func (f *fakePulsarAdmin) UploadSchema(topic string, params *admin.SchemaParams) error {
	f.uploadSchemaCalls++
	return f.uploadSchemaFn(topic, params)
}

func TestApplySchema_NoSchemaInfo(t *testing.T) {
	fake := &fakePulsarAdmin{
		getSchemaInfoWithVersionFn: func(string) (*resourcev1alpha1.SchemaInfo, int64, error) {
			return nil, 0, nil
		},
		getSchemaVersionBySchemaInfoFn: func(string, *resourcev1alpha1.SchemaInfo) (int64, error) {
			return 0, nil
		},
		uploadSchemaFn: func(string, *admin.SchemaParams) error {
			return nil
		},
	}

	topic := &resourcev1alpha1.PulsarTopic{
		Spec: resourcev1alpha1.PulsarTopicSpec{
			Name: "persistent://tenant/ns/topic",
		},
	}

	if err := applySchema(fake, topic, logr.Discard()); err != nil {
		t.Fatalf("applySchema returned error: %v", err)
	}
	if fake.getSchemaInfoWithVersionCalls != 0 {
		t.Fatalf("expected GetSchemaInfoWithVersion not called, got %d", fake.getSchemaInfoWithVersionCalls)
	}
	if fake.getSchemaVersionBySchemaInfoCalls != 0 {
		t.Fatalf("expected GetSchemaVersionBySchemaInfo not called, got %d", fake.getSchemaVersionBySchemaInfoCalls)
	}
	if fake.uploadSchemaCalls != 0 {
		t.Fatalf("expected UploadSchema not called, got %d", fake.uploadSchemaCalls)
	}
}

func TestApplySchema_UploadWhenSchemaMissing(t *testing.T) {
	fake := &fakePulsarAdmin{
		getSchemaInfoWithVersionFn: func(string) (*resourcev1alpha1.SchemaInfo, int64, error) {
			return nil, 0, rest.Error{Code: 404, Reason: "schema not found"}
		},
		getSchemaVersionBySchemaInfoFn: func(string, *resourcev1alpha1.SchemaInfo) (int64, error) {
			t.Fatalf("GetSchemaVersionBySchemaInfo should not be called when schema is missing")
			return 0, nil
		},
		uploadSchemaFn: func(string, *admin.SchemaParams) error {
			return nil
		},
	}

	topic := &resourcev1alpha1.PulsarTopic{
		Spec: resourcev1alpha1.PulsarTopicSpec{
			Name: "persistent://tenant/ns/topic",
			SchemaInfo: &resourcev1alpha1.SchemaInfo{
				Type:   "STRING",
				Schema: "",
			},
		},
	}

	if err := applySchema(fake, topic, logr.Discard()); err != nil {
		t.Fatalf("applySchema returned error: %v", err)
	}
	if fake.uploadSchemaCalls != 1 {
		t.Fatalf("expected UploadSchema called once, got %d", fake.uploadSchemaCalls)
	}
}

func TestApplySchema_SkipWhenVersionMatches(t *testing.T) {
	fake := &fakePulsarAdmin{
		getSchemaInfoWithVersionFn: func(string) (*resourcev1alpha1.SchemaInfo, int64, error) {
			return &resourcev1alpha1.SchemaInfo{Type: "STRING", Schema: ""}, 3, nil
		},
		getSchemaVersionBySchemaInfoFn: func(string, *resourcev1alpha1.SchemaInfo) (int64, error) {
			return 3, nil
		},
		uploadSchemaFn: func(string, *admin.SchemaParams) error {
			return nil
		},
	}

	topic := &resourcev1alpha1.PulsarTopic{
		Spec: resourcev1alpha1.PulsarTopicSpec{
			Name: "persistent://tenant/ns/topic",
			SchemaInfo: &resourcev1alpha1.SchemaInfo{
				Type:   "STRING",
				Schema: "",
			},
		},
	}

	if err := applySchema(fake, topic, logr.Discard()); err != nil {
		t.Fatalf("applySchema returned error: %v", err)
	}
	if fake.uploadSchemaCalls != 0 {
		t.Fatalf("expected UploadSchema not called, got %d", fake.uploadSchemaCalls)
	}
}

func TestApplySchema_UploadWhenVersionDiffers(t *testing.T) {
	fake := &fakePulsarAdmin{
		getSchemaInfoWithVersionFn: func(string) (*resourcev1alpha1.SchemaInfo, int64, error) {
			return &resourcev1alpha1.SchemaInfo{Type: "STRING", Schema: ""}, 3, nil
		},
		getSchemaVersionBySchemaInfoFn: func(string, *resourcev1alpha1.SchemaInfo) (int64, error) {
			return 2, nil
		},
		uploadSchemaFn: func(string, *admin.SchemaParams) error {
			return nil
		},
	}

	topic := &resourcev1alpha1.PulsarTopic{
		Spec: resourcev1alpha1.PulsarTopicSpec{
			Name: "persistent://tenant/ns/topic",
			SchemaInfo: &resourcev1alpha1.SchemaInfo{
				Type:   "STRING",
				Schema: "",
			},
		},
	}

	if err := applySchema(fake, topic, logr.Discard()); err != nil {
		t.Fatalf("applySchema returned error: %v", err)
	}
	if fake.uploadSchemaCalls != 1 {
		t.Fatalf("expected UploadSchema called once, got %d", fake.uploadSchemaCalls)
	}
}

func TestApplySchema_UploadWhenVersionNotFound(t *testing.T) {
	fake := &fakePulsarAdmin{
		getSchemaInfoWithVersionFn: func(string) (*resourcev1alpha1.SchemaInfo, int64, error) {
			return &resourcev1alpha1.SchemaInfo{Type: "STRING", Schema: ""}, 3, nil
		},
		getSchemaVersionBySchemaInfoFn: func(string, *resourcev1alpha1.SchemaInfo) (int64, error) {
			return -1, nil
		},
		uploadSchemaFn: func(string, *admin.SchemaParams) error {
			return nil
		},
	}

	topic := &resourcev1alpha1.PulsarTopic{
		Spec: resourcev1alpha1.PulsarTopicSpec{
			Name: "persistent://tenant/ns/topic",
			SchemaInfo: &resourcev1alpha1.SchemaInfo{
				Type:   "STRING",
				Schema: "",
			},
		},
	}

	if err := applySchema(fake, topic, logr.Discard()); err != nil {
		t.Fatalf("applySchema returned error: %v", err)
	}
	if fake.uploadSchemaCalls != 1 {
		t.Fatalf("expected UploadSchema called once, got %d", fake.uploadSchemaCalls)
	}
}
