package psql

import (
	"banner-api/internal/model"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"strconv"
	"strings"
	"time"
)

type BannerRepo struct {
	DB *sql.DB
}

func NewBannerRepo(db *sql.DB) *BannerRepo {
	return &BannerRepo{db}
}

func PGDBInit(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS banners (
        id SERIAL PRIMARY KEY,
        tag_ids INTEGER[],
        feature_id INTEGER,
        content JSONB,
        is_active BOOLEAN,
        created_at TIMESTAMP DEFAULT current_timestamp,
        updated_at TIMESTAMP DEFAULT current_timestamp)`,
	)
	if err != nil {
		return err
	}
	return nil
}

func NewPostgresDB(connStr string) *sql.DB {

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	for {
		err := db.Ping()
		if err == nil {
			break
		}
		fmt.Println("Waiting for the database to become ready...")
		time.Sleep(1 * time.Second)
	}
	return db
}

func (r *BannerRepo) GetUserBanner(tagId int16, featureId int16) (*model.Banner, error) {
	banner := model.Banner{}
	rows, err := r.DB.Query(
		"SELECT * FROM banners WHERE $1 = ANY(tag_ids) AND feature_id = $2",
		tagId, featureId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tagIds pq.Int32Array
		if err := rows.Scan(&banner.Id, &tagIds, &banner.FeatureId, &banner.Content, &banner.IsActive,
			&banner.CreatedAt, &banner.UpdatedAt); err != nil {
			return nil, err
		}
		if tagIds == nil {
			banner.TagIds = []int16{}
		} else {
			banner.TagIds = int16ArrayToSlice(tagIds)
		}
	}

	return &banner, nil
}

func int16ArrayToSlice(arr pq.Int32Array) []int16 {
	slice := make([]int16, len(arr))
	for i, v := range arr {
		slice[i] = int16(v)
	}
	return slice
}

func (r *BannerRepo) GetBanners(tagId int16, featureId int16, limit int64, offset int64) ([]model.Banner, error) {

	var banners []model.Banner

	query, args := getBannersQueryGen(tagId, featureId, limit, offset)

	rows, err := r.DB.Query(query, args...)
	defer rows.Close()

	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var banner model.Banner
		if err := rows.Scan(&banner.Id, &banner.TagIds, &banner.FeatureId, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt); err != nil {
			return nil, err
		}
		banners = append(banners, banner)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return banners, nil
}

func (r *BannerRepo) NewBanner(br model.BannerRequest) (*model.Banner, error) {

	banner := &model.Banner{}
	tagIdsArr := pq.Array(br.TagIds)
	err := r.DB.QueryRow(`INSERT INTO banners (tag_ids, feature_id, content, is_active)
								VALUES ($1, $2, $3, $4)
								RETURNING id, created_at, updated_at`, tagIdsArr, br.FeatureId, br.Content,
		br.IsActive).Scan(&banner.Id, &banner.CreatedAt, &banner.UpdatedAt)
	if err != nil {
		return nil, err
	}

	banner.TagIds = br.TagIds
	banner.FeatureId = br.FeatureId
	banner.Content = br.Content
	banner.IsActive = br.IsActive

	return banner, nil
}

func (r *BannerRepo) UpdateBanner(id int64, params map[string]interface{}) error {
	query, args := updateBannerQueryGen(id, params)
	_, err := r.DB.Exec(query, args...)
	return err
}

func (r *BannerRepo) DeleteBanner(id int64) (int64, error) {
	result, err := r.DB.Exec("DELETE FROM banners WHERE id = $1", id)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, err
}

func getBannersQueryGen(tagId int16, featureId int16, limit int64, offset int64) (string, []interface{}) {
	query := "SELECT * FROM banners WHERE 1=1"
	var args []interface{}
	if tagId > 0 {
		query = query + " AND tag_ids = $1"
		args = append(args, tagId)
	}
	if featureId > 0 {
		query = query + " AND feature_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, featureId)
	}
	if limit > 0 {
		query = query + " LIMIT $" + strconv.Itoa(len(args)+1)
		args = append(args, limit)
	}
	if offset > 0 {
		query = query + " OFFSET $" + strconv.Itoa(len(args)+1)
		args = append(args, offset)
	}
	return query, args
}

func updateBannerQueryGen(id int64, params map[string]interface{}) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	query.WriteString("UPDATE banners SET")
	i := 1
	for key, value := range params {
		query.WriteString(fmt.Sprintf(" %s = $%d,", key, i))
		args = append(args, value)
		i++
	}
	query.WriteString(" updated_at = CURRENT_TIMESTAMP WHERE id = $")
	query.WriteString(strconv.FormatInt(int64(len(params)+1), 10))
	args = append(args, id)

	return query.String(), args
}
