package sqlite

import "fmt"

type PermissionError struct {
	Path string
	Mode uint32
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("unsafe permissions on %s: mode %#o permits access by other users", e.Path, e.Mode)
}

type UnsafePathError struct {
	Path   string
	Reason string
}

func (e *UnsafePathError) Error() string {
	return fmt.Sprintf("unsafe storage path %s: %s", e.Path, e.Reason)
}

type CorruptError struct {
	Path  string
	Cause error
}

func (e *CorruptError) Error() string {
	return fmt.Sprintf("database %s is not healthy: %v", e.Path, e.Cause)
}
func (e *CorruptError) Unwrap() error { return e.Cause }

type FutureSchemaError struct {
	Path    string
	Found   int
	Current int
}

func (e *FutureSchemaError) Error() string {
	return fmt.Sprintf("database %s uses schema %d, newer than supported schema %d", e.Path, e.Found, e.Current)
}

type UnsupportedSchemaError struct{ Path string }

func (e *UnsupportedSchemaError) Error() string {
	return fmt.Sprintf("database %s has no recognized schema version", e.Path)
}

type LockedError struct{ Cause error }

func (e *LockedError) Error() string { return fmt.Sprintf("database remained locked: %v", e.Cause) }
func (e *LockedError) Unwrap() error { return e.Cause }

type ConflictError struct {
	Name  string
	Cause error
}

func (e *ConflictError) Error() string { return fmt.Sprintf("snippet name %q already exists", e.Name) }
func (e *ConflictError) Unwrap() error { return e.Cause }
