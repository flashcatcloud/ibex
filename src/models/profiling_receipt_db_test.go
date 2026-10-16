package models

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/flashcatcloud/ibex/src/server/config"
	"github.com/flashcatcloud/ibex/src/storage"
	"github.com/flashcatcloud/ibex/src/types"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
)

func TestProfilingReceiptPersistenceChecks(t *testing.T) {
	const parent = "cpt-123e4567-e89b-12d3-a456-426614174000"
	for _, scenario := range []string{"stored", "missing-meta", "ordinary", "still-assigned", "missing-result", "mismatch", "db-error", "edge"} {
		t.Run(scenario, func(t *testing.T) {
			sql, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer sql.Close()
			db, err := gorm.Open(mysql.New(mysql.Config{Conn: sql, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			oldDB, oldCenter := storage.DB, config.C.IsCenter
			storage.DB = db
			config.C.IsCenter = scenario != "edge"
			defer func() { storage.DB = oldDB; config.C.IsCenter = oldCenter }()
			r := types.ReportTask{Id: 42, Clock: 100, ProfilingTaskID: parent, Status: "success", Stdout: "full"}
			if scenario != "edge" {
				q := mock.ExpectQuery("SELECT .*task_meta.*").WithArgs(int64(42))
				if scenario == "db-error" {
					q.WillReturnError(errors.New("db unavailable"))
				} else {
					rows := sqlmock.NewRows([]string{"id", "system_caller", "script"})
					caller := "continuous-profiling"
					if scenario == "ordinary" {
						caller = ""
					}
					if scenario != "missing-meta" {
						rows.AddRow(42, caller, "# CP_RETENTION_V1 task="+parent+"\n")
					}
					q.WillReturnRows(rows)
					if scenario != "missing-meta" && scenario != "ordinary" {
						n := 0
						if scenario == "still-assigned" {
							n = 1
						}
						mock.ExpectQuery("SELECT count.*task_host_doing.*").WithArgs(int64(42), "host").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(n))
						if n == 0 {
							rows := sqlmock.NewRows([]string{"id", "host", "status", "stdout", "stderr"})
							out := "full"
							if scenario == "mismatch" {
								out = "different"
							}
							if scenario != "missing-result" {
								rows.AddRow(42, "host", "success", out, "")
							}
							mock.ExpectQuery("SELECT .*task_host_42.*").WithArgs(int64(42), "host").WillReturnRows(rows)
						}
					}
				}
			}
			got, err := ProfilingResultStored("host", r)
			if got != (scenario == "stored") {
				t.Fatalf("confirmed=%v err=%v", got, err)
			}
			if (err != nil) != (scenario == "db-error") {
				t.Fatalf("unexpected error %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
