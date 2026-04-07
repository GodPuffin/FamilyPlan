package userprofile

import (
	"testing"

	pbmodels "github.com/pocketbase/pocketbase/models"
)

func TestAvatarURL(t *testing.T) {
	t.Parallel()

	collection := &pbmodels.Collection{Name: "users"}
	collection.Id = "_pb_users_auth_"
	record := pbmodels.NewRecord(collection)
	record.Id = "user 123"
	record.Set("avatar", "profile photo.png")

	got := AvatarURL(record)
	want := "/api/files/_pb_users_auth_/user%20123/profile%20photo.png"
	if got != want {
		t.Fatalf("AvatarURL() = %q, want %q", got, want)
	}
}

func TestAvatarURLEmpty(t *testing.T) {
	t.Parallel()

	if got := AvatarURL(nil); got != "" {
		t.Fatalf("AvatarURL(nil) = %q, want empty string", got)
	}

	collection := &pbmodels.Collection{Name: "users"}
	collection.Id = "_pb_users_auth_"
	record := pbmodels.NewRecord(collection)
	record.Id = "user_123"
	if got := AvatarURL(record); got != "" {
		t.Fatalf("AvatarURL(record without avatar) = %q, want empty string", got)
	}
}

func TestAvatarURLNilCollection(t *testing.T) {
	t.Parallel()

	record := &pbmodels.Record{}
	if got := AvatarURL(record); got != "" {
		t.Fatalf("AvatarURL(record without collection) = %q, want empty string", got)
	}
}
