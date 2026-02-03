# Built-in Benchmark Templates

This directory contains built-in benchmark templates for DB-BenchMind.

## Template List

### Sysbench Templates

#### MySQL Templates

| ID | Name | Description | Default | Tables | Table Size |
|----|------|-------------|---------|--------|------------|
| `sysbench-mysql-test` ⭐ | Test | Lightweight test template for quick testing | Yes | 10 | 10,000 |
| `sysbench-mysql-cpu-bound` | CPU Bound | CPU-bound test (data fits in memory) | No | 10 | 10,000,000 |
| `sysbench-mysql-disk-bound` | Disk Bound | Disk-bound test (data exceeds memory) | No | 50 | 10,000,000 |

#### PostgreSQL Templates

| ID | Name | Description | Default | Tables | Table Size |
|----|------|-------------|---------|--------|------------|
| `sysbench-postgresql-test` ⭐ | Test | Lightweight test template for quick testing | Yes | 10 | 10,000 |
| `sysbench-postgresql-cpu-bound` | CPU Bound | CPU-bound test (data fits in memory) | No | 10 | 10,000,000 |
| `sysbench-postgresql-disk-bound` | Disk Bound | Disk-bound test (data exceeds memory) | No | 50 | 10,000,000 |

#### Legacy Templates (Deprecated)

| ID | Name | Description | Supported Databases |
|----|------|-------------|---------------------|
| `sysbench-oltp-read-write` | Sysbench OLTP Read-Write | Standard OLTP mixed test (70% read, 30% write) | MySQL, PostgreSQL |
| `sysbench-oltp-read-only` | Sysbench OLTP Read-Only | Pure read test (100% SELECT) | MySQL, PostgreSQL |
| `sysbench-oltp-write-only` | Sysbench OLTP Write-Only | Pure write test (INSERT/UPDATE/DELETE) | MySQL, PostgreSQL |

### Swingbench Templates

| ID | Name | Description | Supported Databases |
|----|------|-------------|---------------------|
| `swingbench-soe` | Swingbench Order Entry | Simulates order processing system | Oracle |
| `swingbench-calling` | Swingbench Calling Circle | Simulates telecom calling system | Oracle |

### HammerDB Templates

| ID | Name | Description | Supported Databases |
|----|------|-------------|---------------------|
| `hammerdb-tpcc` | HammerDB TPROC-C | Standard TPC-C benchmark | MySQL, PostgreSQL, Oracle, SQL Server |
| `hammerdb-tpcb` | HammerDB TPROC-B | Standard TPC-B benchmark | MySQL, PostgreSQL, Oracle, SQL Server |

## Template Schema

Each template file follows this JSON schema:

```json
{
  "$schema": "https://db-benchmind.dev/schemas/template/v1.json",
  "id": "unique-template-id",
  "name": "Template Display Name",
  "description": "Template description",
  "tool": "sysbench|swingbench|hammerdb|tpcc",
  "database_types": ["mysql", "postgresql", "oracle", "sqlserver"],
  "version": "1.0.0",
  "parameters": {
    "param_name": {
      "type": "integer|string|boolean|enum",
      "label": "Display label",
      "default": "default value",
      "min": 1,
      "max": 1000,
      "options": ["option1", "option2"]
    }
  },
  "command_template": {
    "prepare": "prepare command template",
    "run": "run command template",
    "cleanup": "cleanup command template"
  },
  "output_parser": {
    "type": "regex|json|csv",
    "patterns": {
      "metric_name": "regex pattern"
    }
  }
}
```

## Parameter Types

- **integer**: Numeric value with optional min/max constraints
- **string**: Free-form text input
- **boolean**: Checkbox (true/false)
- **enum**: Dropdown selection from predefined options

## Command Template Variables

The following variables are automatically substituted in command templates:

- `{db_type}`: Database type (mysql, postgresql, oracle, sqlserver)
- `{connection_string}`: Database connection string
- `{user}`: Database username
- `{password}`: Database password (injected only during execution)
- `{param_name}`: Any parameter defined in the template

## Output Parser Types

- **regex**: Extract metrics using regular expression patterns
- **json**: Parse JSON output (requires structured format)
- **csv**: Parse CSV output (requires header row)

## Adding New Templates

To add a new template:

1. Create a new JSON file in this directory
2. Follow the template schema above
3. Validate using: `go run ./cmd/validate-templates/main.go`
4. Test the template before committing

## Related Requirements

- REQ-TMPL-001: Display built-in template list
- REQ-TMPL-002: Display template details
- REQ-TMPL-003: Import custom templates
- REQ-TMPL-004: Validate template definitions
- REQ-TMPL-007: Save template snapshot for reproducibility
