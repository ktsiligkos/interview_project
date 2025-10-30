package company

import "context"

// EventPublisher defines the capability needed for publishing company events.
type EventPublisher interface {
	PublishCompanyEvent(ctx context.Context, event CompanyEvent) error
}

// Models the event to be sent to Kafka
type CompanyEvent struct {
	Operation string       `json:"operation"`
	Company   EventCompany `json:"company"`
}

type EventCompany struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       *string `json:"description,omitempty"`
	AmountOfEmployees int     `json:"amount_of_employees"`
	Registered        bool    `json:"registered"`
	Type              string  `json:"type"`
}
