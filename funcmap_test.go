package main

import "testing"

func TestCreateSelectByPkScan(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	sts := testSetupStruct(t, conn)
	for _, s := range sts {
		sql := createSelectByPkScan(s)
		t.Logf("%s", sql)
	}
}

func TestCreateSelectByPkSQLParams(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	sts := testSetupStruct(t, conn)
	for _, s := range sts {
		sql := createSelectByPkSQLParams(s)
		t.Logf("%s", sql)
	}
}

func TestCreatePkParams(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	sts := testSetupStruct(t, conn)
	for _, s := range sts {
		sql := createSelectByPkFuncParams(s)
		t.Logf("%s", sql)
	}
}

func TestCreateSelectByPkSQL(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	sts := testSetupStruct(t, conn)
	for _, s := range sts {
		sql := createSelectByPkSQL(s)
		t.Logf("%s", sql)
	}
}

func TestCreateInsertSQL(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	sts := testSetupStruct(t, conn)
	for _, s := range sts {
		sql := createInsertSQL(s)
		t.Logf("%s", sql)
	}
}

func TestCreateInsertParams(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	sts := testSetupStruct(t, conn)
	for _, s := range sts {
		params := createInsertParams(s)
		t.Logf("%s", params)
	}
}

func TestCreateInsertScan(t *testing.T) {
	conn, cleanup := testPgSetup(t)
	defer cleanup()

	sts := testSetupStruct(t, conn)
	for _, s := range sts {
		params := createInsertScan(s)
		t.Logf("%s", params)
	}
}
