package main

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type ArticleID = int

type Article struct {
	ID ArticleID
}

type searchParams struct {
	UseAfter  bool      `db:"use_after"`
	After     ArticleID `db:"id_after"`
	UseBefore bool      `db:"use_before"`
	Before    ArticleID `db:"id_before"`
	NumRows   int       `db:"num_rows"`
}

var ErrNewSearchParamsInput = errors.New("{first}, {after, first}, {before, last}のいずれかの組み合わせで指定してください")

func NewSearchParams(after *ArticleID, before *ArticleID, first *int, last *int) (searchParams, error) {
	var sp = searchParams{}

	sp.UseAfter = (after != nil)
	sp.UseBefore = (before != nil)
	useFirst := (first != nil)
	useLast := (last != nil)

	if useFirst && !sp.UseAfter && !useLast && !sp.UseBefore {
		sp.NumRows = *first
	} else if useFirst && sp.UseAfter && !useLast && !sp.UseBefore {
		sp.NumRows = *first
		sp.After = *after
	} else if useLast && sp.UseBefore && !useFirst && !sp.UseAfter {
		sp.NumRows = *last
		sp.Before = *before
	} else {
		return searchParams{}, ErrNewSearchParamsInput
	}

	return sp, nil
}

func (sp searchParams) order() string {
	if sp.UseBefore {
		return "asc"
	}
	return "desc"
}

//go:embed sql/search_articles.sql
var querySearchArticles string

func SeachArticles(db *sqlx.DB, sp searchParams) ([]Article, error) {
	as := make([]Article, 0, sp.NumRows)

	query := fmt.Sprintf(querySearchArticles, sp.order())
	query, args, err := sqlx.Named(query, sp)
	if err != nil {
		return nil, err
	}

	err = db.Select(&as, query, args...)
	if err != nil {
		return nil, err
	}

	return as, nil
}

func PreviousPageExists(db *sqlx.DB, sp searchParams) (bool, error) {
	if sp.UseAfter {
		return behindRowExists(db, sp)
	}
	if sp.UseBefore {
		return additionalRowExists(db, sp)
	}
	return false, nil
}

func NextPageExists(db *sqlx.DB, sp searchParams) (bool, error) {
	if sp.UseBefore {
		behindRowExists(db, sp)
	}
	return additionalRowExists(db, sp)
}

//go:embed sql/additional_row_exists.sql
var queryAdditionalRowExists string

func additionalRowExists(db *sqlx.DB, sp searchParams) (bool, error) {
	var exists bool

	query := fmt.Sprintf(queryAdditionalRowExists, sp.order())
	query, args, err := sqlx.Named(query, sp)
	if err != nil {
		return false, err
	}

	err = db.Get(&exists, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}

//go:embed sql/behind_row_exists.sql
var queryBehindRowExists string

func behindRowExists(db *sqlx.DB, sp searchParams) (bool, error) {
	var exists bool

	query, args, err := sqlx.Named(queryBehindRowExists, sp)
	if err != nil {
		return false, err
	}

	err = db.Get(&exists, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func main() {}
