/// <reference path="../../../server/pb_data/types.d.ts" />
// Organization admins need to manage mailbox inventory without automatically
// gaining access to mailbox contents. The previous rules only exposed a
// mailbox and its membership rows to mailbox members, so the administrative
// Mailboxes screen reported zero personal mailboxes for admins who correctly
// were not members of employees' personal mailboxes.
//
// These rules expose mailbox and membership metadata to an admin/owner in the
// mailbox's organization and allow them to maintain that metadata. Thread,
// message, and thread-state rules remain unchanged and membership-gated.
migrate(
  (app) => {
    const mailboxMember =
      "mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id";
    const mailboxOwner =
      'mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id && mail_mailbox_members_via_mailbox.role ?= "owner"';
    const orgAdminViaDomain =
      'domain.org.user_org_via_org.user ?= @request.auth.id && (domain.org.user_org_via_org.role ?= "admin" || domain.org.user_org_via_org.role ?= "owner")';

    const mailboxes = app.findCollectionByNameOrId("mail_mailboxes");
    mailboxes.listRule = `(${mailboxMember}) || (${orgAdminViaDomain})`;
    mailboxes.viewRule = `(${mailboxMember}) || (${orgAdminViaDomain})`;
    mailboxes.updateRule = `(${mailboxOwner}) || (${orgAdminViaDomain})`;
    mailboxes.deleteRule = `(${mailboxOwner}) || (${orgAdminViaDomain})`;
    app.save(mailboxes);

    const selfMember = "user_org.user = @request.auth.id";
    const mailboxOwnerViaMailbox =
      'mailbox.mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id && mailbox.mail_mailbox_members_via_mailbox.role ?= "owner"';
    const orgAdminViaMailbox =
      'mailbox.domain.org.user_org_via_org.user ?= @request.auth.id && (mailbox.domain.org.user_org_via_org.role ?= "admin" || mailbox.domain.org.user_org_via_org.role ?= "owner")';
    const sameOrg = "user_org.org = mailbox.domain.org";
    const ownerCanAdd = `${mailboxOwnerViaMailbox} && ${sameOrg}`;
    const adminCanAdd = `${orgAdminViaMailbox} && ${sameOrg}`;
    const bootstrapFirstOwner =
      'user_org.user = @request.auth.id && role = "owner" && mailbox.mail_mailbox_members_via_mailbox.id = "" && mailbox.domain.org.user_org_via_org.user ?= @request.auth.id && mailbox.domain.org.user_org_via_org.role ?!= "guest"';

    const members = app.findCollectionByNameOrId("mail_mailbox_members");
    members.listRule = `(${selfMember}) || (${orgAdminViaMailbox})`;
    members.viewRule = `(${selfMember}) || (${orgAdminViaMailbox})`;
    members.createRule = `(${ownerCanAdd}) || (${adminCanAdd}) || (${bootstrapFirstOwner})`;
    members.updateRule = `(${mailboxOwnerViaMailbox}) || (${orgAdminViaMailbox})`;
    members.deleteRule = `(${selfMember}) || (${mailboxOwnerViaMailbox}) || (${orgAdminViaMailbox})`;
    app.save(members);
  },
  (app) => {
    const mailboxMember =
      "mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id";
    const mailboxOwner =
      'mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id && mail_mailbox_members_via_mailbox.role ?= "owner"';

    const mailboxes = app.findCollectionByNameOrId("mail_mailboxes");
    mailboxes.listRule = mailboxMember;
    mailboxes.viewRule = mailboxMember;
    mailboxes.updateRule = mailboxOwner;
    mailboxes.deleteRule = mailboxOwner;
    app.save(mailboxes);

    const selfMember = "user_org.user = @request.auth.id";
    const ownerCanAdd =
      'mailbox.mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id && mailbox.mail_mailbox_members_via_mailbox.role ?= "owner" && user_org.org = mailbox.domain.org';
    const bootstrapFirstOwner =
      'user_org.user = @request.auth.id && role = "owner" && mailbox.mail_mailbox_members_via_mailbox.id = "" && mailbox.domain.org.user_org_via_org.user ?= @request.auth.id && mailbox.domain.org.user_org_via_org.role ?!= "guest"';

    const members = app.findCollectionByNameOrId("mail_mailbox_members");
    members.listRule = selfMember;
    members.viewRule = selfMember;
    members.createRule = `(${ownerCanAdd}) || (${bootstrapFirstOwner})`;
    members.updateRule =
      'mailbox.mail_mailbox_members_via_mailbox.user_org.user ?= @request.auth.id && mailbox.mail_mailbox_members_via_mailbox.role ?= "owner"';
    members.deleteRule = selfMember;
    app.save(members);
  },
);
