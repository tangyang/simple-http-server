package pg_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"

	"gopkg.in/pg.v4"
	"gopkg.in/pg.v4/orm"
)

func benchmarkDB() *pg.DB {
	return pg.Connect(&pg.Options{
		User:         "postgres",
		Database:     "postgres",
		DialTimeout:  30 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
}

func BenchmarkQueryRowsGopgDiscard(b *testing.B) {
	seedDB()

	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := db.Query(pg.Discard, `SELECT * FROM records LIMIT 100`)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkQueryRowsGopgOptimized(b *testing.B) {
	seedDB()

	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var rs OptRecords
			_, err := db.Query(&rs, `SELECT * FROM records LIMIT 100`)
			if err != nil {
				b.Fatal(err)
			}
			if len(rs.C) != 100 {
				b.Fatalf("got %d, wanted 100", len(rs.C))
			}
		}
	})
}

func BenchmarkQueryRowsGopgReflect(b *testing.B) {
	seedDB()

	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var rs []Record
			_, err := db.Query(&rs, `SELECT * FROM records LIMIT 100`)
			if err != nil {
				b.Fatal(err)
			}
			if len(rs) != 100 {
				b.Fatalf("got %d, wanted 100", len(rs))
			}
		}
	})
}

func BenchmarkQueryRowsGopgORM(b *testing.B) {
	seedDB()

	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var rs []Record
			err := db.Model(&rs).Limit(100).Select()
			if err != nil {
				b.Fatal(err)
			}
			if len(rs) != 100 {
				b.Fatalf("got %d, wanted 100", len(rs))
			}
		}
	})
}

func BenchmarkQueryRowsStdlibPq(b *testing.B) {
	seedDB()

	pqdb, err := pqdb()
	if err != nil {
		b.Fatal(err)
	}
	defer pqdb.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rows, err := pqdb.Query(`SELECT * FROM records LIMIT 100`)
			if err != nil {
				b.Fatal(err)
			}

			var rs []Record
			for rows.Next() {
				rs = append(rs, Record{})
				rec := &rs[len(rs)-1]

				err := rows.Scan(&rec.Num1, &rec.Num2, &rec.Num3, &rec.Str1, &rec.Str2, &rec.Str3)
				if err != nil {
					b.Fatal(err)
				}
			}
			rows.Close()

			if len(rs) != 100 {
				b.Fatalf("got %d, wanted 100", len(rs))
			}
		}
	})
}

func BenchmarkQueryRowsGORM(b *testing.B) {
	seedDB()

	db, err := gormdb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var rs []Record
			err := db.Limit(100).Find(&rs).Error
			if err != nil {
				b.Fatal(err)
			}

			if len(rs) != 100 {
				b.Fatalf("got %d, wanted 100", len(rs))
			}
		}
	})
}

func BenchmarkModelHasOneGopg(b *testing.B) {
	seedDB()

	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var books []Book
			err := db.Model(&books).Column("book.*", "Author").Limit(100).Select()
			if err != nil {
				b.Fatal(err)
			}

			if len(books) != 100 {
				b.Fatalf("got %d, wanted 100", len(books))
			}
		}
	})
}

func BenchmarkModelHasOneGORM(b *testing.B) {
	seedDB()

	db, err := gormdb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var books []Book
			err := db.Preload("Author").Limit(100).Find(&books).Error
			if err != nil {
				b.Fatal(err)
			}

			if len(books) != 100 {
				b.Fatalf("got %d, wanted 100", len(books))
			}
		}
	})
}

func BenchmarkModelHasManyGopg(b *testing.B) {
	seedDB()

	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var books []Book
			err := db.Model(&books).Column("book.*", "Translations").Limit(100).Select()
			if err != nil {
				b.Fatal(err)
			}

			if len(books) != 100 {
				b.Fatalf("got %d, wanted 100", len(books))
			}
			for _, book := range books {
				if len(book.Translations) != 10 {
					b.Fatalf("got %d, wanted 10", len(book.Translations))
				}
			}
		}
	})
}

func BenchmarkModelHasManyGORM(b *testing.B) {
	seedDB()

	db, err := gormdb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var books []Book
			err := db.Preload("Translations").Limit(100).Find(&books).Error
			if err != nil {
				b.Fatal(err)
			}

			if len(books) != 100 {
				b.Fatalf("got %d, wanted 100", len(books))
			}
			for _, book := range books {
				if len(book.Translations) != 10 {
					b.Fatalf("got %d, wanted 10", len(book.Translations))
				}
			}
		}
	})
}

