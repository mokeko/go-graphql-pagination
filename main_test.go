package main

import (
	_ "embed"
	"log"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func intToPtr(arg int) *int {
	return &arg
}

//go:embed sql/test_setup.sql
var statementTestSetup string

func setup() *sqlx.DB {
	var db *sqlx.DB
	waitSecond := 0
	for {
		db = sqlx.MustOpen("mysql", "root:@tcp(127.0.0.1:3306)/test_database?multiStatements=true")
		err := db.Ping()
		if err == nil {
			break
		}
		db.Close()
		time.Sleep(1 * time.Second)
		waitSecond += 1
		if waitSecond > 30 {
			log.Panicln("failed to connect db")
		}
	}
	db.MustExec(statementTestSetup)
	return db
}

//go:embed sql/test_teardown.sql
var statementTestTeardown string

func teardown(db *sqlx.DB) {
	db.MustExec(statementTestTeardown)
	db.Close()
}

type newSearchParamsInput struct {
	after  *ArticleID
	before *ArticleID
	first  *int
	last   *int
}

func TestNewSearchParams(t *testing.T) {
	type output struct {
		spExpected  searchParams
		errExpected error
	}

	cases := []struct {
		name string
		in   newSearchParamsInput
		out  output
	}{
		{
			"firstを指定する場合",
			newSearchParamsInput{
				after:  nil,
				before: nil,
				first:  intToPtr(5),
				last:   nil,
			},
			output{
				spExpected: searchParams{
					NumRows: 5,
				},
				errExpected: nil,
			},
		},
		{
			"after, firstを指定する場合",
			newSearchParamsInput{
				after:  intToPtr(3),
				before: nil,
				first:  intToPtr(5),
				last:   nil,
			},
			output{
				spExpected: searchParams{
					UseAfter: true,
					After:    3,
					NumRows:  5,
				},
				errExpected: nil,
			},
		},
		{
			"before, lastを指定する場合",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(4),
				first:  nil,
				last:   intToPtr(2),
			},
			output{
				spExpected: searchParams{
					UseBefore: true,
					Before:    4,
					NumRows:   2,
				},
				errExpected: nil,
			},
		},
		// 以下全てerror
		{
			"何も指定しない場合",
			newSearchParamsInput{nil, nil, nil, nil},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"after",
			newSearchParamsInput{
				after:  intToPtr(3),
				before: nil,
				first:  nil,
				last:   nil,
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"before",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(4),
				first:  nil,
				last:   nil,
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"last",
			newSearchParamsInput{
				after:  nil,
				before: nil,
				first:  nil,
				last:   intToPtr(2),
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"after, before",
			newSearchParamsInput{
				after:  intToPtr(3),
				before: intToPtr(4),
				first:  nil,
				last:   nil,
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"first, last",
			newSearchParamsInput{
				after:  nil,
				before: nil,
				first:  intToPtr(5),
				last:   intToPtr(2),
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"after, last",
			newSearchParamsInput{
				after:  intToPtr(3),
				before: nil,
				first:  nil,
				last:   intToPtr(2),
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"before, first",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(4),
				first:  intToPtr(5),
				last:   nil,
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"after, before, first",
			newSearchParamsInput{
				after:  intToPtr(3),
				before: intToPtr(4),
				first:  intToPtr(5),
				last:   nil,
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"after, before, last",
			newSearchParamsInput{
				after:  intToPtr(3),
				before: intToPtr(4),
				first:  nil,
				last:   intToPtr(2),
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"after, first, last",
			newSearchParamsInput{
				after:  intToPtr(3),
				before: nil,
				first:  intToPtr(5),
				last:   intToPtr(2),
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"before, first, last",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(4),
				first:  intToPtr(5),
				last:   intToPtr(2),
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
		{
			"全て指定する場合",
			newSearchParamsInput{
				after:  intToPtr(3),
				before: intToPtr(4),
				first:  intToPtr(5),
				last:   intToPtr(2),
			},
			output{searchParams{}, ErrNewSearchParamsInput},
		},
	}

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {
				spActual, errActual := NewSearchParams(c.in.after, c.in.before, c.in.first, c.in.last)
				assert.Equal(t, c.out.errExpected, errActual)
				assert.Equal(t, c.out.spExpected, spActual)
			},
		)
	}
}

func TestPreviousPageExists(t *testing.T) {
	db := setup()
	defer teardown(db)

	cases := []struct {
		name           string
		in             newSearchParamsInput
		existsExpected bool
	}{
		{
			"firstのみを指定する場合",
			newSearchParamsInput{
				after:  nil,
				before: nil,
				first:  intToPtr(3),
				last:   nil,
			},
			false,
		},
		{
			"after, firstを指定する場合",
			newSearchParamsInput{
				after:  intToPtr(5),
				before: nil,
				first:  intToPtr(3),
				last:   nil,
			},
			true, // afterで指定したid=5が前pageに残る. 取得するのは4, 3, 2
		},
		{
			"before, lastを指定して余りがある場合",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(1),
				first:  nil,
				last:   intToPtr(2),
			},
			true, // id=3, 2を取得して4, 5が残る
		},
		{
			"before, lastを指定して余りがない場合",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(1),
				first:  nil,
				last:   intToPtr(4),
			},
			false,
		},
	}

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {
				sp, err := NewSearchParams(c.in.after, c.in.before, c.in.first, c.in.last)
				assert.NoError(t, err)
				existsActual, err := PreviousPageExists(db, sp)
				assert.NoError(t, err)
				assert.Equal(t, c.existsExpected, existsActual)
			},
		)
	}
}

func TestNextPageExists(t *testing.T) {
	db := setup()
	defer teardown(db)

	cases := []struct {
		name           string
		in             newSearchParamsInput
		existsExpected bool
	}{
		{
			"firstを指定して余りがある場合",
			newSearchParamsInput{
				after:  nil,
				before: nil,
				first:  intToPtr(3),
				last:   nil,
			},
			true, // id=5, 4, 3を取得して2, 1が残る
		},
		{
			"after, firstを指定して余りがある場合",
			newSearchParamsInput{
				after:  intToPtr(5),
				before: nil,
				first:  intToPtr(3),
				last:   nil,
			},
			true, // id=4, 3, 2を取得して1が残る
		},
		{
			"firstを指定して余りがない場合",
			newSearchParamsInput{
				after:  nil,
				before: nil,
				first:  intToPtr(5),
				last:   nil,
			},
			false,
		},
		{
			"after, firstを指定して余りがない場合",
			newSearchParamsInput{
				after:  intToPtr(5),
				before: nil,
				first:  intToPtr(4),
				last:   nil,
			},
			false,
		},
		{
			"before, lastを指定する場合",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(1),
				first:  nil,
				last:   intToPtr(2),
			},
			true, // beforeで指定したid=1が次pageに残る. 取得するのはid=3, 2
		},
	}

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {
				sp, err := NewSearchParams(c.in.after, c.in.before, c.in.first, c.in.last)
				assert.NoError(t, err)
				existsActual, err := NextPageExists(db, sp)
				assert.NoError(t, err)
				assert.Equal(t, c.existsExpected, existsActual)
			},
		)
	}
}

