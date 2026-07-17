package mail

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// Rule strings mirror 1830000004_allow_org_admin_mailbox_metadata.js.
const (
	mailAdminMailboxMember = `mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id`
	mailAdminMailboxOwner  = `mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id && ` +
		`mail_mailbox_members_via_mailbox.role ?= "owner"`
	mailAdminViaDomain = `domain.org.user_org_via_org.user ?= @request.auth.id && ` +
		`(domain.org.user_org_via_org.role ?= "admin" || domain.org.user_org_via_org.role ?= "owner")`
	mailAdminMailboxRead = `(` + mailAdminMailboxMember + `) || (` + mailAdminViaDomain + `)`

	mailAdminSelfMember      = `user_org.user = @request.auth.id`
	mailAdminOwnerViaMailbox = `mailbox.mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id && ` +
		`mailbox.mail_mailbox_members_via_mailbox.role ?= "owner"`
	mailAdminViaMailbox = `mailbox.domain.org.user_org_via_org.user ?= @request.auth.id && ` +
		`(mailbox.domain.org.user_org_via_org.role ?= "admin" || ` +
		`mailbox.domain.org.user_org_via_org.role ?= "owner")`
	mailAdminMemberRead   = `(` + mailAdminSelfMember + `) || (` + mailAdminViaMailbox + `)`
	mailAdminMemberCreate = `(` + mailAdminOwnerViaMailbox +
		` && user_org.org = mailbox.domain.org) || (` + mailAdminViaMailbox +
		` && user_org.org = mailbox.domain.org)`

	mailContentMemberRead = `mailbox.mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id`
)

type mailboxAdminEnv struct {
	app              *tests.TestApp
	mailbox          *core.Record
	ownerMembership  *core.Record
	nonmemberUserOrg *core.Record
	adminToken       string
	ownerToken       string
	nonmemberToken   string
}

func setupMailboxAdminApp(t *testing.T) *mailboxAdminEnv {
	t.Helper()
	guestEnv := setupMailGuestApp(t)
	app := guestEnv.app

	guestEnv.mailbox.Set("type", "personal")
	if err := app.Save(guestEnv.mailbox); err != nil {
		t.Fatal(err)
	}

	members, _ := app.FindCollectionByNameOrId("mail_mailbox_members")
	ownerMembership := core.NewRecord(members)
	ownerMembership.Set("mailbox", guestEnv.mailbox.Id)
	ownerMembership.Set("user_org", guestEnv.memberUserOrg.Id)
	ownerMembership.Set("role", "owner")
	if err := app.Save(ownerMembership); err != nil {
		t.Fatal(err)
	}

	admin := mailGuestUser(t, app, "admin@test.local")
	mailGuestMembership(t, app, admin, guestEnv.org, "admin")
	adminToken, err := admin.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}

	nonmember := mailGuestUser(t, app, "nonmember@test.local")
	nonmemberUserOrg := mailGuestMembership(t, app, nonmember, guestEnv.org, "member")
	nonmemberToken, err := nonmember.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}

	return &mailboxAdminEnv{
		app:              app,
		mailbox:          guestEnv.mailbox,
		ownerMembership:  ownerMembership,
		nonmemberUserOrg: nonmemberUserOrg,
		adminToken:       adminToken,
		ownerToken:       guestEnv.memberToken,
		nonmemberToken:   nonmemberToken,
	}
}

func mailSetUpdate(t *testing.T, app core.App, name, rule string) {
	t.Helper()
	col, err := app.FindCollectionByNameOrId(name)
	if err != nil {
		t.Fatal(err)
	}
	col.UpdateRule = &rule
	if err := app.Save(col); err != nil {
		t.Fatalf("set update on %s: %v", name, err)
	}
}

func TestMailboxAdminRLS_AdminSeesPersonalMailboxMetadata(t *testing.T) {
	env := setupMailboxAdminApp(t)
	mailSetListView(t, env.app, "mail_mailboxes", mailAdminMailboxRead)

	mailRunList(t, env.app, "mail_mailboxes", env.adminToken,
		[]string{`"totalItems":1`, `"type":"personal"`, `"address":"team"`}, nil)
}

func TestMailboxAdminRLS_NonmemberCannotSeePersonalMailboxMetadata(t *testing.T) {
	env := setupMailboxAdminApp(t)
	mailSetListView(t, env.app, "mail_mailboxes", mailAdminMailboxRead)

	mailRunList(t, env.app, "mail_mailboxes", env.nonmemberToken,
		[]string{`"totalItems":0`}, []string{`"address":"team"`})
}

