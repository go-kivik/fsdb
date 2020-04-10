package test

import (
	"net/http"

	"github.com/go-kivik/kiviktest/v3"
	"github.com/go-kivik/kiviktest/v3/kt"
)

func registerFSSuite() {
	kiviktest.RegisterSuite(kiviktest.SuiteKivikFS, kt.SuiteConfig{
		"AllDBs.expected": []string{},

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"AllDocs.skip": true, // FIXME: Not yet implemented
		// "AllDocs/Admin.databases":  []string{"foo"},
		// "AllDocs/Admin/foo.status": http.StatusNotFound,

		"DBExists/Admin.databases":       []string{"chicken"},
		"DBExists/Admin/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists": true,

		"DestroyDB/RW/Admin/NonExistantDB.status": http.StatusNotFound,

		"Version.version":        `^0\.0\.1$`,
		"Version.vendor":         "Kivik",
		"Version.vendor_version": `^0\.0\.1$`,

		// Replications not to be implemented
		"GetReplications.skip": true,
		"Replicate.skip":       true,

		"Get/RW/group/Admin/bogus.status":  http.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": http.StatusNotFound,

		"GetMeta.skip":           true,                      // FIXME: Unimplemented
		"Flush.skip":             true,                      // FIXME: Unimplemented
		"Delete.skip":            true,                      // FIXME: Unimplemented
		"Stats.skip":             true,                      // FIXME: Unimplemented
		"CreateDoc.skip":         true,                      // FIXME: Unimplemented
		"Compact.skip":           true,                      // FIXME: Unimplemented
		"Security.skip":          true,                      // FIXME: Unimplemented
		"DBUpdates.status":       http.StatusNotImplemented, // FIXME: Unimplemented
		"Changes.skip":           true,                      // FIXME: Unimplemented
		"Copy.skip":              true,                      // FIXME: Unimplemented, depends on Get/Put or Copy
		"BulkDocs.skip":          true,                      // FIXME: Unimplemented
		"GetAttachment.skip":     true,                      // FIXME: Unimplemented
		"GetAttachmentMeta.skip": true,                      // FIXME: Unimplemented
		"PutAttachment.skip":     true,                      // FIXME: Unimplemented
		"DeleteAttachment.skip":  true,                      // FIXME: Unimplemented
		"Query.skip":             true,                      // FIXME: Unimplemented
		"Find.skip":              true,                      // FIXME: Unimplemented
		"Explain.skip":           true,                      // FIXME: Unimplemented
		"CreateIndex.skip":       true,                      // FIXME: Unimplemented
		"GetIndexes.skip":        true,                      // FIXME: Unimplemented
		"DeleteIndex.skip":       true,                      // FIXME: Unimplemented

		"Put/RW/Admin/group/LeadingUnderscoreInID.status":  http.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":               http.StatusConflict,
		"Put/RW/NoAuth/group/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/NoAuth/group/DesignDoc.status":             http.StatusUnauthorized,
		"Put/RW/NoAuth/group/Conflict.status":              http.StatusConflict,

		"SetSecurity.skip": true, // FIXME: Unimplemented
		"ViewCleanup.skip": true, // FIXME: Unimplemented
		"Rev.skip":         true, // FIXME: Unimplemented
	})
}
