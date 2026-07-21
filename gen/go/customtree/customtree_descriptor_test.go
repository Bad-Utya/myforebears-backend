package customtreepb

import "testing"

func TestCustomTreeDescriptorContainsAddParent(t *testing.T) {
	service := File_customtree_customtree_proto.Services().ByName("CustomTreeService")
	if service == nil || service.Methods().ByName("AddParent") == nil {
		t.Fatal("CustomTreeService descriptor does not contain AddParent")
	}
}
