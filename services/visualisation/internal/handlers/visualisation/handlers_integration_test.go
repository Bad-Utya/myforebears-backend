package visualisation

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	familytreepb "github.com/Bad-Utya/myforebears-backend/gen/go/familytree"
	visualisationpb "github.com/Bad-Utya/myforebears-backend/gen/go/visualisation"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/domain/models"
	visservice "github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/services/visualisation"
	"github.com/Bad-Utya/myforebears-backend/services/visualisation/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestRenderCoordinatesForClientRPC(t *testing.T) {
	treeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	rootID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	partnerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	childID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	content := testTreeContent(rootID, partnerID, childID)
	client, cleanup := newVisualisationTestClient(t, content, treeID)
	defer cleanup()

	resp, err := client.RenderCoordinatesForClient(context.Background(), &visualisationpb.RenderCoordinatesForClientRequest{
		TreeId:       treeID.String(),
		RootPersonId: rootID.String(),
		MaxDepth:     0,
	})
	if err != nil {
		t.Fatalf("RenderCoordinatesForClient RPC failed: %v", err)
	}

	if len(resp.GetCoordinatesJson()) == 0 {
		t.Fatal("expected non-empty coordinates json")
	}

	var payload struct {
		Nodes []struct {
			People []struct {
				ID string `json:"id"`
			} `json:"people"`
		} `json:"nodes"`
		Edges []struct {
			EdgeType string `json:"edgeType"`
		} `json:"edges"`
	}
	if err := json.Unmarshal(resp.GetCoordinatesJson(), &payload); err != nil {
		t.Fatalf("failed to parse coordinates json: %v", err)
	}

	if len(payload.Nodes) < 3 {
		t.Fatalf("expected at least 3 nodes, got %d", len(payload.Nodes))
	}

	hasParentChild := false
	for _, edge := range payload.Edges {
		if edge.EdgeType == "parent-child" {
			hasParentChild = true
			break
		}
	}
	if !hasParentChild {
		t.Fatal("expected parent-child edge in coordinates response")
	}
}

func TestCreateVisualisationAndFetchSVGOverRPC(t *testing.T) {
	treeID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	rootID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	partnerID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	childID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	content := testTreeContent(rootID, partnerID, childID)
	client, cleanup := newVisualisationTestClient(t, content, treeID)
	defer cleanup()

	createResp, err := client.CreateAncestorsAndDescendantsVisualisation(context.Background(), &visualisationpb.CreateLineageVisualisationRequest{
		TreeId:       treeID.String(),
		RootPersonId: rootID.String(),
	})
	if err != nil {
		t.Fatalf("CreateAncestorsAndDescendantsVisualisation RPC failed: %v", err)
	}

	visID := createResp.GetVisualisation().GetId()
	if visID == "" {
		t.Fatal("expected visualisation id")
	}

	var getResp *visualisationpb.GetVisualisationContentResponse
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		getResp, err = client.GetVisualisationByID(context.Background(), &visualisationpb.GetVisualisationByIDRequest{
			TreeId:          treeID.String(),
			VisualisationId: visID,
		})
		if err == nil {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("GetVisualisationByID RPC did not become ready: %v", err)
	}

	svg := string(getResp.GetContent())
	if !strings.HasPrefix(svg, "<?xml") {
		t.Fatalf("expected svg xml output, got %q", svg[:minLen(len(svg), 32)])
	}
	if !strings.Contains(svg, "<svg") {
		t.Fatalf("expected svg tag in output")
	}
	if !strings.Contains(svg, "Root") || !strings.Contains(svg, "Partner") || !strings.Contains(svg, "Child") {
		t.Fatalf("expected svg to contain rendered names, got %q", svg[:minLen(len(svg), 256)])
	}
}

func newVisualisationTestClient(t *testing.T, content *familytreepb.GetTreeContentResponse, treeID uuid.UUID) (visualisationpb.VisualisationServiceClient, func()) {
	t.Helper()

	familyClient := &fakeFamilyTreeClient{
		treeID:       treeID.String(),
		rootPersonID: content.GetPersons()[0].GetId(),
		content:      content,
	}
	meta := newFakeMetadataStorage()
	objects := newFakeObjectStorage()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := visservice.New(logger, meta, objects, familyClient)

	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	Register(server, svc)

	go func() {
		_ = server.Serve(lis)
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufconn grpc server: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		server.Stop()
		_ = lis.Close()
	}

	return visualisationpb.NewVisualisationServiceClient(conn), cleanup
}

