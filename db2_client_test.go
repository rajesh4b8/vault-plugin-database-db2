// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package db2

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

func TestNew(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("failed to create new db2 plugin: %v", err)
	}

	if db == nil {
		t.Fatal("expected non-nil database")
	}
}

func TestType(t *testing.T) {
	db := newDB2()
	typ, err := db.Type()
	if err != nil {
		t.Fatalf("failed to get type: %v", err)
	}

	if typ != "db2" {
		t.Errorf("expected type 'db2', got '%s'", typ)
	}
}

func TestInitialize_MissingConnectionURL(t *testing.T) {
	db := newDB2()

	req := dbplugin.InitializeRequest{
		Config: map[string]interface{}{
			"username": "testuser",
			"password": "testpass",
		},
		VerifyConnection: false,
	}

	_, err := db.Initialize(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing connection_url")
	}
}

func TestNewUser_NotSupported(t *testing.T) {
	db := newDB2()

	req := dbplugin.NewUserRequest{
		UsernameConfig: dbplugin.UsernameMetadata{
			DisplayName: "test",
			RoleName:    "test",
		},
	}

	_, err := db.NewUser(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for unsupported NewUser operation")
	}
}

func TestDeleteUser_NotSupported(t *testing.T) {
	db := newDB2()

	req := dbplugin.DeleteUserRequest{
		Username: "testuser",
	}

	_, err := db.DeleteUser(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for unsupported DeleteUser operation")
	}
}

func TestUpdateUser_NotInitialized(t *testing.T) {
	db := newDB2()

	req := dbplugin.UpdateUserRequest{
		Username: "testuser",
		Password: &dbplugin.ChangePassword{
			NewPassword: "newpassword",
		},
	}

	_, err := db.UpdateUser(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for uninitialized connection")
	}
}

func TestClose_NoConnection(t *testing.T) {
	db := newDB2()

	err := db.Close()
	if err != nil {
		t.Fatalf("unexpected error closing nil connection: %v", err)
	}
}

func TestInitialize_Success(t *testing.T) {
	db := newDB2()

	req := dbplugin.InitializeRequest{
		Config: map[string]interface{}{
			"connection_url":       "DATABASE=testdb;HOSTNAME=localhost;PORT=50000",
			"username":             "testuser",
			"password":             "testpass",
			"max_open_connections": 5,
			"max_idle_connections": 2,
		},
		VerifyConnection: false,
	}

	resp, err := db.Initialize(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error during initialization: %v", err)
	}

	if resp.Config == nil {
		t.Fatal("expected non-nil config in response")
	}

	// Verify the connection producer was properly initialized
	if db.db2ConnectionProducer == nil {
		t.Fatal("expected connection producer to be initialized")
	}
}

func TestUpdateUser_NoPassword(t *testing.T) {
	db := newDB2()

	// Initialize first
	req := dbplugin.InitializeRequest{
		Config: map[string]interface{}{
			"connection_url": "DATABASE=testdb;HOSTNAME=localhost;PORT=50000",
			"username":       "testuser",
			"password":       "testpass",
		},
		VerifyConnection: false,
	}

	_, err := db.Initialize(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// Test UpdateUser with no password change
	updateReq := dbplugin.UpdateUserRequest{
		Username: "testuser",
		Password: nil,
	}

	_, err = db.UpdateUser(context.Background(), updateReq)
	if err != nil {
		t.Fatalf("unexpected error when Password is nil: %v", err)
	}
}

func TestUpdateUser_EmptyUsername(t *testing.T) {
	db := newDB2()

	req := dbplugin.UpdateUserRequest{
		Username: "",
		Password: &dbplugin.ChangePassword{
			NewPassword: "newpassword",
		},
	}

	_, err := db.UpdateUser(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty username")
	}

	if err.Error() != "username is required" {
		t.Errorf("expected 'username is required' error, got: %v", err)
	}
}

func TestUpdateUser_EmptyPassword(t *testing.T) {
	db := newDB2()

	req := dbplugin.UpdateUserRequest{
		Username: "testuser",
		Password: &dbplugin.ChangePassword{
			NewPassword: "",
		},
	}

	_, err := db.UpdateUser(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty password")
	}

	if err.Error() != "new password is required" {
		t.Errorf("expected 'new password is required' error, got: %v", err)
	}
}

func TestSecretValues(t *testing.T) {
	db := newDB2()

	// Initialize with test config
	req := dbplugin.InitializeRequest{
		Config: map[string]interface{}{
			"connection_url": "DATABASE=testdb;HOSTNAME=localhost;PORT=50000",
			"username":       "testuser",
			"password":       "testpass",
		},
		VerifyConnection: false,
	}

	_, err := db.Initialize(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}

	// SecretValues is provided by SQLConnectionProducer
	secrets := db.SecretValues()
	if secrets == nil {
		t.Fatal("expected non-nil secret values map")
	}

	// The SQLConnectionProducer should mask the password
	if _, exists := secrets["testpass"]; !exists {
		t.Error("expected password to be in secret values")
	}
}

func TestConnectionProducer_Type(t *testing.T) {
	db := newDB2()

	if db.db2ConnectionProducer == nil {
		t.Fatal("expected connection producer to be initialized")
	}

	if db.db2ConnectionProducer.Type != "db2" {
		t.Errorf("expected connection producer type to be 'db2', got: %s", db.db2ConnectionProducer.Type)
	}
}
