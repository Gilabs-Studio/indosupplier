package seeders

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
	monetizationModels "github.com/gilabs/indosupplier/api/internal/monetization/data/models"
)

type subscriptionPlanSeed struct {
	Code         string
	Name         string
	BillingCycle string
	Price        float64
	Description  string
	BenefitsJSON string
}

func SeedMonetization() error {
	plans := []subscriptionPlanSeed{
		{
			Code:         "free",
			Name:         "Free Basic",
			BillingCycle: "month",
			Price:        0,
			Description:  "Basic supplier access for early onboarding.",
			BenefitsJSON: `["Basic supplier profile","Product discovery listing"]`,
		},
		{
			Code:         "bronze",
			Name:         "Bronze Seller",
			BillingCycle: "year",
			Price:        2000000,
			Description:  "Starter paid supplier plan for growing sellers.",
			BenefitsJSON: `["More product uploads","RFQ bid access","Basic analytics"]`,
		},
		{
			Code:         "silver",
			Name:         "Silver Pro",
			BillingCycle: "year",
			Price:        5000000,
			Description:  "Professional supplier plan with higher usage limits.",
			BenefitsJSON: `["Priority product visibility","Expanded RFQ bids","Advanced analytics"]`,
		},
		{
			Code:         "gold",
			Name:         "Gold Enterprise",
			BillingCycle: "year",
			Price:        12000000,
			Description:  "Enterprise supplier plan for high-volume marketplace activity.",
			BenefitsJSON: `["Premium verification","Featured visibility","Dedicated support"]`,
		},
	}

	for _, seed := range plans {
		var plan monetizationModels.SubscriptionPlan
		err := database.DB.Where("code = ?", seed.Code).First(&plan).Error
		if err == nil {
			plan.Name = seed.Name
			plan.BillingCycle = seed.BillingCycle
			plan.Price = seed.Price
			plan.Description = seed.Description
			plan.BenefitsJSON = seed.BenefitsJSON
			plan.IsActive = true
			if err := database.DB.Save(&plan).Error; err != nil {
				return err
			}
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		plan = monetizationModels.SubscriptionPlan{
			Code:         seed.Code,
			Name:         seed.Name,
			BillingCycle: seed.BillingCycle,
			Price:        seed.Price,
			Description:  seed.Description,
			BenefitsJSON: seed.BenefitsJSON,
			IsActive:     true,
		}
		if err := database.DB.Create(&plan).Error; err != nil {
			return err
		}
	}

	fmt.Printf("seeded %d subscription plans\n", len(plans))
	return nil
}
