package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/google/uuid"

	"azuremyst.org/go-home-sensors/exporters"
	"azuremyst.org/go-home-sensors/log"
	"azuremyst.org/go-home-sensors/sensors"
)

//go:embed table.sql
var tableSQL string

type SqliteExporter struct {
	db *sql.DB
}

func createDBFile(path string) error {
	err := os.MkdirAll(filepath.Dir(path), fs.FileMode(os.O_RDWR))
	if err != nil {
		return fmt.Errorf("failed to create path to db file: %q", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create database file: %q", err)
	}
	return file.Close()
}

func CreateExporter(path string) exporters.Exporter {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := createDBFile(path); err != nil {
			log.ErrorLog.Fatalf("Unable to create db file %q", err)
		}
	} else {
		dir := filepath.Dir(path)
		file := filepath.Base(path)
		ext := filepath.Ext(file)
		fileName := strings.TrimSuffix(file, ext)

		newPath := filepath.Join(dir, fileName+"_"+strconv.FormatInt(time.Now().UnixMilli(), 10)+"_"+uuid.NewString()+"."+ext)
		if err := createDBFile(newPath); err != nil {
			log.ErrorLog.Fatalf("Unable to create new db file %q", err)
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.ErrorLog.Fatalf("Unable to open file %q", err)
	}

	_, err = db.Exec(tableSQL)
	if err != nil {
		log.ErrorLog.Fatalf("Failed to create tables %q: %s", err, tableSQL)
	}

	tx, err := db.Begin()
	if err != nil {
		log.ErrorLog.Fatalf("Failed to start transaction: %q", err)
	}
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO measurement(id, description, unit) VALUES(?, ?, ?)")
	if err != nil {
		log.ErrorLog.Fatalf("Failed create measurement statement: %q", err)
	}
	defer stmt.Close()
	for _, m := range sensors.Measurements {
		_, err = stmt.Exec(m.ID, m.Description, m.Unit)
		if err != nil {
			log.ErrorLog.Fatalf("Failed to insert measurements: %q", err)
		}
	}

	if err = tx.Commit(); err != nil {
		log.ErrorLog.Fatalf("Failed to commit db initialization: %q", err)
	}
	return &SqliteExporter{db: db}
}

func (pe *SqliteExporter) Export(recordings []sensors.MeasurementRecording) {
	tx, err := pe.db.Begin()
	if err != nil {
		log.ErrorLog.Fatalf("Failed to start export transaction: %q", err)
	}

	stmtMeasurement, err := tx.Prepare("INSERT INTO measurement_recording(value, measure_id) VALUES(?, ?)")
	if err != nil {
		log.ErrorLog.Fatalf("Failed create measurement_recording statement: %q", err)
	}
	defer stmtMeasurement.Close()

	stmtMetadata, err := tx.Prepare("INSERT INTO measurement_meta_data(key, value, recording_id) VALUES(?, ?, ?)")
	if err != nil {
		log.ErrorLog.Fatalf("Failed create measurement_meta_data statement: %q", err)
	}
	defer stmtMetadata.Close()

	for _, recording := range recordings {
		res, err := stmtMeasurement.Exec(recording.Value, recording.Measure.ID)
		if err != nil {
			log.ErrorLog.Fatalf("Failed to insert %s: %q\n", recording.Measure.ID, err)
		}

		insertId, err := res.LastInsertId()
		if err != nil {
			log.ErrorLog.Fatalf("Failed to retrieve last insert id: %q", err)
		}
		for k, v := range recording.Metadata {
			if _, err = stmtMetadata.Exec(k, v, insertId); err != nil {
				log.ErrorLog.Fatalf("Failed to insert metadata asociated with recording: %q", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		log.ErrorLog.Fatalf("Failed to commit metric export: %q", err)
	}
}
