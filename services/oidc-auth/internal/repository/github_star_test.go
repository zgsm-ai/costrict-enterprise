package repository

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func starTestDatabase(t *testing.T, dialect string) (*Database, sqlmock.Sqlmock) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = conn.Close()
	})
	var dialector gorm.Dialector = postgres.New(postgres.Config{Conn: conn})
	if dialect == "mysql" {
		dialector = mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true})
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	return &Database{db: db}, mock
}

// Exact SQL matching enforces the safety boundary: no subject_id, devices,
// invite_code, timestamps, or other user fields may appear in the SET clause.
func starBatchSQL(dialect string, count int) string {
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		if dialect == "mysql" {
			placeholders[i] = "?"
		}
	}
	if dialect == "mysql" {
		return regexp.QuoteMeta("UPDATE `auth_users` SET `github_star`=? WHERE github_id IN ("+strings.Join(placeholders, ",")+")") + "$"
	}
	return regexp.QuoteMeta(`UPDATE "auth_users" SET "github_star"=$1 WHERE github_id IN (`+strings.Join(placeholders, ",")+")") + "$"
}

func TestBatchUpdateGithubStarsOnlyWritesStarColumn(t *testing.T) {
	for _, dialect := range []string{"postgres", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db, mock := starTestDatabase(t, dialect)
			// Groups can be emitted in either order.
			mock.MatchExpectationsInOrder(false)
			mock.ExpectBegin()
			mock.ExpectExec(starBatchSQL(dialect, 2)).
				WithArgs("zgsm-ai.costrict", "123", "456").
				WillReturnResult(sqlmock.NewResult(0, 2))
			mock.ExpectExec(starBatchSQL(dialect, 1)).
				WithArgs("", "789").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()
			err := db.BatchUpdateGithubStars(context.Background(), []GithubStarUpdate{
				{GithubID: "123", GithubStar: "zgsm-ai.costrict"},
				{GithubID: "789", GithubStar: ""},
				{GithubID: "456", GithubStar: "zgsm-ai.costrict"},
				{GithubID: "", GithubStar: "zgsm-ai.costrict"},
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBatchUpdateGithubStarsDoesNotRecreateMissingUsers(t *testing.T) {
	db, mock := starTestDatabase(t, "postgres")
	mock.ExpectBegin()
	mock.ExpectExec(starBatchSQL("postgres", 1)).WithArgs("zgsm-ai.costrict", "123").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	if err := db.BatchUpdateGithubStars(context.Background(), []GithubStarUpdate{
		{GithubID: "123", GithubStar: "zgsm-ai.costrict"},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestBatchUpdateGithubStarsBatchesAndRollsBackOnFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprintf("fail=%v", fail), func(t *testing.T) {
			db, mock := starTestDatabase(t, "postgres")
			updates := make([]GithubStarUpdate, 501)
			args := []driver.Value{"zgsm-ai.costrict"}
			for i := range updates {
				id := fmt.Sprint(i + 1)
				updates[i] = GithubStarUpdate{GithubID: id, GithubStar: "zgsm-ai.costrict"}
				if i < 500 {
					args = append(args, id)
				}
			}
			mock.ExpectBegin()
			mock.ExpectExec(starBatchSQL("postgres", 500)).WithArgs(args...).
				WillReturnResult(sqlmock.NewResult(0, 500))
			last := mock.ExpectExec(starBatchSQL("postgres", 1)).WithArgs("zgsm-ai.costrict", "501")
			failure := errors.New("database write failed")
			if fail {
				last.WillReturnError(failure)
				mock.ExpectRollback()
			} else {
				last.WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}
			err := db.BatchUpdateGithubStars(context.Background(), updates)
			if fail && !errors.Is(err, failure) || !fail && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestBatchUpdateGithubStarsEmptyInputDoesNotWrite(t *testing.T) {
	db, _ := starTestDatabase(t, "postgres")
	for _, updates := range [][]GithubStarUpdate{nil, {{GithubID: ""}}} {
		if err := db.BatchUpdateGithubStars(context.Background(), updates); err != nil {
			t.Fatal(err)
		}
	}
}
