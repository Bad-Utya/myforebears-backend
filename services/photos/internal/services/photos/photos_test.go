package photos

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Bad-Utya/myforebears-backend/services/photos/internal/domain/models"
	"github.com/google/uuid"
)

type metaStub struct {
	t                *testing.T
	getPhotoByIDFn   func(ctx context.Context, photoID uuid.UUID) (models.Photo, error)
}

func (m *metaStub) CreatePhoto(ctx context.Context, photo models.Photo) error {
	m.t.Fatalf("unexpected CreatePhoto call")
	return nil
}

func (m *metaStub) GetPhotoByID(ctx context.Context, photoID uuid.UUID) (models.Photo, error) {
	if m.getPhotoByIDFn == nil {
		m.t.Fatalf("unexpected GetPhotoByID call")
	}
	return m.getPhotoByIDFn(ctx, photoID)
}

func (m *metaStub) GetUserAvatar(ctx context.Context, ownerUserID int) (models.Photo, error) {
	m.t.Fatalf("unexpected GetUserAvatar call")
	return models.Photo{}, nil
}

func (m *metaStub) GetPersonAvatar(ctx context.Context, personID uuid.UUID) (models.Photo, error) {
	m.t.Fatalf("unexpected GetPersonAvatar call")
	return models.Photo{}, nil
}

func (m *metaStub) UnsetPersonAvatar(ctx context.Context, personID uuid.UUID) error {
	m.t.Fatalf("unexpected UnsetPersonAvatar call")
	return nil
}

func (m *metaStub) ListPersonPhotos(ctx context.Context, personID uuid.UUID) ([]models.Photo, error) {
	m.t.Fatalf("unexpected ListPersonPhotos call")
	return nil, nil
}

func (m *metaStub) ListEventPhotos(ctx context.Context, eventID uuid.UUID) ([]models.Photo, error) {
	m.t.Fatalf("unexpected ListEventPhotos call")
	return nil, nil
}

func (m *metaStub) DeletePhotoByID(ctx context.Context, photoID uuid.UUID) (models.Photo, error) {
	m.t.Fatalf("unexpected DeletePhotoByID call")
	return models.Photo{}, nil
}

func (m *metaStub) Close() {}

type objectStub struct {
	t           *testing.T
	getObjectFn func(ctx context.Context, key string) ([]byte, error)
}

func (o *objectStub) PutObject(ctx context.Context, key string, content []byte, mimeType string) error {
	o.t.Fatalf("unexpected PutObject call")
	return nil
}

func (o *objectStub) GetObject(ctx context.Context, key string) ([]byte, error) {
	if o.getObjectFn == nil {
		o.t.Fatalf("unexpected GetObject call")
	}
	return o.getObjectFn(ctx, key)
}

func (o *objectStub) DeleteObject(ctx context.Context, key string) error {
	o.t.Fatalf("unexpected DeleteObject call")
	return nil
}

type familyTreeStub struct {
	t *testing.T
}

func (f *familyTreeStub) GetPerson(ctx context.Context, treeID string, personID string) error {
	f.t.Fatalf("unexpected GetPerson call")
	return nil
}

func (f *familyTreeStub) UpdatePersonAvatarPhoto(ctx context.Context, personID string, avatarPhotoID string) error {
	f.t.Fatalf("unexpected UpdatePersonAvatarPhoto call")
	return nil
}

func (f *familyTreeStub) GetTreeCreatorID(ctx context.Context, treeID string) (int, error) {
	f.t.Fatalf("unexpected GetTreeCreatorID call")
	return 0, nil
}

type eventsStub struct {
	t *testing.T
}

func (e *eventsStub) IsEventFromTree(ctx context.Context, treeID string, eventID string) error {
	e.t.Fatalf("unexpected IsEventFromTree call")
	return nil
}

func TestValidateFileInputErrors(t *testing.T) {
	if err := validateFileInput("", "image/png", []byte{1}); !errors.Is(err, ErrInvalidFileName) {
		t.Fatalf("expected ErrInvalidFileName, got %v", err)
	}
	if err := validateFileInput("a.png", "text/plain", []byte{1}); !errors.Is(err, ErrInvalidMIMEType) {
		t.Fatalf("expected ErrInvalidMIMEType, got %v", err)
	}
	if err := validateFileInput("a.png", "image/png", nil); !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("expected ErrEmptyContent, got %v", err)
	}
	tooLarge := make([]byte, maxPhotoSizeBytes+1)
	if err := validateFileInput("a.png", "image/png", tooLarge); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestNormalizeFileName(t *testing.T) {
	if got := normalizeFileName("../photo.jpg"); got != "photo.jpg" {
		t.Fatalf("expected base file name, got %q", got)
	}
	if got := normalizeFileName(" "); got != "image" {
		t.Fatalf("expected default name, got %q", got)
	}
}

func TestBuildObjectKeyIncludesSegments(t *testing.T) {
	key := buildObjectKey("users", "123", "avatar", "../photo.jpg")
	if len(key) == 0 {
		t.Fatal("expected non-empty key")
	}
	if !strings.HasPrefix(key, "users/123/avatar/") {
		t.Fatalf("expected users/123/avatar/ prefix, got %q", key)
	}
	if !strings.HasSuffix(key, "_photo.jpg") {
		t.Fatalf("expected key to end with _photo.jpg, got %q", key)
	}
}

func TestGetPhotoByIDRejectsUserAvatar(t *testing.T) {
	ctx := context.Background()
	treeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	photoID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	meta := &metaStub{
		t: t,
		getPhotoByIDFn: func(ctx context.Context, id uuid.UUID) (models.Photo, error) {
			return models.Photo{
				ID:           photoID,
				TreeID:       &treeID,
				IsUserAvatar: true,
			}, nil
		},
	}

	objects := &objectStub{
		t: t,
		getObjectFn: func(ctx context.Context, key string) ([]byte, error) {
			t.Fatalf("unexpected GetObject call")
			return nil, nil
		},
	}

	svc := New(nil, meta, objects, &familyTreeStub{t: t}, &eventsStub{t: t})
	_, _, err := svc.GetPhotoByID(ctx, treeID.String(), photoID.String())
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
