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

package sink

import (
	"fmt"

	"github.com/joseluisq/drone/version"
)

func createTags(config Config) []string {
	tags := []string{
		fmt.Sprintf("version:%s", version.Version),
	}

	switch {
	case config.EnableBitbucket:
		tags = append(tags, "remote:bitbucket:cloud")
	case config.EnableStash:
		tags = append(tags, "remote:bitbucket:server")
	case config.EnableGithubEnt:
		tags = append(tags, "remote:github:enterprise")
	case config.EnableGithub:
		tags = append(tags, "remote:github:cloud")
	case config.EnableGitlab:
		tags = append(tags, "remote:gitlab")
	case config.EnableGogs:
		tags = append(tags, "remote:gogs")
	case config.EnableGitea:
		tags = append(tags, "remote:gitea")
	default:
		tags = append(tags, "remote:undefined")
	}

	switch {
	case config.EnableAgents:
		tags = append(tags, "scheduler:internal:agents")
	case config.EnableNomad:
		tags = append(tags, "scheduler:nomad")
	default:
		tags = append(tags, "scheduler:internal:local")
	}

	if config.Subscription != "" {
		tag := fmt.Sprintf("license:%s:%s:%s",
			config.License,
			config.Licensor,
			config.Subscription,
		)
		tags = append(tags, tag)
	} else if config.Licensor != "" {
		tag := fmt.Sprintf("license:%s:%s",
			config.License,
			config.Licensor,
		)
		tags = append(tags, tag)
	} else {
		tag := fmt.Sprintf("license:%s", config.License)
		tags = append(tags, tag)
	}
	return tags
}
