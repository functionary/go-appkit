package appkit

import (
	"io"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/theduke/go-apperror"
	db "github.com/theduke/go-dukedb"
)

type Model interface {
	Collection() string
	GetId() interface{}
	SetId(id interface{}) error
	GetStrId() string
	SetStrId(id string) error
}

/**
 * EventHandler.
 */

type EventHandler func(data interface{})
type AllEventsHandler func(event string, data interface{})

type EventBus interface {
	// Publish registers an event type with the EventBus.
	Publish(event string)

	// Subscribe registers a handler function for events of the given event type.
	Subscribe(event string, handler EventHandler)

	// Trigger triggers an event with the given event data.
	Trigger(event string, data interface{})
}

/**
 * Taskrunner system.
 */

// TaskHandler functions are called to handle a task.
// On success, return (nil, false).
// On error, return an error, and true if the task may be retried, or false
// otherwise.
type TaskHandler func(registry Registry, task Task, progressChan chan Task) (result interface{}, err apperror.Error, canRetry bool)

type TaskOnCompleteHandler func(registry Registry, task Task)

// TaskSpec is a task specification that defines an executable task.
type TaskSpec interface {
	// GetName returns a unique name for the task.
	GetName() string

	// GetAllowedRetries returns the number of allowed retries.
	GetAllowedRetries() int

	// GetRetryInterval returns the time that must pass before a
	// retry is attempted.
	GetRetryInterval() time.Duration

	// GetHandler returns the TaskHandler function that will execute the task.
	GetHandler() TaskHandler

	GetOnCompleteHandler() TaskOnCompleteHandler
}

// Task represents a single task to be executed.
type Task interface {
	// GetId returns the unique task id.
	GetStrId() string

	// GetName Returns the name of the task (see @TaskSpec).
	GetName() string
	SetName(name string)

	GetUserId() interface{}
	SetUserId(id interface{})

	GetRunAt() *time.Time
	SetRunAt(t *time.Time)

	// GetData returns the data associated with the task.
	GetData() interface{}
	SetData(data interface{})

	GetPriority() int
	SetPriority(priority int)

	// GetResult returns the result data omitted by the task.
	GetResult() interface{}

	// SetResult sets the result data omitted by the task.
	SetResult(result interface{})

	GetProgress() int
	SetProgress(p int)

	IsCancelled() bool
	SetIsCancelled(flag bool)

	// TryCount returns the number of times the task has been tried.
	GetTryCount() int
	SetTryCount(count int)

	SetCreatedAt(t time.Time)
	GetCreatedAt() time.Time

	// StartedAt returns a time if the task was started, or zero value otherwise.
	GetStartedAt() *time.Time
	SetStartedAt(t *time.Time)

	// FinishedAt returns the time the task was finished, or zero value.
	GetFinishedAt() *time.Time
	SetFinishedAt(t *time.Time)

	IsRunning() bool
	SetIsRunning(flag bool)

	IsComplete() bool
	SetIsComplete(flag bool)

	IsSuccess() bool
	SetIsSuccess(flag bool)

	// GetError returns the error that occured on the last try, or nil if none.
	GetError() string
	SetError(err string)

	// Returns the log messages the last task run produced.
	GetLog() string
	SetLog(log string)
}

type TaskRunner interface {
	SetRegistry(registry Registry)
	Registry() Registry

	SetBackend(backend db.Backend)
	Backend() db.Backend

	SetMaximumConcurrentTasks(count int)
	MaximumConcurrentTasks() int

	SetTaskCheckInterval(duration time.Duration)
	GetTaskCheckInterval() time.Duration

	RegisterTask(spec TaskSpec)

	// GetTaskSpecs returns a slice with all registered tasks.
	GetTaskSpecs() map[string]TaskSpec

	Run() apperror.Error

	Shutdown() chan bool
}

type TaskService interface {
	Queue(task Task) apperror.Error

	GetTask(id string) (Task, apperror.Error)
}

/**
 * User interfaces.
 */

