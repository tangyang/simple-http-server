package pg_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gopkg.in/pg.v4"
)

func init() {
	//pg.SetLogger(log.New(os.Stderr, "pg: ", log.LstdFlags))
}

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pg")
}

func pgOptions() *pg.Options {
	return &pg.Options{
		User:               "postgres",
		Database:           "postgres",
		DialTimeout:        30 * time.Second,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		PoolSize:           10,
		PoolTimeout:        30 * time.Second,
		IdleTimeout:        10 * time.Second,
		IdleCheckFrequency: 100 * time.Millisecond,
	}
}

var _ = Describe("Collection", func() {
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())
	})

	AfterEach(func() {
		Expect(db.Close()).NotTo(HaveOccurred())
	})

	It("supports slice of structs", func() {
		coll := []struct {
			Id int
		}{}
		_, err := db.Query(&coll, `
			WITH data (id) AS (VALUES (1), (2), (3))
			SELECT id FROM data
		`)
		Expect(err).NotTo(HaveOccurred())
		Expect(coll).To(HaveLen(3))
		Expect(coll[0].Id).To(Equal(1))
		Expect(coll[1].Id).To(Equal(2))
		Expect(coll[2].Id).To(Equal(3))
	})

	It("supports slice of pointers", func() {
		coll := []*struct {
			Id int
		}{}
		_, err := db.Query(&coll, `
			WITH data (id) AS (VALUES (1), (2), (3))
			SELECT id FROM data
		`)
		Expect(err).NotTo(HaveOccurred())
		Expect(coll).To(HaveLen(3))
		Expect(coll[0].Id).To(Equal(1))
		Expect(coll[1].Id).To(Equal(2))
		Expect(coll[2].Id).To(Equal(3))
	})

	It("supports Collection interface", func() {
		var coll pg.Ints
		_, err := db.Query(&coll, `
			WITH data (id) AS (VALUES (1), (2), (3))
			SELECT id FROM data
		`)
		Expect(err).NotTo(HaveOccurred())
		Expect(coll).To(HaveLen(3))
		Expect(coll[0]).To(Equal(int64(1)))
		Expect(coll[1]).To(Equal(int64(2)))
		Expect(coll[2]).To(Equal(int64(3)))
	})

	It("supports slice of values", func() {
		var ints []int
		_, err := db.Query(&ints, `
			WITH data (id) AS (VALUES (1), (2), (3))
			SELECT id FROM data
		`)
		Expect(err).NotTo(HaveOccurred())
		Expect(ints).To(HaveLen(3))
		Expect(ints[0]).To(Equal(1))
		Expect(ints[1]).To(Equal(2))
		Expect(ints[2]).To(Equal(3))
	})

	It("supports slice of time.Time", func() {
		var times []time.Time
		_, err := db.Query(&times, `
			WITH data (time) AS (VALUES (clock_timestamp()), (clock_timestamp()))
			SELECT time FROM data
		`)
		Expect(err).NotTo(HaveOccurred())
		Expect(times).To(HaveLen(2))
	})
})

