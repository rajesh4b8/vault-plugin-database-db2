// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package db2

import (
	"github.com/hashicorp/vault/sdk/database/helper/dbutil"
)

// New creates a new instance of the DB2 database plugin
func New() (interface{}, error) {
	db := newDB2()

	// Wrap with error sanitization middleware
	dbType := dbutil.NewDatabaseErrorSanitizerMiddleware(db, db.SecretValues)

	return dbType, nil
}
