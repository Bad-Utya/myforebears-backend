package familytree

import (
	"github.com/Bad-Utya/myforebears-backend/services/familytree/internal/domain/models"
	"github.com/google/uuid"
	"testing"
)

func TestNormalizePublicEventsPreservesExistingIDs(t *testing.T) {
	personID, eventID := uuid.New(), uuid.New()
	items := normalizePublicEvents(personID, []models.PublicPersonEvent{{ID: eventID, EventTypeName: "Birth"}, {EventTypeName: "Death"}})
	if items[0].ID != eventID {
		t.Fatal("existing event id changed during edit")
	}
	if items[1].ID == uuid.Nil {
		t.Fatal("new event did not receive id")
	}
	for _, item := range items {
		if item.PublicPersonID != personID {
			t.Fatal("event was not attached to public person")
		}
	}
}
