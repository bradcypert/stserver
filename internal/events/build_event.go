package events

import (
	"context"

	"github.com/bradcypert/stserver/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleBuildEvent(context context.Context, event GameEvent, pool *pgxpool.Pool) {
	d := db.New(pool)
	building, err := d.GetBuildingByPortAndType(context, db.GetBuildingByPortAndTypeParams{
		PortID: event.PortID,
		Type:   event.BuildingType,
	})

	if (building == db.Building{}) {
		building, err = d.CreateBuilding(context, db.CreateBuildingParams{
			PortID: event.PortID,
			Type:   event.BuildingType,
		})
		if err != nil {
			panic(err)
		}
	}
}
