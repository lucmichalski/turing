package api

import (
	"errors"
	"testing"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/api/request"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListRouters(t *testing.T) {
	// Create mock services
	// MLP service
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", 1).Return(nil, errors.New("Test project error"))
	mlpSvc.On("GetProject", 2).Return(&mlp.Project{Id: 2}, nil)
	mlpSvc.On("GetProject", 3).Return(&mlp.Project{Id: 3}, nil)
	// Router Service
	routers := []*models.Router{
		{
			Model: models.Model{
				ID: 1,
			},
			ProjectID: 3,
		},
		{
			Model: models.Model{
				ID: 2,
			},
			ProjectID: 3,
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("ListRouters", 2, "").Return(nil, errors.New("Test router error"))
	routerSvc.On("ListRouters", 3, "").Return(routers, nil)

	// Define test cases
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | bad request": {
			vars:     map[string]string{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | not found": {
			vars:     map[string]string{"project_id": "1"},
			expected: NotFound("project not found", "Test project error"),
		},
		"failure | internal server error": {
			vars:     map[string]string{"project_id": "2"},
			expected: InternalServerError("unable to list routers", "Test router error"),
		},
		"success": {
			vars: map[string]string{"project_id": "3"},
			expected: &Response{
				code: 200,
				data: routers,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RoutersController{
				&routerDeploymentController{
					&baseController{
						&AppContext{
							MLPService:     mlpSvc,
							RoutersService: routerSvc,
						},
					},
				},
			}
			response := ctrl.ListRouters(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestGetRouter(t *testing.T) {
	router := &models.Router{
		Model: models.Model{
			ID: 2,
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	routerSvc.On("FindByID", uint(2)).Return(router, nil)

	// Define tests
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | bad request": {
			vars:     map[string]string{},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | not found": {
			vars:     map[string]string{"router_id": "1"},
			expected: NotFound("router not found", "Test router error"),
		},
		"success": {
			vars: map[string]string{"router_id": "2"},
			expected: &Response{
				code: 200,
				data: router,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RoutersController{
				&routerDeploymentController{
					&baseController{
						&AppContext{
							RoutersService: routerSvc,
						},
					},
				},
			}
			response := ctrl.GetRouter(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestCreateRouter(t *testing.T) {
	// Create mock services
	// MLP service
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", 1).Return(nil, errors.New("Test project error"))
	mlpSvc.On("GetProject", 2).Return(&mlp.Project{Id: 2}, nil)
	mlpSvc.On("GetEnvironment", "dev-invalid").Return(nil, errors.New("Test env error"))
	mlpSvc.On("GetEnvironment", "dev").Return(&merlin.Environment{}, nil)
	// Router Service
	router1 := &models.Router{
		Model: models.Model{
			ID: 1,
		},
	}
	router2 := &models.Router{
		Name:            "router2",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusPending,
	}
	router3 := &models.Router{
		Name:            "router3",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusPending,
	}
	router3Saved := &models.Router{
		Name:            "router3",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusPending,
		Model: models.Model{
			ID: 3,
		},
	}
	router3Failure := &models.Router{
		Name:            "router3",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusFailed,
		Model: models.Model{
			ID: 3,
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByProjectAndName", 2, "router1").Return(router1, nil)
	routerSvc.On("FindByProjectAndName", 2, "router2").Return(nil, errors.New("Test router error"))
	routerSvc.On("FindByProjectAndName", 2, "router3").Return(nil, errors.New("Test router error"))
	routerSvc.On("Save", router2).Return(nil, errors.New("Test router save error"))
	routerSvc.On("Save", router3).Return(router3Saved, nil)
	routerSvc.On("Save", router3Failure).Return(router3Failure, nil)
	// For the deployment method
	routerSvc.On("Save", mock.Anything).Return(nil, errors.New("Test Router Deployment Failure"))
	// Router Version Service
	routerVersion := &models.RouterVersion{
		RouterID: uint(3),
		Router:   router3Saved,
		ExperimentEngine: &models.ExperimentEngine{
			Type: models.ExperimentEngineTypeNop,
		},
		LogConfig: &models.LogConfig{
			ResultLoggerType: models.NopLogger,
		},
		Status: models.RouterVersionStatusPending,
	}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.On("Save", routerVersion).Return(routerVersion, nil)

	// Define tests
	tests := map[string]struct {
		body     interface{}
		vars     map[string]string
		expected *Response
	}{
		"failure | bad request": {
			vars:     map[string]string{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | project not found": {
			vars:     map[string]string{"project_id": "1"},
			expected: NotFound("project not found", "Test project error"),
		},
		"failure | router exists": {
			body: &request.CreateOrUpdateRouterRequest{
				Name: "router1",
			},
			vars:     map[string]string{"project_id": "2"},
			expected: BadRequest("invalid router name", "router with name router1 already exists in project 2"),
		},
		"failure | environment missing": {
			body: &request.CreateOrUpdateRouterRequest{
				Name:        "router2",
				Environment: "dev-invalid",
			},
			vars:     map[string]string{"project_id": "2"},
			expected: BadRequest("invalid environment", "environment dev-invalid does not exist"),
		},
		"failure | router save": {
			body: &request.CreateOrUpdateRouterRequest{
				Name:        "router2",
				Environment: "dev",
			},
			vars:     map[string]string{"project_id": "2"},
			expected: InternalServerError("unable to create router", "Test router save error"),
		},
		"failure | build router version": {
			body: &request.CreateOrUpdateRouterRequest{
				Name:        "router3",
				Environment: "dev",
			},
			vars:     map[string]string{"project_id": "2"},
			expected: InternalServerError("unable to create router", "Router Config is empty"),
		},
		"success": {
			body: &request.CreateOrUpdateRouterRequest{
				Name:        "router3",
				Environment: "dev",
				Config: &request.RouterConfig{
					ExperimentEngine: &request.ExperimentEngineConfig{
						Type: "nop",
					},
					LogConfig: &request.LogConfig{
						ResultLoggerType: models.NopLogger,
					},
				},
			},
			vars: map[string]string{"project_id": "2"},
			expected: &Response{
				code: 200,
				data: router3Saved,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RoutersController{
				&routerDeploymentController{
					&baseController{
						&AppContext{
							MLPService:            mlpSvc,
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							RouterDefaults:        &config.RouterDefaults{},
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.CreateRouter(nil, data.vars, data.body)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestUpdateRouter(t *testing.T) {
	// Create mock services
	// MLP service
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", 1).Return(nil, errors.New("Test project error"))
	mlpSvc.On("GetProject", 2).Return(&mlp.Project{Id: 2}, nil)
	mlpSvc.On("GetEnvironment", "dev-invalid").Return(nil, errors.New("Test env error"))
	mlpSvc.On("GetEnvironment", "dev").Return(&merlin.Environment{}, nil)
	// Router Service
	router2 := &models.Router{
		Name:            "router2",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusPending,
	}
	router3 := &models.Router{
		Name:            "router3",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusFailed,
		Model: models.Model{
			ID: 3,
		},
	}
	router4 := &models.Router{
		Name:            "router4",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusFailed,
		Model: models.Model{
			ID: 4,
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	routerSvc.On("FindByID", uint(2)).Return(router2, nil)
	routerSvc.On("FindByID", uint(3)).Return(router3, nil)
	routerSvc.On("FindByID", uint(4)).Return(router4, nil)
	// For the deployment method
	routerSvc.On("Save", mock.Anything).Return(nil, errors.New("Test Router Deployment Failure"))
	// Router Version Service
	routerVersion := &models.RouterVersion{
		RouterID: uint(4),
		Router:   router4,
		ExperimentEngine: &models.ExperimentEngine{
			Type: models.ExperimentEngineTypeNop,
		},
		LogConfig: &models.LogConfig{
			ResultLoggerType: models.NopLogger,
		},
		Status: models.RouterVersionStatusPending,
	}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.On("Save", routerVersion).Return(routerVersion, nil)

	// Define tests
	tests := map[string]struct {
		body     interface{}
		vars     map[string]string
		expected *Response
	}{
		"failure | bad request (missing project_id)": {
			vars:     map[string]string{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | project not found": {
			vars:     map[string]string{"project_id": "1", "router_id": "1"},
			expected: NotFound("project not found", "Test project error"),
		},
		"failure | bad request (missing router_id)": {
			vars:     map[string]string{"project_id": "2"},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | router not found": {
			body: &request.CreateOrUpdateRouterRequest{
				Name: "router1",
			},
			vars:     map[string]string{"project_id": "2", "router_id": "1"},
			expected: NotFound("router not found", "Test router error"),
		},
		"failure | invalid router config": {
			body: &request.CreateOrUpdateRouterRequest{
				Name: "router1",
			},
			vars: map[string]string{"project_id": "2", "router_id": "2"},
			expected: BadRequest(
				"invalid router configuration",
				"Router name and environment cannot be changed after creation",
			),
		},
		"failure | deployment in progress": {
			body: &request.CreateOrUpdateRouterRequest{
				Name:        "router2",
				Environment: "dev",
			},
			vars: map[string]string{"project_id": "2", "router_id": "2"},
			expected: BadRequest(
				"invalid update request",
				"another version is currently pending deployment",
			),
		},
		"failure | build router version": {
			body: &request.CreateOrUpdateRouterRequest{
				Name:        "router3",
				Environment: "dev",
			},
			vars:     map[string]string{"project_id": "2", "router_id": "3"},
			expected: InternalServerError("unable to update router", "Router Config is empty"),
		},
		"success": {
			body: &request.CreateOrUpdateRouterRequest{
				Name:        "router4",
				Environment: "dev",
				Config: &request.RouterConfig{
					ExperimentEngine: &request.ExperimentEngineConfig{
						Type: "nop",
					},
					LogConfig: &request.LogConfig{
						ResultLoggerType: models.NopLogger,
					},
				},
			},
			vars: map[string]string{"project_id": "2", "router_id": "4"},
			expected: &Response{
				code: 200,
				data: router4,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RoutersController{
				&routerDeploymentController{
					&baseController{
						&AppContext{
							MLPService:            mlpSvc,
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							RouterDefaults:        &config.RouterDefaults{},
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.UpdateRouter(nil, data.vars, data.body)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestDeleteRouter(t *testing.T) {
	// Create mock services
	// Router Service
	router2 := &models.Router{
		Name:            "router2",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusDeployed,
	}
	router3 := &models.Router{
		Name:            "router3",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusFailed,
		Model: models.Model{
			ID: 3,
		},
	}
	router4 := &models.Router{
		Name:            "router4",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusUndeployed,
		Model: models.Model{
			ID: 4,
		},
	}
	router5 := &models.Router{
		Name:            "router5",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusUndeployed,
		Model: models.Model{
			ID: 5,
		},
	}
	router6 := &models.Router{
		Name:            "router6",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusUndeployed,
		Model: models.Model{
			ID: 6,
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	routerSvc.On("FindByID", uint(2)).Return(router2, nil)
	routerSvc.On("FindByID", uint(3)).Return(router3, nil)
	routerSvc.On("FindByID", uint(4)).Return(router4, nil)
	routerSvc.On("FindByID", uint(5)).Return(router5, nil)
	routerSvc.On("FindByID", uint(6)).Return(router6, nil)
	routerSvc.On("Delete", router5).Return(errors.New("Test delete router error"))
	routerSvc.On("Delete", router6).Return(nil)
	// For the deployment method
	routerSvc.On("Save", mock.Anything).Return(nil, errors.New("Test Router Deployment Failure"))
	// Router Version Service
	routerVersion := &models.RouterVersion{
		RouterID: uint(3),
		Router:   router3,
	}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.On("Save", routerVersion).Return(routerVersion, nil)
	routerVersionSvc.
		On("ListRouterVersionsWithStatus", uint(3), models.RouterVersionStatusPending).
		Return(nil, errors.New("Test List Router Versions error"))
	routerVersionSvc.
		On("ListRouterVersionsWithStatus", uint(4), models.RouterVersionStatusPending).
		Return([]*models.RouterVersion{{Status: models.RouterVersionStatusPending}}, nil)
	routerVersionSvc.
		On("ListRouterVersionsWithStatus", uint(5), models.RouterVersionStatusPending).
		Return([]*models.RouterVersion{}, nil)
	routerVersionSvc.
		On("ListRouterVersionsWithStatus", uint(6), models.RouterVersionStatusPending).
		Return([]*models.RouterVersion{}, nil)

	// Define tests
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | bad request (missing router_id)": {
			vars:     map[string]string{},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | router not found": {
			vars:     map[string]string{"router_id": "1"},
			expected: NotFound("router not found", "Test router error"),
		},
		"failure | router deployed": {
			vars: map[string]string{"router_id": "2"},
			expected: BadRequest(
				"invalid delete request",
				"router is currently deployed. Undeploy it first.",
			),
		},
		"failure | list router versions": {
			vars: map[string]string{"router_id": "3"},
			expected: InternalServerError(
				"unable to retrieve router versions",
				"Test List Router Versions error",
			),
		},
		"failure | pending router versions": {
			vars: map[string]string{"router_id": "4"},
			expected: BadRequest(
				"invalid delete request",
				"a router version is currently pending deployment",
			),
		},
		"failure | delete failed": {
			vars:     map[string]string{"router_id": "5"},
			expected: InternalServerError("unable to delete router", "Test delete router error"),
		},
		"success": {
			vars: map[string]string{"router_id": "6"},
			expected: &Response{
				code: 200,
				data: map[string]int{"id": 6},
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RoutersController{
				&routerDeploymentController{
					&baseController{
						&AppContext{
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							RouterDefaults:        &config.RouterDefaults{},
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.DeleteRouter(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestDeployRouter(t *testing.T) {
	// Create mock services
	// MLP service
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", 1).Return(nil, errors.New("Test project error"))
	mlpSvc.On("GetProject", 2).Return(&mlp.Project{}, nil)
	mlpSvc.On("GetEnvironment", "dev-invalid").Return(nil, errors.New("Test env error"))
	mlpSvc.On("GetEnvironment", "dev").Return(&merlin.Environment{}, nil)
	// Router Service
	router2 := &models.Router{
		Name:            "router2",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusPending,
		Model: models.Model{
			ID: 2,
		},
	}
	router3 := &models.Router{
		Name:            "router3",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusDeployed,
		Model: models.Model{
			ID: 3,
		},
	}
	router4 := &models.Router{
		Name:            "router4",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusFailed,
		Model: models.Model{
			ID: 4,
		},
	}
	router5 := &models.Router{
		Name:            "router5",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusFailed,
		Model: models.Model{
			ID: 5,
		},
		CurrRouterVersion: &models.RouterVersion{
			Model: models.Model{
				ID: 1,
			},
		},
	}
	router6 := &models.Router{
		Name:            "router5",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusFailed,
		Model: models.Model{
			ID: 6,
		},
		CurrRouterVersion: &models.RouterVersion{
			Model: models.Model{
				ID: 2,
			},
			Version: 2,
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	routerSvc.On("FindByID", uint(2)).Return(router2, nil)
	routerSvc.On("FindByID", uint(3)).Return(router3, nil)
	routerSvc.On("FindByID", uint(4)).Return(router4, nil)
	routerSvc.On("FindByID", uint(5)).Return(router5, nil)
	routerSvc.On("FindByID", uint(6)).Return(router6, nil)
	// For the deployment method
	routerSvc.On("Save", mock.Anything).Return(nil, errors.New("Test Router Deployment Failure"))
	// Router Version Service
	routerVersion := &models.RouterVersion{
		RouterID: uint(2),
		Router:   router6,
		Version:  2,
	}
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test router version error"))
	routerVersionSvc.On("FindByID", uint(2)).Return(routerVersion, nil)

	// Define tests
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | bad request (missing project_id)": {
			vars:     map[string]string{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | project not found": {
			vars:     map[string]string{"project_id": "1", "router_id": "1"},
			expected: NotFound("project not found", "Test project error"),
		},
		"failure | bad request (missing router_id)": {
			vars:     map[string]string{"project_id": "2"},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | router not found": {
			vars:     map[string]string{"project_id": "2", "router_id": "1"},
			expected: NotFound("router not found", "Test router error"),
		},
		"failure | router status pending": {
			vars: map[string]string{"project_id": "2", "router_id": "2"},
			expected: BadRequest(
				"invalid deploy request",
				"router is currently deploying, cannot do another deployment",
			),
		},
		"failure | router status deployed": {
			vars:     map[string]string{"project_id": "2", "router_id": "3"},
			expected: BadRequest("invalid deploy request", "router is already deployed"),
		},
		"failure | no current version": {
			vars:     map[string]string{"project_id": "2", "router_id": "4"},
			expected: BadRequest("invalid deploy request", "Router has no current configuration"),
		},
		"failure | router version not found": {
			vars:     map[string]string{"project_id": "2", "router_id": "5"},
			expected: NotFound("router version not found", "Test router version error"),
		},
		"success": {
			vars: map[string]string{"project_id": "2", "router_id": "6"},
			expected: &Response{
				code: 202,
				data: map[string]int{
					"router_id": 6,
					"version":   2,
				},
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RoutersController{
				&routerDeploymentController{
					&baseController{
						&AppContext{
							MLPService:            mlpSvc,
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							RouterDefaults:        &config.RouterDefaults{},
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.DeployRouter(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestUndeployRouter(t *testing.T) {
	// Create mock services
	// Event Service
	eventSvc := &mocks.EventService{}
	// For the undeployment method
	eventSvc.On("Save", mock.Anything).Return(nil)
	// MLP service
	project := &mlp.Project{}
	environment := &merlin.Environment{}
	mlpSvc := &mocks.MLPService{}
	mlpSvc.On("GetProject", 1).Return(nil, errors.New("Test project error"))
	mlpSvc.On("GetProject", 2).Return(project, nil)
	mlpSvc.On("GetEnvironment", "dev-invalid").Return(nil, errors.New("Test env error"))
	mlpSvc.On("GetEnvironment", "dev").Return(environment, nil)
	// Router Service
	router2 := &models.Router{
		Name:            "router2",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusDeployed,
		Model: models.Model{
			ID: 2,
		},
		CurrRouterVersion: &models.RouterVersion{
			Model: models.Model{
				ID: 1,
			},
		},
	}
	router3 := &models.Router{
		Name:            "router3",
		ProjectID:       2,
		EnvironmentName: "dev",
		Status:          models.RouterStatusDeployed,
		Model: models.Model{
			ID: 3,
		},
		CurrRouterVersion: &models.RouterVersion{
			Model: models.Model{
				ID: 2,
			},
		},
	}
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	routerSvc.On("FindByID", uint(2)).Return(router2, nil)
	routerSvc.On("FindByID", uint(3)).Return(router3, nil)
	// For the deployment method
	routerSvc.On("Save", router2).Return(nil, errors.New("Test Router Deployment Failure"))
	routerSvc.On("Save", router3).Return(router3, nil)
	// Router Version Service
	routerVersionSvc := &mocks.RouterVersionsService{}
	routerVersionSvc.On("ListRouterVersions", uint(2)).Return([]*models.RouterVersion{}, nil)
	routerVersionSvc.On("ListRouterVersions", uint(3)).Return([]*models.RouterVersion{}, nil)
	// Deployment Service
	deploymentSvc := &mocks.DeploymentService{}
	deploymentSvc.
		On("DeleteRouterEndpoint", project, environment, &models.RouterVersion{Router: router2}).
		Return(errors.New("Test undeploy error"))
	deploymentSvc.
		On("DeleteRouterEndpoint", project, environment, &models.RouterVersion{Router: router3}).
		Return(nil)

	// Define tests
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | bad request (missing project_id)": {
			vars:     map[string]string{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | project not found": {
			vars:     map[string]string{"project_id": "1", "router_id": "1"},
			expected: NotFound("project not found", "Test project error"),
		},
		"failure | bad request (missing router_id)": {
			vars:     map[string]string{"project_id": "2"},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | router not found": {
			vars:     map[string]string{"project_id": "2", "router_id": "1"},
			expected: NotFound("router not found", "Test router error"),
		},
		"failure | undeploy error": {
			vars: map[string]string{"project_id": "2", "router_id": "2"},
			expected: InternalServerError(
				"unable to undeploy router",
				"Test undeploy error. Test Router Deployment Failure",
			),
		},
		"success": {
			vars: map[string]string{"project_id": "2", "router_id": "3"},
			expected: &Response{
				code: 200,
				data: map[string]int{"router_id": 3},
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RoutersController{
				&routerDeploymentController{
					&baseController{
						&AppContext{
							MLPService:            mlpSvc,
							RoutersService:        routerSvc,
							RouterVersionsService: routerVersionSvc,
							RouterDefaults:        &config.RouterDefaults{},
							EventService:          eventSvc,
							DeploymentService:     deploymentSvc,
						},
					},
				},
			}
			// Run test method and validate
			response := ctrl.UndeployRouter(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestListRouterEvents(t *testing.T) {
	router2 := &models.Router{
		Model: models.Model{
			ID: 2,
		},
	}
	router3 := &models.Router{
		Model: models.Model{
			ID: 3,
		},
	}
	events := []*models.Event{
		{
			Model: models.Model{
				ID: 10,
			},
		},
		{
			Model: models.Model{
				ID: 20,
			},
		},
	}

	// Set up mock services
	routerSvc := &mocks.RoutersService{}
	routerSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	routerSvc.On("FindByID", uint(2)).Return(router2, nil)
	routerSvc.On("FindByID", uint(3)).Return(router3, nil)
	eventSvc := &mocks.EventService{}
	eventSvc.On("ListEvents", 2).Return(nil, errors.New("Test event error"))
	eventSvc.On("ListEvents", 3).Return(events, nil)

	// Define tests
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | bad request": {
			vars:     map[string]string{},
			expected: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | router not found": {
			vars:     map[string]string{"router_id": "1"},
			expected: NotFound("router not found", "Test router error"),
		},
		"failure | events not found": {
			vars:     map[string]string{"router_id": "2"},
			expected: NotFound("events not found", "Test event error"),
		},
		"success": {
			vars: map[string]string{"router_id": "3"},
			expected: &Response{
				code: 200,
				data: map[string][]*models.Event{"events": events},
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &RoutersController{
				&routerDeploymentController{
					&baseController{
						&AppContext{
							RoutersService: routerSvc,
							EventService:   eventSvc,
						},
					},
				},
			}
			response := ctrl.ListRouterEvents(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}