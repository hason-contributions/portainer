package migrator

import portainer "github.com/portainer/portainer/api"

func (m *Migrator) updateResourceControlsToDBVersion22() error {
	legacyResourceControls, err := m.resourceControlService.ResourceControls()
	if err != nil {
		return err
	}

	for _, resourceControl := range legacyResourceControls {
		resourceControl.AdministratorsOnly = false

		err := m.resourceControlService.UpdateResourceControl(resourceControl.ID, &resourceControl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateUsersAndRolesToDBVersion22() error {
	legacyUsers, err := m.userService.Users()
	if err != nil {
		return err
	}

	for _, user := range legacyUsers {
		user.PortainerAuthorizations = portainer.DefaultPortainerAuthorizations()
		err = m.userService.UpdateUser(user.ID, &user)
		if err != nil {
			return err
		}
	}

	endpointAdministratorRole, err := m.roleService.Role(portainer.RoleID(1))
	if err != nil {
		return err
	}
	endpointAdministratorRole.Authorizations = portainer.DefaultEndpointAuthorizationsForEndpointAdministratorRole()

	err = m.roleService.UpdateRole(endpointAdministratorRole.ID, endpointAdministratorRole)

	helpDeskRole, err := m.roleService.Role(portainer.RoleID(1))
	if err != nil {
		return err
	}
	helpDeskRole.Authorizations = portainer.DefaultEndpointAuthorizationsForHelpDeskRole()

	err = m.roleService.UpdateRole(helpDeskRole.ID, helpDeskRole)

	standardUserRole, err := m.roleService.Role(portainer.RoleID(1))
	if err != nil {
		return err
	}
	standardUserRole.Authorizations = portainer.DefaultEndpointAuthorizationsForStandardUserRole()

	err = m.roleService.UpdateRole(standardUserRole.ID, standardUserRole)

	readOnlyUserRole, err := m.roleService.Role(portainer.RoleID(1))
	if err != nil {
		return err
	}
	readOnlyUserRole.Authorizations = portainer.DefaultEndpointAuthorizationsForReadOnlyUserRole()

	err = m.roleService.UpdateRole(readOnlyUserRole.ID, readOnlyUserRole)

	authorizationServiceParameters := &portainer.AuthorizationServiceParameters{
		EndpointService:       m.endpointService,
		EndpointGroupService:  m.endpointGroupService,
		RegistryService:       m.registryService,
		RoleService:           m.roleService,
		TeamMembershipService: m.teamMembershipService,
		UserService:           m.userService,
	}

	authorizationService := portainer.NewAuthorizationService(authorizationServiceParameters)
	return authorizationService.UpdateUsersAuthorizations()
}