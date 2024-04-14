package model

import (
	"encoding/json"
	"github.com/lib/pq"
)

type Banner struct {
	Id        int64           `json:"banner_id"`
	TagIds    []int16         `json:"tag_ids"`
	FeatureId int16           `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

type BannerRequest struct {
	TagIds    []int16         `json:"tag_ids"`
	FeatureId int16           `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
}

func (br *BannerRequest) PostValidate() bool {
	if len(br.TagIds) == 0 {
		return false
	}
	if br.FeatureId < 0 {
		return false
	}
	if len(br.Content) == 0 {
		return false
	}
	return true
}

func (br *BannerRequest) MakeParamsMap() map[string]interface{} {
	params := make(map[string]interface{})

	if len(br.TagIds) != 0 {
		params["tag_ids"] = pq.Array(br.TagIds)
	}
	if br.FeatureId != 0 {
		params["feature_id"] = br.FeatureId
	}
	if len(br.Content) != 0 {
		params["content"] = br.Content
	}

	params["is_active"] = br.IsActive

	return params
}
