package connection

import (
	"chat/globals"
	"database/sql"
	"strings"
)

func validSqlError(err error) bool {
	if err == nil {
		return false
	}

	content := err.Error()

	// Error 1060: Duplicate column name
	// Error 1050: Table already exists
	// Error 1061: Duplicate key name

	return !(strings.Contains(content, "Error 1060") || strings.Contains(content, "Error 1050") || strings.Contains(content, "Error 1061"))
}

func checkSqlError(_ sql.Result, err error) error {
	if validSqlError(err) {
		return err
	}

	return nil
}

func execSql(db *sql.DB, sql string, args ...interface{}) error {
	return checkSqlError(globals.ExecDb(db, sql, args...))
}

func doMigration(db *sql.DB) error {
	if globals.SqliteEngine {
		return doSqliteMigration(db)
	}

	// v3.10 migration

	// update `quota`, `used` field in `quota` table
	// migrate `DECIMAL(16, 4)` to `DECIMAL(24, 6)`

	if err := execSql(db, `
		ALTER TABLE quota
		MODIFY COLUMN quota DECIMAL(24, 6),
		MODIFY COLUMN used DECIMAL(24, 6);
	`); err != nil {
		return err
	}

	// add new field `is_banned` in `auth` table
	if err := execSql(db, `
		ALTER TABLE auth
		ADD COLUMN is_banned BOOLEAN DEFAULT FALSE;
	`); err != nil {
		return err
	}

	// add new field `task_id` in `conversation` table to store task id (e.g., video job id)
	if err := execSql(db, `
		ALTER TABLE conversation
		ADD COLUMN task_id VARCHAR(255) NULL;
	`); err != nil {
		return err
	}

	// add wechat identifiers for WeChat login
	if err := execSql(db, `
		ALTER TABLE auth
		ADD COLUMN wechat_openid VARCHAR(128) NULL;
	`); err != nil {
		return err
	}

	if err := execSql(db, `
		ALTER TABLE auth
		ADD COLUMN wechat_unionid VARCHAR(128) NULL;
	`); err != nil {
		return err
	}

	// reserve phone/id_card for future SMS login and KYC
	if err := execSql(db, `
		ALTER TABLE auth
		ADD COLUMN phone VARCHAR(32) NULL;
	`); err != nil {
		return err
	}

	if err := execSql(db, `
		ALTER TABLE auth
		ADD COLUMN id_card VARCHAR(32) NULL;
	`); err != nil {
		return err
	}

	if err := execSql(db, `
		CREATE UNIQUE INDEX idx_auth_wechat_openid ON auth(wechat_openid);
	`); err != nil {
		return err
	}

	if err := execSql(db, `
		CREATE UNIQUE INDEX idx_auth_wechat_unionid ON auth(wechat_unionid);
	`); err != nil {
		return err
	}

	if err := execSql(db, `
		CREATE UNIQUE INDEX idx_auth_phone ON auth(phone);
	`); err != nil {
		return err
	}

	if err := execSql(db, `
		CREATE UNIQUE INDEX idx_auth_id_card ON auth(id_card);
	`); err != nil {
		return err
	}

	return nil
}

func doSqliteMigration(db *sql.DB) error {
	// v3.10 added sqlite support, no migration needed before this version

	// v4 migration
	// add new field `task_id` in `conversation` table to store task id (e.g., video job id)
	if err := execSql(db, `
		ALTER TABLE conversation
		ADD COLUMN task_id VARCHAR(255) NULL;
	`); err != nil {
		return err
	}

	return nil
}
