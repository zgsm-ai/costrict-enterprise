package repository

import (
	"reflect"
	"testing"
)

func TestBatchUpsertSizeStaysBelowPostgresParameterLimit(t *testing.T) {
	const postgresParameterLimit = 65535
	authUserFields := reflect.TypeOf(AuthUser{}).NumField()

	if batchUpsertSize*authUserFields >= postgresParameterLimit {
		t.Fatalf("batch size %d can exceed PostgreSQL's parameter limit", batchUpsertSize)
	}
	if batchUpsertSize != 500 {
		t.Fatalf("batchUpsertSize = %d, want 500", batchUpsertSize)
	}
}
