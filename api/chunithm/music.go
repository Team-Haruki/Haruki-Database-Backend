package chunithm

import (
	"context"
	"haruki-database/api"
	"haruki-database/config"
	"sort"
	"time"

	entchuniMusic "haruki-database/database/schema/chunithm/music"
	"haruki-database/database/schema/chunithm/music/chunithmchartdata"
	"haruki-database/database/schema/chunithm/music/chunithmmusic"
	"haruki-database/database/schema/chunithm/music/chunithmmusicdifficulty"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func (h *MusicHandler) GetAllMusic(c fiber.Ctx) error {
	ctx := context.Background()
	now := time.Now()
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSMusic)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, err := h.svc.client.ChunithmMusic.
		Query().
		Where(
			chunithmmusic.Or(
				chunithmmusic.ReleaseDateLTE(now),
				chunithmmusic.ReleaseDateIsNil(),
			),
		).
		All(ctx)
	if err != nil {
		return api.InternalError(c)
	}
	result := make([]MusicInfoSchema, len(rows))
	for i, row := range rows {
		deleted := row.IsDeleted
		result[i] = MusicInfoSchema{
			MusicID:        row.MusicID,
			Title:          row.Title,
			Artist:         row.Artist,
			Category:       row.Category,
			Version:        row.Version,
			ReleaseDate:    row.ReleaseDate,
			IsDeleted:      &deleted,
			DeletedVersion: row.DeletedVersion,
		}
	}
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", result)
}

func (h *MusicHandler) GetDifficultyInfo(c fiber.Ctx) error {
	ctx := context.Background()
	musicID := fiber.Params[int](c, "music_id", -1)
	if musicID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid music_id")
	}
	version := c.Query("version")
	if version == "" {
		return api.JSONResponse(c, fiber.StatusBadRequest, "version required")
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSMusic)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	record, _ := h.svc.client.ChunithmMusicDifficulty.
		Query().
		Where(chunithmmusicdifficulty.MusicIDEQ(musicID), chunithmmusicdifficulty.VersionEQ(version)).
		First(ctx)
	if record != nil {
		payload := MusicDifficultySchema{
			MusicID: record.MusicID,
			Version: record.Version,
			Diff0:   record.Diff0Const,
			Diff1:   record.Diff1Const,
			Diff2:   record.Diff2Const,
			Diff3:   record.Diff3Const,
			Diff4:   record.Diff4Const,
		}
		return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", payload)
	}
	latest, _ := h.svc.client.ChunithmMusicDifficulty.
		Query().
		Where(chunithmmusicdifficulty.MusicIDEQ(musicID)).
		Order(entchuniMusic.Desc(chunithmmusicdifficulty.FieldVersion)).
		First(ctx)
	if latest != nil {
		payload := MusicDifficultySchema{
			MusicID: latest.MusicID,
			Version: latest.Version,
			Diff0:   latest.Diff0Const,
			Diff1:   latest.Diff1Const,
			Diff2:   latest.Diff2Const,
			Diff3:   latest.Diff3Const,
			Diff4:   latest.Diff4Const,
		}
		return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", payload)
	}
	return api.JSONResponse(c, fiber.StatusNotFound, "No difficulty data")
}

func (h *MusicHandler) GetBasicInfo(c fiber.Ctx) error {
	ctx := context.Background()
	musicID := fiber.Params[int](c, "music_id", -1)
	if musicID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid music_id")
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSMusic)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	row, _ := h.svc.client.ChunithmMusic.
		Query().
		Where(chunithmmusic.MusicIDEQ(musicID)).
		First(ctx)
	if row == nil {
		return api.JSONResponse(c, fiber.StatusNotFound, "Music not found")
	}
	deleted := row.IsDeleted
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", MusicInfoSchema{
		MusicID:        row.MusicID,
		Title:          row.Title,
		Artist:         row.Artist,
		Category:       row.Category,
		Version:        row.Version,
		ReleaseDate:    row.ReleaseDate,
		IsDeleted:      &deleted,
		DeletedVersion: row.DeletedVersion,
	})
}

