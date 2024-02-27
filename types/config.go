// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"time"

	"github.com/harness/gitness/blob"
	"github.com/harness/gitness/events"
	gitenum "github.com/harness/gitness/git/enum"
	"github.com/harness/gitness/lock"
	"github.com/harness/gitness/pubsub"
)

// Config stores the system configuration.
type Config struct {
	// InstanceID specifis the ID of the gitness instance.
	// NOTE: If the value is not provided the hostname of the machine is used.
	InstanceID string `envconfig:"GITNESS_INSTANCE_ID"`

	Debug bool `envconfig:"GITNESS_DEBUG"`
	Trace bool `envconfig:"GITNESS_TRACE"`

	// GracefulShutdownTime defines the max time we wait when shutting down a server.
	// 5min should be enough for most git clones to complete.
	GracefulShutdownTime time.Duration `envconfig:"GITNESS_GRACEFUL_SHUTDOWN_TIME" default:"300s"`

	UserSignupEnabled   bool `envconfig:"GITNESS_USER_SIGNUP_ENABLED" default:"true"`
	NestedSpacesEnabled bool `envconfig:"GITNESS_NESTED_SPACES_ENABLED" default:"false"`

	// PublicResourceCreationEnabled specifies whether a user can create publicly accessible resources.
	PublicResourceCreationEnabled bool `envconfig:"GITNESS_PUBLIC_RESOURCE_CREATION_ENABLED" default:"true"`

	Profiler struct {
		Type        string `envconfig:"GITNESS_PROFILER_TYPE"`
		ServiceName string `envconfig:"GITNESS_PROFILER_SERVICE_NAME" default:"gitness"`
	}

	// URL defines the URLs via which the different parts of the service are reachable by.
	URL struct {
		// Base is used to generate external facing URLs in case they aren't provided explicitly.
		// Value is derived from HTTP.Server unless explicitly specified (e.g. http://localhost:3000).
		Base string `envconfig:"GITNESS_URL_BASE"`

		// Git defines the external URL via which the GIT API is reachable.
		// NOTE: for routing to work properly, the request path & hostname reaching gitness
		// have to statisfy at least one of the following two conditions:
		// - Path ends with `/git`
		// - Hostname is different to API hostname
		// (this could be after proxy path / header rewrite).
		// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000/git).
		Git string `envconfig:"GITNESS_URL_GIT"`

		// API defines the external URL via which the rest API is reachable.
		// NOTE: for routing to work properly, the request path reaching gitness has to end with `/api`
		// (this could be after proxy path rewrite).
		// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000/api).
		API string `envconfig:"GITNESS_URL_API"`

		// UI defines the external URL via which the UI is reachable.
		// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000).
		UI string `envconfig:"GITNESS_URL_UI"`

		// Internal defines the internal URL via which the service is reachable.
		// Value is derived from HTTP.Server unless explicitly specified (e.g. http://localhost:3000).
		Internal string `envconfig:"GITNESS_URL_INTERNAL"`

		// Container is the endpoint that can be used by running container builds to communicate
		// with gitness (for example while performing a clone on a local repo).
		// host.docker.internal allows a running container to talk to services exposed on the host
		// (either running directly or via a port exposed in a docker container).
		// Value is derived from HTTP.Server unless explicitly specified (e.g. http://host.docker.internal:3000).
		Container string `envconfig:"GITNESS_URL_CONTAINER"`
	}

	// Git defines the git configuration parameters
	Git struct {
		// Trace specifies whether git operations should be traces.
		// NOTE: Currently limited to 'push' operation until we move to internal command package.
		Trace bool `envconfig:"GITNESS_GIT_TRACE"`
		// DefaultBranch specifies the default branch for new repositories.
		DefaultBranch string `envconfig:"GITNESS_GIT_DEFAULTBRANCH" default:"main"`
		// Root specifies the directory containing git related data (e.g. repos, ...)
		Root string `envconfig:"GITNESS_GIT_ROOT"`
		// TmpDir (optional) specifies the directory for temporary data (e.g. repo clones, ...)
		TmpDir string `envconfig:"GITNESS_GIT_TMP_DIR"`
		// HookPath points to the binary used as git server hook.
		HookPath string `envconfig:"GITNESS_GIT_HOOK_PATH"`

		// LastCommitCache holds configuration options for the last commit cache.
		LastCommitCache struct {
			// Mode determines where the cache will be. Valid values are "inmemory" (default), "redis" or "none".
			Mode gitenum.LastCommitCacheMode `envconfig:"GITNESS_GIT_LAST_COMMIT_CACHE_MODE" default:"inmemory"`

			// Duration defines cache duration of last commit.
			Duration time.Duration `envconfig:"GITNESS_GIT_LAST_COMMIT_CACHE_DURATION" default:"12h"`
		}
	}

	// Encrypter defines the parameters for the encrypter
	Encrypter struct {
		Secret       string `envconfig:"GITNESS_ENCRYPTER_SECRET"` // key used for encryption
		MixedContent bool   `envconfig:"GITNESS_ENCRYPTER_MIXED_CONTENT"`
	}

	// Server defines the server configuration parameters.
	Server struct {
		// HTTP defines the http configuration parameters
		HTTP struct {
			Port  int    `envconfig:"GITNESS_HTTP_PORT" default:"3000"`
			Proto string `envconfig:"GITNESS_HTTP_PROTO" default:"http"`
		}

		// Acme defines Acme configuration parameters.
		Acme struct {
			Enabled bool   `envconfig:"GITNESS_ACME_ENABLED"`
			Endpont string `envconfig:"GITNESS_ACME_ENDPOINT"`
			Email   bool   `envconfig:"GITNESS_ACME_EMAIL"`
			Host    string `envconfig:"GITNESS_ACME_HOST"`
		}
	}

	// CI defines configuration related to build executions.
	CI struct {
		ParallelWorkers int `envconfig:"GITNESS_CI_PARALLEL_WORKERS" default:"2"`
		// PluginsZipURL is a pointer to a zip containing all the plugins schemas.
		// This could be a local path or an external location.
		//nolint:lll
		PluginsZipURL string `envconfig:"GITNESS_CI_PLUGINS_ZIP_URL" default:"https://github.com/bradrydzewski/plugins/archive/refs/heads/master.zip"`

		// ContainerNetworks is a list of networks that all containers created as part of CI
		// should be attached to.
		// This can be needed when we don't want to use host.docker.internal (eg when a service mesh
		// or proxy is being used) and instead want all the containers to run on the same network as
		// the gitness container so that they can interact via the container name.
		// In that case, GITNESS_URL_CONTAINER should also be changed
		// (eg to http://<gitness_container_name>:<port>).
		ContainerNetworks []string `envconfig:"GITNESS_CI_CONTAINER_NETWORKS"`
	}

	// Database defines the database configuration parameters.
	Database struct {
		Driver     string `envconfig:"GITNESS_DATABASE_DRIVER" default:"sqlite3"`
		Datasource string `envconfig:"GITNESS_DATABASE_DATASOURCE" default:"database.sqlite3"`
	}

	// BlobStore defines the blob storage configuration parameters.
	BlobStore struct {
		// Provider is a name of blob storage service like filesystem or gcs
		Provider blob.Provider `envconfig:"GITNESS_BLOBSTORE_PROVIDER" default:"filesystem"`
		// Bucket is a path to the directory where the files will be stored when using filesystem blob storage,
		// in case of gcs provider this will be the actual bucket where the images are stored.
		Bucket string `envconfig:"GITNESS_BLOBSTORE_BUCKET"`

		// In case of GCS provider, this is expected to be the path to the service account key file.
		KeyPath string `envconfig:"GITNESS_BLOBSTORE_KEY_PATH" default:""`

		// Email ID of the google service account that needs to be impersonated
		TargetPrincipal string `envconfig:"GITNESS_BLOBSTORE_TARGET_PRINCIPAL" default:""`

		ImpersonationLifetime time.Duration `envconfig:"GITNESS_BLOBSTORE_IMPERSONATION_LIFETIME" default:"12h"`
	}

	// Token defines token configuration parameters.
	Token struct {
		CookieName string        `envconfig:"GITNESS_TOKEN_COOKIE_NAME" default:"token"`
		Expire     time.Duration `envconfig:"GITNESS_TOKEN_EXPIRE" default:"720h"`
	}

	Logs struct {
		// S3 provides optional storage option for logs.
		S3 struct {
			Bucket    string `envconfig:"GITNESS_LOGS_S3_BUCKET"`
			Prefix    string `envconfig:"GITNESS_LOGS_S3_PREFIX"`
			Endpoint  string `envconfig:"GITNESS_LOGS_S3_ENDPOINT"`
			PathStyle bool   `envconfig:"GITNESS_LOGS_S3_PATH_STYLE"`
		}
	}

	// Cors defines http cors parameters
	Cors struct {
		AllowedOrigins   []string `envconfig:"GITNESS_CORS_ALLOWED_ORIGINS"   default:"*"`
		AllowedMethods   []string `envconfig:"GITNESS_CORS_ALLOWED_METHODS"   default:"GET,POST,PATCH,PUT,DELETE,OPTIONS"`
		AllowedHeaders   []string `envconfig:"GITNESS_CORS_ALLOWED_HEADERS"   default:"Origin,Accept,Accept-Language,Authorization,Content-Type,Content-Language,X-Requested-With,X-Request-Id"` //nolint:lll // struct tags can't be multiline
		ExposedHeaders   []string `envconfig:"GITNESS_CORS_EXPOSED_HEADERS"   default:"Link"`
		AllowCredentials bool     `envconfig:"GITNESS_CORS_ALLOW_CREDENTIALS" default:"true"`
		MaxAge           int      `envconfig:"GITNESS_CORS_MAX_AGE"           default:"300"`
	}

	// Secure defines http security parameters.
	Secure struct {
		AllowedHosts          []string          `envconfig:"GITNESS_HTTP_ALLOWED_HOSTS"`
		HostsProxyHeaders     []string          `envconfig:"GITNESS_HTTP_PROXY_HEADERS"`
		SSLRedirect           bool              `envconfig:"GITNESS_HTTP_SSL_REDIRECT"`
		SSLTemporaryRedirect  bool              `envconfig:"GITNESS_HTTP_SSL_TEMPORARY_REDIRECT"`
		SSLHost               string            `envconfig:"GITNESS_HTTP_SSL_HOST"`
		SSLProxyHeaders       map[string]string `envconfig:"GITNESS_HTTP_SSL_PROXY_HEADERS"`
		STSSeconds            int64             `envconfig:"GITNESS_HTTP_STS_SECONDS"`
		STSIncludeSubdomains  bool              `envconfig:"GITNESS_HTTP_STS_INCLUDE_SUBDOMAINS"`
		STSPreload            bool              `envconfig:"GITNESS_HTTP_STS_PRELOAD"`
		ForceSTSHeader        bool              `envconfig:"GITNESS_HTTP_STS_FORCE_HEADER"`
		BrowserXSSFilter      bool              `envconfig:"GITNESS_HTTP_BROWSER_XSS_FILTER"    default:"true"`
		FrameDeny             bool              `envconfig:"GITNESS_HTTP_FRAME_DENY"            default:"true"`
		ContentTypeNosniff    bool              `envconfig:"GITNESS_HTTP_CONTENT_TYPE_NO_SNIFF"`
		ContentSecurityPolicy string            `envconfig:"GITNESS_HTTP_CONTENT_SECURITY_POLICY"`
		ReferrerPolicy        string            `envconfig:"GITNESS_HTTP_REFERRER_POLICY"`
	}

	Principal struct {
		// System defines the principal information used to create the system service.
		System struct {
			UID         string `envconfig:"GITNESS_PRINCIPAL_SYSTEM_UID"          default:"gitness"`
			DisplayName string `envconfig:"GITNESS_PRINCIPAL_SYSTEM_DISPLAY_NAME" default:"Gitness"`
			Email       string `envconfig:"GITNESS_PRINCIPAL_SYSTEM_EMAIL"        default:"system@gitness.io"`
		}
		// Pipeline defines the principal information used to create the pipeline service.
		Pipeline struct {
			UID         string `envconfig:"GITNESS_PRINCIPAL_PIPELINE_UID"          default:"pipeline"`
			DisplayName string `envconfig:"GITNESS_PRINCIPAL_PIPELINE_DISPLAY_NAME" default:"Gitness Pipeline"`
			Email       string `envconfig:"GITNESS_PRINCIPAL_PIPELINE_EMAIL"        default:"pipeline@gitness.io"`
		}
		// Admin defines the principal information used to create the admin user.
		// NOTE: The admin user is only auto-created in case a password and an email is provided.
		Admin struct {
			UID         string `envconfig:"GITNESS_PRINCIPAL_ADMIN_UID"           default:"admin"`
			DisplayName string `envconfig:"GITNESS_PRINCIPAL_ADMIN_DISPLAY_NAME"  default:"Administrator"`
			Email       string `envconfig:"GITNESS_PRINCIPAL_ADMIN_EMAIL"`    // No default email
			Password    string `envconfig:"GITNESS_PRINCIPAL_ADMIN_PASSWORD"` // No default password
		}
	}

	Redis struct {
		Endpoint           string `envconfig:"GITNESS_REDIS_ENDPOINT"              default:"localhost:6379"`
		MaxRetries         int    `envconfig:"GITNESS_REDIS_MAX_RETRIES"           default:"3"`
		MinIdleConnections int    `envconfig:"GITNESS_REDIS_MIN_IDLE_CONNECTIONS"  default:"0"`
		Password           string `envconfig:"GITNESS_REDIS_PASSWORD"`
		SentinelMode       bool   `envconfig:"GITNESS_REDIS_USE_SENTINEL"          default:"false"`
		SentinelMaster     string `envconfig:"GITNESS_REDIS_SENTINEL_MASTER"`
		SentinelEndpoint   string `envconfig:"GITNESS_REDIS_SENTINEL_ENDPOINT"`
	}

	Events struct {
		Mode                  events.Mode `envconfig:"GITNESS_EVENTS_MODE"                     default:"inmemory"`
		Namespace             string      `envconfig:"GITNESS_EVENTS_NAMESPACE"                default:"gitness"`
		MaxStreamLength       int64       `envconfig:"GITNESS_EVENTS_MAX_STREAM_LENGTH"        default:"10000"`
		ApproxMaxStreamLength bool        `envconfig:"GITNESS_EVENTS_APPROX_MAX_STREAM_LENGTH" default:"true"`
	}

	Lock struct {
		// Provider is a name of distributed lock service like redis, memory, file etc...
		Provider      lock.Provider `envconfig:"GITNESS_LOCK_PROVIDER"          default:"inmemory"`
		Expiry        time.Duration `envconfig:"GITNESS_LOCK_EXPIRE"            default:"8s"`
		Tries         int           `envconfig:"GITNESS_LOCK_TRIES"             default:"8"`
		RetryDelay    time.Duration `envconfig:"GITNESS_LOCK_RETRY_DELAY"       default:"250ms"`
		DriftFactor   float64       `envconfig:"GITNESS_LOCK_DRIFT_FACTOR"      default:"0.01"`
		TimeoutFactor float64       `envconfig:"GITNESS_LOCK_TIMEOUT_FACTOR"    default:"0.25"`
		// AppNamespace is just service app prefix to avoid conflicts on key definition
		AppNamespace string `envconfig:"GITNESS_LOCK_APP_NAMESPACE"     default:"gitness"`
		// DefaultNamespace is when mutex doesn't specify custom namespace for their keys
		DefaultNamespace string `envconfig:"GITNESS_LOCK_DEFAULT_NAMESPACE" default:"default"`
	}

	PubSub struct {
		// Provider is a name of distributed lock service like redis, memory, file etc...
		Provider pubsub.Provider `envconfig:"GITNESS_PUBSUB_PROVIDER"                default:"inmemory"`
		// AppNamespace is just service app prefix to avoid conflicts on channel definition
		AppNamespace string `envconfig:"GITNESS_PUBSUB_APP_NAMESPACE"                default:"gitness"`
		// DefaultNamespace is custom namespace for their channels
		DefaultNamespace string        `envconfig:"GITNESS_PUBSUB_DEFAULT_NAMESPACE" default:"default"`
		HealthInterval   time.Duration `envconfig:"GITNESS_PUBSUB_HEALTH_INTERVAL"   default:"3s"`
		SendTimeout      time.Duration `envconfig:"GITNESS_PUBSUB_SEND_TIMEOUT"      default:"60s"`
		ChannelSize      int           `envconfig:"GITNESS_PUBSUB_CHANNEL_SIZE"      default:"100"`
	}

	BackgroundJobs struct {
		// MaxRunning is maximum number of jobs that can be running at once.
		MaxRunning int `envconfig:"GITNESS_JOBS_MAX_RUNNING" default:"10"`

		// RetentionTime is the duration after which non-recurring,
		// finished and failed jobs will be purged from the DB.
		RetentionTime time.Duration `envconfig:"GITNESS_JOBS_RETENTION_TIME" default:"120h"` // 5 days
	}

	Webhook struct {
		// UserAgentIdentity specifies the identity used for the user agent header
		// IMPORTANT: do not include version.
		UserAgentIdentity string `envconfig:"GITNESS_WEBHOOK_USER_AGENT_IDENTITY" default:"Gitness"`
		// HeaderIdentity specifies the identity used for headers in webhook calls (e.g. X-Gitness-Trigger, ...).
		// NOTE: If no value is provided, the UserAgentIdentity will be used.
		HeaderIdentity      string `envconfig:"GITNESS_WEBHOOK_HEADER_IDENTITY"`
		Concurrency         int    `envconfig:"GITNESS_WEBHOOK_CONCURRENCY" default:"4"`
		MaxRetries          int    `envconfig:"GITNESS_WEBHOOK_MAX_RETRIES" default:"3"`
		AllowPrivateNetwork bool   `envconfig:"GITNESS_WEBHOOK_ALLOW_PRIVATE_NETWORK" default:"false"`
		AllowLoopback       bool   `envconfig:"GITNESS_WEBHOOK_ALLOW_LOOPBACK" default:"false"`
		// RetentionTime is the duration after which webhook executions will be purged from the DB.
		RetentionTime time.Duration `envconfig:"GITNESS_WEBHOOK_RETENTION_TIME" default:"168h"` // 7 days
	}

	Trigger struct {
		Concurrency int `envconfig:"GITNESS_TRIGGER_CONCURRENCY" default:"4"`
		MaxRetries  int `envconfig:"GITNESS_TRIGGER_MAX_RETRIES" default:"3"`
	}

	Metric struct {
		Enabled  bool   `envconfig:"GITNESS_METRIC_ENABLED" default:"true"`
		Endpoint string `envconfig:"GITNESS_METRIC_ENDPOINT" default:"https://stats.drone.ci/api/v1/gitness"`
		Token    string `envconfig:"GITNESS_METRIC_TOKEN"`
	}

	RepoSize struct {
		Enabled     bool          `envconfig:"GITNESS_REPO_SIZE_ENABLED" default:"true"`
		CRON        string        `envconfig:"GITNESS_REPO_SIZE_CRON" default:"0 0 * * *"`
		MaxDuration time.Duration `envconfig:"GITNESS_REPO_SIZE_MAX_DURATION" default:"15m"`
		NumWorkers  int           `envconfig:"GITNESS_REPO_SIZE_NUM_WORKERS" default:"5"`
	}

	CodeOwners struct {
		FilePaths []string `envconfig:"GITNESS_CODEOWNERS_FILEPATH" default:"CODEOWNERS,.harness/CODEOWNERS"`
	}

	SMTP struct {
		Host     string `envconfig:"GITNESS_SMTP_HOST"`
		Port     int    `envconfig:"GITNESS_SMTP_PORT"`
		Username string `envconfig:"GITNESS_SMTP_USERNAME"`
		Password string `envconfig:"GITNESS_SMTP_PASSWORD"`
		FromMail string `envconfig:"GITNESS_SMTP_FROM_MAIL"`
		Insecure bool   `envconfig:"GITNESS_SMTP_INSECURE"`
	}

	Notification struct {
		MaxRetries  int `envconfig:"GITNESS_NOTIFICATION_MAX_RETRIES" default:"3"`
		Concurrency int `envconfig:"GITNESS_NOTIFICATION_CONCURRENCY" default:"4"`
	}

	KeywordSearch struct {
		Concurrency int `envconfig:"GITNESS_KEYWORD_SEARCH_CONCURRENCY" default:"4"`
		MaxRetries  int `envconfig:"GITNESS_KEYWORD_SEARCH_MAX_RETRIES" default:"3"`
	}

	Repos struct {
		// DeletedRetentionTime is the duration after which deleted repositories will be purged.
		DeletedRetentionTime time.Duration `envconfig:"GITNESS_REPOS_DELETED_RETENTION_TIME" default:"2160h"` // 90 days
	}
}
