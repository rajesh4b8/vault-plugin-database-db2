# Vault Plugin Database DB2

A [HashiCorp Vault](https://www.vaultproject.io/) database secrets engine plugin for IBM DB2 databases. This plugin supports **static credential rotation** for DB2 database users.

## Features

- Static credential rotation for DB2 database users
- Customizable password rotation statements
- Connection pooling support
- Secure credential masking in logs

## Requirements

- HashiCorp Vault 1.12.0+
- IBM DB2 database
- Go 1.21+ (for building from source)
- IBM DB2 ODBC/CLI driver installed on the Vault server

## Installation

### Building from Source

1. Clone this repository:
   ```bash
   git clone https://github.com/hashicorp/vault-plugin-database-db2.git
   cd vault-plugin-database-db2
   ```

2. Build the plugin:
   ```bash
   make build
   ```

   The binary will be created at `bin/vault-plugin-database-db2`.

3. Move the binary to your Vault plugins directory:
   ```bash
   mv bin/vault-plugin-database-db2 /path/to/vault/plugins/
   ```

### Installing IBM DB2 Driver

The plugin requires the IBM DB2 ODBC/CLI driver to be installed. Follow the [IBM documentation](https://www.ibm.com/docs/en/db2/11.5?topic=apis-call-level-interface-guide-reference) to install the appropriate driver for your platform.

Set the following environment variables:
```bash
export IBM_DB_HOME=/path/to/clidriver
export CGO_CFLAGS=-I$IBM_DB_HOME/include
export CGO_LDFLAGS=-L$IBM_DB_HOME/lib
export LD_LIBRARY_PATH=$IBM_DB_HOME/lib:$LD_LIBRARY_PATH
```

## Configuration

### 1. Register the Plugin

```bash
vault plugin register -sha256=$(sha256sum /path/to/vault/plugins/vault-plugin-database-db2 | cut -d' ' -f1) \
    database vault-plugin-database-db2
```

### 2. Enable the Database Secrets Engine

```bash
vault secrets enable database
```

### 3. Configure the DB2 Connection

```bash
vault write database/config/my-db2-database \
    plugin_name=vault-plugin-database-db2 \
    allowed_roles="my-static-role" \
    connection_url="DATABASE=mydb;HOSTNAME=db2.example.com;PORT=50000;PROTOCOL=TCPIP" \
    username="admin" \
    password="admin-password"
```

#### Connection Configuration Options

| Parameter | Description | Required |
|-----------|-------------|----------|
| `connection_url` | DB2 connection string | Yes |
| `username` | Database username for connection | No (can be in connection_url) |
| `password` | Database password for connection | No (can be in connection_url) |
| `max_open_connections` | Maximum number of open connections | No |
| `max_idle_connections` | Maximum number of idle connections | No |
| `max_connection_lifetime` | Maximum lifetime of connections | No |

#### Connection URL Format

The connection URL follows the IBM DB2 CLI connection string format:
```
DATABASE=<database>;HOSTNAME=<host>;PORT=<port>;PROTOCOL=TCPIP;UID=<username>;PWD=<password>
```

You can either embed credentials in the connection URL or provide them separately via the `username` and `password` parameters.

### 4. Create a Static Role

```bash
vault write database/static-roles/my-static-role \
    db_name=my-db2-database \
    username="app_user" \
    rotation_period=86400
```

#### Static Role Configuration Options

| Parameter | Description | Required |
|-----------|-------------|----------|
| `db_name` | Name of the database connection | Yes |
| `username` | DB2 username to manage | Yes |
| `rotation_period` | How often to rotate the password (in seconds) | Yes |
| `rotation_statements` | Custom SQL for password rotation | No |

### 5. Custom Rotation Statements

By default, the plugin uses:
```sql
ALTER USER "{{username}}" IDENTIFIED BY "{{password}}"
```

You can customize this by providing your own rotation statements:
```bash
vault write database/static-roles/my-static-role \
    db_name=my-db2-database \
    username="app_user" \
    rotation_period=86400 \
    rotation_statements="CALL SYSPROC.AUTH_SET_PASSWORD('{{username}}', '{{password}}')"
```

## Usage

### Get Static Credentials

```bash
vault read database/static-creds/my-static-role
```

Example output:
```
Key                    Value
---                    -----
last_vault_rotation    2024-01-15T10:30:00.000000Z
password               A1a-xxxxxxxxxxxxxxxx
rotation_period        24h
ttl                    23h59m55s
username               app_user
```

### Manually Rotate Credentials

```bash
vault write -f database/rotate-static-creds/my-static-role
```

## Development

### Prerequisites

- Go 1.21+
- IBM DB2 ODBC/CLI driver
- Make

### Build

```bash
make build
```

### Test

```bash
make test
```

### Format Code

```bash
make fmt
```

### Cross-compile

```bash
make build-all
```

## Troubleshooting

### Common Issues

1. **Connection Failed**: Ensure the IBM DB2 driver is properly installed and environment variables are set.

2. **Permission Denied**: The configured database user must have permission to alter passwords for the target users.

3. **Plugin Registration Failed**: Verify the SHA256 hash matches the plugin binary.

### Enabling Debug Logging

Set Vault's log level to trace:
```bash
vault server -log-level=trace
```

## Limitations

- **Static roles only**: This plugin only supports static credential rotation. Dynamic credential creation (NewUser) is not supported.
- **User deletion not supported**: The DeleteUser operation is not implemented as this is a static credentials plugin.

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please read the contributing guidelines and submit pull requests to the repository.

## Related Projects

- [vault-plugin-database-redis-elasticache](https://github.com/hashicorp/vault-plugin-database-redis-elasticache)
- [HashiCorp Vault](https://github.com/hashicorp/vault)
- [Vault Database Secrets Engine](https://www.vaultproject.io/docs/secrets/databases)
