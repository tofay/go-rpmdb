package rpmdb

import (
	"github.com/anchore/go-rpmdb/pkg/bdb"
	dbi "github.com/anchore/go-rpmdb/pkg/db"
	"github.com/anchore/go-rpmdb/pkg/sqlite3"
	"golang.org/x/xerrors"
)

type RpmDB struct {
	db dbi.RpmDBInterface
}

func Open(path string) (*RpmDB, error) {
	// SQLite3 Open() returns nil, nil in case of DB format other than SQLite3
	sqldb, err := sqlite3.Open(path)
	if err != nil && !xerrors.Is(err, sqlite3.ErrorInvalidSQLite3) {
		return nil, err
	}
	if sqldb != nil {
		return &RpmDB{db: sqldb}, nil
	}

	db, err := bdb.Open(path)
	if err != nil {
		return nil, err
	}

	return &RpmDB{
		db: db,
	}, nil

}

func (d *RpmDB) ListPackages() ([]*PackageInfo, error) {
	var pkgList []*PackageInfo

	for entry := range d.db.Read() {
		if entry.Err != nil {
			return nil, entry.Err
		}

		indexEntries, err := headerImport(entry.Value)
		if err != nil {
			return nil, xerrors.Errorf("error during importing header: %w", err)
		}
		pkg, err := newPackage(indexEntries)
		if err != nil {
			return nil, xerrors.Errorf("invalid package info: %w", err)
		}

		pkgList = append(pkgList, pkg)
	}

	return pkgList, nil
}
