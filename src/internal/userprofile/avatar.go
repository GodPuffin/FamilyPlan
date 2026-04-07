package userprofile

import (
	"net/url"

	pbmodels "github.com/pocketbase/pocketbase/models"
)

// AvatarURL returns the public PocketBase file URL for a user's avatar.
func AvatarURL(record *pbmodels.Record) string {
	if record == nil || record.Collection() == nil {
		return ""
	}

	avatar := record.GetString("avatar")
	if avatar == "" || record.Id == "" || record.Collection().Id == "" {
		return ""
	}

	return "/api/files/" +
		url.PathEscape(record.Collection().Id) + "/" +
		url.PathEscape(record.Id) + "/" +
		url.PathEscape(avatar)
}
