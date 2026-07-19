package mail

import (
	"testing"
)

func TestAppendStateUserOrgIDsRequiresConfiguredSyncUser(t *testing.T) {
	t.Setenv("IMAP_SYNC_FANOUT_USER", "mail-sync@example.com")

	userOrgIDs, err := appendStateUserOrgIDs(nil, "regular@example.com", "unused", "current-user-org")
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

	userOrgIDs, err := appendStateUserOrgIDs(
		app,
		"MAIL-SYNC@example.com",
		padID("fanout_mailbox"),
		"sync-user-org",
	)
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