type Session interface {
	Model
	UserModel

	SetType(string)
	GetType() string

	SetToken(string)
	GetToken() string

	SetStartedAt(time.Time)
	GetStartedAt() time.Time

	SetValidUntil(time.Time)
	GetValidUntil() time.Time

	IsAnonymous() bool
}

type Role interface {
	Model

	GetName() string
	SetName(string)

	GetPermissions() []string
	SetPermissions(permissions []string)
	AddPermission(perm ...string)
	RemovePermission(perm ...string)
	ClearPermissions()
	HasPermission(perm ...string) bool
}

type Permission interface {
	Model

	GetName() string
	SetName(string)
}

type User interface {
	Model

	SetIsActive(bool)
	IsActive() bool

	SetUsername(string)
	GetUsername() string

	SetEmail(string)
	GetEmail() string

	SetIsEmailConfirmed(bool)
	IsEmailConfirmed() bool

	SetLastLogin(time.Time)
	GetLastLogin() time.Time

	SetCreatedAt(time.Time)
	GetCreatedAt() time.Time

	SetUpdatedAt(time.Time)
	GetUpdatedAt() time.Time

	SetProfile(UserProfile)
	GetProfile() UserProfile

	GetData() (interface{}, apperror.Error)
	SetData(interface{}) apperror.Error

	GetRoles() []string
	SetRoles(roles []string)
	ClearRoles()
	AddRole(role ...string)
	RemoveRole(role ...string)
	HasRole(role ...string) bool

	HasPermission(perm ...string) bool
}

type UserModel interface {
	GetUser() User
	SetUser(User)

	GetUserId() interface{}
	SetUserId(id interface{}) error
}

type UserProfile interface {
	Model
	UserModel
}

type AuthItem interface {
	Model
	UserModel
}

type UserToken interface {
	Model
	UserModel

	GetType() string
	SetType(string)

	GetToken() string
	SetToken(string)

	GetExpiresAt() time.Time
	SetExpiresAt(time.Time)

	IsValid() bool
}

type AuthAdaptor interface {
	Name() string

	Backend() db.Backend
	SetBackend(db.Backend)

	RegisterUser(user User, data map[string]interface{}) (AuthItem, apperror.Error)

	// Authenticate  a user based on data map, and return userId or an error.
	// The userId argument may be an empty string if the adaptor has to
	// map the userId.
	Authenticate(userId string, data map[string]interface{}) (string, apperror.Error)
}

/**
 * Api interfaces.
 */

type TransferData interface {
	GetData() interface{}
	SetData(data interface{})

	GetModels() []Model
	SetModels(models []Model)

	GetExtraModels() []Model
	SetExtraModels(models []Model)

	GetMeta() map[string]interface{}
	SetMeta(meta map[string]interface{})

	GetErrors() []apperror.Error
	SetErrors(errors []apperror.Error)
}

type Request interface {
	// GetFrontend returns the name of the frontend this request is from.
	GetFrontend() string
	SetFrontend(name string)

	GetPath() string
	SetPath(path string)

	GetHttpMethod() string
	SetHttpMethod(method string)

	GetContext() *Context
	SetContext(context *Context)

	GetTransferData() TransferData
	SetTransferData(data TransferData)

	// Convenience helper for .GetTransferData().GetMeta().
	GetMeta() *Context

	GetData() interface{}
	SetData(data interface{})

	GetRawData() []byte
	SetRawData(data []byte)

	// Parse json contained in RawData and extract data and meta.
	ParseJsonData() apperror.Error

	Unserialize(serializer Serializer) apperror.Error

	GetUser() User
	SetUser(User)

	GetSession() Session
	SetSession(Session)

	GetHttpRequest() *http.Request
	SetHttpRequest(request *http.Request)

	ReadHttpBody() apperror.Error

	GetHttpResponseWriter() http.ResponseWriter
	SetHttpResponseWriter(writer http.ResponseWriter)
}