func testTreeContent(rootID, partnerID, childID uuid.UUID) *familytreepb.GetTreeContentResponse {
	return &familytreepb.GetTreeContentResponse{
		Persons: []*familytreepb.Person{
			{Id: rootID.String(), TreeId: "tree-1", FirstName: "Root", Gender: familytreepb.Gender_GENDER_MALE},
			{Id: partnerID.String(), TreeId: "tree-1", FirstName: "Partner", Gender: familytreepb.Gender_GENDER_FEMALE},
			{Id: childID.String(), TreeId: "tree-1", FirstName: "Child", Gender: familytreepb.Gender_GENDER_UNSPECIFIED},
		},
		Relationships: []*familytreepb.Relationship{
			{PersonIdFrom: rootID.String(), PersonIdTo: partnerID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARTNER_MARRIED},
			{PersonIdFrom: rootID.String(), PersonIdTo: childID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
			{PersonIdFrom: partnerID.String(), PersonIdTo: childID.String(), Type: familytreepb.RelationshipType_RELATIONSHIP_PARENT_CHILD},
		},
	}
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type fakeFamilyTreeClient struct {
	treeID       string
	rootPersonID string
	content      *familytreepb.GetTreeContentResponse
}

func (f *fakeFamilyTreeClient) GetPerson(ctx context.Context, treeID string, personID string) error {
	for _, person := range f.content.GetPersons() {
		if person.GetId() == personID {
			return nil
		}
	}
	return context.Canceled
}

func (f *fakeFamilyTreeClient) GetTreeContent(ctx context.Context, treeID string) (*familytreepb.GetTreeContentResponse, error) {
	return f.content, nil
}

func (f *fakeFamilyTreeClient) GetTreeContentWithinDepth(ctx context.Context, treeID string, rootPersonID string, maxDepth int32) (*familytreepb.GetTreeContentResponse, error) {
	return f.content, nil
}

func (f *fakeFamilyTreeClient) GetTreeCreatorID(ctx context.Context, treeID string) (int, error) {
	return 42, nil
}

type fakeMetadataStorage struct {
	mu    sync.Mutex
	items map[uuid.UUID]models.Visualisation
}

func newFakeMetadataStorage() *fakeMetadataStorage {
	return &fakeMetadataStorage{items: make(map[uuid.UUID]models.Visualisation)}
}

func (s *fakeMetadataStorage) CreateVisualisation(ctx context.Context, vis models.Visualisation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[vis.ID] = vis
	return nil
}

func (s *fakeMetadataStorage) GetVisualisationByID(ctx context.Context, visualisationID uuid.UUID) (models.Visualisation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vis, ok := s.items[visualisationID]
	if !ok {
		return models.Visualisation{}, storage.ErrVisualisationNotFound
	}
	return vis, nil
}

func (s *fakeMetadataStorage) ListTreeVisualisations(ctx context.Context, treeID uuid.UUID) ([]models.Visualisation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]models.Visualisation, 0)
	for _, vis := range s.items {
		if vis.TreeID == treeID {
			out = append(out, vis)
		}
	}
	return out, nil
}

func (s *fakeMetadataStorage) DeleteVisualisationByID(ctx context.Context, visualisationID uuid.UUID) (models.Visualisation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vis, ok := s.items[visualisationID]
	if !ok {
		return models.Visualisation{}, storage.ErrVisualisationNotFound
	}
	delete(s.items, visualisationID)
	return vis, nil
}

func (s *fakeMetadataStorage) SetVisualisationProcessing(ctx context.Context, visualisationID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vis, ok := s.items[visualisationID]
	if !ok {
		return storage.ErrVisualisationNotFound
	}
	vis.Status = models.VisualisationStatusProcessing
	s.items[visualisationID] = vis
	return nil
}

func (s *fakeMetadataStorage) SetVisualisationFailed(ctx context.Context, visualisationID uuid.UUID, errorMessage string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vis, ok := s.items[visualisationID]
	if !ok {
		return storage.ErrVisualisationNotFound
	}
	vis.Status = models.VisualisationStatusFailed
	vis.ErrorMessage = errorMessage
	s.items[visualisationID] = vis
	return nil
}

func (s *fakeMetadataStorage) SetVisualisationReady(ctx context.Context, visualisationID uuid.UUID, sizeBytes int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	vis, ok := s.items[visualisationID]
	if !ok {
		return storage.ErrVisualisationNotFound
	}
	now := time.Now()
	vis.Status = models.VisualisationStatusReady
	vis.SizeBytes = sizeBytes
	vis.CompletedAt = &now
	s.items[visualisationID] = vis
	return nil
}

func (s *fakeMetadataStorage) Close() {}

type fakeObjectStorage struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func newFakeObjectStorage() *fakeObjectStorage {
	return &fakeObjectStorage{objects: make(map[string][]byte)}
}

func (s *fakeObjectStorage) PutObject(ctx context.Context, key string, content []byte, mimeType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = append([]byte(nil), content...)
	return nil
}

func (s *fakeObjectStorage) GetObject(ctx context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.objects[key]...), nil
}

func (s *fakeObjectStorage) DeleteObject(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}
