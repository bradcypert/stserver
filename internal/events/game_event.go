package events

type GameEventType int

const (
	GameEventPortBuilding GameEventType = iota
	GameEventResourceCollect
	GameEventShipConstruct
)

type GameEvent struct {
	PortID       int32         `json:"port_id"`
	BuildingType string        `json:"building_type"`
	EventType    GameEventType `json:"event_type"`
}
