package chunithm

import (
	"context"
	"haruki-database/config"
	"net/http"
	"sort"
	"time"

	"haruki-database/api"
	entchuniMusic "haruki-database/database/schema/chunithm/music"
	"haruki-database/database/schema/chunithm/music/chunithmchartdata"
	"haruki-database/database/schema/chunithm/music/chunithmmusic"
	"haruki-database/database/schema/chunithm/music/chunithmmusicdifficulty"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func getAllMusic(client *entchuniMusic.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		now := time.Now()

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "chunithm-music")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, err := client.ChunithmMusic.
			Query().
			Where(
				chunithmmusic.Or(
					chunithmmusic.ReleaseDateLTE(now),
					chunithmmusic.ReleaseDateIsNil(),
				),
			).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		var result []MusicInfoSchema
		for _, row := range rows {
			deleted := row.IsDeleted
			result = append(result, MusicInfoSchema{
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
		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", result)
	}
}

func getDifficultyInfo(client *entchuniMusic.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}
		version := c.Query("version")
		if version == "" {
			return api.JSONResponse(c, http.StatusBadRequest, "version required")
		}

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "chunithm-music")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		record, _ := client.ChunithmMusicDifficulty.
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
			return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", payload)
		}

		latest, _ := client.ChunithmMusicDifficulty.
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
			return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", payload)
		}

		return api.JSONResponse(c, http.StatusNotFound, "No difficulty data")
	}
}

func getBasicInfo(client *entchuniMusic.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "chunithm-music")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		row, _ := client.ChunithmMusic.
			Query().
			Where(chunithmmusic.MusicIDEQ(musicID)).
			First(ctx)
		if row == nil {
			return api.JSONResponse(c, http.StatusNotFound, "Music not found")
		}
		deleted := row.IsDeleted
		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", MusicInfoSchema{
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
}

func getChartData(client *entchuniMusic.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "chunithm-music")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		rows, _ := client.ChunithmChartData.
			Query().
			Where(chunithmchartdata.MusicIDEQ(musicID)).
			All(ctx)
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "No chart data found")
		}
		var result []ChartDataSchema
		for _, r := range rows {
			result = append(result, ChartDataSchema{
				Difficulty: r.Difficulty,
				Creator:    r.Creator,
				BPM:        r.Bpm,
				TapCount:   r.TapCount,
				HoldCount:  r.HoldCount,
				SlideCount: r.SlideCount,
				AirCount:   r.AirCount,
				FlickCount: r.FlickCount,
				TotalCount: r.TotalCount,
			})
		}
		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", result)
	}
}

func queryBatch(client *entchuniMusic.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		var req struct {
			MusicIDs []int  `json:"music_ids"`
			Version  string `json:"version"`
		}
		if err := c.BodyParser(&req); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid request body")
		}

		musicRows, _ := client.ChunithmMusic.
			Query().
			Where(chunithmmusic.MusicIDIn(req.MusicIDs...)).
			All(ctx)
		musicMap := make(map[int]*entchuniMusic.ChunithmMusic)
		for _, m := range musicRows {
			musicMap[m.MusicID] = m
		}

		diffRows, _ := client.ChunithmMusicDifficulty.
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

		return api.JSONResponse(c, http.StatusOK, "success", result)
	}
}

func registerMusicRoutes(r fiber.Router, client *entchuniMusic.Client, redisClient *redis.Client) {
	apiGroup := r.Group("/music")

	apiGroup.Get("/all-music", getAllMusic(client, redisClient))
	apiGroup.Get("/:music_id/difficulty-info", getDifficultyInfo(client, redisClient))
	apiGroup.Get("/:music_id/basic-info", getBasicInfo(client, redisClient))
	apiGroup.Get("/:music_id/chart-data", getChartData(client, redisClient))
	apiGroup.Post("/query-batch", queryBatch(client))
}
