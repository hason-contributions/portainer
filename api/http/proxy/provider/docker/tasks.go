package docker

import (
	"net/http"

	"github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/responseutils"
)

const (
	errDockerTaskServiceIdentifierNotFound = portainer.Error("Docker task service identifier not found")
	taskServiceIdentifier                  = "ServiceID"
	taskLabelForStackIdentifier            = "com.docker.stack.namespace"
)

// taskListOperation extracts the response as a JSON object, loop through the tasks array
// and filter the tasks based on resource controls before rewriting the response
func taskListOperation(response *http.Response, executor *operationExecutor) error {
	var err error

	// TaskList response is a JSON array
	// https://docs.docker.com/engine/api/v1.28/#operation/TaskList
	responseArray, err := responseutils.GetResponseAsJSONArray(response)
	if err != nil {
		return err
	}

	if !executor.operationContext.isAdmin && !executor.operationContext.endpointResourceAccess {
		responseArray, err = filterTaskList(responseArray, executor.operationContext)
		if err != nil {
			return err
		}
	}

	return responseutils.RewriteResponse(response, responseArray, http.StatusOK)
}

// extractTaskLabelsFromTaskListObject retrieve the Labels of the task if present.
// Task schema reference: https://docs.docker.com/engine/api/v1.28/#operation/TaskList
func extractTaskLabelsFromTaskListObject(responseObject map[string]interface{}) map[string]interface{} {
	// Labels are stored under Spec.ContainerSpec.Labels
	taskSpecObject := responseutils.GetJSONObject(responseObject, "Spec")
	if taskSpecObject != nil {
		containerSpecObject := responseutils.GetJSONObject(taskSpecObject, "ContainerSpec")
		if containerSpecObject != nil {
			return responseutils.GetJSONObject(containerSpecObject, "Labels")
		}
	}
	return nil
}

// filterTaskList loops through all tasks and filters public tasks (no associated resource control)
// as well as authorized tasks (access granted to the user based on existing resource control).
// Resource controls checks are based on: service identifier, stack identifier (from label).
// Task object schema reference: https://docs.docker.com/engine/api/v1.28/#operation/TaskList
// any resource control giving access to the user based on the associated service identifier.
func filterTaskList(taskData []interface{}, context *restrictedDockerOperationContext) ([]interface{}, error) {
	filteredTaskData := make([]interface{}, 0)

	for _, task := range taskData {
		taskObject := task.(map[string]interface{})
		if taskObject[taskServiceIdentifier] == nil {
			return nil, errDockerTaskServiceIdentifierNotFound
		}

		serviceID := taskObject[taskServiceIdentifier].(string)
		taskObject, access := applyResourceAccessControl(taskObject, serviceID, context, portainer.ServiceResourceControl)
		if !access {
			taskLabels := extractTaskLabelsFromTaskListObject(taskObject)
			taskObject, access = applyResourceAccessControlFromLabel(taskLabels, taskObject, taskLabelForStackIdentifier, context, portainer.StackResourceControl)
		}

		if access {
			filteredTaskData = append(filteredTaskData, taskObject)
		}
	}

	return filteredTaskData, nil
}