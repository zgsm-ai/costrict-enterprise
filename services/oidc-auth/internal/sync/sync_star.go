package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/zgsm-ai/oidc-auth/internal/constants"
	"github.com/zgsm-ai/oidc-auth/internal/repository"
	"github.com/zgsm-ai/oidc-auth/pkg/log"
)

// SyncStar GitHub star synchronization service
type SyncStar struct {
	Enabled       bool
	PersonalToken string        `json:"PersonalToken" mapstructure:"personalToken" validate:"required"`
	Owner         string        `json:"owner" mapstructure:"owner" validate:"required"`
	Repo          string        `json:"repo" mapstructure:"repo" validate:"required"`
	Interval      time.Duration `json:"interval" mapstructure:"interval" validate:"required"`
	HTTPClient    *http.Client
}

// UserInfo GitHub user information
type UserInfo struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

// StargazerEntry Stargazer entry
type StargazerEntry struct {
	User      UserInfo  `json:"user"`
	StarredAt time.Time `json:"starred_at"`
}

// ProcessedStargazer Processed stargazer
type ProcessedStargazer struct {
	UserLogin     string
	UserID        int
	StarredAtDB   time.Time
	StarredAtUnix int64
}

var (
	Owner string
	Repo  string
)

func (s *SyncStar) StarCount() (int, error) {
	starURL := fmt.Sprintf("%s/%s/%s", constants.GitHubStarBaseURL, s.Owner, s.Repo)

	req, err := http.NewRequest("GET", starURL, nil)
	if err != nil {
		log.Error(nil, "failed to create request: %v", err)
		return 0, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", s.PersonalToken))
	req.Header.Set("Accept", "application/vnd.github.star+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := s.HTTPClient
	resp, err := client.Do(req)

	if err != nil {
		log.Error(nil, "Failed to get stargazers: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error(nil, "failed to get stargazers: %s", resp.Status)
		return 0, fmt.Errorf("failed to get stargazers: %s", resp.Status)
	}

	var stargazers map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&stargazers); err != nil {
		log.Error(nil, "failed to decode stargazers: %v", err)
		return 0, err
	}

	num := int(stargazers["stargazers_count"].(float64))

	log.Info(nil, "got stargazers count: %d", num)
	return num, nil
}

// Stargazers synchronizes GitHub stargazer data
func (s *SyncStar) Stargazers() error {
	starURL := fmt.Sprintf("%s/%s/%s/stargazers", constants.GitHubStarBaseURL, s.Owner, s.Repo)
	starCount, err := s.StarCount()
	if err != nil {
		log.Error(nil, "failed to get star count: %v", err)
		return err
	}

	maxPage := (starCount + constants.DefaultPageSize - 1) / constants.DefaultPageSize
	if maxPage > constants.MaxPageLimit {
		maxPage = constants.MaxPageLimit
	}
	log.Info(nil, "calculating pages, maxPage: %d", maxPage)

	client := s.HTTPClient
	processedMap := make(map[string]*repository.StarUser, maxPage*constants.DefaultPageSize)

	var processedData []*repository.StarUser
	for page := maxPage; page >= 1; page-- {
		pageURL := fmt.Sprintf("%s?per_page=%d&page=%d", starURL, constants.DefaultPageSize, page)
		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			log.Error(nil, "failed to create request: %v", err)
			return err
		}
		req.Header.Set("Authorization", fmt.Sprintf("token %s", s.PersonalToken))
		req.Header.Set("Accept", "application/vnd.github.star+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := client.Do(req)
		if err != nil {
			log.Error(nil, "failed to get stargazers: %v", err)
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error(nil, "failed to get stargazers status: %s", resp.Status)
			return err
		}

		var rawStargazers []StargazerEntry
		if err := json.NewDecoder(resp.Body).Decode(&rawStargazers); err != nil {
			log.Error(nil, "Failed to decode stargazers: %v", err)
			return err
		}
		for _, entry := range rawStargazers {
			dbTime := entry.StarredAt
			starTimeUnixMillis := entry.StarredAt.UnixMilli()

			data := ProcessedStargazer{
				UserLogin:     entry.User.Login,
				UserID:        entry.User.ID,
				StarredAtDB:   dbTime,
				StarredAtUnix: starTimeUnixMillis,
			}
			//processedData = append(processedData, &repository.StarUser{
			//	ID:         int64(data.UserID),
			//	Name:       data.UserLogin,
			//	GitHubID:   strconv.Itoa(data.UserID),
			//	GitHubName: data.UserLogin,
			//	StarredAt:  data.StarredAtDB.Format(time.DateTime),
			//})
			processedMap[strconv.Itoa(data.UserID)] =
				&repository.StarUser{
					ID:         int64(data.UserID),
					Name:       data.UserLogin,
					GitHubID:   strconv.Itoa(data.UserID),
					GitHubName: data.UserLogin,
					StarredAt:  data.StarredAtDB.Format(time.DateTime),
				}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	users, err := repository.GetDB().GetAllUsersByConditions(ctx, map[string]any{
		"github_id": "__NOT_NULL__",
		//"github_star": "__NULL__",    #  Considering that canceling a star requires canceling this condition
	})
	if err != nil {
		log.Error(nil, "Failed to get users who have no star: %v", err)
		return err
	}

	for _, p := range processedData {
		processedMap[p.GitHubID] = p
	}

	for i, u := range users {
		if _, ok := processedMap[u.GithubID]; ok {
			users[i].GithubStar = fmt.Sprintf("%s.%s", s.Owner, s.Repo)
		} else {
			users[i].GithubStar = ""
		}
	}

	err = repository.GetDB().BatchUpsert(ctx, users, constants.DBIndexField)
	//errs = repository.GetDB().BatchUpsert(ctx, processedData, constants.DBIndexField)  // chose to use users instead of processedData
	if err != nil {
		log.Error(nil, "Failed to batch upsert stargazers: %v", err)
		return err
	}
	return nil
}

// withSyncLock  using this lock in k8s or multiple instances
func (s *SyncStar) withSyncLock(ctx context.Context, lock *repository.SyncLock, fn func() error) error {
	tmp, err := repository.GetDB().GetByField(ctx, &repository.SyncLock{}, "name", "github_sync_lock")

	if err != nil {
		return err
	}

	if tmp != nil {
		lock_, ok := tmp.(*repository.SyncLock)
		// if not using timestamptz will result in time comparison errs
		// a lock in an error state that needs to be deleted
		if ok && lock_.LockedAt.Add(2*s.Interval*time.Minute).Before(time.Now().Local()) {
			log.Error(ctx, "Expired lock detected: Lock name=%s, expired at=%v", lock_.Name, lock_.LockedAt)
			if err := repository.GetDB().RemoveSyncLock(ctx, lock); err != nil {
				log.Error(ctx, "Failed to remove expired sync lock: %v", err)
			} else {
				log.Info(ctx, "Already expired lock detected: Lock name=%s, expired at=%v", lock_.Name, lock_.LockedAt)
			}
		}
	}

	if err := repository.GetDB().AddSyncLock(ctx, lock); err != nil {
		log.Info(ctx, "Failed to add sync lock: %v", err)
		return err
	}

	defer func() {
		if err := repository.GetDB().RemoveSyncLock(ctx, lock); err != nil {
			log.Error(ctx, "Failed to remove sync lock: %v", err)
		}
	}()

	if err := fn(); err != nil {
		log.Error(ctx, "Error during GitHub star sync: %v", err)
		return err
	}

	return nil
}

// StarSyncTimer star sync timer
func (s *SyncStar) StarSyncTimer(ctx context.Context) {
	var lock = repository.SyncLock{
		Name:     "github_sync_lock",
		LockedAt: time.Now(),
	}

	if !s.Enabled {
		log.Info(ctx, "GitHub star sync is disabled")
		return
	}

	log.Info(ctx, "Starting initial GitHub star sync...")

	if err := s.withSyncLock(ctx, &lock, func() error {
		if err := s.Stargazers(); err != nil {
			return fmt.Errorf("failed to sync GitHub stars initially: %v", err)
		}
		return nil
	}); err != nil {
		log.Error(ctx, "Error occurred during initial sync: %v", err)
	}

	ticker := time.NewTicker(s.Interval * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info(ctx, "Stopping GitHub star sync: %v", ctx.Err())
			return
		case <-ticker.C:
			log.Info(ctx, "Starting periodic GitHub star sync...")
			if err := s.withSyncLock(ctx, &lock, func() error {
				if err := s.Stargazers(); err != nil {
					return fmt.Errorf("failed to sync GitHub stars: %v", err)
				}
				return nil
			}); err != nil {
				log.Error(ctx, "Error occurred during periodic sync: %v", err)
			}
		}
	}
}