func TestSearchArticles(t *testing.T) {
	db := setup()
	defer teardown(db)

	cases := []struct {
		name       string
		in         newSearchParamsInput
		asExpected []Article
	}{
		{
			"firstを指定する場合",
			newSearchParamsInput{
				after:  nil,
				before: nil,
				first:  intToPtr(3),
				last:   nil,
			},
			[]Article{{5}, {4}, {3}},
		},
		{
			"after, firstを指定して余りがある場合",
			newSearchParamsInput{
				after:  intToPtr(4),
				before: nil,
				first:  intToPtr(2),
				last:   nil,
			},
			[]Article{{3}, {2}},
		},
		{
			"after, firstを指定して余りがない場合",
			newSearchParamsInput{
				after:  intToPtr(4),
				before: nil,
				first:  intToPtr(4),
				last:   nil,
			},
			[]Article{{3}, {2}, {1}},
		},
		{
			"after, firstを指定して該当記事がない場合",
			newSearchParamsInput{
				after:  intToPtr(1),
				before: nil,
				first:  intToPtr(3),
				last:   nil,
			},
			[]Article{},
		},
		{
			"before, lastを指定して余りがある場合",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(2),
				first:  nil,
				last:   intToPtr(2),
			},
			[]Article{{4}, {3}},
		},
		{
			"before, lastを指定して余りがない場合",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(2),
				first:  nil,
				last:   intToPtr(4),
			},
			[]Article{{5}, {4}, {3}},
		},
		{
			"before, lastを指定して該当記事がない場合",
			newSearchParamsInput{
				after:  nil,
				before: intToPtr(5),
				first:  nil,
				last:   intToPtr(4),
			},
			[]Article{},
		},
	}

	for _, c := range cases {
		t.Run(
			c.name,
			func(t *testing.T) {
				sp, err := NewSearchParams(c.in.after, c.in.before, c.in.first, c.in.last)
				assert.NoError(t, err)
				asActual, err := SeachArticles(db, sp)
				assert.NoError(t, err)
				assert.Equal(t, c.asExpected, asActual)
			},
		)
	}
}
