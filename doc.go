/*
Package fs provides a filesystem-backed Kivik driver. This driver is very
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

would look for document files in `/home/usr/some/path/foo`.

Connection Strings

This driver supports three types of connection strings to the New() method:

 - Local filesystem path. This may be relative or absolute. Examples:
   `/home/user/some/path` and './some/path'
 - A full file:// URL. Example: 'file:///home/user/some/path'
 - An empty string (""), which requires the full path to the database is passed
   to the `DB()` method. In this case, the argument to `DB()` accepts the first
   two formats, with the final path element being the database name. Some
   client-level methods, such as AllDBs(), are unavailable, when using an empty
   connection string.

Handling of Filenames

CouchDB allows databases and document IDs to contain a slash (/)
character. This is not permitted in most operating systems/filenames, to
be stored directly on the filesystem this way. Therefore, it is necessary
for this package to escape certain characters in filenames. This is done
as conservatively as possible. The escaping rules are:

 - It contains a slash (i.e. '_design/index'), or a URL-encoded slash
   (i.e. '%2F' or '%2f')
 - When escaping a literal slash (/) or a literal percent sign (%), are
   escaped using standard URL escaping. No other characters are escaped.
*/
package fs