func (h *MusicHandler) GetChartData(c fiber.Ctx) error {
	ctx := context.Background()
	musicID := fiber.Params[int](c, "music_id", -1)
	if musicID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "invalid music_id")
	}
	key, cached, hit, err := api.CacheQuery(ctx, c, h.svc.redisClient, CacheNSMusic)
	if err != nil {
		return api.InternalError(c)
	}
	if hit {
		return c.Status(fiber.StatusOK).JSON(cached)
	}
	rows, _ := h.svc.client.ChunithmChartData.
		Query().
		Where(chunithmchartdata.MusicIDEQ(musicID)).
		All(ctx)
	if len(rows) == 0 {
		return api.JSONResponse(c, fiber.StatusNotFound, "No chart data found")
	}
	result := make([]ChartDataSchema, len(rows))
	for i, r := range rows {
		result[i] = ChartDataSchema{
			Difficulty: r.Difficulty,
			Creator:    r.Creator,
			BPM:        r.Bpm,
			TapCount:   r.TapCount,
			HoldCount:  r.HoldCount,
			SlideCount: r.SlideCount,
			AirCount:   r.AirCount,
			FlickCount: r.FlickCount,
			TotalCount: r.TotalCount,
		}
	}
	return api.CachedJSONResponse(ctx, c, h.svc.redisClient, config.Cfg.Backend.APICacheTTL, key, fiber.StatusOK, "ok", result)
}

func (h *MusicHandler) QueryBatch(c fiber.Ctx) error {
	ctx := context.Background()
	var req struct {
		MusicIDs []int  `json:"music_ids"`
		Version  string `json:"version"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return api.JSONResponse(c, fiber.StatusBadRequest, api.ErrInvalidRequest)
	}
	musicRows, _ := h.svc.client.ChunithmMusic.
		Query().
		Where(chunithmmusic.MusicIDIn(req.MusicIDs...)).
		All(ctx)
	musicMap := make(map[int]*entchuniMusic.ChunithmMusic)
	for _, m := range musicRows {
		musicMap[m.MusicID] = m
	}
	diffRows, _ := h.svc.client.ChunithmMusicDifficulty.
		Query().
		Where(chunithmmusicdifficulty.MusicIDIn(req.MusicIDs...)).
		All(ctx)
	sort.Slice(diffRows, func(i, j int) bool {
		return diffRows[i].Version > diffRows[j].Version
	})
	diffMap := make(map[int]*entchuniMusic.ChunithmMusicDifficulty)
	for _, d := range diffRows {
		if _, ok := diffMap[d.MusicID]; !ok || d.Version == req.Version {
			diffMap[d.MusicID] = d
		}
	}
	result := make(map[int]MusicBatchItemSchema)
	for _, mid := range req.MusicIDs {
		music := musicMap[mid]
		diff := diffMap[mid]
		var diffList []*float64
		if diff != nil {
			diffList = []*float64{
				diff.Diff0Const, diff.Diff1Const, diff.Diff2Const,
				diff.Diff3Const, diff.Diff4Const,
			}
		} else {
			diffList = []*float64{nil, nil, nil, nil, nil}
		}
		var info MusicInfoSchema
		var version *string
		if music != nil {
			deleted := music.IsDeleted
			info = MusicInfoSchema{
				MusicID:        music.MusicID,
				Title:          music.Title,
				Artist:         music.Artist,
				Category:       music.Category,
				Version:        music.Version,
				ReleaseDate:    music.ReleaseDate,
				IsDeleted:      &deleted,
				DeletedVersion: music.DeletedVersion,
			}
			version = music.Version
		} else {
			title := "Unknown"
			artist := "Unknown"
			category := "Unknown"
			isDeleted := false
			info = MusicInfoSchema{
				MusicID:   mid,
				Title:     title,
				Artist:    artist,
				Category:  &category,
				IsDeleted: &isDeleted,
			}
			version = nil
		}
		result[mid] = MusicBatchItemSchema{
			Version:    version,
			Difficulty: diffList,
			Info:       info,
		}
	}
	return api.JSONResponse(c, fiber.StatusOK, "success", result)
}

func registerMusicRoutes(r fiber.Router, client *entchuniMusic.Client, redisClient *redis.Client) {
	svc := NewMusicService(client, redisClient)
	h := NewMusicHandler(svc)
	apiGroup := r.Group("/music")

	apiGroup.Get("/all-music", h.GetAllMusic)
	apiGroup.Get("/:music_id/difficulty-info", h.GetDifficultyInfo)
	apiGroup.Get("/:music_id/basic-info", h.GetBasicInfo)
	apiGroup.Get("/:music_id/chart-data", h.GetChartData)
	apiGroup.Post("/query-batch", h.QueryBatch)
}
