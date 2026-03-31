# Repository Layer Tests

## Overview

This directory contains repository layer tests using SQLite in-memory database. These tests provide code coverage for the repository CRUD operations while maintaining fast test execution.

## SQLite vs PostgreSQL Differences

### ILIKE Operator
The production repository code uses PostgreSQL's `ILIKE` operator for case-insensitive pattern matching. SQLite does not support `ILIKE`.

**Impact on Tests:**
- Tests that verify filtering/search functionality using ILIKE will fail in SQLite
- Core CRUD operations (Create, Read by ID, Update, Delete) are fully tested
- List operations without search filters are tested
- Filter tests are documented as SQLite-incompatible

**Workaround:**
The repository tests focus on:
1. ✅ Create operations - fully tested
2. ✅ GetByID operations - fully tested  
3. ✅ List operations without filters - fully tested
4. ✅ Update operations - fully tested
5. ⚠️  List operations with search queries - limited by ILIKE incompatibility

### UUID Generation
PostgreSQL supports `UUID` type with `gen_random_uuid()` default. SQLite uses TEXT for IDs.

**Solution:** Test database tables are manually created with TEXT PRIMARY KEY instead of using GORM AutoMigrate.

##Coverage

The following repository packages have tests:
- ✅ `internal/unit` - Unit repository tests
- ✅ `internal/department` - Department repository tests  
- ✅ `internal/patient` - Patient repository tests
- ✅ `internal/encounter` - Encounter repository tests
- ✅ `internal/bed` - Bed, BedStatusEvent, BedRequest repository tests  
- ✅ `internal/transport` - TransportRequest, TransportChangeEvent repository tests

## Running Tests

```bash
# Run all repository tests
go test ./internal/unit/... ./internal/department/... ./internal/patient/... ./internal/encounter/... ./internal/bed/... ./internal/transport/... -v -run TestRepository

# Run specific package
go test ./internal/unit/... -v -run TestRepository

# With coverage
go test ./internal/unit/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Structure

Each repository test file follows this pattern:

```go
func newRepositoryTestDB(t *testing.T) *database.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        DisableForeignKeyConstraintWhenMigrating: true,
    })
    require.NoError(t, err)
    
    // Create tables manually for SQLite compatibility
    err = db.Exec(`CREATE TABLE ...`).Error
    require.NoError(t, err)
    
    return &database.DB{DB: db}
}
```

## Notes

- Tests use simple string IDs (e.g., "test-id-1") instead of UUIDs
- Foreign key constraints are disabled in test database
- Tests are white-box (same package as production code)
- Each test uses a fresh in-memory database
