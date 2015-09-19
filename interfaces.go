	SetIsEmailConfirmed(bool)
	IsEmailConfirmed() bool

	GetUser() User
	GetUserID() string
	SetData(interface{}) Error
	GetData() (interface{}, Error)
}

type UserToken interface {
	db.Model
	GetType() string
	SetType(string)

	GetToken() string
	SetToken(string)

	GetUserID() string
	SetUserID(string) error

	GetExpiresAt() time.Time
	SetExpiresAt(time.Time)
	Name() string
	SetName(string)

	Debug() bool
	SetDebug(bool)

	Dependencies() Dependencies
	SetDependencies(Dependencies)
	Collection() string
	Hooks() interface{}
	SetHooks(interface{})

	Q() db.Query
	Find(db.Query) ([]db.Model, Error)
	ApiFind(db.Query, Request) Response
	ApiFindPaginated(db.Query, Request) Response
/**
 * Generic service.
 */

type Service interface {
	SetDebug(bool)
	Debug() bool

	Dependencies() Dependencies
	SetDependencies(Dependencies)
}

/**
 * ResourceService.
 */

type ResourceService interface {
	Service

	Q(modelType string) (db.Query, Error)
	FindOne(modelType string, id string) (db.Model, Error)

	Create(db.Model, User) Error
	Update(db.Model, User) Error
	Delete(db.Model, User) Error
}

/**
 * FileService.
 */

type FileService interface {
	Service

	Resource() Resource
	SetResource(Resource)

	Backend(string) FileBackend
	AddBackend(FileBackend)

	DefaultBackend() FileBackend
	SetDefaultBackend(string)

	Model() interface{}
	SetModel(interface{})

	// Given a file instance with a specified bucket, read the file from filePath, upload it
	// to the backend and then store it in the database.
	// If no file.GetBackendName() is empty, the default backend will be used.
	// The file will be deleted if everything succeeds. Otherwise,
	// it will be left in the file system.
	// If deleteDir is true, the directory holding the file will be deleted
	// also.
	BuildFile(file File, user User, filePath string, deleteDir bool) Error

	// Resource callthroughs.
	// The following methods map resource methods for convenience.

	// Create a new file model.
	New() File

	FindOne(id string) (File, Error)
	Find(db.Query) ([]File, Error)

	Create(File, User) Error
	Update(File, User) Error
	Delete(File, User) Error
}

/**
 * EmailService.
 */

type EmailService interface {
	Service

	SetDefaultFrom(EmailRecipient)

	Send(Email) Error
	SendMultiple(...Email) (Error, []Error)
}

	Service

	SetUserResource(Resource)
	SetSessionResource(Resource)
/**
 * Deps.
 */
type Dependencies interface {
	Logger() *logrus.Logger
	SetLogger(*logrus.Logger)
	Config() *config.Config
	SetConfig(*config.Config)
	Cache(name string) Cache
	Caches() map[string]Cache
	AddCache(cache Cache)
	SetCaches(map[string]Cache)
	DefaultBackend() db.Backend
	SetDefaultBackend(db.Backend)
	Backend(name string) db.Backend
	Backends() map[string]db.Backend
	AddBackend(b db.Backend)
	SetBackends(map[string]db.Backend)
	Resource(name string) Resource
	Resources() map[string]Resource
	AddResource(res Resource)
	SetResources(map[string]Resource)
	EmailService() EmailService
	SetEmailService(EmailService)
	FileService() FileService
	SetFileService(FileService)
	ResourceService() ResourceService
	SetResourceService(ResourceService)

	UserService() UserService
	SetUserService(UserService)

	TemplateEngine() TemplateEngine
	SetTemplateEngine(TemplateEngine)
	Dependencies() Dependencies

	Config() *config.Config
	SetConfig(*config.Config)
	RegisterBackend(backend db.Backend)
	RegisterCache(c Cache)
	RegisterResource(Resource)