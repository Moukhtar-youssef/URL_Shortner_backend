package Storage_test

import (
	"testing"

	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func Setup(t *testing.T) *Storage.URLDB {
	var DB *Storage.URLDB
	var err error

	mr, err := miniredis.Run()
	assert.NoError(t, err)

	dbpath := "file::memory:?cache=shared"
	DB, err = Storage.ConnectToDB(dbpath, mr.Addr())
	assert.NoError(t, err)

	err = DB.CreateTable()
	assert.NoError(t, err)

	t.Cleanup(func() {
		DB.Close()
		mr.Close()
	})

	return DB
}

func TestConnectToDB(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	DB, err := Storage.ConnectToDB("file::memory:?cache=shared", mr.Addr())
	assert.NoError(t, err)
	assert.NotNil(t, DB)

	err = DB.Close()
	assert.NoError(t, err)
	mr.Close()
}

func TestCreateTable(t *testing.T) {
	db := Setup(t)
	err := db.CreateTable()
	assert.NoError(t, err)
}

func TestSavingURLandGetURL(t *testing.T) {
	db := Setup(t)

	short := "abc"
	long := "https://example.com"

	err := db.SaveURL(short, long)
	assert.NoError(t, err)

	got, err := db.GetURL(short)
	assert.NoError(t, err)
	assert.Equal(t, long, got)
}

func TestEditURL(t *testing.T) {
	db := Setup(t)

	short := "xyz"
	oldlong := "https://example.com"
	newlong := "https://New.example.com"

	err := db.SaveURL(short, oldlong)
	assert.NoError(t, err)

	err = db.EditURL(short, newlong)
	assert.NoError(t, err)

	got, err := db.GetURL(short)
	assert.NoError(t, err)
	assert.Equal(t, newlong, got)
	assert.NotEqual(t, oldlong, got)
}

func TestDeleteURL(t *testing.T) {
	db := Setup(t)

	short := "del"
	long := "https://example.com"

	err := db.SaveURL(short, long)
	assert.NoError(t, err)

	err = db.DeleteURL(short)
	assert.NoError(t, err)

	_, err = db.GetURL(short)
	assert.Error(t, err)
}

func TestCheckShortURLExists(t *testing.T) {
	db := Setup(t)

	short := "check"
	long := "https://example.com"

	exists, err := db.CheckShortURLExists(short)
	assert.NoError(t, err)
	assert.False(t, exists)

	err = db.SaveURL(short, long)
	assert.NoError(t, err)

	exists, err = db.CheckShortURLExists(short)
	assert.NoError(t, err)
	assert.True(t, exists)
}
