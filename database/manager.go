package database

import (
	"context"
	"log"

	entchuniMain "haruki-database/database/schema/chunithm/maindb"
	entchuniMusic "haruki-database/database/schema/chunithm/music"
	entpjsk "haruki-database/database/schema/pjsk"
)

type DBManager struct {
	PJSKDBClient          *entpjsk.Client
	ChunithmMainDBClient  *entchuniMain.Client
	ChunithmMusicDBClient *entchuniMusic.Client
}

func (m *DBManager) InitClient(name, dialect, dsn string) error {
	switch name {
	case "pjsk":
		client, err := entpjsk.Open(dialect, dsn)
		if err != nil {
			return err
		}
		if err := client.Schema.Create(context.Background()); err != nil {
			return err
		}
		m.PJSKDBClient = client
	case "chunithm_main":
		client, err := entchuniMain.Open(dialect, dsn)
		if err != nil {
			return err
		}
		if err := client.Schema.Create(context.Background()); err != nil {
			return err
		}
		m.ChunithmMainDBClient = client
	case "chunithm_music":
		client, err := entchuniMusic.Open(dialect, dsn)
		if err != nil {
			return err
		}
		if err := client.Schema.Create(context.Background()); err != nil {
			return err
		}
		m.ChunithmMusicDBClient = client
	default:
		return nil
	}
	return nil
}

func (m *DBManager) Close() {
	if m.PJSKDBClient != nil {
		if err := m.PJSKDBClient.Close(); err != nil {
			log.Printf("error closing PJSK client: %v", err)
		}
	}
	if m.ChunithmMainDBClient != nil {
		if err := m.ChunithmMainDBClient.Close(); err != nil {
			log.Printf("error closing Chunithm Main client: %v", err)
		}
	}
	if m.ChunithmMusicDBClient != nil {
		if err := m.ChunithmMusicDBClient.Close(); err != nil {
			log.Printf("error closing Chunithm Music client: %v", err)
		}
	}
}