func TestMailboxAdminRLS_AdminSeesMailboxMembers(t *testing.T) {
	env := setupMailboxAdminApp(t)
	mailSetListView(t, env.app, "mail_mailbox_members", mailAdminMemberRead)

	mailRunList(t, env.app, "mail_mailbox_members", env.adminToken,
		[]string{`"totalItems":1`, env.ownerMembership.Id}, nil)
}

func TestMailboxAdminRLS_AdminCanAddMailboxMember(t *testing.T) {
	env := setupMailboxAdminApp(t)
	mailSetCreate(t, env.app, "mail_mailbox_members", mailAdminMemberCreate)

	body := `{"mailbox":"` + env.mailbox.Id + `","user_org":"` +
		env.nonmemberUserOrg.Id + `","role":"member"}`
	scenario := &tests.ApiScenario{
		Method:                http.MethodPost,
		URL:                   "/api/collections/mail_mailbox_members/records",
		Body:                  strings.NewReader(body),
		Headers:               map[string]string{"Authorization": env.adminToken, "Content-Type": "application/json"},
		ExpectedStatus:        http.StatusOK,
		ExpectedContent:       []string{`"role":"member"`, `"user_org":"` + env.nonmemberUserOrg.Id + `"`},
		TestAppFactory:        func(_ testing.TB) *tests.TestApp { return env.app },
		DisableTestAppCleanup: true,
	}
	scenario.Test(t)
}

func TestMailboxAdminRLS_AdminCanUpdateMailboxMetadata(t *testing.T) {
	env := setupMailboxAdminApp(t)
	mailSetUpdate(t, env.app, "mail_mailboxes", `(`+mailAdminMailboxOwner+`) || (`+mailAdminViaDomain+`)`)

	scenario := &tests.ApiScenario{
		Method:                http.MethodPatch,
		URL:                   "/api/collections/mail_mailboxes/records/" + env.mailbox.Id,
		Body:                  strings.NewReader(`{"address":"renamed"}`),
		Headers:               map[string]string{"Authorization": env.adminToken, "Content-Type": "application/json"},
		ExpectedStatus:        http.StatusOK,
		ExpectedContent:       []string{`"address":"renamed"`},
		TestAppFactory:        func(_ testing.TB) *tests.TestApp { return env.app },
		DisableTestAppCleanup: true,
	}
	scenario.Test(t)
}

func TestMailboxAdminRLS_AdminStillCannotReadMailboxContent(t *testing.T) {
	env := setupMailboxAdminApp(t)

	threads := core.NewBaseCollection("mail_threads")
	threads.Fields.Add(&core.RelationField{
		Name: "mailbox", Required: true, CollectionId: "pbc_mail_mailboxes_01",
		CascadeDelete: true, MaxSelect: 1,
	})
	threads.Fields.Add(&core.TextField{Name: "subject", Required: true})
	if err := env.app.Save(threads); err != nil {
		t.Fatal(err)
	}

	thread := core.NewRecord(threads)
	thread.Set("mailbox", env.mailbox.Id)
	thread.Set("subject", "Private message")
	if err := env.app.Save(thread); err != nil {
		t.Fatal(err)
	}
	mailSetListView(t, env.app, "mail_threads", mailContentMemberRead)

	mailRunList(t, env.app, "mail_threads", env.adminToken,
		[]string{`"totalItems":0`}, []string{"Private message"})
}

func TestMailboxAdminRLS_MailboxOwnerCanReadMailboxContent(t *testing.T) {
	env := setupMailboxAdminApp(t)

	threads := core.NewBaseCollection("mail_threads")
	threads.Fields.Add(&core.RelationField{
		Name: "mailbox", Required: true, CollectionId: "pbc_mail_mailboxes_01",
		CascadeDelete: true, MaxSelect: 1,
	})
	threads.Fields.Add(&core.TextField{Name: "subject", Required: true})
	if err := env.app.Save(threads); err != nil {
		t.Fatal(err)
	}

	thread := core.NewRecord(threads)
	thread.Set("mailbox", env.mailbox.Id)
	thread.Set("subject", "Private message")
	if err := env.app.Save(thread); err != nil {
		t.Fatal(err)
	}
	mailSetListView(t, env.app, "mail_threads", mailContentMemberRead)

	mailRunList(t, env.app, "mail_threads", env.ownerToken,
		[]string{`"totalItems":1`, "Private message"}, nil)
}
