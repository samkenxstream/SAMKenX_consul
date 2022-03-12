//go:build consulent
// +build consulent

package acl

import (
	"hash"
	"strings"
)

// EnterpriseMeta contains common Consul Enterprise specific metadata
//
// NOTES:
// * We rely on this struct being "comparable" in the agent/local state
//   Basically its used as a part of a struct that is used as a map key
//   and we do equality operations on it with ==. The compiler SHOULD
//   complain if incompatible fields are added but if it does, this is why.
type EnterpriseMeta struct {
	// Partition is the Consul Enterprise Admin Partition the data resides within - this is a layer of
	// isolation above namespaces.
	Partition string `json:",omitempty"`
	// Namespace is the Consul Enterprise Namespace the data resides within
	Namespace string `json:",omitempty"`
}

var (
	defaultEnterpriseMeta = &EnterpriseMeta{
		Partition: DefaultPartitionName,
		Namespace: DefaultNamespaceName,
	}

	wildcardEnterpriseMeta = &EnterpriseMeta{
		Namespace: WildcardName,
	}
)

func (m *EnterpriseMeta) ToEnterprisePolicyMeta() *EnterprisePolicyMeta {
	if m == nil {
		return &EnterprisePolicyMeta{}
	}
	return &EnterprisePolicyMeta{
		Namespace: m.Namespace,
		Partition: m.Partition,
	}
}

func DefaultEnterpriseMeta() *EnterpriseMeta {
	return &EnterpriseMeta{Partition: DefaultPartitionName, Namespace: DefaultNamespaceName}
}

func WildcardEnterpriseMeta() *EnterpriseMeta {
	return &EnterpriseMeta{Namespace: WildcardName}
}

func (m *EnterpriseMeta) EstimateSize() int {
	return len(m.Namespace)
}

func (m *EnterpriseMeta) AddToHash(hasher hash.Hash, ignoreDefault bool) {
	partition := m.PartitionOrDefault()
	ns := m.NamespaceOrDefault()

	// In some circumstances we don't want to include the default partition/namespace
	// in a hash. The one big one is when generating the names of service/check
	// registration files for on-disk persistence
	if !ignoreDefault || partition != DefaultPartitionName {
		hasher.Write([]byte(partition))
	}
	if !ignoreDefault || ns != DefaultNamespaceName {
		hasher.Write([]byte(ns))
	}
}

func (m *EnterpriseMeta) PartitionOrDefault() string {
	if m == nil || m.Partition == "" {
		return DefaultPartitionName
	}

	return strings.ToLower(m.Partition)
}

func EqualPartitions(a, b string) bool {
	if IsDefaultPartition(a) && IsDefaultPartition(b) {
		return true
	}

	return strings.EqualFold(a, b)
}

func IsDefaultPartition(partition string) bool {
	return partition == "" || strings.EqualFold(partition, DefaultPartitionName)
}

func PartitionOrDefault(partition string) string {
	if partition == "" {
		return DefaultPartitionName
	}

	return strings.ToLower(partition)
}

func (m *EnterpriseMeta) PartitionOrEmpty() string {
	if m == nil || m.Partition == "" {
		return ""
	}

	return strings.ToLower(m.Partition)
}

func (m *EnterpriseMeta) InDefaultPartition() bool {
	return m.PartitionOrDefault() == DefaultPartitionName
}

func (m *EnterpriseMeta) NamespaceOrDefault() string {
	if m == nil || m.Namespace == "" {
		return DefaultNamespaceName
	}

	return strings.ToLower(m.Namespace)
}

func NamespaceOrDefault(namespace string) string {
	if namespace == "" {
		return DefaultNamespaceName
	}

	return strings.ToLower(namespace)
}

func (m *EnterpriseMeta) NamespaceOrEmpty() string {
	if m == nil || m.Namespace == "" {
		return ""
	}

	return strings.ToLower(m.Namespace)
}

func (m *EnterpriseMeta) InDefaultNamespace() bool {
	return m.NamespaceOrDefault() == DefaultNamespaceName
}

