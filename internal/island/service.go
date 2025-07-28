package island

import (
	"context"
	"fmt"
	"time"

	"github.com/bradcypert/stserver/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		queries: db.New(pool),
		pool:    pool,
	}
}

type BuildingConstructionRequest struct {
	PortID      int32  `json:"port_id"`
	BuildingType string `json:"building_type"`
}

type IslandOverview struct {
	Port      db.GetPortWithResourcesRow `json:"port"`
	Buildings []db.GetPortBuildingsRow   `json:"buildings"`
}

func (s *Service) GetIslandOverview(ctx context.Context, portID int32) (*IslandOverview, error) {
	// Get port with resources
	port, err := s.queries.GetPortWithResources(ctx, portID)
	if err != nil {
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	// Get buildings
	buildings, err := s.queries.GetPortBuildings(ctx, portID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buildings: %w", err)
	}

	return &IslandOverview{
		Port:      port,
		Buildings: buildings,
	}, nil
}

func (s *Service) ConstructBuilding(ctx context.Context, req BuildingConstructionRequest) (*db.Building, error) {
	// Get building type info
	buildingType, err := s.queries.GetBuildingTypeByName(ctx, req.BuildingType)
	if err != nil {
		return nil, fmt.Errorf("invalid building type: %w", err)
	}

	// Check if player has enough resources
	hasResources, err := s.queries.CheckResourceAvailability(ctx, db.CheckResourceAvailabilityParams{
		PortID:   req.PortID,
		HasWood:  buildingType.BaseCostWood,
		HasIron:  buildingType.BaseCostIron,
		HasGold:  buildingType.BaseCostGold,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check resources: %w", err)
	}

	if !hasResources.HasWood || !hasResources.HasIron || !hasResources.HasGold {
		return nil, fmt.Errorf("insufficient resources for construction")
	}

	// Consume resources
	err = s.queries.ConsumeResourcesFromPort(ctx, db.ConsumeResourcesFromPortParams{
		PortID: req.PortID,
		Wood:   buildingType.BaseCostWood,
		Iron:   buildingType.BaseCostIron,
		Gold:   buildingType.BaseCostGold,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to consume resources: %w", err)
	}

	// Calculate completion time
	completionTime := time.Now().Add(time.Duration(buildingType.BaseBuildTime) * time.Second)

	// Create building under construction
	building, err := s.queries.CreateBuildingConstruction(ctx, db.CreateBuildingConstructionParams{
		PortID:                 req.PortID,
		Type:                   req.BuildingType,
		ConstructionCompleteAt: pgtype.Timestamptz{Time: completionTime, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create building: %w", err)
	}

	return &building, nil
}

func (s *Service) UpgradeBuilding(ctx context.Context, buildingID int32) error {
	// Get building info
	building, err := s.queries.GetBuilding(ctx, buildingID)
	if err != nil {
		return fmt.Errorf("building not found: %w", err)
	}

	if building.UnderConstruction {
		return fmt.Errorf("building is already under construction")
	}

	// Get building type info
	buildingType, err := s.queries.GetBuildingTypeByName(ctx, building.Type)
	if err != nil {
		return fmt.Errorf("invalid building type: %w", err)
	}

	if building.Level >= buildingType.MaxLevel {
		return fmt.Errorf("building is already at maximum level")
	}

	// Calculate upgrade cost (increases with level)
	upgradeCostMultiplier := int32(building.Level + 1)
	upgradeCostWood := buildingType.BaseCostWood * upgradeCostMultiplier
	upgradeCostIron := buildingType.BaseCostIron * upgradeCostMultiplier
	upgradeCostGold := buildingType.BaseCostGold * upgradeCostMultiplier

	// Check resources
	hasResources, err := s.queries.CheckResourceAvailability(ctx, db.CheckResourceAvailabilityParams{
		PortID:   building.PortID,
		HasWood:  upgradeCostWood,
		HasIron:  upgradeCostIron,
		HasGold:  upgradeCostGold,
	})
	if err != nil {
		return fmt.Errorf("failed to check resources: %w", err)
	}

	if !hasResources.HasWood || !hasResources.HasIron || !hasResources.HasGold {
		return fmt.Errorf("insufficient resources for upgrade")
	}

	// Consume resources
	err = s.queries.ConsumeResourcesFromPort(ctx, db.ConsumeResourcesFromPortParams{
		PortID: building.PortID,
		Wood:   upgradeCostWood,
		Iron:   upgradeCostIron,
		Gold:   upgradeCostGold,
	})
	if err != nil {
		return fmt.Errorf("failed to consume resources: %w", err)
	}

	// Calculate upgrade time (longer than base construction)
	upgradeTime := time.Duration(buildingType.BaseBuildTime) * time.Duration(building.Level+1) * time.Second
	completionTime := time.Now().Add(upgradeTime)

	// Start upgrade
	err = s.queries.UpgradeBuilding(ctx, db.UpgradeBuildingParams{
		ID:                     buildingID,
		ConstructionCompleteAt: pgtype.Timestamptz{Time: completionTime, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to start upgrade: %w", err)
	}

	return nil
}

func (s *Service) ProcessResourceGeneration(ctx context.Context) error {
	// Get all buildings ready for production
	cutoffTime := time.Now().Add(-5 * time.Second) // Last tick was 5 seconds ago
	buildings, err := s.queries.GetBuildingsReadyForProduction(ctx, pgtype.Timestamptz{Time: cutoffTime, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to get buildings ready for production: %w", err)
	}

	for _, building := range buildings {
		err := s.processProductionForBuilding(ctx, building)
		if err != nil {
			// Log error but continue processing other buildings
			fmt.Printf("Error processing production for building %d: %v\n", building.ID, err)
			continue
		}

		// Update last production time
		err = s.queries.UpdateBuildingLastProduction(ctx, building.ID)
		if err != nil {
			fmt.Printf("Error updating last production time for building %d: %v\n", building.ID, err)
		}
	}

	return nil
}

func (s *Service) processProductionForBuilding(ctx context.Context, building db.GetBuildingsReadyForProductionRow) error {
	// Get production rates for this building type and level
	productions, err := s.queries.GetProductionRatesForBuilding(ctx, db.GetProductionRatesForBuildingParams{
		TypeName: building.Type,
		Level:    building.Level,
	})
	if err != nil {
		return fmt.Errorf("failed to get production rates: %w", err)
	}

	if len(productions) == 0 {
		// This building type doesn't produce resources
		return nil
	}

	// Calculate resource additions
	var wood, iron, rum, sugar, tobacco, cotton, coffee, grain, gold, silver int32

	for _, prod := range productions {
		switch prod.ResourceType {
		case "wood":
			wood += prod.ProductionRate
		case "iron":
			iron += prod.ProductionRate
		case "rum":
			rum += prod.ProductionRate
		case "sugar":
			sugar += prod.ProductionRate
		case "tobacco":
			tobacco += prod.ProductionRate
		case "cotton":
			cotton += prod.ProductionRate
		case "coffee":
			coffee += prod.ProductionRate
		case "grain":
			grain += prod.ProductionRate
		case "gold":
			gold += prod.ProductionRate
		case "silver":
			silver += prod.ProductionRate
		}
	}

	// Add resources to port
	err = s.queries.AddResourcesToPort(ctx, db.AddResourcesToPortParams{
		PortID:  building.PortID,
		Wood:    wood,
		Iron:    iron,
		Rum:     rum,
		Sugar:   sugar,
		Tobacco: tobacco,
		Cotton:  cotton,
		Coffee:  coffee,
		Grain:   grain,
		Gold:    gold,
		Silver:  silver,
	})
	if err != nil {
		return fmt.Errorf("failed to add resources: %w", err)
	}

	return nil
}

func (s *Service) CompleteConstructions(ctx context.Context) error {
	// Get buildings that should be completed
	completedBuildings, err := s.queries.GetBuildingsUnderConstruction(ctx)
	if err != nil {
		return fmt.Errorf("failed to get completed buildings: %w", err)
	}

	for _, building := range completedBuildings {
		err := s.queries.CompleteBuildingConstruction(ctx, building.ID)
		if err != nil {
			fmt.Printf("Error completing construction for building %d: %v\n", building.ID, err)
		}
	}

	return nil
}