// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/joseluisq/drone/cmd/drone-server/config"
	"github.com/joseluisq/drone/handler/api"
	"github.com/joseluisq/drone/handler/web"
	"github.com/joseluisq/drone/livelog"
	"github.com/joseluisq/drone/operator/manager"
	"github.com/joseluisq/drone/pubsub"
	"github.com/joseluisq/drone/service/canceler"
	"github.com/joseluisq/drone/service/commit"
	"github.com/joseluisq/drone/service/hook/parser"
	"github.com/joseluisq/drone/service/license"
	"github.com/joseluisq/drone/service/linker"
	"github.com/joseluisq/drone/service/token"
	"github.com/joseluisq/drone/service/transfer"
	"github.com/joseluisq/drone/service/user"
	"github.com/joseluisq/drone/store/cron"
	"github.com/joseluisq/drone/store/perm"
	"github.com/joseluisq/drone/store/secret"
	"github.com/joseluisq/drone/store/secret/global"
	"github.com/joseluisq/drone/store/step"
	"github.com/joseluisq/drone/trigger"
	cron2 "github.com/joseluisq/drone/trigger/cron"
)

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Injectors from wire.go:

func InitializeApplication(config2 config.Config) (application, error) {
	client := provideClient(config2)
	refresher := provideRefresher(config2)
	db, err := provideDatabase(config2)
	if err != nil {
		return application{}, err
	}
	userStore := provideUserStore(db)
	renewer := token.Renewer(refresher, userStore)
	commitService := commit.New(client, renewer)
	cronStore := cron.New(db)
	repositoryStore := provideRepoStore(db)
	buildStore := provideBuildStore(db)
	corePubsub := pubsub.New()
	stageStore := provideStageStore(db)
	scheduler := provideScheduler(stageStore, config2)
	statusService := provideStatusService(client, renewer, config2)
	stepStore := step.New(db)
	system := provideSystem(config2)
	webhookSender := provideWebhookPlugin(config2, system)
	coreCanceler := canceler.New(buildStore, corePubsub, repositoryStore, scheduler, stageStore, statusService, stepStore, userStore, webhookSender)
	fileService := provideContentService(client, renewer)
	configService := provideConfigPlugin(client, fileService, config2)
	convertService := provideConvertPlugin(client, config2)
	validateService := provideValidatePlugin(config2)
	triggerer := trigger.New(coreCanceler, configService, convertService, commitService, statusService, buildStore, scheduler, repositoryStore, userStore, validateService, webhookSender)
	cronScheduler := cron2.New(commitService, cronStore, repositoryStore, userStore, triggerer)
	reaper := provideReaper(repositoryStore, buildStore, stageStore, coreCanceler, config2)
	coreLicense := provideLicense(client, config2)
	datadog := provideDatadog(userStore, repositoryStore, buildStore, system, coreLicense, config2)
	logStore := provideLogStore(db, config2)
	logStream := livelog.New()
	netrcService := provideNetrcService(client, renewer, config2)
	encrypter, err := provideEncrypter(config2)
	if err != nil {
		return application{}, err
	}
	secretStore := secret.New(db, encrypter)
	globalSecretStore := global.New(db, encrypter)
	buildManager := manager.New(buildStore, configService, convertService, corePubsub, logStore, logStream, netrcService, repositoryStore, scheduler, secretStore, globalSecretStore, statusService, stageStore, stepStore, system, userStore, webhookSender)
	secretService := provideSecretPlugin(config2)
	registryService := provideRegistryPlugin(config2)
	runner := provideRunner(buildManager, secretService, registryService, config2)
	hookService := provideHookService(client, renewer, config2)
	licenseService := license.NewService(userStore, repositoryStore, buildStore, coreLicense)
	organizationService := provideOrgService(client, renewer)
	permStore := perm.New(db)
	repositoryService := provideRepositoryService(client, renewer, config2)
	session, err := provideSession(userStore, config2)
	if err != nil {
		return application{}, err
	}
	batcher := provideBatchStore(db, config2)
	syncer := provideSyncer(repositoryService, repositoryStore, userStore, batcher, config2)
	transferer := transfer.New(repositoryStore, permStore)
	userService := user.New(client, renewer)
	server := api.New(buildStore, commitService, cronStore, corePubsub, globalSecretStore, hookService, logStore, coreLicense, licenseService, organizationService, permStore, repositoryStore, repositoryService, scheduler, secretStore, stageStore, stepStore, statusService, session, logStream, syncer, system, transferer, triggerer, userStore, userService, webhookSender)
	admissionService := provideAdmissionPlugin(client, organizationService, userService, config2)
	hookParser := parser.New(client)
	coreLinker := linker.New(client)
	middleware := provideLogin(config2)
	options := provideServerOptions(config2)
	webServer := web.New(admissionService, buildStore, client, hookParser, coreLicense, licenseService, coreLinker, middleware, repositoryStore, session, syncer, triggerer, userStore, userService, webhookSender, options, system)
	mainRpcHandlerV1 := provideRPC(buildManager, config2)
	mainRpcHandlerV2 := provideRPC2(buildManager, config2)
	mainHealthzHandler := provideHealthz()
	metricServer := provideMetric(session, config2)
	mainPprofHandler := providePprof(config2)
	mux := provideRouter(server, webServer, mainRpcHandlerV1, mainRpcHandlerV2, mainHealthzHandler, metricServer, mainPprofHandler)
	serverServer := provideServer(mux, config2)
	mainApplication := newApplication(cronScheduler, reaper, datadog, runner, serverServer, userStore)
	return mainApplication, nil
}