type Response interface {
	GetError() apperror.Error

	GetHttpStatus() int
	SetHttpStatus(int)

	GetMeta() map[string]interface{}
	SetMeta(meta map[string]interface{})

	GetTransferData() TransferData
	SetTransferData(data TransferData)

	GetData() interface{}
	SetData(interface{})

	GetRawData() []byte
	SetRawData([]byte)

	GetRawDataReader() io.ReadCloser
	SetRawDataReader(io.ReadCloser)
}

type RequestHandler func(Registry, Request) (Response, bool)
type AfterRequestMiddleware func(Registry, Request, Response) (Response, bool)

type HttpRoute interface {
	Route() string
	Method() string
	Handler() RequestHandler
}

/**
 * Serializers.
 */

type Serializer interface {
	Name() string

	// SerializeModel converts a model into the target format.
	SerializeModel(model Model) (modelData interface{}, extraModels []interface{}, err apperror.Error)

	// UnserializeModel converts serialized data into a Model.
	// collection argument is optional, but has to be supplied if the
	// collection can not be extracted from data.
	UnserializeModel(collection string, data interface{}) (Model, apperror.Error)

	SerializeTransferData(data TransferData) (interface{}, apperror.Error)
	UnserializeTransferData(data interface{}) (TransferData, apperror.Error)

	// SerializeResponse converts a response with model data into the target format.
	SerializeResponse(response Response) (interface{}, apperror.Error)

	// MustSerializeResponse serializes a response, and returns properly serialized error data
	// if any error occurs.
	// In addition, an error is returned if one occured, to allow logging the error.
	MustSerializeResponse(response Response) (interface{}, apperror.Error)

	// UnserializeRequest converts request data into a request object.
	UnserializeRequest(data interface{}, request Request) apperror.Error
}

/**
 * Caches.
 */

type CacheItem interface {
	Model

	GetKey() string
	SetKey(string)

	GetValue() interface{}
	SetValue(interface{})

	ToString() (string, apperror.Error)
	FromString(string) apperror.Error

	GetExpiresAt() time.Time
	SetExpiresAt(time.Time)
	IsExpired() bool

	GetTags() []string
	SetTags([]string)
}

type Cache interface {
	Name() string
	SetName(string)

	// Save a new item into the cache.
	Set(CacheItem) apperror.Error
	SetString(key string, value string, expiresAt *time.Time, tags []string) apperror.Error

	// Retrieve a cache item from the cache.
	Get(key string, item ...CacheItem) (CacheItem, apperror.Error)
	GetString(key string) (string, apperror.Error)

	// Delete item from the cache.
	Delete(key ...string) apperror.Error

	// Get all keys stored in the cache.
	Keys() ([]string, apperror.Error)

	// Return all keys that have a certain tag.
	KeysByTags(tag ...string) ([]string, apperror.Error)

	// Clear all items from the cache.
	Clear() apperror.Error

	// Clear all items with the specified tags.
	ClearTag(tag string) apperror.Error

	// Clean up all expired entries.
	Cleanup() apperror.Error
}

/**
 * Emails.
 */

type EmailRecipient interface {
	GetEmail() string
	GetName() string
}

type EmailPart interface {
	GetMimeType() string
	GetContent() []byte
	GetFilePath() string
	GetReader() io.ReadCloser
}

type Email interface {
	SetFrom(email, name string)
	GetFrom() EmailRecipient

	AddTo(email, name string)
	GetTo() []EmailRecipient

	AddCc(email, name string)
	GetCc() []EmailRecipient

	AddBcc(email, name string)
	GetBcc() []EmailRecipient

	SetSubject(string)
	GetSubject() string

	SetBody(contentType string, body []byte)
	AddBody(contentType string, body []byte)
	GetBodyParts() []EmailPart

	Attach(contentType string, data []byte) apperror.Error
	AttachReader(contentType string, reader io.ReadCloser) apperror.Error
	AttachFile(path string) apperror.Error

	GetAttachments() []EmailPart

	Embed(contentType string, data []byte) apperror.Error
	EmbedReader(contentType string, reader io.ReadCloser) apperror.Error
	EmbedFile(path string) apperror.Error

	GetEmbeddedAttachments() []EmailPart

	SetHeader(name string, values ...string)
	SetHeaders(map[string][]string)
}

