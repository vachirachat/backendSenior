package model

import (
	"time"
)

type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}

type Organization struct {
	OrganizationID string    `json:"organizationId" bson:"_id,omitempty"`
	TimeStamp      time.Time `json:"timestamp" bson:"timestamp"`
	UserIDList     []string  `json:"userIdList" bson:"userIdList"`
	AdminIDList    []string  `json:"adminIdList" bson:"adminIdList"`
	Name           string    `json:"name" bson:"name"`
}
