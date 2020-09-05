package slackbot

import (
	"fmt"

	"github.com/dailyburn/bigquery/client"
	"github.com/victorouse/slackbot/config"
)

type DAO struct {
	bq        *client.Client
	projectID string
}

func NewDAO() *DAO {
	config := config.ParseConfig()
	bq := client.New(config.GoogleApplicationCredentials)

	return &DAO{
		bq:        bq,
		projectID: config.ProjectID,
	}
}

func (d *DAO) GetShakespear() ([]string, [][]string, error) {
	query := "select * from publicdata:samples.shakespeare limit 5;"
	rows, headers, err := d.bq.Query("shakespeare", d.projectID, query)

	if err != nil {
		fmt.Println("[ERROR]: ", err)
		return nil, nil, err
	}

	return headers, ResultSet(rows), nil
}