func (m *EnterpriseMeta) Merge(other *EnterpriseMeta) {
	if m == nil || other == nil {
		return
	}

	// overwrite the partition if its empty
	if m.Partition == "" {
		m.Partition = other.Partition
	}
	// overwrite the namespace if its empty
	if m.Namespace == "" {
		m.Namespace = other.Namespace
	}
}

func (m *EnterpriseMeta) MergeNoWildcard(other *EnterpriseMeta) {
	if m == nil || other == nil {
		return
	}

	// overwrite the partition if its empty and the new partition isn't the wildcard
	if m.Partition == "" && other.Partition != WildcardName {
		m.Partition = other.Partition
	}
	// overwrite the namespace if its empty and the new ns isn't the wildcard
	if m.Namespace == "" && other.Namespace != WildcardName {
		m.Namespace = other.Namespace
	}
}

func (m *EnterpriseMeta) Normalize() {
	if m == nil {
		return
	}

	if m.Partition == "" {
		m.Partition = DefaultPartitionName
	} else {
		m.Partition = strings.ToLower(m.Partition)
	}

	if m.Namespace == "" {
		m.Namespace = DefaultNamespaceName
	} else {
		m.Namespace = strings.ToLower(m.Namespace)
	}
}

func (m *EnterpriseMeta) Matches(other *EnterpriseMeta) bool {
	partition := m.PartitionOrDefault()
	ns := m.NamespaceOrDefault()

	otherPartition := other.PartitionOrDefault()
	// TODO(partitions): should we allow wildcards here?
	if partition == otherPartition || partition == WildcardName || otherPartition == WildcardName {
		otherNS := other.NamespaceOrDefault()
		if ns == otherNS || ns == WildcardName || otherNS == WildcardName {
			return true
		}
	}

	return false
}

func (m *EnterpriseMeta) IsSame(other *EnterpriseMeta) bool {
	return m.NamespaceOrDefault() == other.NamespaceOrDefault() &&
		m.PartitionOrDefault() == other.PartitionOrDefault()
}

func (m *EnterpriseMeta) LessThan(other *EnterpriseMeta) bool {
	if m.PartitionOrDefault() != other.PartitionOrDefault() {
		return m.PartitionOrDefault() < other.PartitionOrDefault()
	}
	return m.NamespaceOrDefault() < other.NamespaceOrDefault()
}

// WithWildcardNamespace returns a new EnterpriseMeta with
// the same partition, and with the namespace set to the wildcard namespace.
func (m *EnterpriseMeta) WithWildcardNamespace() *EnterpriseMeta {
	// TODO: using PartitionOrDefault here seems like it could be problematic
	// in some cases. Can we avoid yet another place where we set defaults and
	// instead use t.Partition directly?
	return &EnterpriseMeta{Partition: m.PartitionOrDefault(), Namespace: WildcardName}
}

// UnsetPartition returns a new EnterpriseMeta with the partition set
// to the empty string. This is used for cases like forwarding RPCs to a remote
// datacenter where it's uncertain whether the remote server has knowledge of
// partitions and so should decide how to behave.
func (m *EnterpriseMeta) UnsetPartition() {
	m.Partition = ""
}

func NewEnterpriseMetaWithPartition(partition, ns string) EnterpriseMeta {
	return EnterpriseMeta{Partition: partition, Namespace: ns}
}

// FillAuthzContext is used to fill in bits of the ACL AuthorizerContext. This fills in values
// in an existing struct instead of allocating and returning a new one to avoid lots of extra allocations
// during ACL enforcement. It also will be easier to layer these up so that data that needs a Sentinel
// scope can define this Function on the type that embeds the EnterpriseMeta and in that implmentation it
// can ask the embedded EnterpriseMeta to fill in its own bits.
func (m *EnterpriseMeta) FillAuthzContext(ctx *AuthorizerContext) {
	if ctx == nil {
		return
	}
	ctx.Namespace = m.NamespaceOrDefault()
	ctx.Partition = m.PartitionOrDefault()
}
