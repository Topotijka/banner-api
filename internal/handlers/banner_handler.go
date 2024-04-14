package handlers

import (
	"banner-api/internal/middleware"
	"banner-api/internal/model"
	"banner-api/internal/service"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type BannerHandler struct {
	BannerService *service.BannerService
}

func NewBannerHandler(bannerService *service.BannerService) *BannerHandler {
	return &BannerHandler{bannerService}
}

func (h *BannerHandler) RegisterHandlers(r chi.Router) {
	r.Use(middleware.AuthUserPermission)
	r.Get("/user_banner", h.getUserBanner)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthAdminPermission)

		r.Get("/banner", h.getBanners)
		r.Post("/banner", h.newBanner)
		r.Patch("/banner/{id}", h.updateBanner)
		r.Delete("/banner/{id}", h.deleteBanner)
	})

}

func (h *BannerHandler) getUserBanner(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	tagIdStr := params.Get("tag_id")
	tagId64, err := strconv.ParseInt(tagIdStr, 10, 16)
	tagId := int16(tagId64)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		return
	}

	featureIDStr := params.Get("feature_id")
	featureID64, err := strconv.ParseInt(featureIDStr, 10, 16)
	featureID := int16(featureID64)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		return
	}

	useLastRevisionStr := params.Get("use_last_revision")
	var useLastRevision bool
	if useLastRevisionStr != "" {
		useLastRevision, err = strconv.ParseBool(useLastRevisionStr)
		if err != nil {
			http.Error(w, "invalid query params", http.StatusBadRequest)
			return
		}
	} else {
		useLastRevision = false
	}

	role, ok := r.Context().Value("role").(int8)
	if !ok {
		http.Error(w, "missing role", http.StatusInternalServerError)
		return
	}

	response, err := h.BannerService.GetUserBanner(tagId, featureID, useLastRevision, role)

	if errors.Is(err, sql.ErrNoRows) || len(response) == 2 {
		http.Error(w, "there is no data with the given params", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "json marshal error", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, "json write error", http.StatusInternalServerError)
		return
	}
}

func (h *BannerHandler) getBanners(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	tagIdStr := params.Get("tag_id")
	featureIdStr := params.Get("feature_id")
	if tagIdStr == "" && featureIdStr == "" {
		http.Error(w, "at least one of feature_id or tag_id is required", http.StatusBadRequest)
	}

	tagId64, err := strconv.ParseInt(tagIdStr, 10, 16)
	if err != nil {
		http.Error(w, "parsing tag_id error", http.StatusBadRequest)
	}
	tagId := int16(tagId64)

	featureId64, err := strconv.ParseInt(featureIdStr, 10, 16)
	if err != nil {
		http.Error(w, "parsing feature_id error", http.StatusInternalServerError)
	}
	featureId := int16(featureId64)

	limitStr := params.Get("limit")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		http.Error(w, "parsing limit error", http.StatusInternalServerError)
	}

	offsetStr := params.Get("offset")
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		http.Error(w, "parsing offset error", http.StatusInternalServerError)
	}
	w.Header().Set("Content-type", "application/json")

	banners, err := h.BannerService.GetBanners(tagId, featureId, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	jsonData, err := json.Marshal(banners)
	if err != nil {
		http.Error(w, "failed to marshal json", http.StatusInternalServerError)
	}
	w.Write(jsonData)
}

func (h *BannerHandler) newBanner(w http.ResponseWriter, r *http.Request) {

	var nb model.BannerRequest
	err := json.NewDecoder(r.Body).Decode(&nb)
	if err != nil || !nb.PostValidate() {
		http.Error(w, "failed to unmarshal json", http.StatusBadRequest)
	}

	id, err := h.BannerService.NewBanner(nb)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jsonData := []byte(`{"banner_id": ` + fmt.Sprint(id) + `}`)
	w.Write(jsonData)
}

func (h *BannerHandler) updateBanner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 16)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
	}

	var br model.BannerRequest
	err = json.NewDecoder(r.Body).Decode(&br)
	if err != nil {
		http.Error(w, "failed to unmarshal json", http.StatusBadRequest)
	}
	params := br.MakeParamsMap()
	err = h.BannerService.UpdateBanner(id, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	w.WriteHeader(http.StatusOK)
}
func (h *BannerHandler) deleteBanner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
	}

	w.Header().Set("Content-type", "application/json")

	var rowsAffected int64
	rowsAffected, err = h.BannerService.DeleteBanner(id)
	if err != nil && rowsAffected == 0 {
		http.Error(w, "internal erros", http.StatusInternalServerError)
	}
	if rowsAffected == 0 {
		http.Error(w, "there is no data with the given params", http.StatusNotFound)
	}
	w.WriteHeader(http.StatusNoContent)
}