func BenchmarkModelHasMany2ManyGopg(b *testing.B) {
	seedDB()

	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var books []Book
			err := db.Model(&books).
				Column("book.*", "Genres").
				Limit(100).
				Select()

			if err != nil {
				b.Fatal(err)
			}

			if len(books) != 100 {
				b.Fatalf("got %d, wanted 100", len(books))
			}
			for _, book := range books {
				if len(book.Genres) != 10 {
					b.Fatalf("got %d, wanted 10", len(book.Genres))
				}
			}
		}
	})
}

func BenchmarkModelHasMany2ManyGORM(b *testing.B) {
	seedDB()

	db, err := gormdb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var books []Book
			err = db.Preload("Genres").Limit(100).Find(&books).Error
			if err != nil {
				b.Fatal(err)
			}

			if len(books) != 100 {
				b.Fatalf("got %d, wanted 100", len(books))
			}
			for _, book := range books {
				if len(book.Genres) != 10 {
					b.Fatalf("got %d, wanted 10", len(book.Genres))
				}
			}
		}
	})
}

func BenchmarkQueryRow(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var dst numLoader
		_, err := db.QueryOne(&dst, `SELECT ?::bigint AS num`, 1)
		if err != nil {
			b.Fatal(err)
		}
		if dst.Num != 1 {
			b.Fatalf("got %d, wanted 1", dst.Num)
		}
	}
}

func BenchmarkQueryRowStmt(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	stmt, err := db.Prepare(`SELECT $1::bigint AS num`)
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var dst numLoader
		_, err := stmt.QueryOne(&dst, 1)
		if err != nil {
			b.Fatal(err)
		}
		if dst.Num != 1 {
			b.Fatalf("got %d, wanted 1", dst.Num)
		}
	}
}

func BenchmarkQueryRowScan(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var n int64
			_, err := db.QueryOne(pg.Scan(&n), `SELECT ? AS num`, 1)
			if err != nil {
				b.Fatal(err)
			}
			if n != 1 {
				b.Fatalf("got %d, wanted 1", n)
			}
		}
	})
}

func BenchmarkQueryRowStmtScan(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	stmt, err := db.Prepare(`SELECT $1::bigint AS num`)
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var n int64
		_, err := stmt.QueryOne(pg.Scan(&n), 1)
		if err != nil {
			b.Fatal(err)
		}
		if n != 1 {
			b.Fatalf("got %d, wanted 1", n)
		}
	}
}

func BenchmarkQueryRowStdlibPq(b *testing.B) {
	db, err := pqdb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var n int64
			r := db.QueryRow(`SELECT $1::bigint AS num`, 1)
			if err := r.Scan(&n); err != nil {
				b.Fatal(err)
			}
			if n != 1 {
				b.Fatalf("got %d, wanted 1", n)
			}
		}
	})
}

func BenchmarkQueryRowWithoutParamsStdlibPq(b *testing.B) {
	db, err := pqdb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var n int64
			r := db.QueryRow("SELECT 1::bigint AS num")
			if err := r.Scan(&n); err != nil {
				b.Fatal(err)
			}
			if n != 1 {
				b.Fatalf("got %d, wanted 1", n)
			}
		}
	})
}

func BenchmarkQueryRowStdlibMySQL(b *testing.B) {
	db, err := mysqldb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var n int64
			r := db.QueryRow("SELECT ? AS num", 1)
			if err := r.Scan(&n); err != nil {
				b.Fatal(err)
			}
			if n != 1 {
				b.Fatalf("got %d, wanted 1", n)
			}
		}
	})
}

func BenchmarkQueryRowStmtStdlibPq(b *testing.B) {
	db, err := pqdb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare(`SELECT $1::bigint AS num`)
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var n int64
		r := stmt.QueryRow(1)
		if err := r.Scan(&n); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExec(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	qs := []string{
		`DROP TABLE IF EXISTS exec_test`,
		`CREATE TABLE exec_test(id bigint, name varchar(500))`,
	}
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := db.Exec(`INSERT INTO exec_test (id, name) VALUES (?, ?)`, 1, "hello world")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkExecWithError(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	qs := []string{
		`DROP TABLE IF EXISTS exec_with_error_test`,
		`CREATE TABLE exec_with_error_test(id bigint PRIMARY KEY, name varchar(500))`,
	}
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			b.Fatal(err)
		}
	}

	_, err := db.Exec(`
		INSERT INTO exec_with_error_test(id, name) VALUES(?, ?)
	`, 1, "hello world")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := db.Exec(`INSERT INTO exec_with_error_test(id) VALUES(?)`, 1)
			if err == nil {
				b.Fatalf("got nil error, expected integrity violation")
			} else if pgErr, ok := err.(pg.Error); !ok || !pgErr.IntegrityViolation() {
				b.Fatalf("got %s, expected integrity violation", err)
			}
		}
	})
}

