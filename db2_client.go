// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package db2

import (
	"context"
	"database/sql"
	"fmt"

	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
	"github.com/hashicorp/vault/sdk/database/helper/connutil"
	"github.com/hashicorp/vault/sdk/database/helper/dbutil"
	_ "github.com/ibmdb/go_ibm_db"
)

const (
	db2TypeName = "db2"

	defaultChangePasswordStatement = `ALTER USER "{{username}}" IDENTIFIED BY "{{password}}"`
)

var _ dbplugin.Database = (*db2DB)(nil)

// db2DB implements the Database interface for IBM DB2
type db2DB struct {
	*db2ConnectionProducer
}

// db2ConnectionProducer implements ConnectionProducer and provides a connection producer for DB2
type db2ConnectionProducer struct {
	*connutil.SQLConnectionProducer
}

// newDB2 creates a new DB2 database instance
func newDB2() *db2DB {
	connProducer := &db2ConnectionProducer{
		SQLConnectionProducer: &connutil.SQLConnectionProducer{},
	}
	connProducer.Type = db2TypeName

	return &db2DB{
		db2ConnectionProducer: connProducer,
	}
}

// Type returns the type name of the database
func (d *db2DB) Type() (string, error) {
	return db2TypeName, nil
}

// Initialize configures the database connection
func (d *db2DB) Initialize(ctx context.Context, req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
	newConf, err := d.db2ConnectionProducer.Init(ctx, req.Config, req.VerifyConnection)
	if err != nil {
		return dbplugin.InitializeResponse{}, err
	}

	resp := dbplugin.InitializeResponse{
		Config: newConf,
	}

	return resp, nil
}

// NewUser creates a new user - not supported for static credentials
func (d *db2DB) NewUser(ctx context.Context, req dbplugin.NewUserRequest) (dbplugin.NewUserResponse, error) {
	return dbplugin.NewUserResponse{}, fmt.Errorf("NewUser is not supported for DB2 static credentials plugin")
}

// UpdateUser updates user credentials (password rotation for static roles)
func (d *db2DB) UpdateUser(ctx context.Context, req dbplugin.UpdateUserRequest) (dbplugin.UpdateUserResponse, error) {
	if req.Password == nil {
		return dbplugin.UpdateUserResponse{}, nil
	}

	username := req.Username
	newPassword := req.Password.NewPassword

	if username == "" {
		return dbplugin.UpdateUserResponse{}, fmt.Errorf("username is required")
	}

	if newPassword == "" {
		return dbplugin.UpdateUserResponse{}, fmt.Errorf("new password is required")
	}

	// Get connection from the connection producer
	dbConn, err := d.Connection(ctx)
	if err != nil {
		return dbplugin.UpdateUserResponse{}, err
	}

	// Type assert to *sql.DB
	db, ok := dbConn.(*sql.DB)
	if !ok {
		return dbplugin.UpdateUserResponse{}, fmt.Errorf("unable to use connection")
	}

	// Get the password change statements
	statements := req.Password.Statements.Commands
	if len(statements) == 0 {
		statements = []string{defaultChangePasswordStatement}
	}

	// Execute password change statements
	for _, stmt := range statements {
		// Replace placeholders
		query := dbutil.QueryHelper(stmt, map[string]string{
			"username": username,
			"password": newPassword,
		})

		if _, err := db.ExecContext(ctx, query); err != nil {
			return dbplugin.UpdateUserResponse{}, fmt.Errorf("failed to update password for user %s: %w", username, err)
		}
	}

	return dbplugin.UpdateUserResponse{}, nil
}

// DeleteUser deletes a user - not supported for static credentials
func (d *db2DB) DeleteUser(ctx context.Context, req dbplugin.DeleteUserRequest) (dbplugin.DeleteUserResponse, error) {
	return dbplugin.DeleteUserResponse{}, fmt.Errorf("DeleteUser is not supported for DB2 static credentials plugin")
}

// secretValues returns the secret values as a map of string to string for error sanitization
func (d *db2DB) secretValues() map[string]string {
	secretValuesMap := d.db2ConnectionProducer.SecretValues()
	result := make(map[string]string)
	for k, v := range secretValuesMap {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}
