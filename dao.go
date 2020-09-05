package slackbot

import (
	"fmt"

	"github.com/dailyburn/bigquery/client"
	"github.com/victorouse/slackbot/config"
)

type BQ struct {
	client    *client.Client
	projectID string
}

type DAO struct {
	bq    *BQ
	Store *Store
}

func NewDAO(store *Store) *DAO {
	config := config.ParseConfig()
	bq := client.New(config.GoogleApplicationCredentials)

	return &DAO{
		bq:    &BQ{bq, config.ProjectID},
		Store: store,
	}
}

func (d *DAO) GetShakespear() ([]string, [][]string, error) {
	query := `select * from publicdata:samples.shakespeare limit 5;`
	rows, headers, err := d.bq.client.Query("shakespeare", d.bq.projectID, query)

	if err != nil {
		fmt.Println("[ERROR]: ", err)
		return nil, nil, err
	}

	return headers, ResultSet(rows), nil
}
