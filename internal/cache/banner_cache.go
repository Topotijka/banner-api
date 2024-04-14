package cache

import (
	"banner-api/internal/model"
	"encoding/json"
	"sync"
	"time"
)

type BannerCache struct {
	rw      sync.RWMutex
	banners map[int64]*cachedBanner
	ttl     time.Duration
}

type cachedBanner struct {
	banner     *model.Banner
	expiration time.Time
}

func NewBannerCache() *BannerCache {
	return &BannerCache{
		rw:      sync.RWMutex{},
		banners: make(map[int64]*cachedBanner),
		ttl:     (5 * time.Minute),
	}
}

func (bc *BannerCache) SetBanner(id int64, params map[string]interface{}) {
	bc.rw.Lock()
	defer bc.rw.Unlock()
	banner := bc.banners[id]
	if banner != nil {
		for key, value := range params {
			switch key {
			case "TagIds":
				banner.banner.TagIds = value.([]int16)
			case "FeatureId":
				banner.banner.FeatureId = value.(int16)
			case "Content":
				banner.banner.Content = value.(json.RawMessage)
			case "IsActive":
				banner.banner.IsActive = value.(bool)
			}
		}
		expiration := time.Now().Add(bc.ttl)
		bc.banners[id].expiration = expiration
	}
}

func (bc *BannerCache) NewBanner(id int64, banner *model.Banner) {
	bc.rw.Lock()
	defer bc.rw.Unlock()
	bc.banners[id] = &cachedBanner{banner: banner,
		expiration: time.Now().Add(bc.ttl)}
}

func (bc *BannerCache) GetUserBanner(tagId int16, featureId int16) (*model.Banner, bool) {
	bc.rw.RLock()
	defer bc.rw.RUnlock()
	for _, b := range bc.banners {

		if contains(b.banner.TagIds, tagId) && b.banner.FeatureId == featureId {

			b.expiration = time.Now().Add(bc.ttl)
			return b.banner, true
		}
	}
	return nil, false
}

func (bc *BannerCache) StartCleanup() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bc.cleanup()
		}
	}
}

func (bc *BannerCache) cleanup() {
	bc.rw.Lock()
	defer bc.rw.Unlock()

	now := time.Now()
	for id, banner := range bc.banners {
		if now.After(banner.expiration) {
			delete(bc.banners, id)
		}
	}
}

func contains(slice []int16, item int16) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}
