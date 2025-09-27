package chunithm

import (
	"context"
	"net/http"
	"sort"
	"time"

	"haruki-database/api"
	entchuniMusic "haruki-database/database/schema/chunithm/music"
	"haruki-database/database/schema/chunithm/music/chunithmchartdata"
	"haruki-database/database/schema/chunithm/music/chunithmmusic"
	"haruki-database/database/schema/chunithm/music/chunithmmusicdifficulty"

	"github.com/gofiber/fiber/v2"
)

func RegisterMusicRoutes(r fiber.Router, client *entchuniMusic.Client) {
	apiGroup := r.Group("/music")

	apiGroup.Get("/all-music", func(c *fiber.Ctx) error {
		ctx := context.Background()
		now := time.Now()
		rows, err := client.ChunithmMusic.
			Query().
			Where(chunithmmusic.ReleaseDateLTE(now)).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		var result []MusicInfoSchema
		for _, row := range rows {
			deleted := row.IsDeleted == 1
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
		return api.JSONResponse(c, http.StatusOK, "success", result)
	})

	apiGroup.Get("/:music_id/difficulty-info", func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}
		version := c.Query("version")
		if version == "" {
			return api.JSONResponse(c, http.StatusBadRequest, "version required")
		}

		record, _ := client.ChunithmMusicDifficulty.
			Query().
			Where(chunithmmusicdifficulty.MusicIDEQ(musicID), chunithmmusicdifficulty.VersionEQ(version)).
			First(ctx)
		if record != nil {
			return api.JSONResponse(c, http.StatusOK, "success", MusicDifficultySchema{
				MusicID: record.MusicID,
				Version: record.Version,
				Diff0:   record.Diff0Const,
				Diff1:   record.Diff1Const,
				Diff2:   record.Diff2Const,
				Diff3:   record.Diff3Const,
				Diff4:   record.Diff4Const,
			})
		}

		latest, _ := client.ChunithmMusicDifficulty.
			Query().
			Where(chunithmmusicdifficulty.MusicIDEQ(musicID)).
			Order(entchuniMusic.Desc(chunithmmusicdifficulty.FieldVersion)).
			First(ctx)
		if latest != nil {
			return api.JSONResponse(c, http.StatusOK, "success", MusicDifficultySchema{
				MusicID: latest.MusicID,
				Version: latest.Version,
				Diff0:   latest.Diff0Const,
				Diff1:   latest.Diff1Const,
				Diff2:   latest.Diff2Const,
				Diff3:   latest.Diff3Const,
				Diff4:   latest.Diff4Const,
			})
		}
		return api.JSONResponse(c, http.StatusNotFound, "No difficulty data")
	})

	apiGroup.Get("/:music_id/basic-info", func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		row, _ := client.ChunithmMusic.
			Query().
			Where(chunithmmusic.MusicIDEQ(musicID)).
			First(ctx)
		if row == nil {
			return api.JSONResponse(c, http.StatusNotFound, "Music not found")
		}
		deleted := row.IsDeleted == 1
		return api.JSONResponse(c, http.StatusOK, "success", MusicInfoSchema{
			MusicID:        row.MusicID,
			Title:          row.Title,
			Artist:         row.Artist,
			Category:       row.Category,
			Version:        row.Version,
			ReleaseDate:    row.ReleaseDate,
			IsDeleted:      &deleted,
			DeletedVersion: row.DeletedVersion,
		})
	})

	apiGroup.Get("/:music_id/chart-data", func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
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
		return api.JSONResponse(c, http.StatusOK, "success", result)
	})

	apiGroup.Post("/query-batch", func(c *fiber.Ctx) error {
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
				deleted := music.IsDeleted == 1
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
	})
}
