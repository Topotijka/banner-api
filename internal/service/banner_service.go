package service

import (
	"banner-api/internal/cache"
	"banner-api/internal/middleware"
	"banner-api/internal/model"
	"banner-api/internal/repository/psql"
	"encoding/json"
)

type BannerService struct {
	BannerRepo  *psql.BannerRepo
	BannerCache *cache.BannerCache
}

func NewBannerService(bannerRepo *psql.BannerRepo, bannerCache *cache.BannerCache) *BannerService {
	return &BannerService{bannerRepo, bannerCache}
}

func (s *BannerService) GetUserBanner(tagId int16, featureId int16, useLastRevision bool, role int8) (json.RawMessage, error) {

	var (
		banner *model.Banner
		err    error
		found  bool
	)

	switch useLastRevision {
	case true:
		banner, err = s.BannerRepo.GetUserBanner(tagId, featureId)
		if err != nil && banner != nil {
			s.BannerCache.NewBanner(banner.Id, banner)
		}
	case false:
		banner, found = s.BannerCache.GetUserBanner(tagId, featureId)
		if !found {
			banner, err = s.BannerRepo.GetUserBanner(tagId, featureId)
			if err != nil && banner != nil {
				s.BannerCache.NewBanner(banner.Id, banner)
			}
		}
	}
	if banner != nil && err == nil {
		if !banner.IsActive {
			switch role {

			case middleware.AdminRole:
				return banner.Content, err

			case middleware.UserRole:
				return []byte(`{}`), err
			}
		}
	}

	return nil, err
}

func (s *BannerService) GetBanners(tagId int16, featureId int16, limit int64, offset int64) ([]model.Banner, error) {
	banners, err := s.BannerRepo.GetBanners(tagId, featureId, limit, offset)
	return banners, err
}

func (s *BannerService) NewBanner(br model.BannerRequest) (int64, error) {
	banner, err := s.BannerRepo.NewBanner(br)
	if err != nil {
		return 0, err
	}
	if banner != nil {
		s.BannerCache.NewBanner(banner.Id, banner)
	}

	return banner.Id, err
}

func (s *BannerService) UpdateBanner(id int64, params map[string]interface{}) error {
	err := s.BannerRepo.UpdateBanner(id, params)
	return err

}
func (s *BannerService) DeleteBanner(id int64) (int64, error) {
	rowsAffected, err := s.BannerRepo.DeleteBanner(id)
	return rowsAffected, err
}
