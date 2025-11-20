// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package db2

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/database/dbplugin/v5"
	"github.com/hashicorp/vault/sdk/database/helper/dbutil"
)

// New creates a new instance of the DB2 database plugin
func New() (interface{}, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	db := &db2DB{
		logger: logger,
	}

	return wrapWithSanitizerMiddleware(db), nil
}

// wrapWithSanitizerMiddleware wraps the database with error sanitization
func wrapWithSanitizerMiddleware(db *db2DB) dbplugin.Database {
	return dbutil.NewDatabaseErrorSanitizerMiddleware(db, db.secretValuesToMask)
}

// secretValuesToMask returns sensitive values that should be masked in logs/errors
func (d *db2DB) secretValuesToMask() map[string]string {
	return map[string]string{
		d.config.Password: "[password]",
		d.config.Username: "[username]",
	}
}
