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
	"github.com/joseluisq/drone/cmd/drone-server/config"
	"github.com/joseluisq/drone/core"
	"github.com/joseluisq/drone/metric"
	"github.com/joseluisq/drone/store/batch"
	"github.com/joseluisq/drone/store/batch2"
	"github.com/joseluisq/drone/store/build"
	"github.com/joseluisq/drone/store/cron"
	"github.com/joseluisq/drone/store/logs"
	"github.com/joseluisq/drone/store/perm"
	"github.com/joseluisq/drone/store/repos"
	"github.com/joseluisq/drone/store/secret"
	"github.com/joseluisq/drone/store/secret/global"
	"github.com/joseluisq/drone/store/shared/db"
	"github.com/joseluisq/drone/store/shared/encrypt"
	"github.com/joseluisq/drone/store/stage"
	"github.com/joseluisq/drone/store/step"
	"github.com/joseluisq/drone/store/user"

	"github.com/google/wire"
)

// wire set for loading the stores.
var storeSet = wire.NewSet(
	provideDatabase,
	provideEncrypter,
	provideBuildStore,
	provideLogStore,
	provideRepoStore,
	provideStageStore,
	provideUserStore,
	provideBatchStore,
	// batch.New,
	cron.New,
	perm.New,
	secret.New,
	global.New,
	step.New,
)

// provideDatabase is a Wire provider function that provides a
// database connection, configured from the environment.
func provideDatabase(config config.Config) (*db.DB, error) {
	return db.Connect(
		config.Database.Driver,
		config.Database.Datasource,
	)
}

// provideEncrypter is a Wire provider function that provides a
// database encrypter, configured from the environment.
func provideEncrypter(config config.Config) (encrypt.Encrypter, error) {
	return encrypt.New(config.Database.Secret)
}

// provideBuildStore is a Wire provider function that provides a
// build datastore, configured from the environment, with metrics
// enabled.
func provideBuildStore(db *db.DB) core.BuildStore {
	builds := build.New(db)
	metric.BuildCount(builds)
	metric.PendingBuildCount(builds)
	metric.RunningBuildCount(builds)
	return builds
}

// provideLogStore is a Wire provider function that provides a
// log datastore, configured from the environment.
func provideLogStore(db *db.DB, config config.Config) core.LogStore {
	s := logs.New(db)
	if config.S3.Bucket != "" {
		p := logs.NewS3Env(
			config.S3.Bucket,
			config.S3.Prefix,
			config.S3.Endpoint,
			config.S3.PathStyle,
		)
		return logs.NewCombined(p, s)
	}
	if config.AzureBlob.ContainerName != "" {
		p := logs.NewAzureBlobEnv(
			config.AzureBlob.ContainerName,
			config.AzureBlob.StorageAccountName,
			config.AzureBlob.StorageAccessKey,
		)
		return logs.NewCombined(p, s)
	}
	return s
}

// provideStageStore is a Wire provider function that provides a
// stage datastore, configured from the environment, with metrics
// enabled.
func provideStageStore(db *db.DB) core.StageStore {
	stages := stage.New(db)
	metric.PendingJobCount(stages)
	metric.RunningJobCount(stages)
	return stages
}

// provideRepoStore is a Wire provider function that provides a
// user datastore, configured from the environment, with metrics
// enabled.
func provideRepoStore(db *db.DB) core.RepositoryStore {
	repos := repos.New(db)
	metric.RepoCount(repos)
	return repos
}

// provideUserStore is a Wire provider function that provides a
// user datastore, configured from the environment, with metrics
// enabled.
func provideUserStore(db *db.DB) core.UserStore {
	users := user.New(db)
	metric.UserCount(users)
	return users
}

// provideBatchStore is a Wire provider function that provides a
// batcher. If the experimental batcher is enabled it is returned.
func provideBatchStore(db *db.DB, config config.Config) core.Batcher {
	if config.Database.LegacyBatch {
		return batch.New(db)
	}
	return batch2.New(db)
}
