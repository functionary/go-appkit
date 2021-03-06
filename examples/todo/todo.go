package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	_ "github.com/lib/pq"

	db "github.com/theduke/go-dukedb"
	"github.com/theduke/go-dukedb/backends/sql"

	kit "github.com/app-kit/go-appkit"
	kitapp "github.com/app-kit/go-appkit/app"
	"github.com/app-kit/go-appkit/caches/fs"
	"github.com/app-kit/go-appkit/email"
	"github.com/app-kit/go-appkit/files"
	"github.com/app-kit/go-appkit/resources"
	"github.com/app-kit/go-appkit/users"
)

type Project struct {
	Id   uint64 `gorm:"primary_key"`
	Name string

	Todos []*Todo

	Todo   *Todo
	TodoId uint64
}

func (b Project) GetId() string {
	return strconv.FormatUint(b.Id, 10)
}

func (b Project) SetId(rawId string) error {
	id, err := strconv.ParseUint(rawId, 10, 64)
	if err != nil {
		return err
	}
	b.Id = id
	return nil
}

func (p Project) Collection() string {
	return "projects"
}

type ProjectHooks struct {
}

func (p ProjectHooks) BeforeCreate(res kit.Resource, obj kit.Model, user kit.User) apperror.Error {
	log.Printf("obj: %+v\n", obj)
	return nil
}

type Todo struct {
	Id uint64 `gorm:"primary_key"`

	Name     string
	Comments string
	DueDate  time.Time
	Priority int

	Project   *Project
	ProjectId uint64
}

func (b Todo) GetId() string {
	return strconv.FormatUint(b.Id, 10)
}

func (b Todo) SetId(rawId string) error {
	id, err := strconv.ParseUint(rawId, 10, 64)
	if err != nil {
		return err
	}
	b.Id = id
	return nil
}

func (t Todo) Collection() string {
	return "todos"
}

func InitMigrations(app kit.App) {
	handler := app.Backend("sql").(db.MigrationBackend).GetMigrationHandler()

	userMigrations := users.GetUserMigrations(app)
	handler.Add(userMigrations[0])
	//handler.Add(userMigrations[1])

	v2 := db.Migration{
		Name: "create tables",
		Up: func(b db.MigrationBackend) error {
			if err := b.CreateCollection("todos"); err != nil {
				return err
			}
			if err := b.CreateCollection("projects"); err != nil {
				return err
			}
			if err := b.CreateCollection("files"); err != nil {
				return err
			}

			return nil
		},
	}
	handler.Add(v2)
}

func start() error {
	app := kitapp.NewApp("")

	// Build backend.
	backend, err := sql.New("postgres", "postgres://test:test@localhost/test?sslmode=disable")
	if err != nil {
		return err
	}
	backend.SetDebug(true)
	app.RegisterBackend(backend)

	// Build cache.
	fsCache, err := fs.New("tmp/cache")
	if err != nil {
		return err
	}
	app.RegisterCache(fsCache)

	userHandler := users.NewService(nil, backend, nil)
	app.RegisterUserService(userHandler)

	fileHandler := files.NewFileServiceWithFs(nil, "data")
	app.RegisterFileService(fileHandler)

	app.RegisterResource(resources.NewResource(&Project{}, ProjectHooks{}, true))
	app.RegisterResource(resources.NewResource(&Todo{}, nil, true))

	app.PrepareBackends()

	todoMethod := kitapp.NewMethod("todo-count", func(app kit.App, request kit.Request, unblock func()) kit.Response {
		//todos := app.GetResource("projects")
		//count, _ := todos.GetQuery().Last()
		//count := 10

		return &kit.AppResponse{
			Data: 22,
		}
	})
	app.RegisterMethod(todoMethod)

	InitMigrations(app)

	fs := fileHandler.Backend("fs")
	if ok, _ := fs.HasBucket("test"); !ok {
		fs.CreateBucket("test", nil)
	}

	dirs, _ := ioutil.ReadDir("tmp/uploads")
	for _, dir := range dirs {
		if dir.IsDir() {
			dirs, _ := ioutil.ReadDir("tmp/uploads/" + dir.Name())

			if len(dirs) < 1 {
				continue
			}

			filepath := "tmp/uploads/" + dir.Name() + "/" + dirs[0].Name()
			log.Printf("Adding file " + filepath)

			f := fileHandler.New()
			f.SetBucket("test")

			err := fileHandler.BuildFile(f, nil, filepath, true, true)
			log.Printf("file: %+v\nerr: %v", f, err)
		}
	}

	// Send an email.
	e := email.NewMail()
	e.SetSubject("testSubject")
	e.AddTo("reg@theduke.at", "Christoph Herzog")
	e.AddBody("text/plain", []byte("Hallo du lulu"))
	e.SetFrom("reg@theduke.at", "pfuscher")
	app.EmailService().Send(e)

	app.RunCli()

	return nil
}

func main() {
	err := start()
	log.Printf("error: %v\n", err)
}
