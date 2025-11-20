// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package db2

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/database/dbplugin/v5"
	"github.com/hashicorp/vault/sdk/database/helper/dbutil"
	"github.com/hashicorp/vault/sdk/helper/strutil"
	_ "github.com/ibmdb/go_ibm_db"
	"github.com/mitchellh/mapstructure"
)

const (
	db2TypeName = "db2"

	defaultChangePasswordStatement = `ALTER USER "{{username}}" IDENTIFIED BY "{{password}}"`
)

var _ dbplugin.Database = (*db2DB)(nil)

// db2DB implements the Database interface for IBM DB2
type db2DB struct {
	logger hclog.Logger
	config config
	db     *sql.DB
	mux    sync.RWMutex
}

// config holds the connection configuration for DB2
type config struct {
	// ConnectionURL is the DB2 connection string
	// Format: DATABASE=<database>;HOSTNAME=<host>;PORT=<port>;PROTOCOL=TCPIP;UID=<username>;PWD=<password>
	ConnectionURL string `mapstructure:"connection_url"`

	// Username for the DB2 connection (can also be embedded in connection_url)
	Username string `mapstructure:"username"`

	// Password for the DB2 connection (can also be embedded in connection_url)
	Password string `mapstructure:"password"`

	// MaxOpenConnections limits the number of open connections to the database
	MaxOpenConnections int `mapstructure:"max_open_connections"`

	// MaxIdleConnections limits the number of idle connections
	MaxIdleConnections int `mapstructure:"max_idle_connections"`

	// MaxConnectionLifetime limits the maximum amount of time a connection may be reused
	MaxConnectionLifetimeRaw interface{} `mapstructure:"max_connection_lifetime"`
}

// Type returns the type name of the database
func (d *db2DB) Type() (string, error) {
	return db2TypeName, nil
}

// Initialize configures the database connection
func (d *db2DB) Initialize(ctx context.Context, req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
	d.mux.Lock()
	defer d.mux.Unlock()

	// Close any existing connection
	if d.db != nil {
		d.db.Close()
		d.db = nil
	}

	// Decode configuration
	var cfg config
	if err := mapstructure.WeakDecode(req.Config, &cfg); err != nil {
		return dbplugin.InitializeResponse{}, fmt.Errorf("failed to decode configuration: %w", err)
	}

	// Validate required fields
	if cfg.ConnectionURL == "" {
		return dbplugin.InitializeResponse{}, fmt.Errorf("connection_url is required")
	}

	d.config = cfg

	// Build connection string
	connStr := d.buildConnectionString()

	// Open DB2 connection
	db, err := sql.Open("go_ibm_db", connStr)
	if err != nil {
		return dbplugin.InitializeResponse{}, fmt.Errorf("failed to open DB2 connection: %w", err)
	}

	// Configure connection pool
	if cfg.MaxOpenConnections > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConnections)
	}

	d.db = db

	// Verify connection if requested
	if req.VerifyConnection {
		if err := d.db.PingContext(ctx); err != nil {
			d.db.Close()
			d.db = nil
			return dbplugin.InitializeResponse{}, fmt.Errorf("failed to verify DB2 connection: %w", err)
		}
		d.logger.Debug("DB2 connection verified successfully")
	}

	resp := dbplugin.InitializeResponse{
		Config: req.Config,
	}

	return resp, nil
}

// buildConnectionString constructs the DB2 connection string
func (d *db2DB) buildConnectionString() string {
	connStr := d.config.ConnectionURL

	// If username and password are provided separately, append them
	if d.config.Username != "" && d.config.Password != "" {
		// Check if connection URL already has credentials
		if !containsCredentials(connStr) {
			connStr = fmt.Sprintf("%s;UID=%s;PWD=%s", connStr, d.config.Username, d.config.Password)
		}
	}

	return connStr
}

// containsCredentials checks if the connection string already contains credentials
func containsCredentials(connStr string) bool {
	return strutil.StrListContainsGlob([]string{connStr}, "*UID=*") ||
		strutil.StrListContainsGlob([]string{connStr}, "*PWD=*")
}

// NewUser creates a new user - not supported for static credentials
func (d *db2DB) NewUser(ctx context.Context, req dbplugin.NewUserRequest) (dbplugin.NewUserResponse, error) {
	return dbplugin.NewUserResponse{}, fmt.Errorf("NewUser is not supported for DB2 static credentials plugin")
}

// UpdateUser updates user credentials (password rotation for static roles)
func (d *db2DB) UpdateUser(ctx context.Context, req dbplugin.UpdateUserRequest) (dbplugin.UpdateUserResponse, error) {
	d.mux.RLock()
	defer d.mux.RUnlock()

	if d.db == nil {
		return dbplugin.UpdateUserResponse{}, fmt.Errorf("DB2 connection not initialized")
	}

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

		d.logger.Debug("executing password rotation statement", "username", username)

		if _, err := d.db.ExecContext(ctx, query); err != nil {
			return dbplugin.UpdateUserResponse{}, fmt.Errorf("failed to update password for user %s: %w", username, err)
		}
	}

	d.logger.Info("successfully rotated password", "username", username)

	return dbplugin.UpdateUserResponse{}, nil
}

// DeleteUser deletes a user - not supported for static credentials
func (d *db2DB) DeleteUser(ctx context.Context, req dbplugin.DeleteUserRequest) (dbplugin.DeleteUserResponse, error) {
	return dbplugin.DeleteUserResponse{}, fmt.Errorf("DeleteUser is not supported for DB2 static credentials plugin")
}

// Close closes the database connection
func (d *db2DB) Close() error {
	d.mux.Lock()
	defer d.mux.Unlock()

	if d.db != nil {
		err := d.db.Close()
		d.db = nil
		return err
	}

	return nil
}