/**
 * TemplateEngine.
 */

type TemplateEngine interface {
	Build(name string, tpl string) (interface{}, apperror.Error)
	BuildFile(name string, paths ...string) (interface{}, apperror.Error)

	GetTemplate(name string) interface{}

	BuildAndRender(name string, tpl string, data interface{}) ([]byte, apperror.Error)
	BuildFileAndRender(name string, data interface{}, paths ...string) ([]byte, apperror.Error)

	Render(name string, data interface{}) ([]byte, apperror.Error)

	// Clean up all templates.
	Clear()
}

/**
 * Method.
 */

type MethodHandler func(registry Registry, r Request, unblock func()) Response

type Method interface {
	GetName() string
	IsBlocking() bool
	GetHandler() MethodHandler
}

/**
 * Resource.
 */

type Resource interface {
	Debug() bool
	SetDebug(bool)

	Registry() Registry
	SetRegistry(Registry)

	Backend() db.Backend
	SetBackend(db.Backend)

	ModelInfo() *db.ModelInfo

	IsPublic() bool

	Collection() string
	Model() Model
	SetModel(Model)
	CreateModel() Model

	Hooks() interface{}
	SetHooks(interface{})

	Q() *db.Query

	Query(query *db.Query, targetSlice ...interface{}) ([]Model, apperror.Error)
	FindOne(id interface{}) (Model, apperror.Error)

	Count(query *db.Query) (int, apperror.Error)

	ApiFindOne(string, Request) Response
	ApiFind(*db.Query, Request) Response

	Create(obj Model, user User) apperror.Error
	ApiCreate(obj Model, r Request) Response

	Update(obj Model, user User) apperror.Error
	// Updates the model by loading the current version from the database
	// and setting the changed values.
	PartialUpdate(obj Model, user User) apperror.Error

	ApiUpdate(obj Model, r Request) Response
	// See PartialUpdate.
	ApiPartialUpdate(obj Model, request Request) Response

	Delete(obj Model, user User) apperror.Error
	ApiDelete(id string, r Request) Response
}

/**
 * Generic service.
 */

type Service interface {
	SetDebug(bool)
	Debug() bool

	Registry() Registry
	SetRegistry(Registry)
}

/**
 * ResourceService.
 */

type ResourceService interface {
	Service

	Q(modelType string) (*db.Query, apperror.Error)
	FindOne(modelType string, id string) (Model, apperror.Error)

	Create(Model, User) apperror.Error
	Update(Model, User) apperror.Error
	Delete(Model, User) apperror.Error
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

	Model() Model
	SetModel(model Model)

	// Given a file instance with a specified bucket, read the file from filePath, upload it
	// to the backend and then store it in the database.
	// If no file.GetBackendName() is empty, the default backend will be used.
	// The file will be deleted if everything succeeds. Otherwise,
	// it will be left in the file system.
	// If deleteDir is true, the directory holding the file will be deleted
	// also.
	BuildFile(file File, user User, deleteDir, deleteFile bool) apperror.Error

	BuildFileFromPath(bucket, path string, deleteFile bool) (File, apperror.Error)

	// Resource callthroughs.
	// The following methods map resource methods for convenience.

	// Create a new file model.
	New() File

	FindOne(id string) (File, apperror.Error)
	Find(*db.Query) ([]File, apperror.Error)

	Create(File, User) apperror.Error
	Update(File, User) apperror.Error
	Delete(File, User) apperror.Error

	DeleteById(id interface{}, user User) apperror.Error
}

/**
 * EmailService.
 */

type EmailService interface {
	Service

	SetDefaultFrom(EmailRecipient)

	Send(Email) apperror.Error
	SendMultiple(...Email) (apperror.Error, []apperror.Error)
}

/**
 * UserService.
 */

