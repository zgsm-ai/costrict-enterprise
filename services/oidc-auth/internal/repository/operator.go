package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/zgsm-ai/oidc-auth/pkg/log"
)

var allowedFields = map[string]map[string]bool{
	"AuthUser": {
		"id":          true,
		"name":        true,
		"github_id":   true,
		"github_name": true,
		"github_star": true,
		"email":       true,
		"provider":    true,
		"phone":       true,
		"invite_code": true,
	},
	"StarUser": {
		"id":          true,
		"name":        true,
		"github_id":   true,
		"github_name": true,
		"github_star": true,
	},
	"SyncLock": {
		"name": true,
	},
	"SmsVerificationCode": {
		"id":      true,
		"phone":   true,
		"code":    true,
		"purpose": true,
		"is_used": true,
	},
}

const batchUpsertSize = 500

func ValidateFieldName(modelName, fieldName string) error {
	for _, char := range fieldName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return fmt.Errorf("invalid field name: %s", fieldName)
		}
	}
	if fields, exists := allowedFields[modelName]; exists {
		if !fields[strings.ToLower(fieldName)] {
			return fmt.Errorf("field %s not allowed for model %s", fieldName, modelName)
		}
	}

	return nil
}

func (d *Database) withTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback().Error; rbErr != nil {
				log.Info(nil, "failed to rollback transaction after panic: %v\n", rbErr)
			}
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *Database) GetByField(ctx context.Context, model any, uniqueField string, value any) (any, error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	modelName := modelType.Name()

	if err := ValidateFieldName(modelName, uniqueField); err != nil {
		return nil, fmt.Errorf("field validation failed: %w", err)
	}

	if err := d.db.WithContext(ctx).
		Where(fmt.Sprintf("%s = ?", uniqueField), value).
		First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	return model, nil
}

// GetUserByField gets a user by a unique field
func (d *Database) GetUserByField(ctx context.Context, uniqueField string, value any) (*AuthUser, error) {
	if err := ValidateFieldName("AuthUser", uniqueField); err != nil {
		return nil, fmt.Errorf("field validation failed: %w", err)
	}

	var user AuthUser
	if err := d.db.WithContext(ctx).
		Where(fmt.Sprintf("%s = ?", uniqueField), value).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}

