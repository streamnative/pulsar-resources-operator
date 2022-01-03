// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

package admin

// NewDummyPulsarAdmin is a dummy initialization function
func NewDummyPulsarAdmin(config PulsarAdminConfig) (PulsarAdmin, error) {
	return &DummyPulsarAdmin{}, nil
}

// DummyPulsarAdmin is a dummy struct of PulsarAdmin
type DummyPulsarAdmin struct {
}

var _ PulsarAdmin = &DummyPulsarAdmin{}

// ApplyTenant is a fake implements of ApplyTenant
func (d *DummyPulsarAdmin) ApplyTenant(name string, params *TenantParams) error {
	return nil
}

// DeleteTenant is a fake implements of DeleteTenant
func (d *DummyPulsarAdmin) DeleteTenant(name string) error {
	return nil
}

// ApplyNamespace is a fake implements of ApplyNamespace
func (d *DummyPulsarAdmin) ApplyNamespace(name string, params *NamespaceParams) error {
	return nil
}

// DeleteNamespace is a fake implements of DeleteNamespace
func (d *DummyPulsarAdmin) DeleteNamespace(name string) error {
	return nil
}

// ApplyTopic is a fake implements of ApplyTopic
func (d *DummyPulsarAdmin) ApplyTopic(name string, params *TopicParams) error {
	return nil
}

// DeleteTopic is a fake implements of DeleteTopic
func (d *DummyPulsarAdmin) DeleteTopic(name string) error {
	return nil
}

// Close is a fake implements of Close
func (d *DummyPulsarAdmin) Close() error {
	return nil
}

// GrantPermissions is a fake implements of GrantPermissions
func (d *DummyPulsarAdmin) GrantPermissions(p Permissioner) error {
	return nil
}

// RevokePermissions is a fake implements of RevokePermissions
func (d *DummyPulsarAdmin) RevokePermissions(p Permissioner) error {
	return nil
}