type UserService interface {
	Service

	AuthAdaptor(name string) AuthAdaptor
	AddAuthAdaptor(a AuthAdaptor)

	UserResource() Resource
	SetUserResource(Resource)

	ProfileResource() Resource
	SetProfileResource(resource Resource)

	ProfileModel() UserProfile

	SessionResource() Resource
	SetSessionResource(Resource)

	SetRoleResource(Resource)
	RoleResource() Resource

	SetPermissionResource(Resource)
	PermissionResource() Resource

	// Build a user token, persist it and return it.
	BuildToken(typ, userId string, expiresAt time.Time) (UserToken, apperror.Error)

	// Return a full user with roles and the profile joined.
	FindUser(userId interface{}) (User, apperror.Error)

	CreateUser(user User, adaptor string, data map[string]interface{}) apperror.Error
	AuthenticateUser(userIdentifier string, adaptor string, data map[string]interface{}) (User, apperror.Error)
	StartSession(user User, sessionType string) (Session, apperror.Error)
	VerifySession(token string) (User, Session, apperror.Error)

	SendConfirmationEmail(User) apperror.Error
	ConfirmEmail(token string) (User, apperror.Error)

	ChangePassword(user User, newPassword string) apperror.Error

	SendPasswordResetEmail(User) apperror.Error
	ResetPassword(token, newPassword string) (User, apperror.Error)
}

/**
 * Files.
 */

type BucketConfig interface {
}

type ReadSeekerCloser interface {
	io.Reader
	io.Closer
	io.Seeker
}

// Interface for a File stored in a database backend.
type File interface {
	Model
	// File can belong to a user.
	UserModel

	// Retrieve the backend the file is stored in or should be stored in.
	// WARNING: can return nil.
	GetBackend() FileBackend
	SetBackend(FileBackend)

	GetBackendName() string
	SetBackendName(string)

	GetBackendId() string
	SetBackendId(string) error

	// File bucket.
	GetBucket() string
	SetBucket(string)

	GetTmpPath() string
	SetTmpPath(path string)

	// File name without extension.
	GetName() string
	SetName(string)

	// File extension if available.
	GetExtension() string
	SetExtension(string)

	// Name with extension.
	GetFullName() string
	SetFullName(string)

	GetTitle() string
	SetTitle(string)

	GetDescription() string
	SetDescription(string)

	// File size in bytes if available.
	GetSize() int64
	SetSize(int64)

	// Mime type if available.
	GetMime() string
	SetMime(string)

	GetMediaType() string
	SetMediaType(t string)

	GetIsImage() bool
	SetIsImage(bool)

	// File width and hight in pixels for images and videos.
	GetWidth() int
	SetWidth(int)

	GetHeight() int
	SetHeight(int)

	GetHash() string
	SetHash(hash string)

	GetData() map[string]interface{}
	SetData(data map[string]interface{})

	GetType() string
	SetType(t string)

	GetWeight() int
	SetWeight(weight int)

	GetParentFile() File
	SetParentFile(f File)

	GetParentFileId() interface{}
	SetParentFileId(id interface{})

	GetRelatedFiles() []File
	SetRelatedFiles(files []File)

	// Get a reader for the file.
	// Might return an error if the file does not exist in the backend,
	// or it is not connected to a backend.
	Reader() (ReadSeekerCloser, apperror.Error)

	// Base64 returns the file contents as a base64 encoded string.
	// Can only be called if Backend is set on the file.
	Base64() (string, apperror.Error)

	// Get a writer for the file.
	// Might return an error if the file is not connected to a backend.
	Writer(create bool) (string, io.WriteCloser, apperror.Error)
}

