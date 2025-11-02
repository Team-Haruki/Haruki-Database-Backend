package pjsk

import (
	"context"
	"fmt"
	"haruki-database/api"
	"haruki-database/config"
	"haruki-database/database/schema/pjsk"
	"haruki-database/database/schema/pjsk/userbinding"
	"haruki-database/database/schema/pjsk/userdefaultbinding"
	"haruki-database/utils"
	harukiRedis "haruki-database/utils/redis"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func getBindings(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		server := c.Query("server")
		if server != "" {
			if _, err := utils.ParseBindingServer(server); err != nil {
				return api.JSONResponse(c, http.StatusBadRequest, err.Error())
			}
		}

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-user-binding")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		q := client.UserBinding.Query().
			Where(userbinding.PlatformEQ(platform), userbinding.ImIDEQ(imID))
		if server != "" {
			q = q.Where(userbinding.ServerEQ(server))
		}

		rows, err := q.All(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return api.JSONResponse(c, http.StatusNotFound, "No bindings found")
		}

		out := make([]BindingSchema, len(rows))
		for i, r := range rows {
			out[i] = BindingSchema{ID: r.ID, Platform: r.Platform, ImID: r.ImID, Server: r.Server, UserID: r.UserID, Visible: r.Visible}
		}
		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", BindingResponse{Bindings: out})
	}
}

func createBinding(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")

		var body struct {
			Server  string `json:"server"`
			UserID  string `json:"user_id"`
			Visible bool   `json:"visible"`
		}
		if err := c.BodyParser(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request")
		}
		if serverEnum, _ := utils.ParseDefaultBindingServer(body.Server); serverEnum == utils.DefaultBindingServerDefault {
			return api.JSONResponse(c, http.StatusBadRequest, "Unacceptable server param")
		}
		if _, err := utils.ParseBindingServer(body.Server); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}

		exists, _ := client.UserBinding.
			Query().
			Where(userbinding.PlatformEQ(platform), userbinding.ImIDEQ(imID), userbinding.ServerEQ(body.Server), userbinding.UserIDEQ(body.UserID)).
			First(ctx)
		if exists != nil {
			return api.JSONResponse(c, http.StatusConflict, "Binding already exists")
		}

		newBind, err := client.UserBinding.
			Create().
			SetPlatform(platform).
			SetImID(imID).
			SetServer(body.Server).
			SetUserID(body.UserID).
			SetVisible(body.Visible).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		_ = harukiRedis.ClearAllCacheForPath(ctx, redisClient, "pjsk-user-binding", fmt.Sprintf("/pjsk/%s/user/%s/binding", platform, imID))
		return api.JSONResponse(c, http.StatusCreated, "ok", AddBindingSuccessResponse{BindingID: newBind.ID})
	}
}

func getDefaultBinding(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		server := c.Query("server", "default")

		if _, err := utils.ParseDefaultBindingServer(server); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}

		key, cached, hit, err := api.CacheQuery(ctx, c, redisClient, "pjsk-user-binding")
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}
		if hit {
			return c.Status(http.StatusOK).JSON(cached)
		}

		row, err := client.UserDefaultBinding.
			Query().
			Where(userdefaultbinding.PlatformEQ(platform), userdefaultbinding.ImIDEQ(imID), userdefaultbinding.ServerEQ(server)).
			WithBinding().
			First(ctx)
		if err != nil || row.Edges.Binding == nil {
			msg := "No global default set"
			if server != "default" {
				msg = "No default for server '" + server + "'"
			}
			return api.JSONResponse(c, http.StatusNotFound, msg)
		}
		b := row.Edges.Binding
		return api.CachedJSONResponse(ctx, c, redisClient, config.Cfg.Backend.APICacheTTL, key, http.StatusOK, "ok", BindingResponse{
			Binding: &BindingSchema{ID: b.ID, Platform: b.Platform, ImID: b.ImID, Server: b.Server, UserID: b.UserID, Visible: b.Visible},
		})
	}
}

