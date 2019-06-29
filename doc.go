/*
Package fsdb provides a filesystem-backed Kivik driver. This driver is very
much a work in progress. Please refer to the GitHub page for current status and
ongoing changes. https://github.com/go-kivik/fsdb

Bug reports, feature requests, and pull requests are always welcome. Current
development is primarily focused around using fsdb for testing of CouchDB
applications, and bootstraping CouchDB applications.

General Usage

Use the `fs` driver name when using this driver. The DSN should be an existing
directory on the local filesystem. Access control is managed by your filesystem
permissions.

    import (
        "github.com/go-kivik/kivik"
        _ "github.com/go-kivik/fsdb" // The Filesystem driver
    )

    client, err := kivik.New("fs", "/home/user/some/path")

Database names represent directories under the path provided to `kivik.New`.
For example:

    db := client.DB(ctx, "foo")

would look for document files in `/home/usr/some/path/foo`
*/
package fs