type FileBackend interface {
	Name() string
	SetName(string)

	// Lists the buckets that currently exist.
	Buckets() ([]string, apperror.Error)

	// Check if a Bucket exists.
	HasBucket(string) (bool, apperror.Error)

	// Create a bucket.
	CreateBucket(string, BucketConfig) apperror.Error

	// Return the configuration for a a bucket.
	BucketConfig(string) BucketConfig

	// Change the configuration for a bucket.
	ConfigureBucket(string, BucketConfig) apperror.Error

	// Delete all files in a bucket.
	ClearBucket(bucket string) apperror.Error

	DeleteBucket(bucket string) apperror.Error

	// Clear all buckets.
	ClearAll() apperror.Error

	// Return the ids of all files in a bucket.
	FileIds(bucket string) ([]string, apperror.Error)

	HasFile(File) (bool, apperror.Error)
	HasFileById(bucket, id string) (bool, apperror.Error)

	FileSize(file File) (int64, apperror.Error)
	FileSizeById(bucket, id string) (int64, apperror.Error)

	DeleteFile(File) apperror.Error
	DeleteFileById(bucket, id string) apperror.Error

	// Retrieve a reader for a file.
	Reader(File) (ReadSeekerCloser, apperror.Error)
	// Retrieve a reader for a file in a bucket.
	ReaderById(bucket, id string) (ReadSeekerCloser, apperror.Error)

	// Retrieve a writer for a file in a bucket.
	Writer(f File, create bool) (string, io.WriteCloser, apperror.Error)
	// Retrieve a writer for a file in a bucket.
	WriterById(bucket, id string, create bool) (string, io.WriteCloser, apperror.Error)
}

/**
 * Registry.
 */

type Registry interface {
	// App.

	App() App
	SetApp(app App)

	// Logger.

	Logger() *logrus.Logger
	SetLogger(logger *logrus.Logger)

	// EventBus.

	EventBus() EventBus
	SetEventBus(bus EventBus)

	// Config.

	Config() Config
	SetConfig(cfg Config)

	// Caches.

	DefaultCache() Cache
	SetDefaultCache(cache Cache)

	Cache(name string) Cache
	Caches() map[string]Cache
	AddCache(cache Cache)
	SetCaches(caches map[string]Cache)

	// Backends.

	DefaultBackend() db.Backend
	SetDefaultBackend(db.Backend)
	Backend(name string) db.Backend
	Backends() map[string]db.Backend
	AddBackend(b db.Backend)
	SetBackends(backends map[string]db.Backend)

	// AllModelInfo returns the model info from all registered backends.
	AllModelInfo() map[string]*db.ModelInfo

	// Resources.

	Resource(name string) Resource
	Resources() map[string]Resource
	AddResource(res Resource)
	SetResources(resources map[string]Resource)

	// Frontends.

	Frontend(name string) Frontend
	Frontends() map[string]Frontend
	AddFrontend(frontend Frontend)
	SetFrontends(frontends map[string]Frontend)
	HttpFrontend() HttpFrontend

	// Methods.

	Method(name string) Method
	Methods() map[string]Method
	AddMethod(method Method)
	SetMethods(methods map[string]Method)

	// Serializers.

	DefaultSerializer() Serializer
	SetDefaultSerializer(serialzier Serializer)

	Serializer(name string) Serializer
	Serializers() map[string]Serializer
	AddSerializer(serializer Serializer)
	SetSerializers(serializers map[string]Serializer)

	// Services.

	TaskService() TaskService
	SetTaskService(service TaskService)

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

	Get(name string) interface{}
	Set(name string, val interface{})
}

