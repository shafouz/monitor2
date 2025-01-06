package repositories

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"monitor2/src/alerts"
	database "monitor2/src/db"
	"monitor2/src/db/models"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func RunBySchedule(schedule int, db *database.Database) (int, []error) {
	var errors []error

	repositories, err := database.DB.GetManyRepositoriesBySchedule(schedule)
	if err != nil {
		log.Err(err).Caller().Msg("")
		return 0, errors
	}

	for _, repository := range repositories {
    diff, commit, err := gitPullAndDiff(repository, git.PullOptions{
			RemoteName: repository.Remote,
		})

		if err != nil {
			log.Err(err).Caller().Msg("")
			continue
		}

		log.Info().
			Caller().
			Str("url", repository.Url).
			Str("diff", diff).
			Msg("Successfully pulled and diffed")

		if len(diff) != 0 {
			id := uuid.New().String()

			err := db.CreateDiff(models.Diff{
				Id:   id,
				Body: diff,
				Url:  repository.Url,
        Commit: commit,
			})

			if err != nil {
				log.Err(err).Caller().Msg("")
				continue
			}

			ngrok_url := os.Getenv("NGROK_URL")
			msg := fmt.Sprintf("repo: %s\n%s/diff/%s", repository.Url, ngrok_url, id)
			alerts.Alert(msg, "", "diff")
		}
	}

	return 0, errors
}

func getRepoDir(url string) string {
	tmp := strings.Split(url, "/")
	return tmp[len(tmp)-1]
}

func gitClone(url string, dir string) (*git.Repository, error) {
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})

	if err != nil {
		return nil, err
	}

	return r, nil
}

func gitPullAndDiff(repository models.Repository, pull_opts git.PullOptions) (string, string, error) {
	var watched_files []string
	var repo *git.Repository

	err := json.Unmarshal(repository.WatchedFiles, &watched_files)
	if err != nil {
		log.Err(err).Caller().Msg("")
		return "", "", err
	}

	log.Info().
		Caller().
		Str("url", repository.Url).
		Str("watched_files", string(repository.WatchedFiles)).
		Msg("")

	repo, err = git.PlainOpen(repository.Directory)
	if err != nil {
		if err.Error() == "repository does not exist" {
			repo, err = gitClone(repository.Url, repository.Directory)

			if err != nil {
				log.Err(err).Caller().Msg("")
				return "", "", err
			}
		} else {
			log.Err(err).Caller().Msg("")
			return "", "", err
		}
	}

	old_head, err := repo.Head()
	if err != nil {
		log.Err(err).Caller().Msg("")
		return "", "", err
	}
	old_head_commit, err := repo.CommitObject(old_head.Hash())
	if err != nil {
		log.Err(err).Caller().Msg("")
		return "", "", err
	}

	w, err := repo.Worktree()
	if err != nil {
		log.Err(err).Caller().Msg("")
		return "", "", err
	}
	err = w.Pull(&pull_opts)
	if err != nil && err.Error() != "already up-to-date" {
		log.Err(err).Caller().Msg("")
		return "", "", err
	}
	ref, err := repo.Head()
	if err != nil {
		log.Err(err).Caller().Msg("")
		return "", "", err
	}
	new_head_commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Err(err).Caller().Msg("")
		return "", "", err
	}
	patch, err := old_head_commit.Patch(new_head_commit)
	if err != nil {
		log.Err(err).Caller().Msg("")
		return "", "", err
	}

	return parse_diff(patch.String(), watched_files), new_head_commit.Hash.String(), nil
}

func parse_diff(diff string, changed_files []string) string {
	re := regexp.MustCompile("(?m)^diff --git ")
	split := re.Split(diff, -1)

	var final_diff string
	var hashes []string

	for _, s := range split {
		if len(s) == 0 {
			continue
		}

		first_line := strings.Split(s, "\n")[0]

		for _, changed_file := range changed_files {
			if strings.Contains(first_line, changed_file) {

				hsh := fmt.Sprintf("%x", md5.Sum([]byte(s)))

				if slices.Contains(hashes, hsh) {
					continue
				}

				hashes = append(hashes, hsh)
				final_diff = final_diff + s
			}
		}
	}

	return final_diff
}