func setDefaultBinding(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")

		var body struct {
			Server    string `json:"server"`
			BindingID int    `json:"binding_id"`
		}
		if err := c.BodyParser(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request")
		}
		if _, err := utils.ParseDefaultBindingServer(body.Server); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}

		binding, err := client.UserBinding.
			Query().
			Where(userbinding.PlatformEQ(platform), userbinding.ImIDEQ(imID), userbinding.IDEQ(body.BindingID)).
			First(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusNotFound, "Binding not found")
		}
		if dfs, _ := utils.ParseDefaultBindingServer(body.Server); dfs != utils.DefaultBindingServerDefault && binding.Server != body.Server {
			return api.JSONResponse(c, http.StatusBadRequest, "Illegal request")
		}

		_, _ = client.UserDefaultBinding.
			Delete().
			Where(userdefaultbinding.PlatformEQ(platform), userdefaultbinding.ImIDEQ(imID), userdefaultbinding.ServerEQ(body.Server)).
			Exec(ctx)

		_, err = client.UserDefaultBinding.
			Create().
			SetPlatform(platform).
			SetImID(imID).
			SetServer(body.Server).
			SetBindingID(body.BindingID).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		_ = harukiRedis.ClearAllCacheForPath(ctx, redisClient, "pjsk-user-binding", fmt.Sprintf("/pjsk/%s/user/%s/binding/default", platform, imID))

		return api.JSONResponse(c, http.StatusOK, "Set default binding for "+body.Server)
	}
}

func deleteDefaultBinding(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")

		var body struct{ Server string }
		if err := c.BodyParser(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request")
		}
		if _, err := utils.ParseDefaultBindingServer(body.Server); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, err.Error())
		}

		_, err := client.UserDefaultBinding.
			Delete().
			Where(userdefaultbinding.PlatformEQ(platform), userdefaultbinding.ImIDEQ(imID), userdefaultbinding.ServerEQ(body.Server)).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		_ = harukiRedis.ClearAllCacheForPath(ctx, redisClient, "pjsk-user-binding", fmt.Sprintf("/pjsk/%s/user/%s/binding/default", platform, imID))

		return api.JSONResponse(c, http.StatusOK, "Deleted default binding for "+body.Server)
	}
}

func updateBindingVisibility(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		bindingID, _ := strconv.Atoi(c.Params("binding_id"))

		var body struct{ Visible bool }
		if err := c.BodyParser(&body); err != nil {
			return api.JSONResponse(c, http.StatusBadRequest, "Invalid request")
		}

		bind, err := client.UserBinding.
			Query().
			Where(userbinding.PlatformEQ(platform), userbinding.ImIDEQ(imID), userbinding.IDEQ(bindingID)).
			First(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusNotFound, "Binding not found")
		}

		_, err = client.UserBinding.
			UpdateOneID(bind.ID).
			SetVisible(body.Visible).
			Save(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		_ = harukiRedis.ClearAllCacheForPath(ctx, redisClient, "pjsk-user-binding", fmt.Sprintf("/pjsk/%s/user/%s/binding", platform, imID))
		_ = harukiRedis.ClearAllCacheForPath(ctx, redisClient, "pjsk-user-binding", fmt.Sprintf("/pjsk/%s/user/%s/binding/default", platform, imID))

		return api.JSONResponse(c, http.StatusOK, "Visibility updated")
	}
}

func deleteBinding(client *pjsk.Client, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		platform := c.Params("platform")
		imID := c.Params("im_id")
		bindingID, _ := strconv.Atoi(c.Params("binding_id"))

		_, _ = client.UserDefaultBinding.
			Delete().
			Where(userdefaultbinding.PlatformEQ(platform), userdefaultbinding.ImIDEQ(imID), userdefaultbinding.BindingIDEQ(bindingID)).
			Exec(ctx)

		_, err := client.UserBinding.
			Delete().
			Where(userbinding.PlatformEQ(platform), userbinding.ImIDEQ(imID), userbinding.IDEQ(bindingID)).
			Exec(ctx)
		if err != nil {
			return api.JSONResponse(c, http.StatusInternalServerError, err.Error())
		}

		_ = harukiRedis.ClearAllCacheForPath(ctx, redisClient, "pjsk-user-binding", fmt.Sprintf("/pjsk/%s/user/%s/binding", platform, imID))
		_ = harukiRedis.ClearAllCacheForPath(ctx, redisClient, "pjsk-user-binding", fmt.Sprintf("/pjsk/%s/user/%s/binding/default", platform, imID))

		return api.JSONResponse(c, http.StatusOK, "Binding deleted")
	}
}

func registerBindingRoutes(router fiber.Router, client *pjsk.Client, redisClient *redis.Client) {
	r := router.Group("/:platform/user", api.VerifyAPIAuthorization())

	r.Get("/:im_id/binding", getBindings(client, redisClient))
	r.Post("/:im_id/binding", createBinding(client, redisClient))
	r.Get("/:im_id/binding/default", getDefaultBinding(client, redisClient))
	r.Put("/:im_id/binding/default", setDefaultBinding(client, redisClient))
	r.Delete("/:im_id/binding/default", deleteDefaultBinding(client, redisClient))
	r.Patch("/:im_id/binding/:binding_id", updateBindingVisibility(client, redisClient))
	r.Delete("/:im_id/binding/:binding_id", deleteBinding(client, redisClient))
}
