// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package admin

func NewDummyPulsarAdmin(config PulsarAdminConfig) (PulsarAdmin, error) {
	return &DummyPulsarAdmin{}, nil
}

type DummyPulsarAdmin struct {
}

var _ PulsarAdmin = &DummyPulsarAdmin{}

func (d *DummyPulsarAdmin) ApplyTenant(name string, params *TenantParams) error {
	return nil
}

func (d *DummyPulsarAdmin) DeleteTenant(name string) error {
	return nil
}

func (d *DummyPulsarAdmin) ApplyNamespace(name string, params *NamespaceParams) error {
	return nil
}

func (d *DummyPulsarAdmin) DeleteNamespace(name string) error {
	return nil
}

func (d *DummyPulsarAdmin) ApplyTopic(name string, params *TopicParams) error {
	return nil
}

func (d *DummyPulsarAdmin) DeleteTopic(name string) error {
	return nil
}

func (d *DummyPulsarAdmin) Close() error {
	return nil
}

func (d *DummyPulsarAdmin) GrantPermissions(p Permissioner) error {
	return nil
}

func (d *DummyPulsarAdmin) RevokePermissions(p Permissioner) error {
	return nil
}
