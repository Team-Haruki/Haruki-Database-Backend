package chunithm

import (
	"context"
	"haruki-database/api"
	"net/http"

	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	"haruki-database/database/schema/chunithm/maindb/chunithmmusicalias"

	"github.com/gofiber/fiber/v2"
)

func RegisterAliasRoutes(router fiber.Router, client *entchuniMain.Client) {
	r := router.Group("/alias")

	r.Get("/music-id", func(c *fiber.Ctx) error {
		ctx := context.Background()
		aliasStr := c.Query("alias")
		if aliasStr == "" {
			return api.JSONResponse(c, http.StatusBadRequest, "alias is required")
		}

		rows, err := client.ChunithmMusicAlias.
			Query().
			Where(chunithmmusicalias.AliasEQ(aliasStr)).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Alias not found")
		}

		ids := make([]int, len(rows))
		for i, r := range rows {
			ids[i] = r.MusicID
		}

		return api.JSONResponse(c, http.StatusOK, "success", AliasToMusicIDResponse{
			Status:  200,
			Message: "success",
			Data:    ids,
		})
	})

	r.Get("/:music_id", func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		rows, err := client.ChunithmMusicAlias.
			Query().
			Where(chunithmmusicalias.MusicIDEQ(musicID)).
			All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		aliases := make([]string, len(rows))
		for i, r := range rows {
			aliases[i] = r.Alias
		}

		return api.JSONResponse(c, http.StatusOK, "success", AllAliasesResponse{
			Status:  200,
			Message: "success",
			Data:    aliases,
		})
	})

	r.Post("/:music_id/add", api.VerifyAPIAuthorization(), func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		var body MusicAliasSchema
		if err := c.BodyParser(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid request body")
		}

		existing, _ := client.ChunithmMusicAlias.
			Query().
			Where(chunithmmusicalias.MusicIDEQ(musicID), chunithmmusicalias.AliasEQ(body.Alias)).
			First(ctx)
		if existing != nil {
			return api.JSONResponse(c, http.StatusConflict, "Alias already exists")
		}

		newAlias, err := client.ChunithmMusicAlias.
			Create().
			SetMusicID(musicID).
			SetAlias(body.Alias).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		return api.JSONResponse(c, http.StatusOK, "Alias added", AddAliasResponse{
			Status:  200,
			Message: "Alias added",
			Data:    &MusicAliasSchema{ID: newAlias.ID, Alias: newAlias.Alias},
		})
	})

	r.Delete("/:music_id", api.VerifyAPIAuthorization(), func(c *fiber.Ctx) error {
		ctx := context.Background()
		musicID, err := c.ParamsInt("music_id")
		if err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid music_id")
		}

		var body MusicAliasSchema
		if err := c.BodyParser(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "invalid request body")
		}

		deleted, err := client.ChunithmMusicAlias.
			Delete().
			Where(chunithmmusicalias.MusicIDEQ(musicID), chunithmmusicalias.AliasEQ(body.Alias)).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if deleted == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "Alias not found")
		}

		return api.JSONResponse(c, http.StatusOK, "Alias deleted")
	})
}
