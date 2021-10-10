// Copyright 2019 Drone IO, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/joseluisq/drone/cmd/drone-server/config"
	"github.com/joseluisq/drone/core"
	"github.com/joseluisq/drone/livelog"
	"github.com/joseluisq/drone/metric/sink"
	"github.com/joseluisq/drone/pubsub"
	"github.com/joseluisq/drone/service/canceler"
	"github.com/joseluisq/drone/service/canceler/reaper"
	"github.com/joseluisq/drone/service/commit"
	contents "github.com/joseluisq/drone/service/content"
	"github.com/joseluisq/drone/service/content/cache"
	"github.com/joseluisq/drone/service/hook"
	"github.com/joseluisq/drone/service/hook/parser"
	"github.com/joseluisq/drone/service/linker"
	"github.com/joseluisq/drone/service/netrc"
	orgs "github.com/joseluisq/drone/service/org"
	"github.com/joseluisq/drone/service/repo"
	"github.com/joseluisq/drone/service/status"
	"github.com/joseluisq/drone/service/syncer"
	"github.com/joseluisq/drone/service/token"
	"github.com/joseluisq/drone/service/transfer"
	"github.com/joseluisq/drone/service/user"
	"github.com/joseluisq/drone/session"
	"github.com/joseluisq/drone/trigger"
	"github.com/joseluisq/drone/trigger/cron"
	"github.com/joseluisq/drone/version"

	"github.com/google/wire"
)

// wire set for loading the services.
var serviceSet = wire.NewSet(
	canceler.New,
	commit.New,
	cron.New,
	livelog.New,
	linker.New,
	parser.New,
	pubsub.New,
	token.Renewer,
	transfer.New,
	trigger.New,
	user.New,

	provideRepositoryService,
	provideContentService,
	provideDatadog,
	provideHookService,
	provideNetrcService,
	provideOrgService,
	provideReaper,
	provideSession,
	provideStatusService,
	provideSyncer,
	provideSystem,
)

// provideContentService is a Wire provider function that
// returns a contents service wrapped with a simple LRU cache.
func provideContentService(client *scm.Client, renewer core.Renewer) core.FileService {
	return cache.Contents(
		contents.New(client, renewer),
	)
}

// provideHookService is a Wire provider function that returns a
// hook service based on the environment configuration.
func provideHookService(client *scm.Client, renewer core.Renewer, config config.Config) core.HookService {
	return hook.New(client, config.Proxy.Addr, renewer)
}

// provideNetrcService is a Wire provider function that returns
// a netrc service based on the environment configuration.
func provideNetrcService(client *scm.Client, renewer core.Renewer, config config.Config) core.NetrcService {
	return netrc.New(
		client,
		renewer,
		config.Cloning.AlwaysAuth,
		config.Cloning.Username,
		config.Cloning.Password,
	)
}

// provideOrgService is a Wire provider function that
// returns an organization service wrapped with a simple cache.
func provideOrgService(client *scm.Client, renewer core.Renewer) core.OrganizationService {
	return orgs.NewCache(orgs.New(client, renewer), 10, time.Minute*5)
}

// provideRepo is a Wire provider function that returns
// a repo based on the environment configuration
func provideRepositoryService(client *scm.Client, renewer core.Renewer, config config.Config) core.RepositoryService {
	return repo.New(
		client,
		renewer,
		config.Repository.Visibility,
		config.Repository.Trusted,
	)
}

// provideSession is a Wire provider function that returns a
// user session based on the environment configuration.
func provideSession(store core.UserStore, config config.Config) (core.Session, error) {
	if config.Session.MappingFile != "" {
		return session.Legacy(store, session.Config{
			Secure:      config.Session.Secure,
			Secret:      config.Session.Secret,
			Timeout:     config.Session.Timeout,
			MappingFile: config.Session.MappingFile,
		})
	}

	return session.New(store, session.NewConfig(
		config.Session.Secret,
		config.Session.Timeout,
		config.Session.Secure),
	), nil
}

// provideUserService is a Wire provider function that returns a
// user service based on the environment configuration.
func provideStatusService(client *scm.Client, renewer core.Renewer, config config.Config) core.StatusService {
	return status.New(client, renewer, status.Config{
		Base:     config.Server.Addr,
		Name:     config.Status.Name,
		Disabled: config.Status.Disabled,
	})
}

// provideSyncer is a Wire provider function that returns a
// repository synchronizer.
func provideSyncer(repoz core.RepositoryService,
	repos core.RepositoryStore,
	users core.UserStore,
	batch core.Batcher,
	config config.Config) core.Syncer {
	sync := syncer.New(repoz, repos, users, batch)
	// the user can define a filter that limits which
	// repositories can be synchronized and stored in the
	// database.
	if filter := config.Repository.Filter; len(filter) > 0 {
		sync.SetFilter(syncer.NamespaceFilter(filter))
	}
	return sync
}

// provideSyncer is a Wire provider function that returns the
// system details structure.
func provideSystem(config config.Config) *core.System {
	return &core.System{
		Proto:   config.Server.Proto,
		Host:    config.Server.Host,
		Link:    config.Server.Addr,
		Version: version.Version.String(),
	}
}

// provideReaper is a Wire provider function that returns the
// zombie build reaper.
func provideReaper(
	repos core.RepositoryStore,
	builds core.BuildStore,
	stages core.StageStore,
	canceler core.Canceler,
	config config.Config,
) *reaper.Reaper {
	return reaper.New(
		repos,
		builds,
		stages,
		canceler,
		config.Cleanup.Running,
		config.Cleanup.Pending,
	)
}

// provideDatadog is a Wire provider function that returns the
// datadog sink.
func provideDatadog(
	users core.UserStore,
	repos core.RepositoryStore,
	builds core.BuildStore,
	system *core.System,
	license *core.License,
	config config.Config,
) *sink.Datadog {
	return sink.New(
		users,
		repos,
		builds,
		*system,
		sink.Config{
			Endpoint:        config.Datadog.Endpoint,
			Token:           config.Datadog.Token,
			License:         license.Kind,
			Licensor:        license.Licensor,
			Subscription:    license.Subscription,
			EnableGithub:    config.IsGitHub(),
			EnableGithubEnt: config.IsGitHubEnterprise(),
			EnableGitlab:    config.IsGitLab(),
			EnableBitbucket: config.IsBitbucket(),
			EnableStash:     config.IsStash(),
			EnableGogs:      config.IsGogs(),
			EnableGitea:     config.IsGitea(),
			EnableAgents:    !config.Agent.Disabled,
			EnableNomad:     config.Nomad.Enabled,
		},
	)
}
