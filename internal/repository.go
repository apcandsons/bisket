package internal

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type Repository struct {
	AppName          string
	RepoUrl          string
	Vers             []Version
	PreviewVers      []Version
	LatestVer        string
	ApiKey           string
	VersionUpdateFn  func(*Repository, *VersionUpdateEvent) error
	VersionDestroyFn func(*Repository, *VersionDestroyEvent) error
}

type Version struct {
	Tag           string
	ExpectedState AppInstanceState
}

type VersionUpdateEvent struct {
	latestVersion string
}

type VersionDestroyEvent struct {
	destroyedVersion string
}

func (repo *Repository) OnVersionUpdate(fn func(*Repository, *VersionUpdateEvent) error) {
	repo.VersionUpdateFn = fn
}

func (repo *Repository) OnVersionDestroy(fn func(*Repository, *VersionDestroyEvent) error) {
	repo.VersionDestroyFn = fn
}

func (repo *Repository) Init(config *RepoConfig) error {
	splitRepoUrl := strings.Split(config.Github.RepoUrl, "/")
	repo.AppName = splitRepoUrl[len(splitRepoUrl)-1]
	repo.RepoUrl = config.Github.RepoUrl
	slog.Info("Initializing repository: " + repo.RepoUrl)
	// Repo.ApiKey = config.Github.ApiKey

	err := CloneOrPullVersion(repo.RepoUrl, "main")
	if err != nil {
		return err
	}

	err = repo.RefreshTags()
	if err != nil {
		return err
	}

	// Refresh tags
	return nil
}

func (repo *Repository) RefreshTags() error {
	versions, err := getTags()
	if err != nil {
		return err
	}
	repo.Vers = versions
	return nil
}

func CloneOrPullVersion(repoUrl string, version string) error {
	destDir := fmt.Sprintf(".bisket/%s", version)
	if _, err := os.Stat(destDir + "/.git"); err != nil {
		slog.Info("Existing repository not found, cloning repository")
		// git clone --depth 1 --branch <tag_name> <repo_url>
		var cmd *exec.Cmd
		if version == "main" {
			cmd = exec.Command("git", "clone", repoUrl, destDir)
		} else {
			cmd = exec.Command("git", "clone", "--depth", "1", "--branch", version, repoUrl, destDir)
		}
		err := cmd.Run()
		if err != nil {
			return err
		}
	} else {
		slog.Info("Existing repository found, pulling repository")
		cmd := exec.Command("git", "pull")
		cmd.Dir = destDir
		err := cmd.Run()
		if err != nil {
			slog.Warn("Failed to pull repository. Using the existing repository")
		}
	}
	return nil
}

func getTags() ([]Version, error) {
	slog.Info("Fetching tags")
	destDir := fmt.Sprintf(".bisket/main")
	cmd := exec.Command("git", "fetch", "--prune", "origin", "+refs/tags/*:refs/tags/*")
	cmd.Dir = destDir
	err := cmd.Run()
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to fetch tags. Continue using the existing list of tags: %v\n", err))
	}

	cmd = exec.Command("git", "tag", "--list")
	cmd.Dir = destDir
	out, err := cmd.Output()
	if err != nil {
		slog.Error("Error reading tags: %v", err)
		return nil, err
	}

	// Erase the existing tags and preview tags
	var previews []Version
	var versions []Version

	tags := strings.Split(string(out), "\n")
	for _, tag := range tags {
		// first filter everything that begins with "@" then if tag begins with "@preview/" then add to PreviewVers, otherwise add to Vers
		if !strings.HasPrefix(tag, "@") {
			continue
		}
		if strings.HasPrefix(tag, "@preview/") {
			slog.Info(fmt.Sprintf("Detected preview version: %s", tag))
			previews = append(previews, Version{Tag: tag, ExpectedState: Running})
		} else {
			slog.Info(fmt.Sprintf("Detected version: %s", tag))
			versions = append(versions, Version{Tag: tag, ExpectedState: Stopped})
		}
	}

	// TODO: Naive implementation, should be sorted by version
	return append(previews, versions[len(versions)-1]), nil
}

func (repo *Repository) FindVersionByTag(version string) (Version, error) {
	for _, ver := range repo.Vers {
		if ver.Tag == version {
			return ver, nil
		}
	}
	return Version{}, fmt.Errorf("Version not found: %s", version)
}