func BenchmarkExecStmt(b *testing.B) {
	db := benchmarkDB()
	defer db.Close()

	_, err := db.Exec(`CREATE TEMP TABLE statement_exec(id bigint, name varchar(500))`)
	if err != nil {
		b.Fatal(err)
	}

	stmt, err := db.Prepare(`INSERT INTO statement_exec (id, name) VALUES ($1, $2)`)
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := stmt.Exec(1, "hello world")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecStmtStdlibPq(b *testing.B) {
	db, err := pqdb()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TEMP TABLE statement_exec(id bigint, name varchar(500))`)
	if err != nil {
		b.Fatal(err)
	}

	stmt, err := db.Prepare(`INSERT INTO statement_exec (id, name) VALUES ($1, $2)`)
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := stmt.Exec(1, "hello world")
		if err != nil {
			b.Fatal(err)
		}
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func pqdb() (*sql.DB, error) {
	return sql.Open("postgres", "user=postgres dbname=postgres sslmode=disable")
}

func mysqldb() (*sql.DB, error) {
	return sql.Open("mysql", "root:root@tcp(localhost:3306)/test")
}

func gormdb() (*gorm.DB, error) {
	return gorm.Open("postgres", "user=postgres dbname=postgres sslmode=disable")
}

type Record struct {
	Num1, Num2, Num3 int64
	Str1, Str2, Str3 string
}

func (r *Record) GetNum1() int64 {
	return r.Num1
}

func (r *Record) GetNum2() int64 {
	return r.Num2
}

func (r *Record) GetNum3() int64 {
	return r.Num3
}

func (r *Record) GetStr1() string {
	return r.Str1
}

func (r *Record) GetStr2() string {
	return r.Str2
}

func (r *Record) GetStr3() string {
	return r.Str3
}

type OptRecord struct {
	Num1, Num2, Num3 int64
	Str1, Str2, Str3 string
}

var _ orm.ColumnScanner = (*OptRecord)(nil)

func (r *OptRecord) ScanColumn(colIdx int, colName string, b []byte) error {
	var err error
	switch colName {
	case "num1":
		r.Num1, err = strconv.ParseInt(string(b), 10, 64)
	case "num2":
		r.Num2, err = strconv.ParseInt(string(b), 10, 64)
	case "num3":
		r.Num3, err = strconv.ParseInt(string(b), 10, 64)
	case "str1":
		r.Str1 = string(b)
	case "str2":
		r.Str2 = string(b)
	case "str3":
		r.Str3 = string(b)
	default:
		return fmt.Errorf("unknown column: %q", colName)
	}
	return err
}

type OptRecords struct {
	C []OptRecord
}

var _ orm.Collection = (*OptRecords)(nil)

func (rs *OptRecords) NewModel() orm.ColumnScanner {
	rs.C = append(rs.C, OptRecord{})
	return &rs.C[len(rs.C)-1]
}

func (rs *OptRecords) AddModel(_ orm.ColumnScanner) error {
	return nil
}

var seedDBOnce sync.Once

func seedDB() {
	seedDBOnce.Do(func() {
		if err := _seedDB(); err != nil {
			panic(err)
		}
	})
}

func _seedDB() error {
	db := benchmarkDB()
	defer db.Close()

	_, err := db.Exec(`DROP TABLE IF EXISTS records`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE records(
			num1 serial,
			num2 serial,
			num3 serial,
			str1 text,
			str2 text,
			str3 text
		)
	`)
	if err != nil {
		return err
	}

	for i := 0; i < 1000; i++ {
		_, err := db.Exec(`
			INSERT INTO records (str1, str2, str3) VALUES (?, ?, ?)
		`, randSeq(100), randSeq(200), randSeq(300))
		if err != nil {
			return err
		}
	}

	err = createTestSchema(db)
	if err != nil {
		return err
	}

	for i := 1; i < 100; i++ {
		genre := Genre{
			Id:   i,
			Name: fmt.Sprintf("genre %d", i),
		}
		err = db.Create(&genre)
		if err != nil {
			return err
		}

		author := Author{
			ID:   i,
			Name: fmt.Sprintf("author %d", i),
		}
		err = db.Create(&author)
		if err != nil {
			return err
		}
	}

	for i := 1; i <= 1000; i++ {
		err = db.Create(&Book{
			Id:        i,
			Title:     fmt.Sprintf("book %d", i),
			AuthorID:  rand.Intn(99) + 1,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}

		for j := 1; j <= 10; j++ {
			err = db.Create(&BookGenre{
				BookId:  i,
				GenreId: rand.Intn(99) + 1,
			})
			if err != nil {
				return err
			}

			err = db.Create(&Translation{
				BookId: i,
				Lang:   fmt.Sprintf("%d", j),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
