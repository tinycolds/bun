package bun

import (
	"context"

	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/internal"
	"github.com/uptrace/bun/sqlfmt"
)

type TruncateTableQuery struct {
	baseQuery
	cascadeQuery

	continueIdentity bool
}

func NewTruncateTableQuery(db *DB) *TruncateTableQuery {
	q := &TruncateTableQuery{
		baseQuery: baseQuery{
			db:  db,
			dbi: db.DB,
		},
	}
	return q
}

func (q *TruncateTableQuery) DB(db DBI) *TruncateTableQuery {
	q.dbi = db
	return q
}

func (q *TruncateTableQuery) Model(model interface{}) *TruncateTableQuery {
	q.setTableModel(model)
	return q
}

//------------------------------------------------------------------------------

func (q *TruncateTableQuery) Table(tables ...string) *TruncateTableQuery {
	for _, table := range tables {
		q.addTable(sqlfmt.UnsafeIdent(table))
	}
	return q
}

func (q *TruncateTableQuery) TableExpr(query string, args ...interface{}) *TruncateTableQuery {
	q.addTable(sqlfmt.SafeQuery(query, args))
	return q
}

//------------------------------------------------------------------------------

func (q *TruncateTableQuery) ContinueIdentity() *TruncateTableQuery {
	q.continueIdentity = true
	return q
}

func (q *TruncateTableQuery) Restrict() *TruncateTableQuery {
	q.restrict = true
	return q
}

//------------------------------------------------------------------------------

func (q *TruncateTableQuery) AppendQuery(
	fmter sqlfmt.QueryFormatter, b []byte,
) (_ []byte, err error) {
	if q.err != nil {
		return nil, q.err
	}

	if !fmter.HasFeature(feature.TableTruncate) {
		b = append(b, "DELETE FROM "...)

		b, err = q.appendTables(fmter, b)
		if err != nil {
			return nil, err
		}

		return b, nil
	}

	b = append(b, "TRUNCATE TABLE "...)

	b, err = q.appendTables(fmter, b)
	if err != nil {
		return nil, err
	}

	if q.db.features.Has(feature.TableIdentity) {
		if q.continueIdentity {
			b = append(b, " CONTINUE IDENTITY"...)
		} else {
			b = append(b, " RESTART IDENTITY"...)
		}
	}

	b = q.appendCascade(fmter, b)

	return b, nil
}

//------------------------------------------------------------------------------

func (q *TruncateTableQuery) Exec(ctx context.Context, dest ...interface{}) (res Result, _ error) {
	bs := getByteSlice()
	defer putByteSlice(bs)

	queryBytes, err := q.AppendQuery(q.db.fmter, bs.b)
	if err != nil {
		return res, err
	}

	bs.b = queryBytes
	query := internal.String(queryBytes)

	res, err = q.exec(ctx, q, query)
	if err != nil {
		return res, err
	}

	return res, nil
}