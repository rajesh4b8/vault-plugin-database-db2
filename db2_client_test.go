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
	db := &db2DB{}
	typ, err := db.Type()
	if err != nil {
		t.Fatalf("failed to get type: %v", err)
	}

	if typ != "db2" {
		t.Errorf("expected type 'db2', got '%s'", typ)
	}
}

func TestInitialize_MissingConnectionURL(t *testing.T) {
	db := &db2DB{}

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
	db := &db2DB{}

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
	db := &db2DB{}

	req := dbplugin.DeleteUserRequest{
		Username: "testuser",
	}

	_, err := db.DeleteUser(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for unsupported DeleteUser operation")
	}
}

func TestUpdateUser_NotInitialized(t *testing.T) {
	db := &db2DB{}

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
	db := &db2DB{}

	err := db.Close()
	if err != nil {
		t.Fatalf("unexpected error closing nil connection: %v", err)
	}
}

func TestSecretValuesToMask(t *testing.T) {
	db := &db2DB{
		config: config{
			Username: "testuser",
			Password: "testpass",
		},
	}

	secrets := db.secretValuesToMask()

	if secrets["testuser"] != "[username]" {
		t.Error("expected username to be masked")
	}

	if secrets["testpass"] != "[password]" {
		t.Error("expected password to be masked")
	}
}

func TestContainsCredentials(t *testing.T) {
	tests := []struct {
		name     string
		connStr  string
		expected bool
	}{
		{
			name:     "contains UID",
			connStr:  "DATABASE=mydb;HOSTNAME=localhost;UID=user;PWD=pass",
			expected: true,
		},
		{
			name:     "contains PWD only",
			connStr:  "DATABASE=mydb;HOSTNAME=localhost;PWD=pass",
			expected: true,
		},
		{
			name:     "no credentials",
			connStr:  "DATABASE=mydb;HOSTNAME=localhost;PORT=50000",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsCredentials(tt.connStr)
			if result != tt.expected {
				t.Errorf("containsCredentials(%s) = %v, expected %v", tt.connStr, result, tt.expected)
			}
		})
	}
}