var _ = Describe("read/write timeout", func() {
	var db *pg.DB

	BeforeEach(func() {
		opt := pgOptions()
		opt.ReadTimeout = time.Millisecond
		db = pg.Connect(opt)
	})

	AfterEach(func() {
		Expect(db.Close()).NotTo(HaveOccurred())
	})

	It("slow query timeouts", func() {
		_, err := db.Exec(`SELECT pg_sleep(1)`)
		Expect(err.(net.Error).Timeout()).To(BeTrue())
	})

	Describe("WithTimeout", func() {
		It("slow query passes", func() {
			_, err := db.WithTimeout(time.Minute).Exec(`SELECT pg_sleep(1)`)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("CopyFrom/CopyTo", func() {
	const n = 1000000
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())

		qs := []string{
			"CREATE TEMP TABLE copy_from(n int)",
			"CREATE TEMP TABLE copy_to(n int)",
			fmt.Sprintf("INSERT INTO copy_from SELECT generate_series(1, %d)", n),
		}
		for _, q := range qs {
			_, err := db.Exec(q)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("copies data from a table and to a table", func() {
		var buf bytes.Buffer
		res, err := db.CopyTo(&buf, "COPY copy_from TO STDOUT")
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Affected()).To(Equal(n))

		res, err = db.CopyFrom(&buf, "COPY copy_to FROM STDIN")
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Affected()).To(Equal(n))

		var count int
		_, err = db.QueryOne(pg.Scan(&count), "SELECT count(*) FROM copy_to")
		Expect(err).NotTo(HaveOccurred())
		Expect(count).To(Equal(n))

		st := db.Pool().Stats()
		Expect(st.Requests).To(Equal(uint32(6)))
		Expect(st.Hits).To(Equal(uint32(5)))
		Expect(st.Timeouts).To(Equal(uint32(0)))
		Expect(st.TotalConns).To(Equal(uint32(1)))
		Expect(st.FreeConns).To(Equal(uint32(1)))
	})

	It("copies corrupted data to a table", func() {
		buf := bytes.NewBufferString("corrupted data")
		res, err := db.CopyFrom(buf, "COPY copy_to FROM STDIN")
		Expect(err).To(MatchError(`ERROR #22P02 invalid input syntax for integer: "corrupted data": `))
		Expect(res).To(BeNil())

		st := db.Pool().Stats()
		Expect(st.Requests).To(Equal(uint32(4)))
		Expect(st.Hits).To(Equal(uint32(3)))
		Expect(st.Timeouts).To(Equal(uint32(0)))
		Expect(st.TotalConns).To(Equal(uint32(1)))
		Expect(st.FreeConns).To(Equal(uint32(1)))
	})
})

var _ = Describe("DB nulls", func() {
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())

		_, err := db.Exec("CREATE TEMP TABLE tests (id int, value int)")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create struct with sql.NullInt64", func() {
		type Test struct {
			Id    int
			Value sql.NullInt64
		}

		It("inserts null value", func() {
			ins := Test{
				Id: 1,
			}
			err := db.Create(&ins)
			Expect(err).NotTo(HaveOccurred())

			sel := Test{
				Id: 1,
			}
			err = db.Select(&sel)
			Expect(err).NotTo(HaveOccurred())
			Expect(sel.Value.Valid).To(BeFalse())
		})

		It("inserts non-null value", func() {
			ins := Test{
				Id: 1,
				Value: sql.NullInt64{
					Int64: 2,
					Valid: true,
				},
			}
			err := db.Create(&ins)
			Expect(err).NotTo(HaveOccurred())

			sel := Test{
				Id: 1,
			}
			err = db.Select(&sel)
			Expect(err).NotTo(HaveOccurred())
			Expect(sel.Value.Valid).To(BeTrue())
			Expect(sel.Value.Int64).To(Equal(int64(2)))
		})
	})

	Context("nil ptr", func() {
		type Test struct {
			Id    int
			Value *int
		}

		It("inserts null value", func() {
			ins := Test{
				Id: 1,
			}
			err := db.Create(&ins)
			Expect(err).NotTo(HaveOccurred())

			sel := Test{
				Id: 1,
			}
			err = db.Select(&sel)
			Expect(err).NotTo(HaveOccurred())
			Expect(sel.Value).To(BeNil())
		})

		It("inserts non-null value", func() {
			value := 2
			ins := Test{
				Id:    1,
				Value: &value,
			}
			err := db.Create(&ins)
			Expect(err).NotTo(HaveOccurred())

			sel := Test{
				Id: 1,
			}
			err = db.Select(&sel)
			Expect(err).NotTo(HaveOccurred())
			Expect(sel.Value).NotTo(BeNil())
			Expect(*sel.Value).To(Equal(2))
		})
	})
})

var _ = Describe("DB.Select", func() {
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())

		qs := []string{
			`CREATE TEMP TABLE tests (col bytea)`,
			fmt.Sprintf(`INSERT INTO tests VALUES ('\x%x')`, []byte("bytes")),
		}
		for _, q := range qs {
			_, err := db.Exec(q)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("selects bytea", func() {
		var col []byte
		err := db.Model(nil).Table("tests").Column("col").Select(pg.Scan(&col))
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("DB.Create", func() {
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error on nil", func() {
		err := db.Create(nil)
		Expect(err).To(MatchError("pg: Model(nil)"))
	})

	It("returns an errors if value is not settable", func() {
		err := db.Create(1)
		Expect(err).To(MatchError("pg: Model(non-pointer int)"))
	})

	It("returns an errors if value is not supported", func() {
		var v int
		err := db.Create(&v)
		Expect(err).To(MatchError("pg: Model(unsupported int)"))
	})
})

var _ = Describe("DB.Update", func() {
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error on nil", func() {
		err := db.Update(nil)
		Expect(err).To(MatchError("pg: Model(nil)"))
	})

	It("returns an error if there are no pks", func() {
		type Test struct{}
		var test Test
		err := db.Update(&test)
		Expect(err).To(MatchError(`model "test" does not have primary keys`))
	})
})

var _ = Describe("DB.Delete", func() {
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error on nil", func() {
		err := db.Delete(nil)
		Expect(err).To(MatchError("pg: Model(nil)"))
	})

	It("returns an error if there are no pks", func() {
		type Test struct{}
		var test Test
		err := db.Delete(&test)
		Expect(err).To(MatchError(`model "test" does not have primary keys`))
	})
})

