package mail

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestAppendStateUserOrgIDsRequiresConfiguredSyncUser(t *testing.T) {
	t.Setenv("IMAP_SYNC_FANOUT_USER", "mail-sync@example.com")

	session := &imapSession{user: core.NewRecord(core.NewAuthCollection("users"))}
	session.user.Set("email", "regular@example.com")

	userOrgIDs, err := session.appendStateUserOrgIDs("unused", "current-user-org")
	if err != nil {
		t.Fatalf("appendStateUserOrgIDs returned an error: %v", err)
	}
	if len(userOrgIDs) != 1 || userOrgIDs[0] != "current-user-org" {
		t.Fatalf("userOrgIDs = %v, want [current-user-org]", userOrgIDs)
	}
}

func TestAppendStateUserOrgIDsFansOutForConfiguredSyncUser(t *testing.T) {
	t.Setenv("IMAP_SYNC_FANOUT_USER", "mail-sync@example.com")
	app := setupInboundTestApp(t)
	seedDomainAndMailbox(t, app, "example.com", "shared", "fanout_mailbox")
	seedMember(t, app, "fanout_mailbox", "member_one")
	seedMember(t, app, "fanout_mailbox", "member_two")

	user := core.NewRecord(core.NewAuthCollection("users"))
	user.Set("email", "MAIL-SYNC@example.com")
	session := &imapSession{app: app.PocketBase, user: user}

	userOrgIDs, err := session.appendStateUserOrgIDs(padID("fanout_mailbox"), "sync-user-org")
	if err != nil {
		t.Fatalf("appendStateUserOrgIDs returned an error: %v", err)
	}
	got := make(map[string]bool, len(userOrgIDs))
	for _, userOrgID := range userOrgIDs {
		got[userOrgID] = true
	}
	if len(userOrgIDs) != 2 || !got["member_one"] || !got["member_two"] {
		t.Fatalf("userOrgIDs = %v, want [member_one member_two]", userOrgIDs)
	}
}