type Config interface {
	GetData() interface{}

	// ENV returns the current env.
	ENV() string

	Debug() bool

	// TmpDir returns an absolute path to the used tmp directory.
	TmpDir() string

	// DataDir returns an absolute path to the used data directory.
	DataDir() string

	Get(path string) (Config, error)

	// Bool returns a bool value stored at path, or an error if not found or not a  bool.
	Bool(path string) (bool, error)

	// UBool returns a bool value stored at path, the supplied default value or false.
	UBool(path string, defaults ...bool) bool

	// Float64 returns the float64 value stored at path, or an error if not found or wrong type.
	Float64(path string) (float64, error)

	// UFloat64 returns a float64 value stored at path, the supplied default, or 0.
	UFloat64(path string, defaults ...float64) float64

	// Int returns the int value stored at path, or an error if not found or wrong type.
	Int(path string) (int, error)

	// UInt returns an int value stored at path, the supplied default, or 0.
	UInt(path string, defaults ...int) int

	// List returns the list stored at path, or an error if not found or wrong type.
	List(path string) ([]interface{}, error)

	// UList returns the list value stored at path, the supplied default, or nil.
	UList(path string, defaults ...[]interface{}) []interface{}

	// Map returns the map stored at path, or an error if not found or wrong type.
	Map(path string) (map[string]interface{}, error)

	// UMap returns the map stored at path, the supplied default, or nil.
	UMap(path string, defaults ...map[string]interface{}) map[string]interface{}

	// String returns the string value stored at path, or an error if not found or wrong type.
	String(path string) (string, error)

	// UString returns the string value stored at path, the supplied default, or "".
	UString(path string, defaults ...string) string

	// Path returns the absolute version of a file system path stored at config path, or an error if not found or wrong type.
	// If the path in the config  is relative, it will be prefixed with either
	// the config.rootPath or the working directory.
	Path(string) (string, error)

	// UPath returns the absolute version of a file system path stored at config path, the supplied default, or "".
	// If the path in the config  is relative, it will be prefixed with either
	// the config.rootPath or the working directory.
	UPath(path string, defaults ...string) string

	// Set updates a config value to the specified value.
	// If the path is already set, and you supply a different value type, an
	// error will be returned.
	Set(path string, val interface{}) error
}

/**
 * Frontend interfaces.
 */

type Frontend interface {
	Name() string

	Registry() Registry
	SetRegistry(registry Registry)

	Debug() bool
	SetDebug(bool)

	Logger() *logrus.Logger

	// Middleware methods.

	RegisterBeforeMiddleware(handler RequestHandler)
	BeforeMiddlewares() []RequestHandler
	SetBeforeMiddlewares(middlewares []RequestHandler)

	RegisterAfterMiddleware(middleware AfterRequestMiddleware)
	AfterMiddlewares() []AfterRequestMiddleware
	SetAfterMiddlewares(middlewares []AfterRequestMiddleware)

	Init() apperror.Error
	Start() apperror.Error

	Shutdown() (shutdownChan chan bool, err apperror.Error)
}

type HttpFrontend interface {
	Frontend

	Router() *httprouter.Router

	ServeFiles(route, path string)

	// HTTP related methods.

	NotFoundHandler() RequestHandler
	SetNotFoundHandler(x RequestHandler)

	RegisterHttpHandler(method, path string, handler RequestHandler)
}

/**
 * App interfaces.
 */

type App interface {
	// InstanceId returns a unique Id for the app instance.
	InstanceId() string

	// SetInstanceId sets the unique instance id for the app instance.
	SetInstanceId(id string)

	Debug() bool
	SetDebug(bool)

	Registry() Registry

	Logger() *logrus.Logger
	SetLogger(*logrus.Logger)

	SetConfig(Config)
	ReadConfig(path string)

	// Backend methods.
	RegisterBackend(backend db.Backend)

	// PrepareBackends prepares all backends for usage by building relationship information.
	PrepareBackends()

	MigrateBackend(name string, version int, force bool) apperror.Error
	MigrateAllBackends(force bool) apperror.Error
	DropBackend(name string) apperror.Error
	DropAllBackends() apperror.Error
	RebuildBackend(name string) apperror.Error
	RebuildAllBackends() apperror.Error

	// Cache methods.

	RegisterCache(cache Cache)

	// UserService methods.

	RegisterUserService(h UserService)

	// FileService methods.

	RegisterFileService(f FileService)

	// Email methods.

	RegisterEmailService(s EmailService)

	// TemplateEngine methods.

	RegisterTemplateEngine(e TemplateEngine)

	RegisterHttpHandler(method, path string, handler RequestHandler)

	// Method methods.

	RegisterMethod(method Method)
	RunMethod(name string, r Request, responder func(Response), withFinishedChannel bool) (chan bool, apperror.Error)

	// Resource methodds.

	RegisterResource(resource Resource)

	// Frontend methods.
	RegisterFrontend(frontend Frontend)

	// Serializer.
	RegisterSerializer(serializer Serializer)

	// Build all default services.
	Defaults()

	Run()
	RunCli()
	Shutdown() (shutdownChan chan bool, err apperror.Error)
}