func (d *Database) DeleteUserByField(ctx context.Context, field string, value any) (int64, error) {
	if err := ValidateFieldName("AuthUser", field); err != nil {
		return 0, fmt.Errorf("field validation failed: %w", err)
	}

	var rowsAffected int64

	err := d.withTransaction(ctx, func(tx *gorm.DB) error {
		result := tx.Where(fmt.Sprintf("%s = ?", field), value).
			Delete(&AuthUser{})
		rowsAffected = result.RowsAffected
		if result.Error != nil {
			return fmt.Errorf("failed to delete user(s): %w", result.Error)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (d *Database) GetUserByDeviceConditions(ctx context.Context, conditions map[string]any) (*AuthUser, error) {
	if len(conditions) == 0 {
		return nil, errors.New("at least one condition is required")
	}

	var user AuthUser

	conditionBytes, err := json.Marshal(conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	// Cannot sort by time in device, unexpected results may occur due to null
	// When devices is a JSON array, devices->>'updated_at' cannot directly extract the updated_at value of the specific device you want from the array for sorting.
	// It usually returns NULL, resulting in invalid sorting or unpredictable results.
	err = d.db.WithContext(ctx).
		Where("devices @> ?", fmt.Sprintf("[%s]", string(conditionBytes))).
		Order("updated_at DESC").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // if not found, return nil
		}
		return nil, fmt.Errorf("failed to query device: %w", err)
	}
	return &user, nil
}

func (d *Database) GetAllUsersByConditions(ctx context.Context, conditions map[string]any) ([]AuthUser, error) {
	if len(conditions) == 0 {
		return nil, errors.New("at least one condition is required")
	}
	var users []AuthUser
	query := d.db.WithContext(ctx)
	for key, value := range conditions {
		if err := ValidateFieldName("AuthUser", key); err != nil {
			return nil, fmt.Errorf("field validation failed: %w", err)
		}
		if strValue, ok := value.(string); ok {
			if strValue == "__NULL__" {
				query = query.Where(fmt.Sprintf("%s IS NULL OR %s = ''", key, key))
			} else if strValue == "__NOT_NULL__" {
				query = query.Where(fmt.Sprintf("%s IS NOT NULL AND %s != ''", key, key))
			} else {
				query = query.Where(fmt.Sprintf("%s = ?", key), value)
			}
		} else {
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	query = query.Order("updated_at DESC NULLS LAST")

	err := query.Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query users by conditions: %w", err)
	}

	return users, nil
}

func (d *Database) Upsert(ctx context.Context, model any, uniqueField string, value any) error {
	if model == nil {
		return errors.New("model cannot be nil")
	}

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	modelName := modelType.Name()

	if err := ValidateFieldName(modelName, uniqueField); err != nil {
		return fmt.Errorf("field validation failed: %w", err)
	}

	return d.withTransaction(ctx, func(tx *gorm.DB) error {
		instanceToRead := reflect.New(modelType).Interface()
		result := tx.Where(fmt.Sprintf("%s = ?", uniqueField), value).First(instanceToRead)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := tx.Create(model).Error; err != nil {
					return fmt.Errorf("failed to create: %w", err)
				}
			} else {
				return fmt.Errorf("query error: %w", result.Error)
			}
		} else {
			// Update existing record
			if err := tx.Model(instanceToRead).
				Where(fmt.Sprintf("%s = ?", uniqueField), value).
				Updates(model).Error; err != nil {
				return fmt.Errorf("failed to update: %w", err)
			}
		}
		return nil
	})
}

// ClaimInviteCode atomically sets the invite code on the user's row only when
// it is not already set (NULL or empty), so concurrent first-time fetches
// converge on a single persisted code: the first writer wins and later writers
// no-op. Upsert cannot do this — its update branch would silently overwrite a
// just-committed code without erroring. Callers must re-read the row afterwards
// to serve the authoritative code (theirs or the winner's).
func (d *Database) ClaimInviteCode(ctx context.Context, userID uuid.UUID, code string) error {
	return d.withTransaction(ctx, func(tx *gorm.DB) error {
		result := tx.Model(&AuthUser{}).
			Where("id = ? AND (invite_code IS NULL OR invite_code = '')", userID).
			Updates(map[string]any{
				"invite_code": code,
				"updated_at":  time.Now(),
			})
		if result.Error != nil {
			return fmt.Errorf("failed to claim invite code: %w", result.Error)
		}
		return nil
	})
}

func (d *Database) BatchUpsert(ctx context.Context, models any, uniqueField string) error {
	val := reflect.ValueOf(models)
	if val.Kind() != reflect.Slice {
		return errors.New("models must be a slice")
	}
	if val.Len() == 0 {
		return nil
	}

	first := val.Index(0).Interface()
	modelType := reflect.TypeOf(first)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	modelName := modelType.Name()

	if err := ValidateFieldName(modelName, uniqueField); err != nil {
		return fmt.Errorf("field validation failed: %w", err)
	}

	return d.withTransaction(ctx, func(tx *gorm.DB) error {
		// Extract columns from first model
		stmt := &gorm.Statement{DB: tx}
		if err := stmt.Parse(first); err != nil {
			return fmt.Errorf("failed to parse model: %w", err)
		}

		columns := make([]string, 0)
		for _, field := range stmt.Schema.Fields {
			if field.DBName != uniqueField && !field.PrimaryKey && field.DBName != "created_at" {
				columns = append(columns, field.DBName)
			}
		}

		conflictClause := clause.OnConflict{
			Columns:   []clause.Column{{Name: uniqueField}},
			DoUpdates: clause.AssignmentColumns(columns),
		}

		for start := 0; start < val.Len(); start += batchUpsertSize {
			end := start + batchUpsertSize
			if end > val.Len() {
				end = val.Len()
			}
			batch := val.Slice(start, end).Interface()
			if err := tx.Clauses(conflictClause).Create(batch).Error; err != nil {
				return fmt.Errorf("upsert batch starting at index %d: %w", start, err)
			}
		}
		return nil
	})
}

// GithubStarUpdate contains only the identity used for matching and its star marker.
type GithubStarUpdate struct {
	GithubID   string
	GithubStar string
}

// BatchUpdateGithubStars updates existing users without writing a stale AuthUser
// snapshot over concurrent login, device, or invite-code changes. Matching the
// current github_id also avoids applying a result to a user whose binding changed.
func (d *Database) BatchUpdateGithubStars(ctx context.Context, updates []GithubStarUpdate) error {
	groups := make(map[string][]string)
	for _, update := range updates {
		if update.GithubID == "" {
			continue
		}
		groups[update.GithubStar] = append(groups[update.GithubStar], update.GithubID)
	}
	if len(groups) == 0 {
		return nil
	}

	return d.withTransaction(ctx, func(tx *gorm.DB) error {
		for marker, ids := range groups {
			for start := 0; start < len(ids); start += batchUpsertSize {
				end := start + batchUpsertSize
				if end > len(ids) {
					end = len(ids)
				}
				if err := tx.WithContext(ctx).Model(&AuthUser{}).
					Where("github_id IN ?", ids[start:end]).
					UpdateColumn("github_star", marker).Error; err != nil {
					return fmt.Errorf("failed to update GitHub star batch starting at index %d: %w", start, err)
				}
			}
		}
		return nil
	})
}

// UpdateGithubStar updates only the synchronized repository marker for a GitHub user.
func (d *Database) UpdateGithubStar(ctx context.Context, githubID, repository string, starred bool) (int64, error) {
	query := d.db.WithContext(ctx).Model(&AuthUser{}).Where("github_id = ?", githubID)
	value := repository
	if !starred {
		query = query.Where("github_star = ?", repository)
		value = ""
	}

	result := query.UpdateColumn("github_star", value)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to update GitHub star status: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (d *Database) AddSyncLock(ctx context.Context, models any) error {
	return d.withTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(models).Error; err != nil {
			return fmt.Errorf("create sync lock error: %w", err)
		}
		return nil
	})
}

func (d *Database) RemoveSyncLock(ctx context.Context, models any) error {
	return d.withTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Delete(models).Error; err != nil {
			return fmt.Errorf("delete sync lock error: %w", err)
		}
		return nil
	})
}