var _ = Describe("scanning unknown column", func() {
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error", func() {
		type Test struct {
			Col1 int
		}

		var test Test
		_, err := db.QueryOne(&test, "SELECT 1 AS col1, 2 AS col2")
		Expect(err).To(MatchError(`pg: can't find column "col2" in model test`))
		Expect(test.Col1).To(Equal(1))
	})
})

type Genre struct {
	// TableName is an optional field that specifies custom table name and alias.
	// By default go-pg generates table name and alias from struct name.
	TableName struct{} `sql:"genres,alias:genre"` // default values are the same

	Id     int // Id is automatically detected as primary key
	Name   string
	Rating int `sql:"-"` // - is used to ignore field

	Books []Book `pg:",many2many:book_genres"` // many to many relation

	ParentId  int     `sql:",null"`
	Subgenres []Genre `pg:",fk:Parent"` // fk specifies prefix for foreign key (ParentId)
}

func (g Genre) String() string {
	return fmt.Sprintf("Genre<Id=%d Name=%q>", g.Id, g.Name)
}

type Author struct {
	ID    int // both "Id" and "ID" are detected as primary key
	Name  string
	Books []Book // has many relation
}

func (a Author) String() string {
	return fmt.Sprintf("Author<ID=%d Name=%q>", a.ID, a.Name)
}

type BookGenre struct {
	TableName struct{} `sql:",alias:bg"` // custom table alias

	BookId  int `sql:",pk"` // pk tag is used to mark field as primary key
	GenreId int `sql:",pk"`

	Genre_Rating int // belongs to and is copied to Genre model
}

type Book struct {
	Id        int
	Title     string
	AuthorID  int
	Author    *Author // has one relation
	EditorID  int
	Editor    *Author   // has one relation
	CreatedAt time.Time `sql:",null"`

	Genres       []Genre       `pg:",many2many:book_genres" gorm:"many2many:book_genres;"` // many to many relation
	Translations []Translation // has many relation
	Comments     []Comment     `pg:",polymorphic:Trackable"` // has many polymorphic relation
}

func (b Book) String() string {
	return fmt.Sprintf("Book<Id=%d Title=%q>", b.Id, b.Title)
}

type Translation struct {
	TableName struct{} `sql:",alias:tr"` // custom table alias

	Id     int
	BookId int
	Book   *Book // belongs to relation
	Lang   string

	Comments []Comment `pg:",polymorphic:Trackable"` // has many polymorphic relation
}

type Comment struct {
	TrackableId   int    `sql:",pk"` // Book.Id or Translation.Id
	TrackableType string `sql:",pk"` // "book" or "translation"
	Text          string
}

