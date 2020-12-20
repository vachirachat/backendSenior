package route

import "backendSenior/domain/service/auth"

type OrganizationRouteHandler struct {
	organizationService *service.organizationService
	authService         *auth.AuthService
}

func NewOrganizationRouteHandler(organizationService *service.organizationService, authService *auth.AuthService) *OrganizationRouteHandler {
	return &OrganizationRouteHandler{
		organizationService: organizationService,
		authService:         authService,
	}
}
