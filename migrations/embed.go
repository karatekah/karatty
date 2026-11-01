package migrations

import _ "embed"

//go:embed 001_initial.sql
var Initial string

//go:embed 002_retention.sql
var Retention string

//go:embed 003_account_deletion.sql
var AccountDeletion string

//go:embed 004_direct_consent.sql
var DirectConsent string

//go:embed 005_email_identity.sql
var EmailIdentity string

//go:embed 006_anonymous_identity.sql
var AnonymousIdentity string

//go:embed 007_e2ee_keys.sql
var E2EEKeys string