func createTestSchema(db *pg.DB) error {
	sql := []string{
		`DROP TABLE IF EXISTS comments`,
		`DROP TABLE IF EXISTS translations`,
		`DROP TABLE IF EXISTS authors`,
		`DROP TABLE IF EXISTS books`,
		`DROP TABLE IF EXISTS genres`,
		`DROP TABLE IF EXISTS book_genres`,
		`CREATE TABLE authors (id serial, name text)`,
		`CREATE TABLE books (id serial PRIMARY KEY, title text, author_id int, editor_id int, created_at timestamptz DEFAULT now())`,
		`CREATE TABLE genres (id serial, name text, parent_id int)`,
		`CREATE TABLE book_genres (book_id int, genre_id int, genre__rating int)`,
		`CREATE TABLE translations (id serial, book_id int, lang varchar(2))`,
		`CREATE TABLE comments (trackable_id int, trackable_type varchar(100), text text)`,
		`CREATE UNIQUE INDEX authors_name ON authors (name)`,
	}
	for _, q := range sql {
		_, err := db.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ = Describe("ORM", func() {
	var db *pg.DB

	BeforeEach(func() {
		db = pg.Connect(pgOptions())

		err := createTestSchema(db)
		Expect(err).NotTo(HaveOccurred())

		genres := []Genre{{
			Id:   1,
			Name: "genre 1",
		}, {
			Id:   2,
			Name: "genre 2",
		}, {
			Id:       3,
			Name:     "subgenre 1",
			ParentId: 1,
		}, {
			Id:       4,
			Name:     "subgenre 2",
			ParentId: 1,
		}}

		err = db.Create(&genres)
		Expect(err).NotTo(HaveOccurred())
		Expect(genres).To(HaveLen(4))

		authors := []Author{{
			ID:   10,
			Name: "author 1",
		}, {
			ID:   11,
			Name: "author 2",
		}, Author{
			ID:   12,
			Name: "author 3",
		}}
		err = db.Create(&authors)
		Expect(err).NotTo(HaveOccurred())
		Expect(authors).To(HaveLen(3))

		books := []Book{{
			Id:       100,
			Title:    "book 1",
			AuthorID: 10,
			EditorID: 11,
		}, {
			Id:       101,
			Title:    "book 2",
			AuthorID: 10,
			EditorID: 12,
		}, Book{
			Id:       102,
			Title:    "book 3",
			AuthorID: 11,
			EditorID: 11,
		}}
		err = db.Create(&books)
		Expect(err).NotTo(HaveOccurred())
		Expect(books).To(HaveLen(3))
		for _, book := range books {
			Expect(book.CreatedAt).To(BeTemporally("~", time.Now(), time.Second))
		}

		bookGenres := []BookGenre{{
			BookId:       100,
			GenreId:      1,
			Genre_Rating: 999,
		}, {
			BookId:       100,
			GenreId:      2,
			Genre_Rating: 9999,
		}, {
			BookId:       101,
			GenreId:      1,
			Genre_Rating: 99999,
		}}
		err = db.Create(&bookGenres)
		Expect(err).NotTo(HaveOccurred())
		Expect(bookGenres).To(HaveLen(3))

		translations := []Translation{{
			Id:     1000,
			BookId: 100,
			Lang:   "ru",
		}, {
			Id:     1001,
			BookId: 100,
			Lang:   "md",
		}, {
			Id:     1002,
			BookId: 101,
			Lang:   "ua",
		}}
		err = db.Create(&translations)
		Expect(err).NotTo(HaveOccurred())
		Expect(translations).To(HaveLen(3))

		comments := []Comment{{
			TrackableId:   100,
			TrackableType: "book",
			Text:          "comment1",
		}, {
			TrackableId:   100,
			TrackableType: "book",
			Text:          "comment2",
		}, {
			TrackableId:   1000,
			TrackableType: "translation",
			Text:          "comment3",
		}}
		err = db.Create(&comments)
		Expect(err).NotTo(HaveOccurred())
		Expect(comments).To(HaveLen(3))
	})

	Describe("struct model", func() {
		It("supports HasOne, HasMany, HasMany2Many, Polymorphic, HasMany -> Polymorphic", func() {
			var book Book
			err := db.Model(&book).
				Column("book.id", "Author.id", "Editor.id", "Genres.id", "Comments", "Translations", "Translations.Comments").
				First()
			Expect(err).NotTo(HaveOccurred())

			Expect(book.Id).To(Equal(100))
			Expect(book.Author.ID).To(Equal(10))

			Expect(book.Genres).To(HaveLen(2))
			genre := book.Genres[0]
			Expect(genre.Id).To(Equal(1))
			genre = book.Genres[1]
			Expect(genre.Id).To(Equal(2))

			Expect(book.Translations).To(HaveLen(2))
			translation := book.Translations[0]
			Expect(translation.Id).To(Equal(1000))
			Expect(translation.BookId).To(Equal(100))
			Expect(translation.Lang).To(Equal("ru"))

			Expect(translation.Comments).To(HaveLen(1))
			comment := translation.Comments[0]
			Expect(comment.Text).To(Equal("comment3"))

			translation = book.Translations[1]
			Expect(translation.Id).To(Equal(1001))
			Expect(translation.BookId).To(Equal(100))
			Expect(translation.Lang).To(Equal("md"))
			Expect(translation.Comments).To(HaveLen(0))

			Expect(book.Comments).To(HaveLen(2))
			comment = book.Comments[0]
			Expect(comment.Text).To(Equal("comment1"))
			comment = book.Comments[1]
			Expect(comment.Text).To(Equal("comment2"))
		})

		It("supports HasMany -> HasOne, HasMany -> HasMany", func() {
			var author Author
			err := db.Model(&author).
				Column("author.*", "Books.Author", "Books.Editor", "Books.Translations").
				First()
			Expect(err).NotTo(HaveOccurred())
			Expect(author.ID).To(Equal(10))

			Expect(author.Books).To(HaveLen(2))

			book := &author.Books[0]
			Expect(book.Id).To(Equal(100))
			Expect(book.Author.ID).To(Equal(10))
			Expect(book.Editor.ID).To(Equal(11))

			Expect(book.Translations).To(HaveLen(2))
			translation := book.Translations[0]
			Expect(translation.BookId).To(Equal(100))
			Expect(translation.Lang).To(Equal("ru"))
			translation = book.Translations[1]
			Expect(translation.BookId).To(Equal(100))
			Expect(translation.Lang).To(Equal("md"))

			book = &author.Books[1]
			Expect(book.Id).To(Equal(101))
			Expect(book.Author.ID).To(Equal(10))
			Expect(book.Editor.ID).To(Equal(12))

			Expect(book.Translations).To(HaveLen(1))
			translation = book.Translations[0]
			Expect(translation.BookId).To(Equal(101))
			Expect(translation.Lang).To(Equal("ua"))
		})

		It("supports HasMany -> HasMany -> HasMany", func() {
			var genre Genre
			err := db.Model(&genre).
				Column("genre.id", "Books.id", "Books.Translations").
				First()
			Expect(err).NotTo(HaveOccurred())
			Expect(genre.Id).To(Equal(1))
			Expect(genre.Rating).To(Equal(0))

			Expect(genre.Books).To(HaveLen(2))
			book := &genre.Books[0]
			Expect(book.Id).To(Equal(100))

			Expect(book.Translations).To(HaveLen(2))
			translation := book.Translations[0]
			Expect(translation.BookId).To(Equal(100))
			Expect(translation.Lang).To(Equal("ru"))
			translation = book.Translations[1]
			Expect(translation.BookId).To(Equal(100))
			Expect(translation.Lang).To(Equal("md"))

			Expect(genre.Books).To(HaveLen(2))
			book = &genre.Books[1]
			Expect(book.Id).To(Equal(101))

			Expect(book.Translations).To(HaveLen(1))
			translation = book.Translations[0]
			Expect(translation.BookId).To(Equal(101))
			Expect(translation.Lang).To(Equal("ua"))
		})
	})

	Describe("slice model", func() {
		It("supports HasOne, HasMany, HasMany2Many", func() {
			var books []Book
			err := db.Model(&books).
				Column("book.id", "Author", "Editor", "Translations", "Genres").
				Order("book.id ASC").
				Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(books).To(HaveLen(3))

			book := &books[0]
			Expect(book.Id).To(Equal(100))
			Expect(book.Author).NotTo(BeNil())
			Expect(book.Author.ID).To(Equal(10))

			Expect(book.Translations).To(HaveLen(2))
			translation := book.Translations[0]
			Expect(translation.BookId).To(Equal(100))
			Expect(translation.Lang).To(Equal("ru"))
			translation = book.Translations[1]
			Expect(translation.BookId).To(Equal(100))
			Expect(translation.Lang).To(Equal("md"))

			Expect(book.Genres).To(HaveLen(2))
			genre := book.Genres[0]
			Expect(genre.Id).To(Equal(1))
			Expect(genre.Rating).To(Equal(999))
			genre = book.Genres[1]
			Expect(genre.Id).To(Equal(2))
			Expect(genre.Rating).To(Equal(9999))

			book = &books[1]
			Expect(book.Id).To(Equal(101))
			Expect(book.Author.ID).To(Equal(10))

			Expect(book.Translations).To(HaveLen(1))
			translation = book.Translations[0]
			Expect(translation.BookId).To(Equal(101))
			Expect(translation.Lang).To(Equal("ua"))

			Expect(book.Genres).To(HaveLen(1))
			genre = book.Genres[0]
			Expect(genre.Id).To(Equal(1))
			Expect(genre.Rating).To(Equal(99999))

			book = &books[2]
			Expect(book.Id).To(Equal(102))
			Expect(book.Author.ID).To(Equal(11))

			Expect(book.Translations).To(HaveLen(0))

			Expect(book.Genres).To(HaveLen(0))
		})

		It("supports HasMany2Many, HasMany2Many -> HasMany", func() {
			var genres []Genre
			err := db.Model(&genres).
				Column("genre.*", "Subgenres", "Books", "Books.Translations").
				Where("genre.parent_id IS NULL").
				Order("genre.id").
				Select()
			Expect(err).NotTo(HaveOccurred())
			Expect(genres).To(HaveLen(2))

			genre := &genres[0]
			Expect(genre.Id).To(Equal(1))
			Expect(genre.Subgenres).To(ConsistOf(
				Genre{Id: 3, Name: "subgenre 1", ParentId: 1},
				Genre{Id: 4, Name: "subgenre 2", ParentId: 1},
			))

			book := genre.Books[0]
			Expect(book.Id).To(Equal(100))

			Expect(book.Translations).To(HaveLen(2))
			Expect(book.Translations).To(ConsistOf(
				Translation{Id: 1000, BookId: 100, Lang: "ru"},
				Translation{Id: 1001, BookId: 100, Lang: "md"},
			))

			book = genre.Books[1]
			Expect(book.Id).To(Equal(101))
			Expect(book.Translations).To(ConsistOf(
				Translation{Id: 1002, BookId: 101, Lang: "ua"},
			))

			genre = &genres[1]
			Expect(genre.Id).To(Equal(2))
			Expect(genre.Subgenres).To(BeEmpty())

			Expect(genre.Books).To(HaveLen(1))
			book = genre.Books[0]
			Expect(book.Id).To(Equal(100))
			Expect(book.Translations).To(ConsistOf(
				Translation{Id: 1000, BookId: 100, Lang: "ru"},
				Translation{Id: 1001, BookId: 100, Lang: "md"},
			))
		})
	})

	It("filters by HasOne", func() {
		var books []Book
		err := db.Model(&books).
			Column("book.id", "Author._").
			Where("author.id = 10").
			Order("book.id ASC").
			Select()
		Expect(err).NotTo(HaveOccurred())
		Expect(books).To(HaveLen(2))
		Expect(books[0].Id).To(Equal(100))
		Expect(books[0].Author).To(BeNil())
		Expect(books[1].Id).To(Equal(101))
		Expect(books[1].Author).To(BeNil())
	})
})
